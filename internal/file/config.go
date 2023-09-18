package file

import (
	"aws-sso-util/internal/prompt"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

const (
	// UserConfigFilePath is the file path under User's config path
	UserConfigFilePath string = "/aws-sso-util"
	// WindowsConfigPath is where windows will store User's config path
	WindowsConfigPath string = "\\aws-sso-util"
)

// AppConfig is used to save yaml config
type AppConfig struct {
	StartURL string `yaml:"start-url"`
	Region   string `yaml:"region"`
}

// ConfigFilePath is the default config path
func ConfigFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
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
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}

	appConfig := AppConfig{}
	err = yaml.Unmarshal(bytes, &appConfig)
	if err != nil {
		log.Fatalf("Something went wrong: %q", err)
	}
	return &appConfig
}

// GenerateConfigAction is used to generate a config yaml
func GenerateConfigAction() error {
	prompter := prompt.Prompter{}
	startURL := prompt.PromptStartURL(prompter, "")
	region := prompt.PromptRegion(prompter)
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

	prompter := prompt.Prompter{}
	config.StartURL = prompt.PromptStartURL(prompter, config.StartURL)
	config.Region = prompt.PromptRegion(prompter)

	if err := writeConfig(ConfigFilePath(), *config); err != nil {
		return fmt.Errorf("Encountered error at EditConfigAction: %w", err)
	}
	return nil

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

	if err = os.WriteFile(filePath, bytes, 0755); err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	log.Printf("Config file generated: %s", filePath)

	return nil
}
