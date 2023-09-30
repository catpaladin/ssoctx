package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/sso"
)

func printEnvironmentVariables(creds *sso.GetRoleCredentialsOutput) {
	// unset and then set
	unsetEnvironmentVariables()
	fmt.Println("Please copy and paste credentials to use environment variables")
	ak := fmt.Sprintf("export AWS_ACCESS_KEY_ID=%s", *creds.RoleCredentials.AccessKeyId)
	sk := fmt.Sprintf("export AWS_SECRET_ACCESS_KEY=%s", *creds.RoleCredentials.SecretAccessKey)
	st := fmt.Sprintf("export AWS_SESSION_TOKEN=%s", *creds.RoleCredentials.SessionToken)
	fullTest := fmt.Sprintf("%s\n%s\n%s\n", ak, sk, st)
	fmt.Println(fullTest)
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
