// Package prompt contains functionality for terminal prompt and search
package prompt

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// AwsRegions contains selectable regions
var AwsRegions = []string{
	"us-east-2",
	"us-east-1",
	"us-west-1",
	"us-west-2",
	"af-south-1",
	"ap-east-1",
	"ap-south-1",
	"ap-northeast-3",
	"ap-northeast-2",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-northeast-1",
	"ca-central-1",
	"eu-central-1",
	"eu-west-1",
	"eu-west-2",
	"eu-south-1",
	"eu-west-3",
	"eu-north-1",
	"me-south-1",
	"sa-east-1",
}

func PromptStartURL(prompt Prompt, defaultValue string) string {
	return prompt.Prompt("SSO Start URL", defaultValue)
}

func PromptRegion(prompt Prompt) string {
	_, region := prompt.Select("Select your AWS Region", AwsRegions, func(input string, index int) bool {
		target := AwsRegions[index]
		return fuzzy.MatchFold(input, target)
	})
	return region
}

// RetrieveRoleInfo is used to return a pointer to the selected Role
func RetrieveRoleInfo(roles *sso.ListAccountRolesOutput, selector Prompt) *types.RoleInfo {
	if len(roles.RoleList) == 1 {
		log.Printf("Only one role available. Selected role: %s\n", *roles.RoleList[0].RoleName)
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

// RetrieveAccountInfo is used to return a pointer to the selected Account
func RetrieveAccountInfo(accounts *sso.ListAccountsOutput, selector Prompt) *types.AccountInfo {
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

	log.Printf("Selected account: %s - %s", *accountInfo.AccountName, *accountInfo.AccountId)
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
