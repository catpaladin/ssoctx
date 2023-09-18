package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/spf13/cobra"
)

const (
	// LastUsagePath is the pathing to the cached last usage
	LastUsagePath string = "/.aws/sso/cache/last-usage.json"
	// WindowsLastUsagePath is the pathing to cached last usage in Windows
	WindowsLastUsagePath string = "\\.aws\\sso\\cache\\last-usage.json"
)

var (
	refreshCmd = &cobra.Command{
		Use:   "refresh",
		Short: "Refresh your previously used credentials",
		Long:  `Refreshes the credentials based on last account and role.`,
		Run: func(cmd *cobra.Command, args []string) {
			conf := ReadConfig(ConfigFilePath())
			startURL = conf.StartURL
			region = conf.Region
			oidcClient, ssoClient := InitClients(region)
			RefreshCredentials(oidcClient, ssoClient)
		},
	}
)

func init() {
	rootCmd.AddCommand(refreshCmd)
}

// LastUsageInformation contains the info on last usage
type LastUsageInformation struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	Role        string `json:"role"`
}

// LastUsageFile returns the path to the last usage file
func LastUsageFile() string {
	homeDir, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s%s", homeDir, WindowsLastUsagePath)
	}
	return fmt.Sprintf("%s%s", homeDir, LastUsagePath)
}

// RefreshCredentials is used to refresh credentials
func RefreshCredentials(oidcClient *ssooidc.Client, ssoClient *sso.Client) {
	oidcInformation := OIDCInformation{
		Client: oidcClient,
		URL:    startURL,
	}
	clientInformation, err := ReadClientInformation(ClientInfoFileDestination())
	if err != nil || clientInformation.StartURL != startURL {
		clientInformation, _ = oidcInformation.ProcessClientInformation()
	}

	log.Printf("Using Start URL %s", clientInformation.StartURL)

	var accountID *string
	var roleName *string

	lui, err := readUsageInformation()
	log.Printf("Attempting to refresh credentials for account [%s] with role [%s]", lui.AccountName, lui.Role)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			log.Println("Nothing to refresh yet.")
			accountInfo := RetrieveAccountInfo(ssoClient, clientInformation, Prompter{})
			roleInfo := RetrieveRoleInfo(ssoClient, accountInfo, clientInformation, Prompter{})
			roleName = roleInfo.RoleName
			accountID = accountInfo.AccountId
			SaveUsageInformation(accountInfo, roleInfo)
		}
	} else {
		accountID = &lui.AccountID
		roleName = &lui.Role
	}

	rci := &sso.GetRoleCredentialsInput{AccountId: accountID, RoleName: roleName, AccessToken: &clientInformation.AccessToken}
	roleCredentials, err := ssoClient.GetRoleCredentials(ctx, rci)
	check(err)

	template := ProcessPersistedCredentialsTemplate(roleCredentials, profile)
	WriteAWSCredentialsFile(template)

	log.Printf("Successful retrieved credentials for account: %s", *accountID)
	log.Printf("Assumed role: %s", *roleName)
	log.Printf("Credentials expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
}

// SaveUsageInformation is used to write usage information to file
func SaveUsageInformation(accountInfo *types.AccountInfo, roleInfo *types.RoleInfo) {
	homeDir, _ := os.UserHomeDir()
	target := homeDir + "/.aws/sso/cache/last-usage.json"
	usageInformation := LastUsageInformation{
		AccountID:   *accountInfo.AccountId,
		AccountName: *accountInfo.AccountName,
		Role:        *roleInfo.RoleName,
	}
	WriteStructToFile(usageInformation, target)
}

func readUsageInformation() (*LastUsageInformation, error) {
	homeDir, _ := os.UserHomeDir()
	bytes, err := os.ReadFile(homeDir + "/.aws/sso/cache/last-usage.json")
	if err != nil {
		return nil, err
	}
	lui := new(LastUsageInformation)
	err = json.Unmarshal(bytes, lui)
	check(err)
	return lui, nil
}
