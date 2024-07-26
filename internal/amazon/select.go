package amazon

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/rs/zerolog"

	"ssoctx/internal/file"
	"ssoctx/internal/terminal"
)

// SelectFlagInputs contains all needed inputs for login to select account and role
type SelectFlagInputs struct {
	Clean      bool   // flag associated for clean
	AccountID  string // flag input from login
	RoleName   string // flag input from login
	StartURL   string // flag input from login
	Region     string // flag input from login
	Profile    string // flag input from login
	Persist    bool   // flag associated with type of credentials
	PrintCreds bool   // flag associated with printing keys
}

// Select is the primary subcommand used to interactively select account and role
func Select(ctx context.Context, o *OIDCClientAPI, s *Client, inputs SelectFlagInputs) {
	logger := zerolog.Ctx(ctx)
	destination := clientInfoFileDestination(inputs.StartURL)

	if inputs.Clean {
		file.RemoveLock(ctx)
		if err := os.Remove(destination); err != nil {
			logger.Panic().Msgf("Failed to remove access token: %q", err)
		}
	}

	clientInformation, err := o.processClientInformation(ctx, destination)
	if err != nil {
		logger.Fatal().Msgf("Encountered error in processClientInformation: %v", err)
	}
	writeStructToFile(ctx, &clientInformation, destination)

	var accountInfo *types.AccountInfo
	if len(inputs.AccountID) == 0 {
		accountsOutput, laErr := s.listAccounts(ctx, clientInformation.AccessToken)
		if laErr != nil {
			logger.Fatal().Msgf("Encountered error in listAccounts: %v", laErr)
		}
		accountInfo, err = terminal.SelectAccount(accountsOutput, terminal.NewSelectForm)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in selectAccount: %v", err)
		}
		inputs.AccountID = *accountInfo.AccountId
	}

	var roleInfo *types.RoleInfo
	if len(inputs.RoleName) == 0 {
		listRolesOutput, larErr := s.listAvailableRoles(ctx, inputs.AccountID, clientInformation.AccessToken)
		if larErr != nil {
			logger.Fatal().Msgf("Encountered error in listAvailableRoles: %v", larErr)
		}
		roleInfo, err = terminal.SelectRole(listRolesOutput, terminal.NewSelectForm)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in selectRole: %v", err)
		}
		inputs.RoleName = *roleInfo.RoleName
	}

	if len(inputs.StartURL) == 0 {
		inputs.StartURL = clientInformation.StartURL
	}

	roleCredentials, err := s.getRolesCredentials(ctx, inputs.AccountID, inputs.RoleName, clientInformation.AccessToken)
	if err != nil {
		logger.Fatal().Msgf("Encountered error attempting to getRoleCredentials: %v", err)
	}

	// does not write to file because folks just want environment variables
	if inputs.PrintCreds {
		printEnvironmentVariables(roleCredentials)
		return
	}

	if inputs.Persist {
		template := getPersistedCredentials(roleCredentials, inputs.Region)
		writeAWSCredentialsFile(ctx, &template, inputs.Profile)
		logger.Info().Msgf(
			"Credentails expire at: %s",
			time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0).Format(time.RFC3339),
		)
	} else {
		template := getCredentialProcess(inputs.AccountID, inputs.RoleName, inputs.Region, inputs.StartURL)
		writeAWSCredentialsFile(ctx, &template, inputs.Profile)
	}
}
