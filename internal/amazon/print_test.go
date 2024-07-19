package amazon

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
)

func Test_printEnvironmentVariables(t *testing.T) {
	type args struct {
		creds *sso.GetRoleCredentialsOutput
	}

	ak := "fakeaccesskey"
	sk := "fakesecretkey"
	st := "faketoken"

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Print Creds",
			args: args{
				creds: &sso.GetRoleCredentialsOutput{
					RoleCredentials: &types.RoleCredentials{
						AccessKeyId:     &ak,
						SecretAccessKey: &sk,
						SessionToken:    &st,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printEnvironmentVariables(tt.args.creds)
		})
	}
}
