package cmd

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var assumeCmd = &cobra.Command{
	Use:   "assume",
	Short: "Assume directly into an account and SSO role",
	Long: `Assume directly into an account and SSO role.
		This is used by the aws default profile.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := ConfigureLogger()
		ctx = logger.WithContext(ctx)

		conf := file.ReadConfig(ctx, file.ConfigFilePath(ctx))
		startURL = conf.StartURL
		region = conf.Region
		oidcClient, ssoClient := CreateClients(ctx, region)
		AssumeDirectly(ctx, oidcClient, ssoClient)
	},
}

func init() {
	rootCmd.AddCommand(assumeCmd)

	// flags
	assumeCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	assumeCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	assumeCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	assumeCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
	assumeCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "set account id for desired aws account")
	assumeCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "set / override with permission set role name")
	assumeCmd.Flags().BoolVarP(&debug, "debug", "", false, "toggle if you want to enable debug logs")
	assumeCmd.Flags().BoolVarP(&jsonFormat, "json", "", false, "toggle if you want to enable json log output")
}

// AssumeDirectly is used to assume sso role directly.
// Directly assumes into a certain account and role, bypassing the prompt and interactive selection.
func AssumeDirectly(ctx context.Context, oidcClient *ssooidc.Client, ssoClient *sso.Client) {
	logger := zerolog.Ctx(ctx)

	oidc := aws.NewOIDCClient(oidcClient, startURL)
	clientInformation, _ := oidc.ProcessClientInformation(ctx)
	rci := &sso.GetRoleCredentialsInput{AccountId: &accountID, RoleName: &roleName, AccessToken: &clientInformation.AccessToken}
	roleCredentials, err := ssoClient.GetRoleCredentials(ctx, rci)
	if err != nil {
		logger.Fatal().Msgf("Something went wrong: %q", err)
	}

	if len(startURL) == 0 {
		startURL = clientInformation.StartURL
	}

	if persist {
		template := file.GetPersistedCredentials(roleCredentials, region)
		file.WriteAWSCredentialsFile(ctx, &template, profile)

		logger.Info().Msgf("Successful retrieved credentials for account: %s", accountID)
		logger.Info().Msgf("Assumed role: %s", roleName)
		logger.Info().Msgf("Credentials expire at: %s", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
	} else {
		template := file.GetCredentialProcess(accountID, roleName, region, startURL)
		file.WriteAWSCredentialsFile(ctx, &template, profile)

		creds := CredentialProcessOutput{
			Version:         1,
			AccessKeyID:     *roleCredentials.RoleCredentials.AccessKeyId,
			Expiration:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			SecretAccessKey: *roleCredentials.RoleCredentials.SecretAccessKey,
			SessionToken:    *roleCredentials.RoleCredentials.SessionToken,
		}
		bytes, _ := json.Marshal(creds)
		os.Stdout.Write(bytes)
	}
}

// CredentialProcessOutput is used to marshal results from GetRoleCredentials
type CredentialProcessOutput struct {
	Version         int    `json:"Version"`
	AccessKeyID     string `json:"AccessKeyId"`
	Expiration      string `json:"Expiration"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
}
