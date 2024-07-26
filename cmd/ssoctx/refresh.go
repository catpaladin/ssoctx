package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"

	"ssoctx/internal/amazon"
	"ssoctx/internal/file"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh your previously used credentials",
	Long: `Refreshes the previously used credentials to the default profile.
  Use the flags to refresh profile credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := configureLogger(debug, jsonFormat)
		ctx = logger.WithContext(ctx)

		conf := file.ReadConfig(ctx, file.GetConfigFilePath(ctx))
		startURL = conf.StartURL
		region = conf.Region
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in loading default aws config: %v", err)
		}
		oidcClient, ssoClient := amazon.NewClients(cfg)
		oidc := amazon.NewOIDCClient(oidcClient, startURL)
		sso := amazon.NewSSOClient(ssoClient)

		amazon.Credentials(ctx, oidc, sso, amazon.RefreshFlagInputs{
			AccountID: accountID,
			RoleName:  roleName,
			Profile:   profile,
			StartURL:  startURL,
			Region:    region,
			Keys:      keys,
		})
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
	refreshCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "set with permission set role name")
	refreshCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "set account id for desired aws account")
	refreshCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	refreshCmd.Flags().BoolVarP(&keys, "keys", "", false, "toggle if you want to write access/secret keys to credentials file")
	refreshCmd.Flags().BoolVarP(&debug, "debug", "", false, "toggle if you want to enable debug logs")
	refreshCmd.Flags().BoolVarP(&jsonFormat, "json", "", false, "toggle if you want to enable json log output")
}
