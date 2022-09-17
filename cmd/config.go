package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"gopkg.in/yaml.v2"
)

const (
	// UserConfigFilePath is the file path under User's config path
	UserConfigFilePath string = "/aws-sso-util/config.yml"
)

// AppConfig is used to save yaml config
type AppConfig struct {
	StartURL string `yaml:"start-url"`
	Region   string `yaml:"region"`
}

// ConfigFilePath is the default config path
func ConfigFilePath() string {
	configDir, err := os.UserConfigDir()
	check(err)
	return configDir + UserConfigFilePath
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
	startUrl := promptStartURL(prompter, "")
	region := promptRegion(prompter)
	appConfig := AppConfig{
		StartURL: startUrl,
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

	base := path.Dir(filePath)
	if err = os.MkdirAll(base, 0755); err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	if err = ioutil.WriteFile(filePath, bytes, 0755); err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	log.Printf("Config file generated: %s", filePath)

	return nil
}
