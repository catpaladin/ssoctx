package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

// CreateClients is used to return sso and ssooidc clients
func CreateClients(ctx context.Context, region string) (*ssooidc.Client, *sso.Client) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		log.Panicf("Encountered error in InitClients: %v", err)
	}

	oidcClient := ssooidc.NewFromConfig(cfg)
	ssoClient := sso.NewFromConfig(cfg)

	return oidcClient, ssoClient
}
