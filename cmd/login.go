package cmd

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to AWS SSO",
		Long:  "Login to AWS SSO by retrieving short-lived credentials.",
		Run: func(cmd *cobra.Command, args []string) {
			GetConfigs(&startURL, &region)
			oidcClient, ssoClient := InitClients(region)
			promptSelector := Prompter{}

			oidcInformation := OIDCInformation{
				Client: oidcClient,
				URL:    startURL,
			}
			clientInformation, _ := oidcInformation.ProcessClientInformation()

			accountInfo := RetrieveAccountInfo(ssoClient, clientInformation, promptSelector)
			roleInfo := RetrieveRoleInfo(ssoClient, accountInfo, clientInformation, promptSelector)
			SaveUsageInformation(accountInfo, roleInfo)

			input := &sso.GetRoleCredentialsInput{
				AccountId:   accountInfo.AccountId,
				RoleName:    roleInfo.RoleName,
				AccessToken: &clientInformation.AccessToken,
			}
			roleCredentials, err := ssoClient.GetRoleCredentials(ctx, input)
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
				template := ProcessPersistedCredentialsTemplate(roleCredentials, profile)
				WriteAWSCredentialsFile(template)
				log.Printf("Credentails expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
			} else {
				template := ProcessCredentialProcessTemplate(CredentialProcessInputs{
					accountID: accountID,
					roleName:  roleName,
					profile:   profile,
					region:    region,
					startURL:  startURL,
				})
				WriteAWSCredentialsFile(template)
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
