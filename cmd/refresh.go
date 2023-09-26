package cmd

import (
	"log"
	"strings"
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
	Long:  `Refreshes the credentials based on last account and role.`,
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

	var accountID *string
	var roleName *string

	lui, err := file.ReadUsageInformation()
	log.Printf(
		"Attempting to refresh credentials for account [%s] with role [%s]",
		lui.AccountName,
		lui.Role,
	)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			log.Println("Nothing to refresh yet.")
			accountsOutput := sso.ListAccounts(ctx, clientInformation.AccessToken)
			accountInfo := prompt.RetrieveAccountInfo(accountsOutput, prompt.Prompter{})
			listRolesOutput := sso.ListAvailableRoles(
				ctx,
				*accountInfo.AccountId,
				clientInformation.AccessToken,
			)
			roleInfo := prompt.RetrieveRoleInfo(listRolesOutput, prompt.Prompter{})
			roleName = roleInfo.RoleName
			accountID = accountInfo.AccountId
			file.SaveUsageInformation(accountInfo, roleInfo)
		}
	} else {
		accountID = &lui.AccountID
		roleName = &lui.Role
	}

	roleCredentials, err := sso.GetRolesCredentials(
		ctx,
		*accountID,
		*roleName,
		clientInformation.AccessToken,
	)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	template := file.GetCredentialProcess(*accountID, *roleName, region)
	file.WriteAWSCredentialsFile(&template, profile)

	log.Printf("Successful retrieved credentials for account: %s", *accountID)
	log.Printf("Assumed role: %s", *roleName)
	log.Printf(
		"Credentials expire at: %s\n",
		time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0),
	)
}
