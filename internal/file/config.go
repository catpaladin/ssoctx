// Package file contains needed functionality for config and files
package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"aws-sso-util/internal/prompt"

	"github.com/rs/zerolog"
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
func ConfigFilePath(ctx context.Context) string {
	logger := zerolog.Ctx(ctx)
	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Fatal().Msgf("Encountered error finding user config dir: %q", err)
	}
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s%s\\config.yml", configDir, WindowsConfigPath)
	}
	return fmt.Sprintf("%s%s/config.yml", configDir, UserConfigFilePath)
}

// GetConfigs reads the config and sets values.
func GetConfigs(ctx context.Context, startURL, region *string) {
	conf := ReadConfig(ctx, ConfigFilePath(ctx))
	if len(*startURL) == 0 {
		*startURL = conf.StartURL
	}
	if len(*region) == 0 {
		*region = conf.Region
	}
}

// ReadConfig is used to read the config by filePath
func ReadConfig(ctx context.Context, filePath string) *AppConfig {
	logger := zerolog.Ctx(ctx)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		logger.Fatal().Msgf("Encountered error reading file path: %q", err)
	}

	appConfig := AppConfig{}
	err = yaml.Unmarshal(bytes, &appConfig)
	if err != nil {
		logger.Fatal().Msgf("Encountered error in unmarshal of config: %q", err)
	}
	return &appConfig
}

// GenerateConfigAction is used to generate a config yaml
func GenerateConfigAction(ctx context.Context) error {
	prompter := prompt.Prompter{}
	startURL := prompt.PromptStartURL(prompter, "")
	region := prompt.PromptRegion(prompter)
	appConfig := AppConfig{
		StartURL: startURL,
		Region:   region,
	}

	configFile := ConfigFilePath(ctx)
	if err := writeConfig(ctx, configFile, appConfig); err != nil {
		return fmt.Errorf("Encountered error at GenerateConfigAction: %w", err)
	}
	return nil
}

// EditConfigAction is used to edit the generated config yaml
func EditConfigAction(ctx context.Context) error {
	config := ReadConfig(ctx, ConfigFilePath(ctx))

	prompter := prompt.Prompter{}
	config.StartURL = prompt.PromptStartURL(prompter, config.StartURL)
	config.Region = prompt.PromptRegion(prompter)

	if err := writeConfig(ctx, ConfigFilePath(ctx), *config); err != nil {
		return fmt.Errorf("Encountered error at EditConfigAction: %w", err)
	}
	return nil
}

func writeConfig(ctx context.Context, filePath string, ac AppConfig) error {
	logger := zerolog.Ctx(ctx)
	bytes, err := yaml.Marshal(ac)
	if err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	base := filepath.Dir(filePath)
	if err = os.MkdirAll(base, 0o755); err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	if err = os.WriteFile(filePath, bytes, 0o755); err != nil {
		return fmt.Errorf("Encountered error at writeConfig: %w", err)
	}

	logger.Info().Msgf("Config file generated: %s", filePath)

	return nil
}
