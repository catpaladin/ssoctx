// Package aws contains all the aws logic
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/rs/zerolog"
)

// SSOClient is used to abstract the client calls for mock testing
type SSOClient interface {
	GetRoleCredentials(ctx context.Context, params *sso.GetRoleCredentialsInput, optFns ...func(*sso.Options)) (*sso.GetRoleCredentialsOutput, error)
	ListAccountRoles(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error)
	ListAccounts(ctx context.Context, params *sso.ListAccountsInput, optFns ...func(*sso.Options)) (*sso.ListAccountsOutput, error)
}

// Client contains everything needed to make SSO api calls
type Client struct {
	client SSOClient
}

// NewSSOClient implements the interface
func NewSSOClient(s SSOClient) *Client {
	return &Client{client: s}
}

// ListAvailableRoles is used to return a ListAccountRolesOutput
func (c *Client) ListAvailableRoles(ctx context.Context, accountID, accessToken string) *sso.ListAccountRolesOutput {
	logger := zerolog.Ctx(ctx)
	lari := &sso.ListAccountRolesInput{AccountId: &accountID, AccessToken: &accessToken}
	roles, err := c.client.ListAccountRoles(ctx, lari)
	if err != nil {
		// pass through for debug
		_ = GetAWSErrorCode(ctx, err)
		logger.Error().Err(err)
	}
	logger.Debug().Msgf("ListAccountRoles returned %d available roles", len(roles.RoleList))

	return roles
}

// ListAccounts is used to return the ListAccountsOutput
func (c *Client) ListAccounts(ctx context.Context, accessToken string) *sso.ListAccountsOutput {
	logger := zerolog.Ctx(ctx)
	var maxSize int32 = 500
	lai := sso.ListAccountsInput{AccessToken: &accessToken, MaxResults: &maxSize}
	accounts, err := c.client.ListAccounts(ctx, &lai)
	if err != nil {
		// pass through for debug
		_ = GetAWSErrorCode(ctx, err)
		logger.Error().Err(err)
	}

	return accounts
}

// GetRolesCredentials is used
func (c *Client) GetRolesCredentials(ctx context.Context, accountID, roleName, accessToken string) (*sso.GetRoleCredentialsOutput, error) {
	rci := &sso.GetRoleCredentialsInput{AccountId: &accountID, RoleName: &roleName, AccessToken: &accessToken}
	roleCredentials, err := c.client.GetRoleCredentials(ctx, rci)
	if err != nil {
		// pass through for debug
		_ = GetAWSErrorCode(ctx, err)
		return &sso.GetRoleCredentialsOutput{}, err
	}
	return roleCredentials, nil
}
