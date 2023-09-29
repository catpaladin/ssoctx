package cmd

import (
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
			logger := ConfigureLogger()
			ctx = logger.WithContext(ctx)

			file.GetConfigs(ctx, &startURL, &region)
			oidcClient, ssoClient := CreateClients(ctx, region)
			promptSelector := prompt.Prompter{Ctx: ctx}

			oidc := aws.NewOIDCClient(oidcClient, startURL)
			sso := aws.NewSSOClient(ssoClient)

			if clean {
				file.RemoveLock(ctx)
				destination, _ := file.ClientInfoFileDestination()
				if err := os.Remove(destination); err != nil {
					logger.Panic().Msgf("Failed to remove access token: %q", err)
				}
			}

			var accountInfo *types.AccountInfo
			clientInformation, err := oidc.ProcessClientInformation(ctx)
			if err != nil {
				clientInformation = reprocessCredentials(ctx, oidcClient, startURL)
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

			if len(startURL) == 0 {
				startURL = clientInformation.StartURL
			}

			roleCredentials, err := sso.GetRolesCredentials(ctx, accountID, roleName, clientInformation.AccessToken)
			if err != nil {
				logger.Fatal().Msgf("Encountered error attempting to GetRoleCredentials: %v", err)
			}

			if persist {
				template := file.GetPersistedCredentials(roleCredentials, region)
				file.WriteAWSCredentialsFile(ctx, &template, profile)
				logger.Info().Msgf("Credentails expire at: %s", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
			} else {
				template := file.GetCredentialProcess(accountID, roleName, region, startURL)
				file.WriteAWSCredentialsFile(ctx, &template, profile)
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
	loginCmd.Flags().BoolVarP(&debug, "debug", "", false, "toggle if you want to enable debug logs")
	loginCmd.Flags().BoolVarP(&jsonFormat, "json", "", false, "toggle if you want to enable json log output")
}
