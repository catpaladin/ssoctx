package amazon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/rs/zerolog"
	ini "gopkg.in/ini.v1"
)

// credentialsFilePath is used to store the credentials path to variable
var credentialsFilePath = getCredentialsFilePath()

// CredentialsTemplate is what is expected in the ini file
type CredentialsTemplate struct {
	AwsAccessKeyID     string `ini:"aws_access_key_id,omitempty"`
	AwsSecretAccessKey string `ini:"aws_secret_access_key,omitempty"`
	AwsSessionToken    string `ini:"aws_session_token,omitempty"`
	CredentialProcess  string `ini:"credential_process,omitempty"`
	Output             string `ini:"output,omitempty"`
	Region             string `ini:"region,omitempty"`
}

// getPersistedCredentials returns a struct containing persisted creds values
func getPersistedCredentials(creds *sso.GetRoleCredentialsOutput, region string) CredentialsTemplate {
	return CredentialsTemplate{
		AwsAccessKeyID:     *creds.RoleCredentials.AccessKeyId,
		AwsSecretAccessKey: *creds.RoleCredentials.SecretAccessKey,
		AwsSessionToken:    *creds.RoleCredentials.SessionToken,
		Region:             region,
	}
}

// getCredentialProcess returns a struct containing credential process values
func getCredentialProcess(accountID, roleName, region, startURL string) CredentialsTemplate {
	return CredentialsTemplate{
		CredentialProcess: fmt.Sprintf(
			"%s assume -a %s -n %s -u %s",
			ProjectFileName,
			accountID,
			roleName,
			startURL,
		),
		Region: region,
	}
}

// getCredentialsFilePath returns the credentials path
func getCredentialsFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".aws", "credentials")
}

// clientInfoFileDestination returns the path to cached access
func clientInfoFileDestination(startURL string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".aws", "sso", "cache", "access-token.json")
}

// writeAWSCredentialsFile is used to write the template to credentials
func writeAWSCredentialsFile(ctx context.Context, template *CredentialsTemplate, profile string) {
	if !exists(ctx, credentialsFilePath) {
		createCredentialsFile(ctx)
	}
	// Write to ini file
	writeTemplateToFile(ctx, template, profile)
}

// readClientInformation is used to read file for ClientInformation
func readClientInformation(ctx context.Context, destination string) (ClientInformation, error) {
	logger := zerolog.Ctx(ctx)
	if exists(ctx, destination) {
		clientInformation := ClientInformation{}
		content, _ := os.ReadFile(destination)
		err := json.Unmarshal(content, &clientInformation)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in unmarshal of client information: %q", err)
		}
		return clientInformation, nil
	}
	return ClientInformation{}, errors.New("no ClientInformation exists")
}

// writeStructToFile is used to write the payload to file
func writeStructToFile(ctx context.Context, payload interface{}, dest string) {
	logger := zerolog.Ctx(ctx)
	targetDir := filepath.Dir(dest)
	if !exists(ctx, targetDir) {
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			logger.Fatal().Msgf("Encountered error in making dir: %q", err)
		}
	}
	file, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		logger.Fatal().Msgf("Encountered error in marshal of payload: %q", err)
	}
	if err = os.WriteFile(dest, file, 0o644); err != nil {
		logger.Fatal().Msgf("Encountered error trying to write file %s: %q", file, err)
	}
}

// exists checks either or not a target file is existing.
// Returns true if the target exists, otherwise false.
func exists(ctx context.Context, target string) bool {
	logger := zerolog.Ctx(ctx)
	if _, err := os.Stat(target); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	logger.Debug().Msgf("Could not determine if file or folder %q exists or not. Assuming not.", target)
	return false
}

// createCredentialsFile is used to create missing directories and file
func createCredentialsFile(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	dir := filepath.Dir(credentialsFilePath)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		logger.Fatal().Msgf("Encountered error in making dir: %q", err)
	}
	f, err := os.OpenFile(credentialsFilePath, os.O_CREATE, 0o644)
	if err != nil {
		logger.Fatal().Msgf("Encountered error opening file: %q", err)
	}
	defer f.Close()
}

// writeTemplateToFile loads credentials to replace and write new.
// it calls replaceProfile to selectively delete a profile and replace it.
// it then saves the changes over the credentials file.
func writeTemplateToFile(ctx context.Context, template *CredentialsTemplate, profile string) {
	logger := zerolog.Ctx(ctx)
	creds, err := ini.Load(credentialsFilePath)
	if err != nil {
		logger.Fatal().Msgf("Encountered error loading credentials file: %q", err)
	}

	replaceProfile(ctx, creds, template, profile)
	if err := creds.SaveTo(credentialsFilePath); err != nil {
		logger.Fatal().Msgf("Encountered error saving credentials: %q", err)
	}
}

// replaceProfile is used to selectively delete a profile from the credentials file.
// it then uses the new struct to populate the new information.
func replaceProfile(ctx context.Context, creds *ini.File, template *CredentialsTemplate, profile string) {
	logger := zerolog.Ctx(ctx)
	creds.DeleteSection(profile)

	newSection, err := creds.NewSection(profile)
	if err != nil {
		logger.Fatal().Msgf("Encountered error creating new section in credentials: %q", err)
	}

	// replace new section wtih template
	if err = newSection.ReflectFrom(template); err != nil {
		logger.Fatal().Msgf("Encountered error reflecting template to new section: %q", err)
	}
}
