// Package aws contains all the aws logic
package aws

import (
	"aws-sso-util/internal/client"
	"aws-sso-util/internal/file"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/smithy-go"
)

const (
	grantType  string = "urn:ietf:params:oauth:grant-type:device_code"
	clientType string = "public"
	clientName string = "aws-sso-util"
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

// ProcessClientInformation tries to read available ClientInformation.
// If no ClientInformation is available it retrieves and creates new one and writes this information to disk
// If the start url is overridden via flag and differs from the previous one, a new Client is registered for the given start url.
// When the ClientInformation.AccessToken is expired, it starts retrieving a new AccessToken
func (o OIDCClientAPI) ProcessClientInformation(ctx context.Context) (client.ClientInformation, error) {
	clientInformation, err := file.ReadClientInformation(file.ClientInfoFileDestination())
	if err != nil || clientInformation.StartURL != o.url {
		var clientInfoPointer *client.ClientInformation
		clientInfoPointer = o.registerClient(ctx)
		clientInfoPointer = o.retrieveToken(ctx, clientInfoPointer)
		file.WriteStructToFile(clientInfoPointer, file.ClientInfoFileDestination())
		clientInformation = *clientInfoPointer
	} else if clientInformation.IsExpired() {
		log.Println("AccessToken expired. Start retrieving a new AccessToken.")
		clientInformation = o.handleOutdatedAccessToken(ctx, clientInformation)
	}
	return clientInformation, err
}

// handleOutdatedAccessToken handles client information if AccessToken is expired
func (o OIDCClientAPI) handleOutdatedAccessToken(ctx context.Context, clientInformation client.ClientInformation) client.ClientInformation {
	registerClientOutput := ssooidc.RegisterClientOutput{ClientId: &clientInformation.ClientID, ClientSecret: &clientInformation.ClientSecret}
	deviceAuth, err := o.startDeviceAuthorization(ctx, &registerClientOutput)
	if err != nil {
		log.Println("Failed to authorize device. Regenerating AccessToken")
		var clientInfoPointer *client.ClientInformation
		clientInfoPointer = o.registerClient(ctx)
		clientInfoPointer = o.retrieveToken(ctx, clientInfoPointer)
		file.WriteStructToFile(clientInfoPointer, file.ClientInfoFileDestination())
		return *clientInfoPointer
	}

	clientInformation.DeviceCode = *deviceAuth.DeviceCode
	var clientInfoPointer *client.ClientInformation
	clientInfoPointer = o.retrieveToken(ctx, &clientInformation)
	file.WriteStructToFile(clientInfoPointer, file.ClientInfoFileDestination())
	return *clientInfoPointer
}

// RegisterClient is used to start device auth
func (o OIDCClientAPI) registerClient(ctx context.Context) *client.ClientInformation {
	cn := clientName
	ct := clientType

	input := ssooidc.RegisterClientInput{ClientName: &cn, ClientType: &ct}
	output, err := o.client.RegisterClient(ctx, &input)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	deviceAuth, err := o.startDeviceAuthorization(ctx, output)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	return &client.ClientInformation{
		ClientID:                *output.ClientId,
		ClientSecret:            *output.ClientSecret,
		ClientSecretExpiresAt:   strconv.FormatInt(output.ClientSecretExpiresAt, 10),
		DeviceCode:              *deviceAuth.DeviceCode,
		VerificationURIComplete: *deviceAuth.VerificationUriComplete,
		StartURL:                o.url,
	}
}

// startDeviceAuthorization is used to start device auth and open browser
func (o OIDCClientAPI) startDeviceAuthorization(ctx context.Context, rco *ssooidc.RegisterClientOutput) (ssooidc.StartDeviceAuthorizationOutput, error) {
	output, err := o.client.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     rco.ClientId,
		ClientSecret: rco.ClientSecret,
		StartUrl:     &o.url,
	})
	if err != nil {
		return ssooidc.StartDeviceAuthorizationOutput{}, fmt.Errorf("Encountered error at startDeviceAuthorization: %w", err)
	}
	log.Println("Please verify your client request: " + *output.VerificationUriComplete)
	OpenURLInBrowser(*output.VerificationUriComplete)
	return *output, nil
}

// retrieveToken is used to create the access token from the sso session
// this is obtained after auth in the browser.
func (o OIDCClientAPI) retrieveToken(ctx context.Context, info *client.ClientInformation) *client.ClientInformation {
	input := generateCreateTokenInput(info)
	// need loop to prevent errors while waiting on auth through browser
	for {
		cto, err := o.client.CreateToken(ctx, &input)
		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) {
				if ae.ErrorCode() == "AuthorizationPendingException" {
					log.Println("Waiting on authorization..")
					time.Sleep(5 * time.Second)
					continue
				}
				log.Fatalf("Encountered an error while retrieveToken: %v", err)
			}
		}
		info.AccessToken = *cto.AccessToken
		info.AccessTokenExpiresAt = time.Now().Add(time.Hour * 8)
		//info.AccessTokenExpiresAt = timer.Now().Add(time.Hour*8 - time.Minute*5)
		return info
	}
}

// generateCreateTokenInput is used to create a CreateTokenInput
func generateCreateTokenInput(clientInformation *client.ClientInformation) ssooidc.CreateTokenInput {
	gtp := grantType
	return ssooidc.CreateTokenInput{
		ClientId:     &clientInformation.ClientID,
		ClientSecret: &clientInformation.ClientSecret,
		DeviceCode:   &clientInformation.DeviceCode,
		GrantType:    &gtp,
	}
}
