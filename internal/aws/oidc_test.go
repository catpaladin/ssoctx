// Package aws contains all the aws logic
package aws

import (
	"aws-sso-util/internal/info"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

var (
	accessToken      string = "1010101010101010101"
	refreshToken     string = "eyyyyEARRRRRRMO"
	code             string = "ABCE-F1JK"
	mockClientID     string = "111111111AAAAAAAAAA"
	mockClientSecret string = "SuperSecretDontLook"
	uriComplete      string = fmt.Sprintf("https://device.sso.us-west-2.amazonaws.com/?user_code=%s", code)
)

type mockOIDCClient struct{}

func newMockOIDCClient() *mockOIDCClient {
	return &mockOIDCClient{}
}

func (m *mockOIDCClient) CreateToken(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
	_, ctxCheck := ctx.Deadline()
	if ctxCheck != false {
		return &ssooidc.CreateTokenOutput{}, errors.New("Context deadline set and true")
	}

	// this is only here to make the linter happy
	if len(optFns) > 0 {
		fmt.Println(optFns)
	}

	deviceCode := params.DeviceCode
	if *deviceCode != code {
		return &ssooidc.CreateTokenOutput{}, errors.New("Not correct code")
	}

	return &ssooidc.CreateTokenOutput{
		AccessToken:  &accessToken,
		RefreshToken: &refreshToken,
	}, nil
}

func (m *mockOIDCClient) RegisterClient(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
	_, ctxCheck := ctx.Deadline()
	if ctxCheck != false {
		return &ssooidc.RegisterClientOutput{}, errors.New("Context deadline set and true")
	}

	// this is only here to make the linter happy
	if len(optFns) > 0 {
		fmt.Println(optFns)
	}
	return &ssooidc.RegisterClientOutput{
		ClientId:     &mockClientID,
		ClientSecret: &mockClientSecret,
	}, nil
}

func (m *mockOIDCClient) StartDeviceAuthorization(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
	_, ctxCheck := ctx.Deadline()
	if ctxCheck != false {
		return &ssooidc.StartDeviceAuthorizationOutput{}, errors.New("Context deadline set and true")
	}

	// this is only here to make the linter happy
	if len(optFns) > 0 {
		fmt.Println(optFns)
	}

	url := params.StartUrl
	if *url == "" {
		return &ssooidc.StartDeviceAuthorizationOutput{}, errors.New("You need a start url")
	}
	// set execCmd to call echo instead of open
	execCmd = fakeExecCmd
	return &ssooidc.StartDeviceAuthorizationOutput{
		DeviceCode:              &code,
		VerificationUriComplete: &uriComplete,
	}, nil
}

func fakeExecCmd(command string, args ...string) *exec.Cmd {
	return exec.Command("echo", "test")
}

func TestOIDCClientAPI_startDeviceAuthorization(t *testing.T) {
	type args struct {
		ctx context.Context
		rco *ssooidc.RegisterClientOutput
	}
	tests := []struct {
		name    string
		client  OIDCClient
		system  string
		url     string
		args    args
		want    ssooidc.StartDeviceAuthorizationOutput
		wantErr bool
	}{
		{
			name:   "TestStartDeviceAuthorizationLinux",
			client: newMockOIDCClient(),
			system: "linux",
			url:    "https://fakebanana.awsapp.com/start",
			args: args{
				ctx: context.Background(),
				rco: &ssooidc.RegisterClientOutput{},
			},
			want: ssooidc.StartDeviceAuthorizationOutput{
				DeviceCode:              &code,
				VerificationUriComplete: &uriComplete,
			},
			wantErr: false,
		},
		{
			name:   "TestStartDeviceAuthorizationDarwin",
			client: newMockOIDCClient(),
			system: "darwin",
			url:    "https://fakebanana.awsapp.com/start",
			args: args{
				ctx: context.Background(),
				rco: &ssooidc.RegisterClientOutput{},
			},
			want: ssooidc.StartDeviceAuthorizationOutput{
				DeviceCode:              &code,
				VerificationUriComplete: &uriComplete,
			},
			wantErr: false,
		},
		{
			name:   "TestStartDeviceAuthorizationWindows",
			client: newMockOIDCClient(),
			system: "windows",
			url:    "https://fakebanana.awsapp.com/start",
			args: args{
				ctx: context.Background(),
				rco: &ssooidc.RegisterClientOutput{},
			},
			want: ssooidc.StartDeviceAuthorizationOutput{
				DeviceCode:              &code,
				VerificationUriComplete: &uriComplete,
			},
			wantErr: false,
		},
		{
			name:   "TestStartDeviceAuthorizationUnknown",
			client: newMockOIDCClient(),
			system: "potato",
			url:    "",
			args: args{
				ctx: context.Background(),
				rco: &ssooidc.RegisterClientOutput{},
			},
			want:    ssooidc.StartDeviceAuthorizationOutput{},
			wantErr: true,
		},
		{
			name:   "TestStartDeviceAuthorizationError",
			client: newMockOIDCClient(),
			system: "linux",
			url:    "",
			args: args{
				ctx: context.Background(),
				rco: &ssooidc.RegisterClientOutput{},
			},
			want:    ssooidc.StartDeviceAuthorizationOutput{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOIDCClient(tt.client, tt.url)
			got, err := o.startDeviceAuthorization(tt.args.ctx, tt.args.rco, tt.system)
			if (err != nil) != tt.wantErr {
				t.Errorf("OIDCClientAPI.startDeviceAuthorization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OIDCClientAPI.startDeviceAuthorization() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOIDCClientAPI_retrieveToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		info *info.ClientInformation
	}
	tests := []struct {
		name   string
		client OIDCClient
		url    string
		args   args
		want   *info.ClientInformation
	}{
		{
			name:   "TestCreateToken",
			client: newMockOIDCClient(),
			url:    "https://banana.awsapp.com/start",
			args: args{
				ctx: context.Background(),
				info: &info.ClientInformation{
					DeviceCode: code,
				},
			},
			want: &info.ClientInformation{
				AccessToken: accessToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOIDCClient(tt.client, tt.url)
			got := o.retrieveToken(tt.args.ctx, tt.args.info)
			if got.AccessToken != accessToken {
				t.Errorf("OIDCClientAPI.retrieveToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOIDCClientAPI_registerClient(t *testing.T) {
	url := "https://newbanana.awsapp.com/start"
	expected := info.ClientInformation{
		ClientID:                mockClientID,
		ClientSecret:            mockClientSecret,
		DeviceCode:              code,
		VerificationURIComplete: uriComplete,
		StartURL:                url,
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		client OIDCClient
		url    string
		args   args
		want   *info.ClientInformation
	}{
		{
			name:   "TestRegisterClient",
			client: newMockOIDCClient(),
			url:    url,
			args: args{
				ctx: context.Background(),
			},
			want: &expected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOIDCClient(tt.client, tt.url)
			got := o.registerClient(tt.args.ctx)
			if got.ClientID != mockClientID {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.ClientSecret != mockClientSecret {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.DeviceCode != code {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.VerificationURIComplete != uriComplete {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.StartURL != url {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOIDCClientAPI_getClientInfoPointer(t *testing.T) {
	url := "https://newbanana.awsapp.com/start"
	expected := info.ClientInformation{
		ClientID:                mockClientID,
		ClientSecret:            mockClientSecret,
		DeviceCode:              code,
		VerificationURIComplete: uriComplete,
		StartURL:                url,
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		client OIDCClient
		url    string
		args   args
		want   *info.ClientInformation
	}{
		// TODO: Add test cases.
		{
			name:   "TestGetClientInfoPointer",
			client: newMockOIDCClient(),
			url:    url,
			args: args{
				ctx: context.Background(),
			},
			want: &expected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOIDCClient(tt.client, tt.url)
			got := o.getClientInfoPointer(tt.args.ctx)
			if got.ClientID != mockClientID {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.ClientSecret != mockClientSecret {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.DeviceCode != code {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.VerificationURIComplete != uriComplete {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
			if got.StartURL != url {
				t.Errorf("OIDCClientAPI.registerClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
