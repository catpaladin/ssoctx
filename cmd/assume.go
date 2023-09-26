package cmd

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/spf13/cobra"
)

var assumeCmd = &cobra.Command{
	Use:   "assume",
	Short: "Assume directly into an account and SSO role",
	Long: `Assume directly into an account and SSO role.
		This is used by the aws default profile.`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := file.ReadConfig(file.ConfigFilePath())
		startURL = conf.StartURL
		region = conf.Region
		oidcClient, ssoClient := CreateClients(ctx, region)
		AssumeDirectly(oidcClient, ssoClient)
	},
}

func init() {
	rootCmd.AddCommand(assumeCmd)

	// flags
	assumeCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	assumeCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	assumeCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	assumeCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
	assumeCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "role name to assume")
	assumeCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "account id where the role exists")
}

// AssumeDirectly is used to assume sso role directly.
// Directly assumes into a certain account and role, bypassing the prompt and interactive selection.
func AssumeDirectly(oidcClient *ssooidc.Client, ssoClient *sso.Client) {
	oidc := aws.NewOIDCClient(oidcClient, startURL)
	clientInformation, _ := oidc.ProcessClientInformation(ctx)
	rci := &sso.GetRoleCredentialsInput{AccountId: &accountID, RoleName: &roleName, AccessToken: &clientInformation.AccessToken}
	roleCredentials, err := ssoClient.GetRoleCredentials(ctx, rci)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	if persist {
		template := file.GetPersistedCredentials(roleCredentials, region)
		file.WriteAWSCredentialsFile(&template, profile)

		log.Printf("Successful retrieved credentials for account: %s", accountID)
		log.Printf("Assumed role: %s", roleName)
		log.Printf("Credentials expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
	} else {
		template := file.GetCredentialProcess(accountID, roleName, region)
		file.WriteAWSCredentialsFile(&template, profile)

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
