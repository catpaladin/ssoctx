/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sso"
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

	ctx = context.Background()

	rootCmd = &cobra.Command{
		Use:   "aws-sso-util",
		Short: "A tool for setting up an AWS SSO session",
		Long: `A tool for seting up AWS SSO.
		Use to login to SSO portal and refresh session.`,
	}

	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Handles configuration",
		Long: `Handles configuration. Config location defaults to
		${HOME}/.config/aws-sso-util/config.yaml`,
	}

	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a config file",
		Long: `Generate a config file. All available properities are interactively prompted.
		Overrides the existing config.`,
		Run: func(cmd *cobra.Command, args []string) {
			GenerateConfigAction()
		},
	}

	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit the config file",
		Long: `Edit the config file. All available properities are interactively prompted.
		Overrides the existing config.`,
		Run: func(cmd *cobra.Command, args []string) {
			EditConfigAction()
		},
	}

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

	assumeCmd = &cobra.Command{
		Use:   "assume",
		Short: "Assume directly into an account and SSO role",
		Long: `Assume directly into an account and SSO role.
		This is used by the aws default profile.`,
		Run: func(cmd *cobra.Command, args []string) {
			conf := ReadConfig(ConfigFilePath())
			startURL = conf.StartURL
			region = conf.Region
			oidcClient, ssoClient := InitClients(region)
			AssumeDirectly(oidcClient, ssoClient)
		},
	}

	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to AWS SSO",
		Long:  "Login to AWS SSO by retrieving short-lived credentials.",
		Run: func(cmd *cobra.Command, args []string) {
			conf := ReadConfig(ConfigFilePath())
			startURL = conf.StartURL
			region = conf.Region
			oidcClient, ssoClient := InitClients(region)
			promptSelector := Prompter{}

			oidcInformation := OIDCInformation{
				Client: oidcClient,
				URL:    startURL,
			}
			clientInformation, _ := oidcInformation.ProcessClientInformation()

			accountInfo := RetrieveAccountInfo(ssoClient, clientInformation, promptSelector)
			roleInfo := RetrieveRoleInfo(ssoClient, accountInfo, clientInformation, promptSelector)
			SaveUsageInformation(accountInfo, roleInfo)

			input := &sso.GetRoleCredentialsInput{
				AccountId:   accountInfo.AccountId,
				RoleName:    roleInfo.RoleName,
				AccessToken: &clientInformation.AccessToken,
			}
			roleCredentials, err := ssoClient.GetRoleCredentials(ctx, input)
			if err != nil {
				log.Fatalf("Encountered error attempting to GetRoleCredentials: %v", err)
			}

			if len(accountID) == 0 {
				accountID = *accountInfo.AccountId
			}
			if len(roleName) == 0 {
				roleName = *roleInfo.RoleName
			}

			if persist {
				template := ProcessPersistedCredentialsTemplate(roleCredentials, profile)
				WriteAWSCredentialsFile(template)
				log.Printf("Credentails expire at: %s\n", time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0))
			} else {
				template := ProcessCredentialProcessTemplate(accountID, roleName, profile, region)
				WriteAWSCredentialsFile(template)
			}
		},
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
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(generateCmd)
	configCmd.AddCommand(editCmd)
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(assumeCmd)
	rootCmd.AddCommand(loginCmd)

	// flags
	assumeCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	assumeCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	assumeCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	assumeCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
	assumeCmd.Flags().StringVarP(&roleName, "role-name", "n", "", "role name to assume")
	assumeCmd.Flags().StringVarP(&accountID, "account-id", "a", "", "account id where the role exists")

	refreshCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	refreshCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	refreshCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	refreshCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")

	loginCmd.Flags().StringVarP(&startURL, "start-url", "u", "", "set / override aws sso url start url")
	loginCmd.Flags().StringVarP(&region, "region", "r", "", "set / override aws region")
	loginCmd.Flags().StringVarP(&profile, "profile", "p", "default", "the profile name to set in credentials file")
	loginCmd.Flags().BoolVarP(&persist, "persist", "", false, "toggle if you want to write short-lived creds to credentials file")
}
