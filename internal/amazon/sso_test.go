package amazon

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	smithy "github.com/aws/smithy-go"
)

type mockSSOClient struct {
	GetRoleCredentialsAPI func(ctx context.Context, params *sso.GetRoleCredentialsInput, optFns ...func(*sso.Options)) (*sso.GetRoleCredentialsOutput, error)
	ListAccountsAPI       func(ctx context.Context, params *sso.ListAccountsInput, optFns ...func(*sso.Options)) (*sso.ListAccountsOutput, error)
	ListAccountRolesAPI   func(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error)
}

// Mock GetRoleCredentials outputs
func (m *mockSSOClient) GetRoleCredentials(ctx context.Context, params *sso.GetRoleCredentialsInput, optFns ...func(*sso.Options)) (*sso.GetRoleCredentialsOutput, error) {
	if m.GetRoleCredentialsAPI != nil {
		return m.GetRoleCredentialsAPI(ctx, params, optFns...)
	}
	return nil, nil
}

// Mock ListAccounts outputs
func (m *mockSSOClient) ListAccounts(ctx context.Context, params *sso.ListAccountsInput, optFns ...func(*sso.Options)) (*sso.ListAccountsOutput, error) {
	if m.ListAccountsAPI != nil {
		return m.ListAccountsAPI(ctx, params, optFns...)
	}
	return nil, nil
}

// Mock ListAccountRoles outputs
func (m *mockSSOClient) ListAccountRoles(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error) {
	if m.ListAccountRolesAPI != nil {
		return m.ListAccountRolesAPI(ctx, params, optFns...)
	}
	return nil, nil
}

func TestClient_ListAvailableRoles(t *testing.T) {
	a := "098765432123"
	av := "222233334444"
	r := "UltimatePowerUser"
	rv := "ViewOnlyUser"
	type args struct {
		accountID   string
		accessToken string
	}
	tests := []struct {
		name    string
		client  SSOClient
		args    args
		want    *sso.ListAccountRolesOutput
		wantErr bool
		errResp string
	}{
		{
			name: "ListAvailableRoles success with one result",
			client: &mockSSOClient{
				ListAccountRolesAPI: func(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error) {
					return &sso.ListAccountRolesOutput{
						RoleList: []types.RoleInfo{
							{
								AccountId: &a,
								RoleName:  &r,
							},
						},
					}, nil
				},
			},
			args: args{
				accountID:   a,
				accessToken: "faketoken",
			},
			want: &sso.ListAccountRolesOutput{
				RoleList: []types.RoleInfo{
					{
						AccountId: &a,
						RoleName:  &r,
					},
				},
			},
			wantErr: false,
			errResp: "",
		},
		{
			name: "ListAvailableRoles success with multiple results",
			client: &mockSSOClient{
				ListAccountRolesAPI: func(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error) {
					return &sso.ListAccountRolesOutput{
						RoleList: []types.RoleInfo{
							{
								AccountId: &av,
								RoleName:  &r,
							},
							{
								AccountId: &av,
								RoleName:  &rv,
							},
						},
					}, nil
				},
			},
			args: args{
				accountID:   "222233334444",
				accessToken: "faketoken",
			},
			want: &sso.ListAccountRolesOutput{
				RoleList: []types.RoleInfo{
					{
						AccountId: &av,
						RoleName:  &r,
					},
					{
						AccountId: &av,
						RoleName:  &rv,
					},
				},
			},
			wantErr: false,
			errResp: "",
		},
		{
			name: "ListAvailableRoles error",
			client: &mockSSOClient{
				ListAccountRolesAPI: func(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error) {
					return &sso.ListAccountRolesOutput{}, &smithy.GenericAPIError{
						Code:    "SomeErrorCode",
						Message: "Some error message you should read",
					}
				},
			},
			args: args{
				accountID:   a,
				accessToken: "badtoken",
			},
			want:    &sso.ListAccountRolesOutput{},
			wantErr: true,
			errResp: "Some error message you should read",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSSOClient(tt.client)
			got, err := c.listAvailableRoles(zerologTestingContext, tt.args.accountID, tt.args.accessToken)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SSOClient.ListAvailableRoles() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("SSOClient.ListAvailableRoles() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("SSOClient.ListAvailableRoles() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}

func TestClient_ListAccounts(t *testing.T) {
	a := "123456789012"
	n := "the_best_account"

	tests := []struct {
		name        string
		client      SSOClient
		accessToken string
		want        *sso.ListAccountsOutput
		wantErr     bool
		errResp     string
	}{
		{
			name: "ListAccounts successful",
			client: &mockSSOClient{
				ListAccountsAPI: func(ctx context.Context, params *sso.ListAccountsInput, optFns ...func(*sso.Options)) (*sso.ListAccountsOutput, error) {
					return &sso.ListAccountsOutput{
						AccountList: []types.AccountInfo{
							{
								AccountId:   &a,
								AccountName: &n,
							},
						},
					}, nil
				},
			},
			accessToken: "goodtoken",
			want: &sso.ListAccountsOutput{
				AccountList: []types.AccountInfo{
					{
						AccountId:   &a,
						AccountName: &n,
					},
				},
			},
			wantErr: false,
			errResp: "",
		},
		{
			name: "ListAccounts error",
			client: &mockSSOClient{
				ListAccountsAPI: func(ctx context.Context, params *sso.ListAccountsInput, optFns ...func(*sso.Options)) (*sso.ListAccountsOutput, error) {
					return &sso.ListAccountsOutput{}, &smithy.GenericAPIError{
						Code:    "SomeErrorCode",
						Message: "Some error message you should read",
					}
				},
			},
			accessToken: "badtoken",
			want:        &sso.ListAccountsOutput{},
			wantErr:     true,
			errResp:     "Some error message you should read",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSSOClient(tt.client)
			got, err := c.listAccounts(zerologTestingContext, tt.accessToken)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SSOClient.ListAccounts() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("SSOClient.ListAccounts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("SSOClient.ListAccounts() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}

func TestClient_GetRolesCredentials(t *testing.T) {
	aki := "111111111111111111"
	sac := "sosecretverysecret"
	a := "123456789012"
	n := "ViewOnlyUser"

	output := types.RoleCredentials{
		AccessKeyId:     &aki,
		SecretAccessKey: &sac,
	}

	type args struct {
		accountID   string
		roleName    string
		accessToken string
	}

	tests := []struct {
		name    string
		client  SSOClient
		args    args
		want    *sso.GetRoleCredentialsOutput
		wantErr bool
		errResp string
	}{
		{
			name: "GetRoleCredentials successful",
			client: &mockSSOClient{
				GetRoleCredentialsAPI: func(ctx context.Context, params *sso.GetRoleCredentialsInput, optFns ...func(*sso.Options)) (*sso.GetRoleCredentialsOutput, error) {
					return &sso.GetRoleCredentialsOutput{
						RoleCredentials: &output,
					}, nil
				},
			},
			args: args{
				accountID:   a,
				roleName:    n,
				accessToken: "goodtoken",
			},
			want: &sso.GetRoleCredentialsOutput{
				RoleCredentials: &output,
			},
			wantErr: false,
			errResp: "",
		},
		{
			name: "GetRoleCredentials error",
			client: &mockSSOClient{
				GetRoleCredentialsAPI: func(ctx context.Context, params *sso.GetRoleCredentialsInput, optFns ...func(*sso.Options)) (*sso.GetRoleCredentialsOutput, error) {
					return &sso.GetRoleCredentialsOutput{}, &smithy.GenericAPIError{
						Code:    "UnauthorizedException",
						Message: "401",
					}
				},
			},
			args: args{
				accountID:   "098765432123",
				roleName:    "UltimatePowerUser",
				accessToken: "badtoken",
			},
			want:    &sso.GetRoleCredentialsOutput{},
			wantErr: true,
			errResp: "401",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSSOClient(tt.client)
			got, err := c.getRolesCredentials(zerologTestingContext, tt.args.accountID, tt.args.roleName, tt.args.accessToken)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SSOClient.GetRolesCredentials() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("SSOClient.GetRolesCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				message := err.Error()
				if !strings.Contains(fmt.Sprint(message), tt.errResp) {
					t.Errorf("SSOClient.GetRolesCredentials() error = %v, errResp %v", err, tt.errResp)
				}
			}
		})
	}
}
