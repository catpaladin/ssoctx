package main

import (
	"github.com/spf13/cobra"

	"ssoctx/internal/file"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Handles configuration",
		Long: `Handles configuration. Config location defaults to
		${HOME}/.config/ssoctx/config.yaml`,
	}

	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a config file",
		Long: `Generate a config file. All available properities are interactively prompted.
		Overrides the existing config.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := configureLogger(debug, jsonFormat)
			ctx = logger.WithContext(ctx)
			if err := file.GenerateConfig(ctx); err != nil {
				logger.Fatal().Err(err)
			}
		},
	}

	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit the config file",
		Long: `Edit the config file. All available properities are interactively prompted.
		Overrides the existing config.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := configureLogger(debug, jsonFormat)
			ctx = logger.WithContext(ctx)
			if err := file.EditConfig(ctx); err != nil {
				logger.Fatal().Err(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(generateCmd)
	configCmd.AddCommand(editCmd)
	configCmd.Flags().BoolVarP(&debug, "debug", "", false, "toggle if you want to enable debug logs")
	configCmd.Flags().BoolVarP(&jsonFormat, "json", "", false, "toggle if you want to enable json log output")
}
