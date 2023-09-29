// Package file contains needed functionality for config and files
package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"aws-sso-util/internal/info"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/rs/zerolog"
	"gopkg.in/ini.v1"
)

// CredentialsFilePath is used to store the credentials path to variable
var CredentialsFilePath, _ = GetCredentialsFilePath()

// CredentialsTemplate is what is expected in the ini file
type CredentialsTemplate struct {
	AwsAccessKeyId     string `ini:"aws_access_key_id,omitempty"`
	AwsSecretAccessKey string `ini:"aws_secret_access_key,omitempty"`
	AwsSessionToken    string `ini:"aws_session_token,omitempty"`
	CredentialProcess  string `ini:"credential_process,omitempty"`
	Output             string `ini:"output,omitempty"`
	Region             string `ini:"region,omitempty"`
}

// GetPersistedCredentials returns a struct containing persisted creds values
func GetPersistedCredentials(creds *sso.GetRoleCredentialsOutput, region string) CredentialsTemplate {
	return CredentialsTemplate{
		AwsAccessKeyId:     *creds.RoleCredentials.AccessKeyId,
		AwsSecretAccessKey: *creds.RoleCredentials.SecretAccessKey,
		AwsSessionToken:    *creds.RoleCredentials.SessionToken,
		Region:             region,
	}
}

// GetCredentialProcess returns a struct containing credential process values
func GetCredentialProcess(accountID, roleName, region, startURL string) CredentialsTemplate {
	return CredentialsTemplate{
		CredentialProcess: fmt.Sprintf(
			"aws-sso-util assume -a %s -n %s -u %s",
			accountID,
			roleName,
			startURL[0],
		),
		Region: region,
	}
}

// GetCredentialsFilePath returns the credentials path
func GetCredentialsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Encountered error in getting users home directory: %q", err)
	}
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.aws\\credentials", homeDir), nil
	}
	return fmt.Sprintf("%s/.aws/credentials", homeDir), nil
}

// ClientInfoFileDestination returns the path to cached access
func ClientInfoFileDestination() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Encountered error in getting users home directory: %q", err)
	}
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("%s\\.aws\\sso\\cache\\access-token.json", homeDir), nil
	default:
		return fmt.Sprintf("%s/.aws/sso/cache/access-token.json", homeDir), nil
	}
}

// WriteAWSCredentialsFile is used to write the template to credentials
func WriteAWSCredentialsFile(ctx context.Context, template *CredentialsTemplate, profile string) {
	if !isFileOrFolderExisting(ctx, CredentialsFilePath) {
		createCredentialsFile(ctx)
	}
	// Write to ini file
	writeTemplateToFile(ctx, template, profile)
}

// ReadClientInformation is used to read file for ClientInformation
func ReadClientInformation(ctx context.Context, file string) (info.ClientInformation, error) {
	logger := zerolog.Ctx(ctx)
	if isFileOrFolderExisting(ctx, file) {
		clientInformation := info.ClientInformation{}
		destination, _ := ClientInfoFileDestination()
		content, _ := os.ReadFile(destination)
		err := json.Unmarshal(content, &clientInformation)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in unmarshal of client information: %q", err)
		}
		return clientInformation, nil
	}
	return info.ClientInformation{}, errors.New("No ClientInformation exists")
}

// WriteStructToFile is used to write the payload to file
func WriteStructToFile(ctx context.Context, payload interface{}, dest string) {
	logger := zerolog.Ctx(ctx)
	targetDir := filepath.Dir(dest)
	if !isFileOrFolderExisting(ctx, targetDir) {
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

// isFileOrFolderExisting checks either or not a target file is existing.
// Returns true if the target exists, otherwise false.
func isFileOrFolderExisting(ctx context.Context, target string) bool {
	logger := zerolog.Ctx(ctx)
	if _, err := os.Stat(target); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	logger.Panic().Msgf("Could not determine if file or folder %q exists or not. Exiting.", target)
	return false
}

// createCredentialsFile is used to create missing directories and file
func createCredentialsFile(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	dir := filepath.Dir(CredentialsFilePath)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		logger.Fatal().Msgf("Encountered error in making dir: %q", err)
	}
	f, err := os.OpenFile(CredentialsFilePath, os.O_CREATE, 0o644)
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
	creds, err := ini.Load(CredentialsFilePath)
	if err != nil {
		logger.Fatal().Msgf("Encountered error loading credentials file: %q", err)
	}

	replaceProfile(ctx, creds, template, profile)
	if err := creds.SaveTo(CredentialsFilePath); err != nil {
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
