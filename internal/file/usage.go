// Package file contains needed functionality for config and files
package file

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/sso/types"
)

const (
	// LastUsagePath is the pathing to the cached last usage
	LastUsagePath string = "/.aws/sso/cache/last-usage.json"
	// WindowsLastUsagePath is the pathing to cached last usage in Windows
	WindowsLastUsagePath string = "\\.aws\\sso\\cache\\last-usage.json"
)

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

// SaveUsageInformation is used to write usage information to file
func SaveUsageInformation(accountInfo *types.AccountInfo, roleInfo *types.RoleInfo) {
	homeDir, _ := os.UserHomeDir()
	target := fmt.Sprintf("%s%s", homeDir, LastUsageFile())
	usageInformation := LastUsageInformation{
		AccountID:   *accountInfo.AccountId,
		AccountName: *accountInfo.AccountName,
		Role:        *roleInfo.RoleName,
	}
	WriteStructToFile(usageInformation, target)
}

// ReadUsageInformation is used to read usage information frm file
func ReadUsageInformation() (*LastUsageInformation, error) {
	homeDir, _ := os.UserHomeDir()
	target := fmt.Sprintf("%s%s", homeDir, LastUsageFile())
	bytes, err := os.ReadFile(target)
	if err != nil {
		return nil, err
	}
	lui := new(LastUsageInformation)
	err = json.Unmarshal(bytes, lui)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	return lui, nil
}
