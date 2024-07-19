// Package file contains needed functionality for config and files
package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"

	"ssoctx/internal/prompt"
)

const (
	// ProjectFileName is the name of built binary
	ProjectFileName string = "ssoctx"
)

// AppConfig is used to save yaml config
type AppConfig struct {
	StartURL string `yaml:"start-url"`
	Region   string `yaml:"region"`
}

// GetConfigFilePath is the default config path
func GetConfigFilePath(ctx context.Context) string {
	logger := zerolog.Ctx(ctx)
	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Fatal().Msgf("Encountered error finding user config dir: %q", err)
	}
	return filepath.Join(configDir, ProjectFileName, "config.yml")
}

// GetConfigs reads the config and sets values.
func GetConfigs(ctx context.Context, startURL, region *string) {
	conf := ReadConfig(ctx, GetConfigFilePath(ctx))
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

// GenerateConfig is used to generate a config yaml
func GenerateConfig(ctx context.Context) error {
	prompter := prompt.Prompter{}
	startURL := prompter.Enter("SSO Start URL", "")
	region := prompt.GetRegion(prompter)
	appConfig := AppConfig{
		StartURL: startURL,
		Region:   region,
	}

	configFile := GetConfigFilePath(ctx)
	if err := writeConfig(ctx, configFile, appConfig); err != nil {
		return fmt.Errorf("Encountered error at GenerateConfigAction: %w", err)
	}
	return nil
}

// EditConfig is used to edit the generated config yaml
func EditConfig(ctx context.Context) error {
	config := ReadConfig(ctx, GetConfigFilePath(ctx))

	prompter := prompt.Prompter{}
	config.StartURL = prompter.Enter("SSO Start URL", config.StartURL)
	config.Region = prompt.GetRegion(prompter)

	if err := writeConfig(ctx, GetConfigFilePath(ctx), *config); err != nil {
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
