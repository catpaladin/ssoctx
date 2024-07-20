package amazon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	smithy "github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

var (
	accessToken      = "1010101010101010101"
	refreshToken     = "eyyyyEARRRRRRMO"
	code             = "ABCE-F1JK"
	mockClientID     = "111111111AAAAAAAAAA"
	mockClientSecret = "SuperSecretDontLook"
	uriComplete      = fmt.Sprintf("https://device.sso.us-west-2.amazonaws.com/?user_code=%s", code)

	zerologTestingContext = log.Logger.WithContext(context.Background())
)

type mockOIDCClient struct {
	CreateTokenAPI              func(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error)
	RegisterClientAPI           func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error)
	StartDeviceAuthorizationAPI func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error)
}

func (m *mockOIDCClient) CreateToken(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
	if m.CreateTokenAPI != nil {
		return m.CreateTokenAPI(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockOIDCClient) RegisterClient(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
	if m.RegisterClientAPI != nil {
		return m.RegisterClientAPI(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockOIDCClient) StartDeviceAuthorization(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
	// set execCmd to call echo instead of open
	execCmd = fakeExecCmd
	if m.StartDeviceAuthorizationAPI != nil {
		return m.StartDeviceAuthorizationAPI(ctx, params, optFns...)
	}
	return nil, nil
}

func fakeExecCmd(command string, args ...string) *exec.Cmd {
	var argsString string
	for _, arg := range args {
		argsString = argsString + arg
	}
	return exec.Command("echo", fmt.Sprintf("test: %s %s", command, argsString))
}

func fakeAccessTokenPath(fileName string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.json", fileName))
}

func removeTempFile(target string) {
	// try to remove file and ignore if it does not exist
	_ = os.Remove(target)
}

func TestOIDCClientAPI_startDeviceAuthorization(t *testing.T) {
	rco := &ssooidc.RegisterClientOutput{}

	validAPI := &mockOIDCClient{
		StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
			return &ssooidc.StartDeviceAuthorizationOutput{
				DeviceCode:              &code,
				VerificationUriComplete: &uriComplete,
			}, nil
		},
	}

	tests := []struct {
		name    string
		client  OIDCClient
		url     string
		rco     *ssooidc.RegisterClientOutput
		want    *ssooidc.StartDeviceAuthorizationOutput
		wantErr bool
		errResp string
	}{
		{
			name:   "StartDeviceAuthorization success",
			client: validAPI,
			url:    "https://fakebanana.awsapp.com/start",
			rco:    rco,
			want: &ssooidc.StartDeviceAuthorizationOutput{
				DeviceCode:              &code,
				VerificationUriComplete: &uriComplete,
			},
			wantErr: false,
			errResp: "",
		},
		{
			name: "StartDeviceAuthorization error",
			client: &mockOIDCClient{
				StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
					return &ssooidc.StartDeviceAuthorizationOutput{}, &smithy.GenericAPIError{
						Code:    "AuthorizationPendingException",
						Message: "This is a fake error test",
					}
				},
			},
			url:     "https://fakebanana.awsapp.com/start",
			rco:     rco,
			want:    &ssooidc.StartDeviceAuthorizationOutput{},
			wantErr: true,
			errResp: "This is a fake error test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOIDCClient(tt.client, tt.url)
			got, err := o.startDeviceAuthorization(zerologTestingContext, tt.rco)
			if !reflect.DeepEqual(got.DeviceCode, tt.want.DeviceCode) {
				t.Errorf("OIDCClientAPI.startDeviceAuthorization() got.DeviceCode = %v, want.DeviceCode %v", got.DeviceCode, tt.want.DeviceCode)
			}
			if !reflect.DeepEqual(got.VerificationUriComplete, tt.want.VerificationUriComplete) {
				t.Errorf("OIDCClientAPI.startDeviceAuthorization() got.VerificationUriComplete = %v, want.VerificationUriComplete %v", got.VerificationUriComplete, tt.want.VerificationUriComplete)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OIDCClientAPI.startDeviceAuthorization() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("OIDCClientAPI.startDeviceAuthorization() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}

func TestOIDCClientAPI_retrieveToken(t *testing.T) {
	tests := []struct {
		name    string
		client  OIDCClient
		url     string
		info    *ClientInformation
		want    *ClientInformation
		wantErr bool
		errResp string
	}{
		{
			name: "CreateToken successful",
			client: &mockOIDCClient{
				CreateTokenAPI: func(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
					return &ssooidc.CreateTokenOutput{
						AccessToken:  &accessToken,
						RefreshToken: &refreshToken,
					}, nil
				},
			},
			url: "https://banana.awsapp.com/start",
			info: &ClientInformation{
				DeviceCode: code,
			},
			want: &ClientInformation{
				AccessToken: accessToken,
			},
			wantErr: false,
			errResp: "",
		},
		{
			name: "CreateToken error timeout",
			client: &mockOIDCClient{
				CreateTokenAPI: func(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
					return &ssooidc.CreateTokenOutput{}, &smithy.GenericAPIError{
						Code:    "AuthorizationPendingException",
						Message: "This is a fake error test",
					}
				},
			},
			url:  "https://banana.awsapp.com/start",
			want: &ClientInformation{},
			info: &ClientInformation{
				DeviceCode: "1234-5678",
			},
			wantErr: true,
			errResp: "Encountered timeout in createToken",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTokenTimeout = 1 * time.Second
			o := NewOIDCClient(tt.client, tt.url)
			input := generateCreateTokenInput(tt.info)

			// override action to pass mock
			var cto *ssooidc.CreateTokenOutput
			var err error
			action = func() {
				cto, err = o.createToken(zerologTestingContext, &input)
			}
			fmt.Print(cto) // for testing
			got, err := o.retrieveToken(zerologTestingContext, tt.info)
			if !reflect.DeepEqual(got.AccessToken, tt.want.AccessToken) {
				t.Errorf("OIDCClientAPI.retrieveToken() got.AccessToken = %v, want.AccessToken %v", got.AccessToken, tt.want.AccessToken)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OIDCClientAPI.retrieveToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("OIDCClientAPI.retrieveToken() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}

func TestOIDCClientAPI_registerClient(t *testing.T) {
	url := "https://newbanana.awsapp.com/start"
	expected := ClientInformation{
		ClientID:                mockClientID,
		ClientSecret:            mockClientSecret,
		DeviceCode:              code,
		VerificationURIComplete: uriComplete,
		StartURL:                url,
	}

	tests := []struct {
		name    string
		client  OIDCClient
		url     string
		want    *ClientInformation
		wantErr bool
		errResp string
	}{
		{
			name: "RegisterClient successful",
			client: &mockOIDCClient{
				RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
					return &ssooidc.RegisterClientOutput{
						ClientId:     &mockClientID,
						ClientSecret: &mockClientSecret,
					}, nil
				},
				StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
					return &ssooidc.StartDeviceAuthorizationOutput{
						DeviceCode:              &code,
						VerificationUriComplete: &uriComplete,
					}, nil
				},
			},
			url:     url,
			want:    &expected,
			wantErr: false,
			errResp: "",
		},
		{
			name: "RegisterClient error",
			client: &mockOIDCClient{
				RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
					return &ssooidc.RegisterClientOutput{}, &smithy.GenericAPIError{
						Code:    "RegisterClientError",
						Message: "This is a fake error test",
					}
				},
			},
			url:     url,
			want:    &ClientInformation{},
			wantErr: true,
			errResp: "This is a fake error test",
		},
		{
			name: "RegisterClient error in StartDevice",
			client: &mockOIDCClient{
				RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
					return &ssooidc.RegisterClientOutput{
						ClientId:     &mockClientID,
						ClientSecret: &mockClientSecret,
					}, nil
				},
				StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
					return &ssooidc.StartDeviceAuthorizationOutput{}, fmt.Errorf("Could not open %s on unsupported platform. Please open the URL manually", uriComplete)
				},
			},
			url:     url,
			want:    &ClientInformation{},
			wantErr: true,
			errResp: fmt.Sprintf("Could not open %s on unsupported platform. Please open the URL manually", uriComplete),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOIDCClient(tt.client, tt.url)
			got, err := o.registerClient(zerologTestingContext)
			if !reflect.DeepEqual(got.ClientID, tt.want.ClientID) {
				t.Errorf("OIDCClientAPI.registerClient() got = %v, want %v", got.ClientID, tt.want.ClientID)
			}
			if !reflect.DeepEqual(got.ClientSecret, tt.want.ClientSecret) {
				t.Errorf("OIDCClientAPI.registerClient() got = %v, want %v", got.ClientSecret, tt.want.ClientSecret)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OIDCClientAPI.registerClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("OIDCClientAPI.registerClient() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}

func TestOIDCClientAPI_getClientInfoPointer(t *testing.T) {
	url := "https://newbanana.awsapp.com/start"
	expected := ClientInformation{
		ClientID:                mockClientID,
		ClientSecret:            mockClientSecret,
		DeviceCode:              code,
		VerificationURIComplete: uriComplete,
		StartURL:                url,
	}
	tests := []struct {
		name    string
		client  OIDCClient
		url     string
		want    *ClientInformation
		wantErr bool
		errResp string
	}{
		{
			name: "GetClientInfoPointer successful",
			client: &mockOIDCClient{
				CreateTokenAPI: func(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
					return &ssooidc.CreateTokenOutput{
						AccessToken:  &accessToken,
						RefreshToken: &refreshToken,
					}, nil
				},
				RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
					return &ssooidc.RegisterClientOutput{
						ClientId:     &mockClientID,
						ClientSecret: &mockClientSecret,
					}, nil
				},
				StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
					return &ssooidc.StartDeviceAuthorizationOutput{
						DeviceCode:              &code,
						VerificationUriComplete: &uriComplete,
					}, nil
				},
			},
			url:     url,
			want:    &expected,
			wantErr: false,
			errResp: "",
		},
		{
			name: "GetClientInfoPointer error in RegisterClient",
			client: &mockOIDCClient{
				RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
					return &ssooidc.RegisterClientOutput{}, &smithy.GenericAPIError{
						Code:    "RegisterClientError",
						Message: "This is a fake error test",
					}
				},
			},
			url:     url,
			want:    &ClientInformation{},
			wantErr: true,
			errResp: "This is a fake error test",
		},
		{
			name: "GetClientInfoPointer error in CreateToken",
			client: &mockOIDCClient{
				CreateTokenAPI: func(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
					return &ssooidc.CreateTokenOutput{}, &smithy.GenericAPIError{
						Code:    "AuthorizationPendingException",
						Message: "This is a fake error test",
					}
				},
				RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
					return &ssooidc.RegisterClientOutput{
						ClientId:     &mockClientID,
						ClientSecret: &mockClientSecret,
					}, nil
				},
				StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
					return &ssooidc.StartDeviceAuthorizationOutput{
						DeviceCode:              &code,
						VerificationUriComplete: &uriComplete,
					}, nil
				},
			},
			url:     url,
			want:    &ClientInformation{},
			wantErr: true,
			errResp: "Encountered timeout in createToken",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTokenTimeout = 1 * time.Second
			o := NewOIDCClient(tt.client, tt.url)
			got, err := o.getClientInfoPointer(zerologTestingContext)
			if got.ClientID != tt.want.ClientID {
				t.Errorf("OIDCClientAPI.getClientInfoPointer() = %v, want %v", got.ClientID, tt.want.ClientID)
			}
			if got.ClientSecret != tt.want.ClientSecret {
				t.Errorf("OIDCClientAPI.getClientInfoPointer() = %v, want %v", got.ClientSecret, tt.want.ClientSecret)
			}
			if got.DeviceCode != tt.want.DeviceCode {
				t.Errorf("OIDCClientAPI.getClientInfoPointer() = %v, want %v", got.DeviceCode, tt.want.DeviceCode)
			}
			if got.VerificationURIComplete != tt.want.VerificationURIComplete {
				t.Errorf("OIDCClientAPI.getClientInfoPointer() = %v, want %v", got.VerificationURIComplete, tt.want.VerificationURIComplete)
			}
			if got.StartURL != tt.want.StartURL {
				t.Errorf("OIDCClientAPI.getClientInfoPointer() = %v, want %v", got.StartURL, tt.want.StartURL)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OIDCClientAPI.getClientInfoPointer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("OIDCClientAPI.getClientInfoPointer() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}

func TestOIDCClientAPI_ProcessClientInformation(t *testing.T) {
	type fields struct {
		client OIDCClient
		url    string
	}

	url := "https://newbanana.awsapp.com/start"

	tests := []struct {
		name       string
		fields     fields
		exists     bool
		clientInfo ClientInformation
		want       ClientInformation
		wantErr    bool
		errResp    string
	}{
		{
			name: "ProcessClientInformation successful with no file",
			fields: fields{
				client: &mockOIDCClient{
					CreateTokenAPI: func(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
						return &ssooidc.CreateTokenOutput{
							AccessToken:  &accessToken,
							RefreshToken: &refreshToken,
						}, nil
					},
					RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
						return &ssooidc.RegisterClientOutput{
							ClientId:     &mockClientID,
							ClientSecret: &mockClientSecret,
						}, nil
					},
					StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
						return &ssooidc.StartDeviceAuthorizationOutput{
							DeviceCode:              &code,
							VerificationUriComplete: &uriComplete,
						}, nil
					},
				},
				url: url,
			},
			exists:     true,
			clientInfo: ClientInformation{},
			want: ClientInformation{
				ClientID:                mockClientID,
				ClientSecret:            mockClientSecret,
				DeviceCode:              code,
				VerificationURIComplete: uriComplete,
				StartURL:                url,
			},
			wantErr: false,
			errResp: "",
		},
		{
			name: "ProcessClientInformation error in StartDevice",
			fields: fields{
				client: &mockOIDCClient{
					CreateTokenAPI: func(ctx context.Context, params *ssooidc.CreateTokenInput, optFns ...func(*ssooidc.Options)) (*ssooidc.CreateTokenOutput, error) {
						return &ssooidc.CreateTokenOutput{
							AccessToken:  &accessToken,
							RefreshToken: &refreshToken,
						}, nil
					},
					RegisterClientAPI: func(ctx context.Context, params *ssooidc.RegisterClientInput, optFns ...func(*ssooidc.Options)) (*ssooidc.RegisterClientOutput, error) {
						return &ssooidc.RegisterClientOutput{
							ClientId:     &mockClientID,
							ClientSecret: &mockClientSecret,
						}, nil
					},
					StartDeviceAuthorizationAPI: func(ctx context.Context, params *ssooidc.StartDeviceAuthorizationInput, optFns ...func(*ssooidc.Options)) (*ssooidc.StartDeviceAuthorizationOutput, error) {
						return &ssooidc.StartDeviceAuthorizationOutput{}, &smithy.GenericAPIError{
							Code:    "GenericErrorCode",
							Message: "This is a mock error",
						}
					},
				},
				url: url,
			},
			exists:     false,
			clientInfo: ClientInformation{},
			want:       ClientInformation{},
			wantErr:    true,
			errResp:    "This is a mock error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OIDCClientAPI{
				client: tt.fields.client,
				url:    tt.fields.url,
			}
			target := fakeAccessTokenPath(tt.name)
			if tt.exists {
				writeStructToFile(zerologTestingContext, tt.clientInfo, target)
			}
			defer removeTempFile(target)

			got, err := o.processClientInformation(zerologTestingContext, target)
			if (err != nil) != tt.wantErr {
				t.Errorf("OIDCClientAPI.processClientInformation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ClientID != tt.want.ClientID {
				t.Errorf("OIDCClientAPI.processClientInformation() = %v, want %v", got.ClientID, tt.want.ClientID)
			}
			if got.ClientSecret != tt.want.ClientSecret {
				t.Errorf("OIDCClientAPI.processClientInformation() = %v, want %v", got.ClientSecret, tt.want.ClientSecret)
			}
			if got.DeviceCode != tt.want.DeviceCode {
				t.Errorf("OIDCClientAPI.processClientInformation() = %v, want %v", got.DeviceCode, tt.want.DeviceCode)
			}
			if got.VerificationURIComplete != tt.want.VerificationURIComplete {
				t.Errorf("OIDCClientAPI.processClientInformation() = %v, want %v", got.VerificationURIComplete, tt.want.VerificationURIComplete)
			}
			if got.StartURL != tt.want.StartURL {
				t.Errorf("OIDCClientAPI.processClientInformation() = %v, want %v", got.StartURL, tt.want.StartURL)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OIDCClientAPI.processClientInformation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("OIDCClientAPI.processClientInformation() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}
