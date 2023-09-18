package file

import (
	"aws-sso-util/internal/client"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/valyala/fasttemplate"
)

// CredentialsFilePath is used to store the credentials path to variable
var CredentialsFilePath = GetCredentialsFilePath()

// CredentialProcessInputs contains inputs needed to write credentials
type CredentialProcessInputs struct {
	accountID string
	roleName  string
	profile   string
	region    string
	startURL  string
}

// GetCredentialFileInputs takes inputs and returns a CredentialProcessInputs struct
func GetCredentialFileInputs(accountID, roleName, profile, region, startURL string) CredentialProcessInputs {
	return CredentialProcessInputs{
		accountID: accountID,
		roleName:  roleName,
		profile:   profile,
		region:    region,
		startURL:  startURL,
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

// ProcessPersistedCredentialsTemplate is used to template the persisted credentials file
func ProcessPersistedCredentialsTemplate(credentials *sso.GetRoleCredentialsOutput, profile string) string {
	template := `[{{profile}}]
aws_access_key_id = {{access_key_id}}
aws_secret_access_key = {{secret_access_key}}
aws_session_token = {{session_token}}
output = json
region = us-east-1
`

	engine := fasttemplate.New(template, "{{", "}}")
	filledTemplate := engine.ExecuteString(map[string]interface{}{
		"profile":           profile,
		"access_key_id":     *credentials.RoleCredentials.AccessKeyId,
		"secret_access_key": *credentials.RoleCredentials.SecretAccessKey,
		"session_token":     *credentials.RoleCredentials.SessionToken,
	})
	return filledTemplate
}

// ProcessCredentialProcessTemplate is used to template the direct assume
func ProcessCredentialProcessTemplate(credentialInputs CredentialProcessInputs) string {
	template := `[{{profile}}]
credential_process = aws-sso-util assume -a {{accountId}} -n {{roleName}}
region = {{region}}
`

	engine := fasttemplate.New(template, "{{", "}}")
	filledTemplate := engine.ExecuteString(map[string]interface{}{
		"profile":   credentialInputs.profile,
		"region":    credentialInputs.region,
		"accountId": credentialInputs.accountID,
		"roleName":  credentialInputs.roleName,
	})
	return filledTemplate
}

// WriteAWSCredentialsFile is used to write the template to credentials
func WriteAWSCredentialsFile(template string) {
	if !isFileOrFolderExisting(CredentialsFilePath) {
		dir := filepath.Dir(CredentialsFilePath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Something went wrong: %q", err)
		}
		f, err := os.OpenFile(CredentialsFilePath, os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Something went wrong: %q", err)
		}
		defer f.Close()
	}
	err := os.WriteFile(CredentialsFilePath, []byte(template), 0644)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
}

// ReadClientInformation is used to read file for ClientInformation
func ReadClientInformation(file string) (client.ClientInformation, error) {
	if isFileOrFolderExisting(file) {
		clientInformation := client.ClientInformation{}
		content, _ := os.ReadFile(ClientInfoFileDestination())
		err := json.Unmarshal(content, &clientInformation)
		if err != nil {
			log.Fatalf("Something went wrong: %q", err)
		}
		return clientInformation, nil
	}
	return client.ClientInformation{}, errors.New("no ClientInformation exist")
}

// WriteStructToFile is used to write the payload to file
func WriteStructToFile(payload interface{}, dest string) {
	targetDir := filepath.Dir(dest)
	if !isFileOrFolderExisting(targetDir) {
		err := os.MkdirAll(targetDir, 0700)
		if err != nil {
			log.Fatalf("Something went wrong: %q", err)
		}
	}
	file, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	_ = os.WriteFile(dest, file, 0600)
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
