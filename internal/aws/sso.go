package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sso"
)

// SSOClientAPI contains everything needed to make SSO api calls
type SSOClientAPI struct {
	Client *sso.Client
}

// ListAvailableRoles is used to return a ListAccountRolesOutput
func (c SSOClientAPI) ListAvailableRoles(ctx context.Context, accountID, accessToken string) *sso.ListAccountRolesOutput {
	lari := &sso.ListAccountRolesInput{AccountId: &accountID, AccessToken: &accessToken}
	roles, _ := c.Client.ListAccountRoles(ctx, lari)

	return roles
}

// ListAccounts is used to return the ListAccountsOutput
func (c SSOClientAPI) ListAccounts(ctx context.Context, accessToken string) *sso.ListAccountsOutput {
	var maxSize int32 = 500
	lai := sso.ListAccountsInput{AccessToken: &accessToken, MaxResults: &maxSize}
	accounts, _ := c.Client.ListAccounts(ctx, &lai)

	return accounts
}

// GetRolesCredentials is used
func (c SSOClientAPI) GetRolesCredentials(ctx context.Context, accountID, roleName, accessToken string) (*sso.GetRoleCredentialsOutput, error) {
	rci := &sso.GetRoleCredentialsInput{AccountId: &accountID, RoleName: &roleName, AccessToken: &accessToken}
	roleCredentials, err := c.Client.GetRoleCredentials(ctx, rci)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	return roleCredentials, nil
}
