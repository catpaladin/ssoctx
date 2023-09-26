// Package file contains needed functionality for config and files
package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"aws-sso-util/internal/info"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"gopkg.in/ini.v1"
)

// CredentialsFilePath is used to store the credentials path to variable
var CredentialsFilePath = GetCredentialsFilePath()

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
func GetCredentialProcess(accountID, roleName, region string) CredentialsTemplate {
	return CredentialsTemplate{
		CredentialProcess: fmt.Sprintf(
			"aws-sso-util assume -a %s -n %s",
			accountID,
			roleName,
		),
		Region: region,
	}
}

// GetCredentialsFilePath returns the credentials path
func GetCredentialsFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.aws\\credentials", homeDir)
	}
	return fmt.Sprintf("%s/.aws/credentials", homeDir)
}

// ClientInfoFileDestination returns the path to cached access
func ClientInfoFileDestination() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("%s\\.aws\\sso\\cache\\access-token.json", homeDir)
	default:
		return fmt.Sprintf("%s/.aws/sso/cache/access-token.json", homeDir)
	}
}

// WriteAWSCredentialsFile is used to write the template to credentials
func WriteAWSCredentialsFile(template *CredentialsTemplate, profile string) {
	if !isFileOrFolderExisting(CredentialsFilePath) {
		createCredentialsFile()
	}
	// Write to ini file
	writeTemplateToFile(template, profile)
}

// ReadClientInformation is used to read file for ClientInformation
func ReadClientInformation(file string) (info.ClientInformation, error) {
	if isFileOrFolderExisting(file) {
		clientInformation := info.ClientInformation{}
		content, _ := os.ReadFile(ClientInfoFileDestination())
		err := json.Unmarshal(content, &clientInformation)
		if err != nil {
			log.Fatalf("Something went wrong: %q", err)
		}
		return clientInformation, nil
	}
	return info.ClientInformation{}, errors.New("no ClientInformation exist")
}

// WriteStructToFile is used to write the payload to file
func WriteStructToFile(payload interface{}, dest string) {
	targetDir := filepath.Dir(dest)
	if !isFileOrFolderExisting(targetDir) {
		err := os.MkdirAll(targetDir, 0o755)
		if err != nil {
			log.Fatalf("Something went wrong: %q", err)
		}
	}
	file, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	if err = os.WriteFile(dest, file, 0o644); err != nil {
		log.Fatalf("Encountered error trying to write file %s: %q", file, err)
	}
}

// isFileOrFolderExisting checks either or not a target file is existing.
// Returns true if the target exists, otherwise false.
func isFileOrFolderExisting(target string) bool {
	if _, err := os.Stat(target); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	log.Panicf("Could not determine if file or folder %q exists or not. Exiting.", target)
	return false
}

// createCredentialsFile is used to create missing directories and file
func createCredentialsFile() {
	dir := filepath.Dir(CredentialsFilePath)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	f, err := os.OpenFile(CredentialsFilePath, os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	defer f.Close()
}

// writeTemplateToFile loads credentials to replace and write new.
// it calls replaceProfile to selectively delete a profile and replace it.
// it then saves the changes over the credentials file.
func writeTemplateToFile(template *CredentialsTemplate, profile string) {
	creds, err := ini.Load(CredentialsFilePath)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	replaceProfile(creds, template, profile)
	if err := creds.SaveTo(CredentialsFilePath); err != nil {
		log.Fatalf("Encountered error saving credentials: %q", err)
	}
}

// replaceProfile is used to selectively delete a profile from the credentials file.
// it then uses the new struct to populate the new information.
func replaceProfile(creds *ini.File, template *CredentialsTemplate, profile string) {
	creds.DeleteSection(profile)

	newSection, err := creds.NewSection(profile)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	// replace new section wtih template
	if err = newSection.ReflectFrom(template); err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
}
