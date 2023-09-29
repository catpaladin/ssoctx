package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/sso"
)

func printEnvironmentVariables(creds *sso.GetRoleCredentialsOutput) {
	// unset and then set
	unsetEnvironmentVariables()
	fmt.Println("Please copy and paste credentials to use environment variables\n")
	fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", *creds.RoleCredentials.AccessKeyId)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", *creds.RoleCredentials.SecretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN=%s\n", *creds.RoleCredentials.SessionToken)
}

func unsetEnvironmentVariables() {
	_, accessKeyExists := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if accessKeyExists {
		_ = os.Unsetenv("AWS_ACCESS_KEY_ID")
	}

	_, secretKeyExists := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if secretKeyExists {
		_ = os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	}

	_, sessionTokenExists := os.LookupEnv("AWS_SESSION_TOKEN")
	if sessionTokenExists {
		_ = os.Unsetenv("AWS_SESSION_TOKEN")
	}
}
