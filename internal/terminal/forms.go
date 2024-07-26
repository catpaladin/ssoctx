package terminal

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/charmbracelet/huh"
)

// SelectorFunc is a generic function type for selection
type SelectorFunc[T comparable] func([]huh.Option[T], string) (T, error)

// SelectRegion allows selecting a region by passing in a SelectorFunc.
// Pass in a NewSelectForm[string]
func SelectRegion(selector SelectorFunc[string]) string {
	label := "Select your region"
	options := GenerateGenericOptions(awsRegions)
	selectedRegion, err := selector(options, label)
	if err != nil {
		return ""
	}
	return selectedRegion
}

// SelectRole is used to return a pointer to the selected Role
// Pass in a NewSelectForm[types.RoleInfo]
func SelectRole(roles *sso.ListAccountRolesOutput, selector SelectorFunc[types.RoleInfo]) (*types.RoleInfo, error) {
	if len(roles.RoleList) == 1 {
		return &roles.RoleList[0], nil
	}

	label := "Select your role"
	options := generateRoleInfoOptions(roles.RoleList)
	selectedRole, err := selector(options, label)
	if err != nil {
		return &types.RoleInfo{}, err
	}
	return &selectedRole, nil
}

// SelectAccount is used to return a pointer to the selected Account
// Pass in a NewSelectForm[types.AccountInfo]
func SelectAccount(accounts *sso.ListAccountsOutput, selector SelectorFunc[types.AccountInfo]) (*types.AccountInfo, error) {
	label := "Select your account"
	sortedAccounts := sortAccounts(accounts.AccountList)
	options := generateAccountInfoOptions(sortedAccounts)
	selectedAccount, err := selector(options, label)
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
