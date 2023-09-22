package cmd

import (
	"aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"
	"aws-sso-util/internal/prompt"
	"log"
	"time"

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
			oidcClient, ssoClient := aws.CreateClients(ctx, region)
			promptSelector := prompt.Prompter{}

			oidc := aws.NewOIDCClient(oidcClient, startURL)
			sso := aws.NewSSOClient(ssoClient)

			clientInformation, _ := oidc.ProcessClientInformation(ctx)
			accountsOutput := sso.ListAccounts(ctx, clientInformation.AccessToken)
			accountInfo := prompt.RetrieveAccountInfo(accountsOutput, promptSelector)
			listRolesOutput := sso.ListAvailableRoles(ctx, *accountInfo.AccountId, clientInformation.AccessToken)
			roleInfo := prompt.RetrieveRoleInfo(listRolesOutput, promptSelector)
			file.SaveUsageInformation(accountInfo, roleInfo)

			roleCredentials, err := sso.GetRolesCredentials(ctx, *accountInfo.AccountId, *roleInfo.RoleName, clientInformation.AccessToken)
			if err != nil {
				log.Fatalf("Encountered error attempting to GetRoleCredentials: %v", err)
			}

			if len(accountID) == 0 {
				accountID = *accountInfo.AccountId
			}
			if len(roleName) == 0 {
				roleName = *roleInfo.RoleName
			}

			if persist {
				template := file.ProcessPersistedCredentialsTemplate(roleCredentials, profile)
				file.WriteAWSCredentialsFile(template)
				log.Printf("Credentails expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
			} else {
				fileInputs := file.GetCredentialFileInputs(
					accountID,
					roleName,
					profile,
					region,
					startURL,
				)
				template := file.ProcessCredentialProcessTemplate(fileInputs)
				file.WriteAWSCredentialsFile(template)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	loginCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	loginCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	loginCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
}
