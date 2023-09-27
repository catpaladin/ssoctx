package cmd

import (
	"context"
	"fmt"
	"os"

	awsSSO "aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"
	"aws-sso-util/internal/info"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/rs/zerolog"
)

// CreateClients is used to return sso and ssooidc clients
func CreateClients(ctx context.Context, region string) (*ssooidc.Client, *sso.Client) {
	logger := zerolog.Ctx(ctx)
	cfg, err := loadConfig(ctx, region)
	if err != nil {
		logger.Fatal().Msgf("Encountered error in CreateClients: %v", err)
	}

	oidcClient := ssooidc.NewFromConfig(cfg)
	ssoClient := sso.NewFromConfig(cfg)

	return oidcClient, ssoClient
}

// loadConfig used to load the default config
func loadConfig(ctx context.Context, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		errCode := awsSSO.GetAWSErrorCode(ctx, err)
		return aws.Config{}, fmt.Errorf("%s in loadConfig", errCode)
	}
	return cfg, nil
}

// reprocessCredentials is used to handle the retry when access token is invalid.
// this makes it so users do not have to manually delete their access token
// or overloading handleOutdatedAccessToken with similar logic.
func reprocessCredentials(ctx context.Context, oidcClient *ssooidc.Client, startURL string) info.ClientInformation {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("Error in Access Token. Reprocessing Credentials.")
	destination, _ := file.ClientInfoFileDestination()
	if err := os.Remove(destination); err != nil {
		logger.Fatal().Msgf("Encountered error in reprocessCredentials: %v", err)
	}
	oidcCli := awsSSO.NewOIDCClient(oidcClient, startURL)

	clientInformation, err := oidcCli.ProcessClientInformation(ctx)
	if err != nil {
		logger.Fatal().Msgf("Encountered error in reprocessCredentials: %v", err)
	}

	return clientInformation
}
