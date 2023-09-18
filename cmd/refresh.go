package cmd

import (
	"aws-sso-util/internal/aws"
	"aws-sso-util/internal/file"
	"aws-sso-util/internal/prompt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/spf13/cobra"
)

//const (
//	// LastUsagePath is the pathing to the cached last usage
//	LastUsagePath string = "/.aws/sso/cache/last-usage.json"
//	// WindowsLastUsagePath is the pathing to cached last usage in Windows
//	WindowsLastUsagePath string = "\\.aws\\sso\\cache\\last-usage.json"
//)

var (
	refreshCmd = &cobra.Command{
		Use:   "refresh",
		Short: "Refresh your previously used credentials",
		Long:  `Refreshes the credentials based on last account and role.`,
		Run: func(cmd *cobra.Command, args []string) {
			conf := file.ReadConfig(file.ConfigFilePath())
			startURL = conf.StartURL
			region = conf.Region
			oidcClient, ssoClient := aws.CreateClients(ctx, region)
			RefreshCredentials(oidcClient, ssoClient)
		},
	}
)

func init() {
	rootCmd.AddCommand(refreshCmd)
}

// LastUsageInformation contains the info on last usage
//type LastUsageInformation struct {
//	AccountID   string `json:"account_id"`
//	AccountName string `json:"account_name"`
//	Role        string `json:"role"`
//}
//
//// LastUsageFile returns the path to the last usage file
//func LastUsageFile() string {
//	homeDir, _ := os.UserHomeDir()
//	if runtime.GOOS == "windows" {
//		return fmt.Sprintf("%s%s", homeDir, WindowsLastUsagePath)
//	}
//	return fmt.Sprintf("%s%s", homeDir, LastUsagePath)
//}

// RefreshCredentials is used to refresh credentials
func RefreshCredentials(oidcClient *ssooidc.Client, ssoClient *sso.Client) {
	oidc := aws.OIDCClientAPI{
		Client: oidcClient,
		URL:    startURL,
	}
	sso := aws.SSOClientAPI{
		Client: ssoClient,
	}

	clientInformation, err := file.ReadClientInformation(file.ClientInfoFileDestination())
	if err != nil || clientInformation.StartURL != startURL {
		clientInformation, _ = oidc.ProcessClientInformation(ctx)
	}
	log.Printf("Using Start URL %s", clientInformation.StartURL)

	var accountID *string
	var roleName *string

	lui, err := file.ReadUsageInformation()
	log.Printf("Attempting to refresh credentials for account [%s] with role [%s]", lui.AccountName, lui.Role)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			log.Println("Nothing to refresh yet.")
			accountsOutput := sso.ListAccounts(ctx, clientInformation.AccessToken)
			accountInfo := prompt.RetrieveAccountInfo(accountsOutput, prompt.Prompter{})
			listRolesOutput := sso.ListAvailableRoles(ctx, *accountInfo.AccountId, clientInformation.AccessToken)
			roleInfo := prompt.RetrieveRoleInfo(listRolesOutput, prompt.Prompter{})
			roleName = roleInfo.RoleName
			accountID = accountInfo.AccountId
			file.SaveUsageInformation(accountInfo, roleInfo)
		}
	} else {
		accountID = &lui.AccountID
		roleName = &lui.Role
	}

	roleCredentials, err := sso.GetRolesCredentials(ctx, *accountID, *roleName, clientInformation.AccessToken)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	template := file.ProcessPersistedCredentialsTemplate(roleCredentials, profile)
	file.WriteAWSCredentialsFile(template)

	log.Printf("Successful retrieved credentials for account: %s", *accountID)
	log.Printf("Assumed role: %s", *roleName)
	log.Printf("Credentials expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
}

// SaveUsageInformation is used to write usage information to file
//func SaveUsageInformation(accountInfo *types.AccountInfo, roleInfo *types.RoleInfo) {
//	homeDir, _ := os.UserHomeDir()
//	target := homeDir + "/.aws/sso/cache/last-usage.json"
//	usageInformation := LastUsageInformation{
//		AccountID:   *accountInfo.AccountId,
//		AccountName: *accountInfo.AccountName,
//		Role:        *roleInfo.RoleName,
//	}
//	WriteStructToFile(usageInformation, target)
//}
//
//func readUsageInformation() (*LastUsageInformation, error) {
//	homeDir, _ := os.UserHomeDir()
//	bytes, err := os.ReadFile(homeDir + "/.aws/sso/cache/last-usage.json")
//	if err != nil {
//		return nil, err
//	}
//	lui := new(LastUsageInformation)
//	err = json.Unmarshal(bytes, lui)
//	check(err)
//	return lui, nil
//}
