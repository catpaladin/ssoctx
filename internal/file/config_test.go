package file

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

// mockUserConfigDir is a function type that matches os.UserConfigDir
type mockUserConfigDir func() (string, error)

func TestGetConfigFilePath(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout)
	ctx = logger.WithContext(ctx)

	// Store the original function and defer its restoration
	originalUserConfigDir := userConfigDir
	defer func() { userConfigDir = originalUserConfigDir }()

	userConfigDir = func() (string, error) {
		return "/home/user", nil
	}

	expected := "/home/user/ssoctx/config.yml"
	result := GetConfigFilePath(ctx)

	if result != expected {
		t.Errorf("GetConfigFilePath() = %v, want %v", result, expected)
	}
}

func TestGetConfigs(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout)
	ctx = logger.WithContext(ctx)

	// Mock ReadConfig
	originalReadConfig := readConfigFunc
	defer func() { readConfigFunc = originalReadConfig }()

	readConfigFunc = func(ctx context.Context, filePath string) *AppConfig {
		return &AppConfig{StartURL: "https://example.com", Region: "us-west-2"}
	}

	tests := []struct {
		name           string
		inputStartURL  string
		inputRegion    string
		expectedConfig AppConfig
	}{
		{"BothEmpty", "", "", AppConfig{StartURL: "https://example.com", Region: "us-west-2"}},
		{"StartURLProvided", "https://provided.com", "", AppConfig{StartURL: "https://provided.com", Region: "us-west-2"}},
		{"RegionProvided", "", "eu-central-1", AppConfig{StartURL: "https://example.com", Region: "eu-central-1"}},
		{"BothProvided", "https://provided.com", "eu-central-1", AppConfig{StartURL: "https://provided.com", Region: "eu-central-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startURL := tt.inputStartURL
			region := tt.inputRegion
			GetConfigs(ctx, &startURL, &region)
			result := AppConfig{StartURL: startURL, Region: region}
			if !reflect.DeepEqual(result, tt.expectedConfig) {
				t.Errorf("GetConfigs() = %v, want %v", result, tt.expectedConfig)
			}
		})
	}
}

func TestReadConfig(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := zerolog.New(os.Stdout)
	ctx = logger.WithContext(ctx)

	// Create a temporary file
	content := []byte("start-url: https://example.com\nregion: us-west-2\n")
	tmpfile, err := os.CreateTemp("", "config.*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	expected := &AppConfig{StartURL: "https://example.com", Region: "us-west-2"}
	result := ReadConfig(ctx, tmpfile.Name())

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ReadConfig() = %v, want %v", result, expected)
	}
}

func TestGenerateConfig(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout)
	ctx = logger.WithContext(ctx)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "testconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock GetConfigFilePath
	originalGetConfigFilePath := GetConfigFilePath
	getConfigFilePathFunc = func(ctx context.Context) string {
		return filepath.Join(tmpDir, "config.yml")
	}
	defer func() { getConfigFilePathFunc = originalGetConfigFilePath }()

	err = GenerateConfig(ctx, "https://example.com", "us-west-2")
	if err != nil {
		t.Errorf("GenerateConfig() error = %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(filepath.Join(tmpDir, "config.yml")); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}
}

func TestEditConfig(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout)
	ctx = logger.WithContext(ctx)

	// Create a temporary file with initial content
	content := []byte("start-url: https://initial.com\nregion: us-east-1\n")
	tmpfile, err := os.CreateTemp("", "config.*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err = tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err = tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Store original functions and defer their restoration
	originalGetConfigFilePath := getConfigFilePathFunc
	originalReadConfig := readConfigFunc
	defer func() {
		getConfigFilePathFunc = originalGetConfigFilePath
		readConfigFunc = originalReadConfig
	}()

	// Mock GetConfigFilePath
	getConfigFilePathFunc = func(ctx context.Context) string {
		return tmpfile.Name()
	}

	// Mock ReadConfig to read from our temporary file
	readConfigFunc = func(ctx context.Context, filePath string) *AppConfig {
		bytes, readErr := os.ReadFile(filePath)
		if readErr != nil {
			t.Fatal(err)
		}
		var config AppConfig
		if yErr := yaml.Unmarshal(bytes, &config); yErr != nil {
			t.Fatal(err)
		}
		return &config
	}

	err = EditConfig(ctx, "https://edited.com", "eu-west-1")
	if err != nil {
		t.Errorf("EditConfig() error = %v", err)
	}

	// Verify the file was edited
	editedConfig := readConfigFunc(ctx, tmpfile.Name())
	expected := &AppConfig{StartURL: "https://edited.com", Region: "eu-west-1"}
	if !reflect.DeepEqual(editedConfig, expected) {
		t.Errorf("EditConfig() resulted in %v, want %v", editedConfig, expected)
	}
}
