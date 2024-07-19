package terminal

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
)

func SelectRegion() string {
	label := "Select your region"
	options := GenerateGenericOptions(awsRegions)
	selectedRegion, err := NewSelectForm(options, label)
	if err != nil {
		return ""
	}
	return selectedRegion
}

// SelectRole is used to return a pointer to the selected Role
func SelectRole(roles *sso.ListAccountRolesOutput) (*types.RoleInfo, error) {
	if len(roles.RoleList) == 1 {
		return &roles.RoleList[0], nil
	}

	label := "Select your role"
	options := generateRoleInfoOptions(roles.RoleList)
	selectedRole, err := NewSelectForm(options, label)
	if err != nil {
		return &types.RoleInfo{}, err
	}
	return &selectedRole, nil
}

// SelectAccount is used to return a pointer to the selected Account
func SelectAccount(accounts *sso.ListAccountsOutput) (*types.AccountInfo, error) {
	label := "Select your account"
	sortedAccounts := sortAccounts(accounts.AccountList)
	options := generateAccountInfoOptions(sortedAccounts)
	selectedAccount, err := NewSelectForm(options, label)
	if err != nil {
		return &types.AccountInfo{}, err
	}

	return &selectedAccount, nil
}

func sortAccounts(accountList []types.AccountInfo) []types.AccountInfo {
	var sortedAccounts []types.AccountInfo

	sortedAccounts = append(sortedAccounts, accountList...)
	sort.Slice(sortedAccounts, func(i, j int) bool {
		return *sortedAccounts[i].AccountName < *sortedAccounts[j].AccountName
	})
	return sortedAccounts
}
