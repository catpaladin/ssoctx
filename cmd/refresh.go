package cmd

import (
	"context"
	"time"

	"aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"
	"aws-sso-util/internal/prompt"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh your previously used credentials",
	Long: `Refreshes the previously used credentials to the default profile.
  Use the flags to refresh profile credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := ConfigureLogger()
		ctx = logger.WithContext(ctx)

		conf := file.ReadConfig(ctx, file.ConfigFilePath(ctx))
		startURL = conf.StartURL
		region = conf.Region
		oidcClient, ssoClient := CreateClients(ctx, region)
		RefreshCredentials(ctx, oidcClient, ssoClient)
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)

	refreshCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "set with permission set role name")
	refreshCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "set account id for desired aws account")
	refreshCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	refreshCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
	refreshCmd.Flags().BoolVarP(&debug, "debug", "", false, "toggle if you want to enable debug logs")
	refreshCmd.Flags().BoolVarP(&jsonFormat, "json", "", false, "toggle if you want to enable json log output")
}

// RefreshCredentials is used to refresh credentials
func RefreshCredentials(ctx context.Context, oidcClient *ssooidc.Client, ssoClient *sso.Client) {
	logger := zerolog.Ctx(ctx)

	oidc := aws.NewOIDCClient(oidcClient, startURL)
	sso := aws.NewSSOClient(ssoClient)

	destination, err := file.ClientInfoFileDestination()
	if err != nil {
		logger.Fatal().Err(err)
	}
	clientInformation, err := file.ReadClientInformation(ctx, destination)
	if err != nil || clientInformation.StartURL != startURL {
		clientInformation, _ = oidc.ProcessClientInformation(ctx)
	}
	logger.Printf("Using Start URL %s", clientInformation.StartURL)

	if len(accountID) == 0 && len(roleName) == 0 {
		logger.Info().Msg("No account-id or role-name provided.")
		logger.Info().Msgf("Refreshing credentials from access to profile: %s", profile)

		promptSelector := prompt.Prompter{Ctx: ctx}
		accountsOutput := sso.ListAccounts(ctx, clientInformation.AccessToken)
		accountInfo := prompt.RetrieveAccountInfo(accountsOutput, promptSelector)
		listRolesOutput := sso.ListAvailableRoles(
			ctx,
			*accountInfo.AccountId,
			clientInformation.AccessToken,
		)
		roleInfo := prompt.RetrieveRoleInfo(listRolesOutput, promptSelector)
		roleName = *roleInfo.RoleName
		accountID = *accountInfo.AccountId
	}
	logger.Info().Msgf("Refreshing for account %s with permission set role %s", accountID, roleName)

	roleCredentials, err := sso.GetRolesCredentials(
		ctx,
		accountID,
		roleName,
		clientInformation.AccessToken,
	)
	if err != nil {
		logger.Fatal().Msgf("Something went wrong: %q", err)
	}

	if len(startURL) == 0 {
		startURL = clientInformation.StartURL
	}

	if persist {
		template := file.GetPersistedCredentials(roleCredentials, region)
		file.WriteAWSCredentialsFile(ctx, &template, profile)
	} else {
		template := file.GetCredentialProcess(accountID, roleName, region, startURL)
		file.WriteAWSCredentialsFile(ctx, &template, profile)
	}

	logger.Info().Msgf("Successful retrieved credentials for account: %s", accountID)
	logger.Info().Msgf("Assumed role: %s", roleName)
	logger.Info().Msgf(
		"Credentials expire at: %s",
		time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0),
	)
}
