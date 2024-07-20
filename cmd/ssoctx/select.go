package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"

	"ssoctx/internal/amazon"
	"ssoctx/internal/file"
)

// selectCmd represents the login command
var (
	selectCmd = &cobra.Command{
		Use:   "select",
		Short: "Login to AWS SSO and select account and role",
		Long:  "Login to AWS SSO by retrieving short-lived credentials for account and role.",
		Run: func(cmd *cobra.Command, args []string) {
			logger := configureLogger(debug, jsonFormat)
			ctx = logger.WithContext(ctx)

			file.GetConfigs(ctx, &startURL, &region)
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

			amazon.Select(ctx, oidc, sso, amazon.SelectFlagInputs{
				AccountID:  accountID,
				RoleName:   roleName,
				Profile:    profile,
				StartURL:   startURL,
				Region:     region,
				Persist:    persist,
				Clean:      clean,
				PrintCreds: printCreds,
			})
		},
	}
)

func init() {
	rootCmd.AddCommand(selectCmd)
	selectCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "set with permission set role name")
	selectCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "set account id for desired aws account")
	selectCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	selectCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	selectCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	selectCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
	selectCmd.Flags().BoolVarP(&clean, "clean", "", false, "toggle if you want to remove lock and access token")
	selectCmd.Flags().BoolVarP(&debug, "debug", "", false, "toggle if you want to enable debug logs")
	selectCmd.Flags().BoolVarP(&jsonFormat, "json", "", false, "toggle if you want to enable json log output")
	selectCmd.Flags().BoolVarP(&printCreds, "print-creds", "", false, "outputs the credentials to stdout and not modifying credentials file")
}
