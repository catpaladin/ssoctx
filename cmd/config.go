package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	// UserConfigFilePath is the file path under User's config path
	UserConfigFilePath string = "/aws-sso-util"
	// WindowsConfigPath is where windows will store User's config path
	WindowsConfigPath string = "\\aws-sso-util"
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
)

func init() {
	// subcommands
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(generateCmd)
	configCmd.AddCommand(editCmd)
}

// AppConfig is used to save yaml config
type AppConfig struct {
	StartURL string `yaml:"start-url"`
	Region   string `yaml:"region"`
}

// ConfigFilePath is the default config path
func ConfigFilePath() string {
	configDir, err := os.UserConfigDir()
	check(err)
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s%s\\config.yml", configDir, WindowsConfigPath)
	}
	return fmt.Sprintf("%s%s/config.yml", configDir, UserConfigFilePath)
}

// GetConfigs reads the config and sets values.
func GetConfigs(startURL, region *string) {
	conf := ReadConfig(ConfigFilePath())
	if len(*startURL) == 0 {
		*startURL = conf.StartURL
	}
	if len(*region) == 0 {
		*region = conf.Region
	}
}

// ReadConfig is used to read the config by filePath
func ReadConfig(filePath string) *AppConfig {
	bytes, err := ioutil.ReadFile(filePath)
	check(err)

	appConfig := AppConfig{}
	err = yaml.Unmarshal(bytes, &appConfig)
	check(err)
	return &appConfig
}

// GenerateConfigAction is used to generate a config yaml
func GenerateConfigAction() error {
	prompter := Prompter{}
	startURL := promptStartURL(prompter, "")
	region := promptRegion(prompter)
	appConfig := AppConfig{
		StartURL: startURL,
		Region:   region,
	}

	configFile := ConfigFilePath()
	if err := writeConfig(configFile, appConfig); err != nil {
		return fmt.Errorf("Encountered error at GenerateConfigAction: %w", err)
	}
	return nil
}

// EditConfigAction is used to edit the generated config yaml
func EditConfigAction() error {
	config := ReadConfig(ConfigFilePath())

	prompter := Prompter{}
	config.StartURL = promptStartURL(prompter, config.StartURL)
	config.Region = promptRegion(prompter)

	if err := writeConfig(ConfigFilePath(), *config); err != nil {
		return fmt.Errorf("Encountered error at EditConfigAction: %w", err)
	}
	return nil

}

func promptStartURL(prompt Prompt, dfault string) string {
	return prompt.Prompt("SSO Start URL", dfault)
}

func promptRegion(prompt Prompt) string {
	_, region := prompt.Select("Select your AWS Region", AwsRegions, func(input string, index int) bool {
		target := AwsRegions[index]
		return fuzzy.MatchFold(input, target)
	})
	return region
}

func writeConfig(filePath string, ac AppConfig) error {
	bytes, err := yaml.Marshal(ac)
	if err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	base := filepath.Dir(filePath)
	if err = os.MkdirAll(base, 0755); err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	if err = ioutil.WriteFile(filePath, bytes, 0755); err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	log.Printf("Config file generated: %s", filePath)

	return nil
}
