/*
Package cmd contains all the cli commands
*/
package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	startURL  string
	region    string
	profile   string
	persist   bool
	roleName  string
	accountID string
	clean     bool

	ctx = context.Background()

	rootCmd = &cobra.Command{
		Use:   "aws-sso-util",
		Short: "A tool for setting up an AWS SSO session",
		Long: `A tool for seting up AWS SSO.
		Use to login to SSO portal and refresh session.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// subcommands
}
