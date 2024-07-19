package amazon

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// AccessToken is used to marshal results from GetRoleCredentials
type AccessToken struct {
	Version         int    `json:"Version"`
	AccessKeyID     string `json:"AccessKeyId"`
	Expiration      string `json:"Expiration"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
}

// AssumeFlagInputs contains all needed inputs for Credentials
type AssumeFlagInputs struct {
	AccountID string // flag input from assume
	RoleName  string // flag input from assume
	StartURL  string // flag input from assume
	Region    string // flag input from assume
	Profile   string // flag input from assume
}

// AssumeCredentialProcess is used to return auth through credential process
// Directly assumes into a certain account and role, bypassing the prompt and interactive selection.
// Sends expected json marshalled response to stdout for the aws credentials file credential process
func AssumeCredentialProcess(ctx context.Context, o *OIDCClientAPI, s *Client, inputs AssumeFlagInputs) {
	logger := zerolog.Ctx(ctx)

	clientInfoDestination := clientInfoFileDestination(inputs.StartURL)
	clientInformation, _ := o.processClientInformation(ctx, clientInfoDestination)
	writeStructToFile(ctx, &clientInformation, clientInfoFileDestination(inputs.StartURL))
	roleCredentials, err := s.getRolesCredentials(
		ctx,
		inputs.AccountID,
		inputs.RoleName,
		clientInformation.AccessToken,
	)
	if err != nil {
		logger.Fatal().Msgf("Something went wrong: %q", err)
	}

	if len(inputs.StartURL) == 0 {
		inputs.StartURL = clientInformation.StartURL
	}

	template := getCredentialProcess(inputs.AccountID, inputs.RoleName, inputs.Region, inputs.StartURL)
	writeAWSCredentialsFile(ctx, &template, inputs.Profile)

	creds := AccessToken{
		Version:         1,
		AccessKeyID:     *roleCredentials.RoleCredentials.AccessKeyId,
		Expiration:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		SecretAccessKey: *roleCredentials.RoleCredentials.SecretAccessKey,
		SessionToken:    *roleCredentials.RoleCredentials.SessionToken,
	}
	bytes, _ := json.Marshal(creds)
	os.Stdout.Write(bytes)
}
