// Package aws contains all the aws logic
package aws

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type mockSSOClient struct{}

func newMockSSOClient() *mockSSOClient {
	return &mockSSOClient{}
}

// Mock GetRoleCredentials outputs
func (m *mockSSOClient) GetRoleCredentials(ctx context.Context, params *sso.GetRoleCredentialsInput, optFns ...func(*sso.Options)) (*sso.GetRoleCredentialsOutput, error) {
	_, ctxCheck := ctx.Deadline()
	if ctxCheck != false {
		return &sso.GetRoleCredentialsOutput{}, errors.New("Context deadline set and true")
	}

	// this is only here to make the linter happy
	if len(optFns) > 0 {
		fmt.Println(optFns)
	}

	accountID := params.AccountId
	roleName := params.RoleName
	token := params.AccessToken

	aki := "111111111111111111"
	sac := "sosecretverysecret"
	output := types.RoleCredentials{
		AccessKeyId:     &aki,
		SecretAccessKey: &sac,
	}

	if *token == "badtoken" {
		return &sso.GetRoleCredentialsOutput{}, &smithy.GenericAPIError{
			Code:    "UnauthorizedException",
			Message: "401",
		}
	}

	if *accountID == "123456789012" && *roleName == "ViewOnlyUser" {
		return &sso.GetRoleCredentialsOutput{
			RoleCredentials: &output,
		}, nil
	}
	return &sso.GetRoleCredentialsOutput{}, nil
}

// Mock ListAccounts outputs
func (m *mockSSOClient) ListAccounts(ctx context.Context, params *sso.ListAccountsInput, optFns ...func(*sso.Options)) (*sso.ListAccountsOutput, error) {
	_, ctxCheck := ctx.Deadline()
	if ctxCheck != false {
		return &sso.ListAccountsOutput{}, errors.New("Context deadline set and true")
	}

	// this is only here to make the linter happy
	if len(optFns) > 0 {
		fmt.Println(optFns)
	}

	token := params.AccessToken
	if *token == "goodtoken" {
		a := "123456789012"
		n := "the_best_account"
		return &sso.ListAccountsOutput{
			AccountList: []types.AccountInfo{
				{
					AccountId:   &a,
					AccountName: &n,
				},
			},
		}, nil
	}
	return &sso.ListAccountsOutput{}, &smithy.GenericAPIError{
		Code:    "SomeErrorCode",
		Message: "Some error message you should read",
	}
}

// Mock ListAccountRoles outputs
func (m *mockSSOClient) ListAccountRoles(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error) {
	_, ctxCheck := ctx.Deadline()
	if ctxCheck != false {
		return &sso.ListAccountRolesOutput{}, errors.New("Context deadline set and true")
	}

	// this is only here to make the linter happy
	if len(optFns) > 0 {
		fmt.Println(optFns)
	}

	if *params.AccessToken == "badtoken" {
		return &sso.ListAccountRolesOutput{}, &smithy.GenericAPIError{
			Code:    "SomeErrorCode",
			Message: "Some error message you should read",
		}
	}

	accountID := params.AccountId
	r := "UltimatePowerUser"
	rv := "ViewOnlyUser"

	switch *accountID {
	case "098765432123":
		return &sso.ListAccountRolesOutput{
			RoleList: []types.RoleInfo{
				{
					AccountId: accountID,
					RoleName:  &r,
				},
			},
		}, nil
	case "222233334444":
		return &sso.ListAccountRolesOutput{
			RoleList: []types.RoleInfo{
				{
					AccountId: accountID,
					RoleName:  &r,
				},
				{
					AccountId: accountID,
					RoleName:  &rv,
				},
			},
		}, nil
	default:
		return &sso.ListAccountRolesOutput{}, nil
	}
}

func TestClient_ListAvailableRoles(t *testing.T) {
	a := "098765432123"
	av := "222233334444"
	r := "UltimatePowerUser"
	rv := "ViewOnlyUser"
	type args struct {
		ctx         context.Context
		accountID   string
		accessToken string
	}
	tests := []struct {
		name    string
		client  SSOClient
		args    args
		want    *sso.ListAccountRolesOutput
		wantErr bool
	}{
		{
			name:   "TestListAvailableRoles",
			client: newMockSSOClient(),
			args: args{
				ctx:         log.Logger.WithContext(context.Background()),
				accountID:   "098765432123",
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
		},
		{
			name:   "TestListAvailableRolesMultiple",
			client: newMockSSOClient(),
			args: args{
				ctx:         log.Logger.WithContext(context.Background()),
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
		},
		{
			name:   "TestListAvailableRolesEmpty",
			client: newMockSSOClient(),
			args: args{
				ctx:         log.Logger.WithContext(context.Background()),
				accountID:   "123456789012",
				accessToken: "faketoken",
			},
			want:    &sso.ListAccountRolesOutput{},
			wantErr: false,
		},
		{
			name:   "TestListAvailableRolesError",
			client: newMockSSOClient(),
			args: args{
				ctx:         log.Logger.WithContext(context.Background()),
				accountID:   "123456789012",
				accessToken: "badtoken",
			},
			want:    &sso.ListAccountRolesOutput{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSSOClient(tt.client)
			if got := c.ListAvailableRoles(tt.args.ctx, tt.args.accountID, tt.args.accessToken); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListAvailableRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListAccounts(t *testing.T) {
	a := "123456789012"
	n := "the_best_account"
	type args struct {
		ctx         context.Context
		accessToken string
	}
	tests := []struct {
		name    string
		client  SSOClient
		args    args
		want    *sso.ListAccountsOutput
		wantErr bool
	}{
		{
			name:   "TestListAccounts",
			client: newMockSSOClient(),
			args: args{
				ctx:         log.Logger.WithContext(context.Background()),
				accessToken: "goodtoken",
			},
			want: &sso.ListAccountsOutput{
				AccountList: []types.AccountInfo{
					{
						AccountId:   &a,
						AccountName: &n,
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "TestListAccountsError",
			client: newMockSSOClient(),
			args: args{
				ctx:         log.Logger.WithContext(context.Background()),
				accessToken: "badtoken",
			},
			want:    &sso.ListAccountsOutput{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSSOClient(tt.client)
			if got := c.ListAccounts(tt.args.ctx, tt.args.accessToken); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListAccounts() = %v, want %v", got, tt.want)
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
		ctx         context.Context
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
	}{
		{
			name:   "TestGetRoleCredentials",
			client: newMockSSOClient(),
			args: args{
				ctx:         context.Background(),
				accountID:   a,
				roleName:    n,
				accessToken: "goodtoken",
			},
			want: &sso.GetRoleCredentialsOutput{
				RoleCredentials: &output,
			},
			wantErr: false,
		},
		{
			name:   "TestGetRoleCredentialsError",
			client: newMockSSOClient(),
			args: args{
				ctx:         log.Logger.WithContext(context.Background()),
				accountID:   "098765432123",
				roleName:    "UltimatePowerUser",
				accessToken: "badtoken",
			},
			want:    &sso.GetRoleCredentialsOutput{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSSOClient(tt.client)
			got, err := c.GetRolesCredentials(tt.args.ctx, tt.args.accountID, tt.args.roleName, tt.args.accessToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetRolesCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetRolesCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
