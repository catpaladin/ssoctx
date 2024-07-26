package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	startURL   string // used to store the sso start url
	region     string // used to store the aws region
	profile    string // used to store the profile name
	keys       bool   // used to determine if using creds with access/secret keys
	roleName   string // used to store permission set name
	accountID  string // used to store the account id selected
	clean      bool   // used to clean lock file
	debug      bool   // used to enable debug logging
	jsonFormat bool   // used to enable json logging
	printCreds bool   // used to print creds

	ctx     = context.Background()
	version = "v0.0.0+unknown"
	commit  = "unknown"

	rootCmd = &cobra.Command{
		Use:   "ssoctx",
		Short: "A tool for setting up an AWS SSO session",
		Long: `A tool for seting up AWS SSO.
Use to login to SSO portal and refresh session.`,
	}
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
