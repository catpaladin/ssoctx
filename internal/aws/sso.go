// Package aws contains all the aws logic
package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sso"
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
func (c Client) ListAvailableRoles(ctx context.Context, accountID, accessToken string) *sso.ListAccountRolesOutput {
	lari := &sso.ListAccountRolesInput{AccountId: &accountID, AccessToken: &accessToken}
	roles, _ := c.client.ListAccountRoles(ctx, lari)

	return roles
}

// ListAccounts is used to return the ListAccountsOutput
func (c Client) ListAccounts(ctx context.Context, accessToken string) *sso.ListAccountsOutput {
	var maxSize int32 = 500
	lai := sso.ListAccountsInput{AccessToken: &accessToken, MaxResults: &maxSize}
	accounts, _ := c.client.ListAccounts(ctx, &lai)

	return accounts
}

// GetRolesCredentials is used
func (c Client) GetRolesCredentials(ctx context.Context, accountID, roleName, accessToken string) (*sso.GetRoleCredentialsOutput, error) {
	rci := &sso.GetRoleCredentialsInput{AccountId: &accountID, RoleName: &roleName, AccessToken: &accessToken}
	roleCredentials, err := c.client.GetRoleCredentials(ctx, rci)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	return roleCredentials, nil
}
