package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"

	"ssoctx/internal/amazon"
	"ssoctx/internal/file"
)

// loginCmd represents the login command
var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to AWS SSO",
		Long:  "Login to AWS SSO by retrieving short-lived credentials.",
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

			amazon.Login(ctx, oidc, sso, amazon.LoginFlagInputs{
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
	loginCmd.Flags().BoolVarP(&printCreds, "print-creds", "", false, "outputs the credentials to stdout and not modifying credentials file")
}
