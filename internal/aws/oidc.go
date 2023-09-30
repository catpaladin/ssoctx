// Package aws contains all the aws logic
package aws

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"aws-sso-util/internal/file"
	"aws-sso-util/internal/info"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/rs/zerolog"
)

const (
	grantType  string = "urn:ietf:params:oauth:grant-type:device_code"
	clientType string = "public"
)

var clientName = "aws-sso-util"

// OIDCClient is used to abstract the client calls for mock testing
type OIDCClient interface {
	CreateToken(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error)
	RegisterClient(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error)
	StartDeviceAuthorization(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error)
}

// OIDCClientAPI contains common info for sso oidc
type OIDCClientAPI struct {
	client OIDCClient
	url    string
}

// NewOIDCClient is used to implement the interface
func NewOIDCClient(c OIDCClient, url string) *OIDCClientAPI {
	return &OIDCClientAPI{
		client: c,
		url:    url,
	}
}

// ProcessClientInformation tries to read available ClientInformation
// If no ClientInformation is available or start url is overrideen, it will process new
// When the AccessToken is expired, it starts retrieving a new AccessToken
// A lock is added during the process to prevent concurrent authorizations
func (o *OIDCClientAPI) ProcessClientInformation(ctx context.Context) (info.ClientInformation, error) {
	logger := zerolog.Ctx(ctx)

	// check for running authorization process by file lock
	if file.LockStatus(ctx) {
		logger.Fatal().Msg("There is already an authorization process running. Please end any concurrent authorizations or run: aws-sso-util login --clean")
	}

	destination, err := file.ClientInfoFileDestination()
	if err != nil {
		logger.Fatal().Err(err)
	}
	clientInformation, err := file.ReadClientInformation(ctx, destination)
	if err != nil || clientInformation.StartURL != o.url {
		logger.Debug().Msg("Unable to read ClientInformation. Processing new ClientInformation")
		file.AddLock(ctx)
		defer file.RemoveLock(ctx)

		clientInfoPointer := o.getClientInfoPointer(ctx)
		file.WriteStructToFile(ctx, clientInfoPointer, destination)
		clientInformation = *clientInfoPointer
	} else if clientInformation.IsExpired(ctx) {
		logger.Debug().Msg("AccessToken expired. Start retrieving a new AccessToken.")
		file.AddLock(ctx)
		defer file.RemoveLock(ctx)

		clientInformation = o.handleOutdatedAccessToken(ctx, clientInformation)
	}
	return clientInformation, err
}

// handleOutdatedAccessToken handles client information if AccessToken is expired
func (o *OIDCClientAPI) handleOutdatedAccessToken(ctx context.Context, clientInformation info.ClientInformation) info.ClientInformation {
	logger := zerolog.Ctx(ctx)

	destination, err := file.ClientInfoFileDestination()
	if err != nil {
		logger.Fatal().Err(err)
	}

	registerClientOutput := ssooidc.RegisterClientOutput{ClientId: &clientInformation.ClientID, ClientSecret: &clientInformation.ClientSecret}
	deviceAuth, err := o.startDeviceAuthorization(ctx, &registerClientOutput, system)
	if err != nil {
		logger.Warn().Msg("Failed to authorize device. Regenerating AccessToken")
		clientInfoPointer := o.getClientInfoPointer(ctx)
		file.WriteStructToFile(ctx, clientInfoPointer, destination)
		return *clientInfoPointer
	}

	clientInformation.DeviceCode = *deviceAuth.DeviceCode
	clientInfoPointer, err := o.retrieveToken(ctx, &clientInformation)
	if err != nil {
		logger.Fatal().Msgf("Failed to create access token: %q", err)
	}
	file.WriteStructToFile(ctx, clientInfoPointer, destination)
	return *clientInfoPointer
}

// getClientInfoPointer handles registering and retrieving the token for client info
func (o *OIDCClientAPI) getClientInfoPointer(ctx context.Context) *info.ClientInformation {
	logger := zerolog.Ctx(ctx)
	clientInfo, err := o.registerClient(ctx)
	if err != nil {
		logger.Fatal().Msgf("Failed to create access token: %q", err)
	}
	clientInfoPointer, err := o.retrieveToken(ctx, clientInfo)
	if err != nil {
		logger.Fatal().Msgf("Failed to create access token: %q", err)
	}
	return clientInfoPointer
}

// RegisterClient is used to start device auth
func (o *OIDCClientAPI) registerClient(ctx context.Context) (*info.ClientInformation, error) {
	logger := zerolog.Ctx(ctx)

	input := ssooidc.RegisterClientInput{
		ClientName: aws.String(clientName),
		ClientType: aws.String(clientType),
	}
	output, err := o.client.RegisterClient(ctx, &input)
	if err != nil {
		awsErr := GetAWSErrorCode(ctx, err)
		logger.Debug().Msgf("Encountered error in RegisterClient: %s", awsErr)
		return &info.ClientInformation{}, err
	}

	deviceAuth, err := o.startDeviceAuthorization(ctx, output, system)
	if err != nil {
		logger.Debug().Msgf("Encountered error in startDeviceAuthorization: %q", err)
		return &info.ClientInformation{}, err
	}

	return &info.ClientInformation{
		ClientID:                *output.ClientId,
		ClientSecret:            *output.ClientSecret,
		ClientSecretExpiresAt:   strconv.FormatInt(output.ClientSecretExpiresAt, 10),
		DeviceCode:              *deviceAuth.DeviceCode,
		VerificationURIComplete: *deviceAuth.VerificationUriComplete,
		StartURL:                o.url,
	}, nil
}

// startDeviceAuthorization is used to start device auth and open browser
func (o *OIDCClientAPI) startDeviceAuthorization(ctx context.Context, rco *ssooidc.RegisterClientOutput, system string) (ssooidc.StartDeviceAuthorizationOutput, error) {
	logger := zerolog.Ctx(ctx)

	output, err := o.client.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     rco.ClientId,
		ClientSecret: rco.ClientSecret,
		StartUrl:     &o.url,
	})
	if err != nil {
		awsErr := GetAWSErrorCode(ctx, err)
		logger.Error().Msgf("Error in StartDeviceAuthorization: %s", awsErr)
		return ssooidc.StartDeviceAuthorizationOutput{}, fmt.Errorf("StartDeviceAuthorization returned %s", awsErr)
	}
	logger.Info().Msgf("Please verify your client request: " + *output.VerificationUriComplete)

	if err := openURLInBrowser(ctx, system, *output.VerificationUriComplete); err != nil {
		return ssooidc.StartDeviceAuthorizationOutput{}, err
	}

	return *output, nil
}

// retrieveToken is used to create the access token from the sso session
// this is obtained after auth in the browser.
func (o *OIDCClientAPI) retrieveToken(ctx context.Context, info *info.ClientInformation) (*info.ClientInformation, error) {
	logger := zerolog.Ctx(ctx)
	input := generateCreateTokenInput(info)

	cto, err := o.createToken(ctx, &input)
	if err != nil {
		logger.Debug().Msg("Encountered error in retrieveToken with createToken call")
		return info, err
	}
	info.AccessToken = *cto.AccessToken
	info.AccessTokenExpiresAt = time.Now().Add(time.Hour * 8)
	return info, nil
}

func (o *OIDCClientAPI) createToken(ctx context.Context, input *ssooidc.CreateTokenInput) (*ssooidc.CreateTokenOutput, error) {
	logger := zerolog.Ctx(ctx)
	// need loop to prevent errors while waiting on auth through browser
	for {
		cto, err := o.client.CreateToken(ctx, input)
		if err != nil {
			awsErr := GetAWSErrorCode(ctx, err)
			if awsErr == "AuthorizationPendingException" {
				logger.Info().Msg("Waiting on authorization..")
				time.Sleep(5 * time.Second)
				continue
			}
			return nil, fmt.Errorf("Encountered an error while createToken: %v", err)
		}
		return cto, nil
	}
}

// generateCreateTokenInput is used to create a CreateTokenInput
func generateCreateTokenInput(clientInformation *info.ClientInformation) ssooidc.CreateTokenInput {
	return ssooidc.CreateTokenInput{
		ClientId:     &clientInformation.ClientID,
		ClientSecret: &clientInformation.ClientSecret,
		DeviceCode:   &clientInformation.DeviceCode,
		GrantType:    aws.String(grantType),
	}
}
