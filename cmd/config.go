package cmd

import (
	"aws-sso-util/internal/file"

	"github.com/spf13/cobra"
)

var (
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
			_ = file.GenerateConfigAction()
		},
	}

	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit the config file",
		Long: `Edit the config file. All available properities are interactively prompted.
		Overrides the existing config.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = file.EditConfigAction()
		},
	}
)

func init() {
	// subcommands
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(generateCmd)
	configCmd.AddCommand(editCmd)
}
