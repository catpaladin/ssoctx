package cmd

import (
	"context"
	"log"
	"os"

	awsSSO "aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"
	"aws-sso-util/internal/info"
	"aws-sso-util/internal/prompt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

// CreateClients is used to return sso and ssooidc clients
func CreateClients(ctx context.Context, region string) (*ssooidc.Client, *sso.Client) {
	cfg, err := loadConfig(ctx, region)
	if err != nil {
		log.Panicf("Encountered error in CreateClients: %v", err)
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
		return aws.Config{}, err
	}
	return cfg, nil
}

// reprocessCredentials is used to handle the retry when access token is invalid.
// this makes it so users do not have to manually delete their access token
// or overloading handleOutdatedAccessToken with similar logic.
func reprocessCredentials(oidcClient *ssooidc.Client, ssoClient *sso.Client, startURL string, selector prompt.Prompter) (info.ClientInformation, *types.AccountInfo) {
	log.Println("Error in Access Token. Reprocessing Credentials.")
	if err := os.Remove(file.ClientInfoFileDestination()); err != nil {
		log.Fatalf("Encountered error in reprocessCredentials: %v", err)
	}
	oidcCli := awsSSO.NewOIDCClient(oidcClient, startURL)
	ssoCli := awsSSO.NewSSOClient(ssoClient)

	clientInformation, err := oidcCli.ProcessClientInformation(ctx)
	if err != nil {
		log.Fatalf("Encountered error in reprocessCredentials: %v", err)
	}
	accountInfoOutput := ssoCli.ListAccounts(ctx, clientInformation.AccessToken)
	accountInfo := prompt.RetrieveAccountInfo(accountInfoOutput, selector)

	return clientInformation, accountInfo
}
