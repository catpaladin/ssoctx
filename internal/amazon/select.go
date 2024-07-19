package amazon

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog"

	"ssoctx/internal/prompt"
)

func fuzzySearchWithPrefixAnchor(itemsToSelect []string, linePrefix string) func(input string, index int) bool {
	return func(input string, index int) bool {
		role := itemsToSelect[index]

		if strings.HasPrefix(input, linePrefix) {
			return strings.HasPrefix(role, input)
		}

		if fuzzy.MatchFold(input, role) {
			return true
		}
		return false
	}
}

// selectRole is used to return a pointer to the selected Role
func selectRole(ctx context.Context, roles *sso.ListAccountRolesOutput, selector prompt.Input) *types.RoleInfo {
	logger := zerolog.Ctx(ctx)
	if len(roles.RoleList) == 1 {
		logger.Info().Msgf("Only one role available. Selected role: %s\n", *roles.RoleList[0].RoleName)
		return &roles.RoleList[0]
	}

	var rolesToSelect []string
	linePrefix := "#"

	for i, info := range roles.RoleList {
		rolesToSelect = append(rolesToSelect, linePrefix+strconv.Itoa(i)+" "+*info.RoleName)
	}

	label := "Select your role"
	indexChoice, _ := selector.Select(label, rolesToSelect, fuzzySearchWithPrefixAnchor(rolesToSelect, linePrefix))
	roleInfo := roles.RoleList[indexChoice]
	return &roleInfo
}

// selectAccount is used to return a pointer to the selected Account
func selectAccount(ctx context.Context, accounts *sso.ListAccountsOutput, selector prompt.Input) *types.AccountInfo {
	logger := zerolog.Ctx(ctx)
	sortedAccounts := sortAccounts(accounts.AccountList)

	var accountsToSelect []string
	linePrefix := "#"

	for i, info := range sortedAccounts {
		accountsToSelect = append(accountsToSelect, linePrefix+strconv.Itoa(i)+" "+*info.AccountName+" "+*info.AccountId)
	}

	label := "Select your account"
	indexChoice, _ := selector.Select(label, accountsToSelect, fuzzySearchWithPrefixAnchor(accountsToSelect, linePrefix))

	fmt.Println()

	accountInfo := sortedAccounts[indexChoice]

	logger.Info().Msgf("Selected account: %s - %s", *accountInfo.AccountName, *accountInfo.AccountId)
	fmt.Println()
	return &accountInfo
}

func sortAccounts(accountList []types.AccountInfo) []types.AccountInfo {
	var sortedAccounts []types.AccountInfo

	sortedAccounts = append(sortedAccounts, accountList...)
	sort.Slice(sortedAccounts, func(i, j int) bool {
		return *sortedAccounts[i].AccountName < *sortedAccounts[j].AccountName
	})
	return sortedAccounts
}
