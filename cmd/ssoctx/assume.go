package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"

	"ssoctx/internal/amazon"
	"ssoctx/internal/file"
)

var assumeCmd = &cobra.Command{
	Use:   "assume",
	Short: "Assume directly into an account and SSO role",
	Long: `Assume directly into an account and SSO role.
		This is used by the aws default profile.`,
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

		amazon.AssumeCredentialProcess(ctx, oidc, sso, amazon.AssumeFlagInputs{
			AccountID: accountID,
			RoleName:  roleName,
			Profile:   profile,
			StartURL:  startURL,
			Region:    region,
		})
	},
}

func init() {
	rootCmd.AddCommand(assumeCmd)
	assumeCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	assumeCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	assumeCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	assumeCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "set account id for desired aws account")
	assumeCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "set / override with permission set role name")
	assumeCmd.Flags().BoolVarP(&debug, "debug", "", false, "toggle if you want to enable debug logs")
	assumeCmd.Flags().BoolVarP(&jsonFormat, "json", "", false, "toggle if you want to enable json log output")
}
