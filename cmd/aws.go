package cmd

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

var (
	// AwsRegions contains selectable regions
	AwsRegions = []string{
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
)

// InitClients is used to return sso and ssooidc clients
func InitClients(region string) (*ssooidc.Client, *sso.Client) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		log.Panicf("Encountered error in InitClients: %v", err)
	}

	oidcClient := ssooidc.NewFromConfig(cfg)
	ssoClient := sso.NewFromConfig(cfg)

	return oidcClient, ssoClient
}

// RetrieveRoleInfo is used to return a pointer to the selected Role
func RetrieveRoleInfo(client *sso.Client, accountInfo *types.AccountInfo, clientInformation ClientInformation, selector Prompt) *types.RoleInfo {
	lari := &sso.ListAccountRolesInput{AccountId: accountInfo.AccountId, AccessToken: &clientInformation.AccessToken}
	roles, _ := client.ListAccountRoles(ctx, lari)

	if len(roles.RoleList) == 1 {
		log.Printf("Only one role available. Selected role: %s\n", *roles.RoleList[0].RoleName)
		return &roles.RoleList[0]
	}

	var rolesToSelect []string
	linePrefix := "#"

	for i, info := range roles.RoleList {
		rolesToSelect = append(rolesToSelect, linePrefix+strconv.Itoa(i)+" "+*info.RoleName)
	}

	label := "Select your role - Note: fuzzy search supported."
	indexChoice, _ := selector.Select(label, rolesToSelect, fuzzySearchWithPrefixAnchor(rolesToSelect, linePrefix))
	roleInfo := roles.RoleList[indexChoice]
	return &roleInfo
}

// RetrieveAccountInfo is used to return a pointer to the selected Account
func RetrieveAccountInfo(client *sso.Client, clientInformation ClientInformation, selector Prompt) *types.AccountInfo {
	var maxSize int32 = 500
	lai := sso.ListAccountsInput{AccessToken: &clientInformation.AccessToken, MaxResults: &maxSize}
	accounts, _ := client.ListAccounts(ctx, &lai)

	sortedAccounts := sortAccounts(accounts.AccountList)

	var accountsToSelect []string
	linePrefix := "#"

	for i, info := range sortedAccounts {
		accountsToSelect = append(accountsToSelect, linePrefix+strconv.Itoa(i)+" "+*info.AccountName+" "+*info.AccountId)
	}

	label := "Select your account - Note: fuzzy search supported."
	indexChoice, _ := selector.Select(label, accountsToSelect, fuzzySearchWithPrefixAnchor(accountsToSelect, linePrefix))

	fmt.Println()

	accountInfo := sortedAccounts[indexChoice]

	log.Printf("Selected account: %s - %s", *accountInfo.AccountName, *accountInfo.AccountId)
	fmt.Println()
	return &accountInfo
}

func sortAccounts(accountList []types.AccountInfo) []types.AccountInfo {
	var sortedAccounts []types.AccountInfo

	for _, info := range accountList {
		sortedAccounts = append(sortedAccounts, info)
	}
	sort.Slice(sortedAccounts, func(i, j int) bool {
		return *sortedAccounts[i].AccountName < *sortedAccounts[j].AccountName
	})
	return sortedAccounts
}
