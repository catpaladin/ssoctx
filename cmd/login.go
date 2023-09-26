package cmd

import (
	"log"
	"os"
	"time"

	"aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"
	"aws-sso-util/internal/prompt"

	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to AWS SSO",
		Long:  "Login to AWS SSO by retrieving short-lived credentials.",
		Run: func(cmd *cobra.Command, args []string) {
			file.GetConfigs(&startURL, &region)
			oidcClient, ssoClient := CreateClients(ctx, region)
			promptSelector := prompt.Prompter{}

			oidc := aws.NewOIDCClient(oidcClient, startURL)
			sso := aws.NewSSOClient(ssoClient)

			if clean {
				file.RemoveLock()
				if err := os.Remove(file.ClientInfoFileDestination()); err != nil {
					log.Panicf("Failed to remove access token: %q", err)
				}
			}

			var accountInfo *types.AccountInfo
			clientInformation, err := oidc.ProcessClientInformation(ctx)
			if err != nil {
				clientInformation = reprocessCredentials(oidcClient, startURL)
			}

			if len(accountID) == 0 {
				accountsOutput := sso.ListAccounts(ctx, clientInformation.AccessToken)
				accountInfo = prompt.RetrieveAccountInfo(accountsOutput, promptSelector)
				accountID = *accountInfo.AccountId
			}

			if len(roleName) == 0 {
				listRolesOutput := sso.ListAvailableRoles(ctx, accountID, clientInformation.AccessToken)
				roleInfo := prompt.RetrieveRoleInfo(listRolesOutput, promptSelector)
				roleName = *roleInfo.RoleName
			}

			roleCredentials, err := sso.GetRolesCredentials(ctx, accountID, roleName, clientInformation.AccessToken)
			if err != nil {
				log.Fatalf("Encountered error attempting to GetRoleCredentials: %v", err)
			}

			if persist {
				template := file.GetPersistedCredentials(roleCredentials, region)
				file.WriteAWSCredentialsFile(&template, profile)
				log.Printf("Credentails expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
			} else {
				template := file.GetCredentialProcess(accountID, roleName, region)
				file.WriteAWSCredentialsFile(&template, profile)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "set with permission set role name")
	loginCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "set account id for desired aws account")
	loginCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	loginCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	loginCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	loginCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
	loginCmd.Flags().BoolVarP(&clean, "clean", "", false, "toggle if you want to remove lock and access token")
}
