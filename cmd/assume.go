package cmd

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

// AssumeDirectly is used to assume sso role directly.
// Directly assumes into a certain account and role, bypassing the prompt and interactive selection.
func AssumeDirectly(oidcClient *ssooidc.Client, ssoClient *sso.Client) {
	oidcInformation := OIDCInformation{
		Client: oidcClient,
		URL:    startURL,
	}
	clientInformation, _ := oidcInformation.ProcessClientInformation()
	rci := &sso.GetRoleCredentialsInput{AccountId: &accountID, RoleName: &roleName, AccessToken: &clientInformation.AccessToken}
	roleCredentials, err := ssoClient.GetRoleCredentials(ctx, rci)
	check(err)

	if persist {
		template := ProcessPersistedCredentialsTemplate(roleCredentials, profile)
		WriteAWSCredentialsFile(template)

		log.Printf("Successful retrieved credentials for account: %s", accountID)
		log.Printf("Assumed role: %s", roleName)
		log.Printf("Credentials expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
	} else {
		template := ProcessCredentialProcessTemplate(accountID, roleName, profile, region)
		WriteAWSCredentialsFile(template)

		creds := CredentialProcessOutput{
			Version:         1,
			AccessKeyID:     *roleCredentials.RoleCredentials.AccessKeyId,
			Expiration:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			SecretAccessKey: *roleCredentials.RoleCredentials.SecretAccessKey,
			SessionToken:    *roleCredentials.RoleCredentials.SessionToken,
		}
		bytes, _ := json.Marshal(creds)
		os.Stdout.Write(bytes)
	}

}

// CredentialProcessOutput is used to marshal results from GetRoleCredentials
type CredentialProcessOutput struct {
	Version         int    `json:"Version"`
	AccessKeyID     string `json:"AccessKeyId"`
	Expiration      string `json:"Expiration"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
}
