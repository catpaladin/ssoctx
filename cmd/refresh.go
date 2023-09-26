package cmd

import (
	"log"
	"time"

	"aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"
	"aws-sso-util/internal/prompt"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh your previously used credentials",
	Long: `Refreshes the previously used credentials to the default profile.
  Use the flags to refresh profile credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := file.ReadConfig(file.ConfigFilePath())
		startURL = conf.StartURL
		region = conf.Region
		oidcClient, ssoClient := CreateClients(ctx, region)
		RefreshCredentials(oidcClient, ssoClient)
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)

	refreshCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "set with permission set role name")
	refreshCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "set account id for desired aws account")
	refreshCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	refreshCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
}

// RefreshCredentials is used to refresh credentials
func RefreshCredentials(oidcClient *ssooidc.Client, ssoClient *sso.Client) {
	oidc := aws.NewOIDCClient(oidcClient, startURL)
	sso := aws.NewSSOClient(ssoClient)

	clientInformation, err := file.ReadClientInformation(file.ClientInfoFileDestination())
	if err != nil || clientInformation.StartURL != startURL {
		clientInformation, _ = oidc.ProcessClientInformation(ctx)
	}
	log.Printf("Using Start URL %s", clientInformation.StartURL)

	if len(accountID) == 0 && len(roleName) == 0 {
		log.Println("No account-id or role-name provided.")
		log.Printf("Refreshing credentials from access to profile: %s", profile)
		accountsOutput := sso.ListAccounts(ctx, clientInformation.AccessToken)
		accountInfo := prompt.RetrieveAccountInfo(accountsOutput, prompt.Prompter{})
		listRolesOutput := sso.ListAvailableRoles(
			ctx,
			*accountInfo.AccountId,
			clientInformation.AccessToken,
		)
		roleInfo := prompt.RetrieveRoleInfo(listRolesOutput, prompt.Prompter{})
		roleName = *roleInfo.RoleName
		accountID = *accountInfo.AccountId
	}
	log.Printf("Refreshing for account %s with permission set role %s", accountID, roleName)

	roleCredentials, err := sso.GetRolesCredentials(
		ctx,
		accountID,
		roleName,
		clientInformation.AccessToken,
	)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	if persist {
		template := file.GetPersistedCredentials(roleCredentials, region)
		file.WriteAWSCredentialsFile(&template, profile)
		log.Printf("Credentails expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
	} else {
		template := file.GetCredentialProcess(accountID, roleName, region)
		file.WriteAWSCredentialsFile(&template, profile)
	}

	log.Printf("Successful retrieved credentials for account: %s", accountID)
	log.Printf("Assumed role: %s", roleName)
	log.Printf(
		"Credentials expire at: %s\n",
		time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0),
	)
}
