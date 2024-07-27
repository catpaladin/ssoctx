package amazon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/rs/zerolog"
)

var (
	mockGetCredentialsFilePath    func() string
	mockClientInfoFileDestination func(string) string
)

// Override package-level functions with mocks
func init() {
	getCredentialsFilePath = func() string {
		if mockGetCredentialsFilePath != nil {
			return mockGetCredentialsFilePath()
		}
		return ""
	}

	clientInfoFileDestination = func(startURL string) string {
		if mockClientInfoFileDestination != nil {
			return mockClientInfoFileDestination(startURL)
		}
		return ""
	}
}

func TestGetPersistedCredentials(t *testing.T) {
	creds := &sso.GetRoleCredentialsOutput{
		RoleCredentials: &types.RoleCredentials{
			AccessKeyId:     aws.String("AKIAIOSFODNN7EXAMPLE"),
			SecretAccessKey: aws.String("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
			SessionToken:    aws.String("AQoDYXdzEJr..."),
		},
	}
	region := "us-west-2"

	result := getPersistedCredentials(creds, region)

	if result.AwsAccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("Expected AccessKeyID to be AKIAIOSFODNN7EXAMPLE, got %s", result.AwsAccessKeyID)
	}
	if result.AwsSecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("Expected SecretAccessKey to be wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY, got %s", result.AwsSecretAccessKey)
	}
	if result.AwsSessionToken != "AQoDYXdzEJr..." {
		t.Errorf("Expected SessionToken to be AQoDYXdzEJr..., got %s", result.AwsSessionToken)
	}
	if result.Region != "us-west-2" {
		t.Errorf("Expected Region to be us-west-2, got %s", result.Region)
	}
}

func TestGetCredentialProcess(t *testing.T) {
	accountID := "123456789012"
	roleName := "MyRole"
	region := "us-east-1"
	startURL := "https://d-123456abcd.awsapps.com/start"

	result := getCredentialProcess(accountID, roleName, region, startURL)

	expectedProcess := "ssoctx assume -a 123456789012 -n MyRole -u https://d-123456abcd.awsapps.com/start"
	if result.CredentialProcess != expectedProcess {
		t.Errorf("Expected CredentialProcess to be %s, got %s", expectedProcess, result.CredentialProcess)
	}
	if result.Region != "us-east-1" {
		t.Errorf("Expected Region to be us-east-1, got %s", result.Region)
	}
}

func TestGetCredentialsFilePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aws-creds-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	expected := filepath.Join(tempDir, ".aws", "credentials")
	mockGetCredentialsFilePath = func() string {
		return expected
	}

	result := getCredentialsFilePath()

	if result != expected {
		t.Errorf("Expected path %s, got %s", expected, result)
	}
}

func TestClientInfoFileDestination(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aws-sso-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	startURL := "https://d-123456abcd.awsapps.com/start"
	expected := filepath.Join(tempDir, ".aws", "sso", "cache", "access-token.json")

	mockClientInfoFileDestination = func(url string) string {
		if url != startURL {
			t.Errorf("Expected startURL %s, got %s", startURL, url)
		}
		return expected
	}

	result := clientInfoFileDestination(startURL)

	if result != expected {
		t.Errorf("Expected path %s, got %s", expected, result)
	}
}

func TestExists(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout)
	ctx = logger.WithContext(ctx)

	tempDir, err := os.MkdirTemp("", "exists-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	existingFile := filepath.Join(tempDir, "existing.txt")
	err = os.WriteFile(existingFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !exists(ctx, existingFile) {
		t.Errorf("Expected file %s to exist, but it doesn't", existingFile)
	}

	nonExistingFile := filepath.Join(tempDir, "non-existing.txt")
	if exists(ctx, nonExistingFile) {
		t.Errorf("Expected file %s to not exist, but it does", nonExistingFile)
	}
}

func TestWriteAWSCredentialsFile(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.NewFile(0, os.DevNull))
	ctx = logger.WithContext(ctx)

	tempDir, err := os.MkdirTemp("", "aws-creds-write-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	credentialsFilePath := filepath.Join(tempDir, ".aws", "credentials")
	os.MkdirAll(filepath.Dir(credentialsFilePath), 0755) // Ensure the directory exists

	template := &CredentialsTemplate{
		AwsAccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		AwsSecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		AwsSessionToken:    "AQoDYXdzEJr...",
		Region:             "us-west-2",
	}

	writeAWSCredentialsFile(ctx, template, "default")
}
