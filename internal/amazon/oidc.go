package amazon

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/rs/zerolog"

	"ssoctx/internal/file"
)

var (
	createTokenTimeout = 300 * time.Second
	execCmd            = exec.Command
)

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
func (o *OIDCClientAPI) processClientInformation(ctx context.Context, fileDestination string) (ClientInformation, error) {
	logger := zerolog.Ctx(ctx)

	// check for running authorization process by file lock
	if file.IsLocked(ctx) {
		return ClientInformation{}, fmt.Errorf("There is already an authorization process running. Please end any concurrent authorizations or run: %s login --clean", ProjectFileName)
	}

	clientInfo, err := readClientInformation(ctx, fileDestination)
	if err != nil || clientInfo.StartURL != o.url || clientInfo.isExpired(ctx) {
		logger.Debug().Msg("Issue with AccessToken. Generating new AccessToken.")
		file.AddLock(ctx)
		defer file.RemoveLock(ctx)

		clientInfo, err := o.getClientInfoPointer(ctx)
		if err != nil {
			return ClientInformation{}, err
		}
		return *clientInfo, nil
	}
	return clientInfo, nil
}

// getClientInfoPointer handles registering and retrieving the token for client info
func (o *OIDCClientAPI) getClientInfoPointer(ctx context.Context) (*ClientInformation, error) {
	clientInfo, err := o.registerClient(ctx)
	if err != nil {
		return &ClientInformation{}, err
	}
	clientInfoPointer, err := o.retrieveToken(ctx, clientInfo)
	if err != nil {
		return &ClientInformation{}, err
	}
	return clientInfoPointer, nil
}

// RegisterClient is used to start device auth
func (o *OIDCClientAPI) registerClient(ctx context.Context) (*ClientInformation, error) {
	input := ssooidc.RegisterClientInput{
		ClientName: aws.String(ProjectFileName),
		ClientType: aws.String(clientType),
	}
	output, err := o.client.RegisterClient(ctx, &input)
	if err != nil {
		_ = GetAWSErrorCode(ctx, err)
		return &ClientInformation{}, err
	}

	deviceAuth, err := o.startDeviceAuthorization(ctx, output)
	if err != nil {
		return &ClientInformation{}, err
	}

	return &ClientInformation{
		ClientID:                *output.ClientId,
		ClientSecret:            *output.ClientSecret,
		ClientSecretExpiresAt:   strconv.FormatInt(output.ClientSecretExpiresAt, 10),
		DeviceCode:              *deviceAuth.DeviceCode,
		VerificationURIComplete: *deviceAuth.VerificationUriComplete,
		StartURL:                o.url,
	}, nil
}

// startDeviceAuthorization is used to start device auth and open browser
func (o *OIDCClientAPI) startDeviceAuthorization(ctx context.Context, rco *ssooidc.RegisterClientOutput) (ssooidc.StartDeviceAuthorizationOutput, error) {
	logger := zerolog.Ctx(ctx)

	output, err := o.client.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     rco.ClientId,
		ClientSecret: rco.ClientSecret,
		StartUrl:     &o.url,
	})
	if err != nil {
		_ = GetAWSErrorCode(ctx, err)
		return ssooidc.StartDeviceAuthorizationOutput{}, err
	}
	logger.Info().Msgf("Please verify your client request: " + *output.VerificationUriComplete)

	if err := openURLInBrowser(*output.VerificationUriComplete); err != nil {
		return ssooidc.StartDeviceAuthorizationOutput{}, err
	}

	return *output, nil
}

// retrieveToken is used to create the access token from the sso session
// this is obtained after auth in the browser.
func (o *OIDCClientAPI) retrieveToken(ctx context.Context, info *ClientInformation) (*ClientInformation, error) {
	input := generateCreateTokenInput(info)

	cto, err := o.createToken(ctx, &input)
	if err != nil {
		return info, err
	}
	info.AccessToken = *cto.AccessToken
	info.AccessTokenExpiresAt = time.Now().Add(time.Hour * 8)
	return info, nil
}

func (o *OIDCClientAPI) createToken(ctx context.Context, input *ssooidc.CreateTokenInput) (*ssooidc.CreateTokenOutput, error) {
	logger := zerolog.Ctx(ctx)
	// need loop to prevent errors while waiting on auth through browser
	for start := time.Now(); time.Since(start) < createTokenTimeout; {
		cto, err := o.client.CreateToken(ctx, input)
		if err != nil {
			awsErrCode := GetAWSErrorCode(ctx, err)
			if awsErrCode == "AuthorizationPendingException" {
				logger.Info().Msg("Waiting on authorization..")
				time.Sleep(5 * time.Second)
				continue
			}
		}
		return cto, nil
	}
	return &ssooidc.CreateTokenOutput{}, errors.New("Encountered timeout in createToken")
}

// generateCreateTokenInput is used to create a CreateTokenInput
func generateCreateTokenInput(clientInformation *ClientInformation) ssooidc.CreateTokenInput {
	return ssooidc.CreateTokenInput{
		ClientId:     &clientInformation.ClientID,
		ClientSecret: &clientInformation.ClientSecret,
		DeviceCode:   &clientInformation.DeviceCode,
		GrantType:    aws.String(grantType),
	}
}
