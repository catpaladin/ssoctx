package amazon

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sso"
)

// printEnvironmentVariables is used to print creds to stdout
func printEnvironmentVariables(creds *sso.GetRoleCredentialsOutput) {
	fmt.Println("Please copy and paste credentials to use environment variables")
	ak := fmt.Sprintf("export AWS_ACCESS_KEY_ID=%s", *creds.RoleCredentials.AccessKeyId)
	sk := fmt.Sprintf("export AWS_SECRET_ACCESS_KEY=%s", *creds.RoleCredentials.SecretAccessKey)
	st := fmt.Sprintf("export AWS_SESSION_TOKEN=%s", *creds.RoleCredentials.SessionToken)
	fmt.Println("export AWS_REGION=us-east-1")
	fullTest := fmt.Sprintf("%s\n%s\n%s\n", ak, sk, st)
	fmt.Println(fullTest)
}
