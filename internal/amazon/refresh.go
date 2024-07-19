package amazon

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"ssoctx/internal/terminal"
)

// RefreshFlagInputs contains all needed inputs for Credentials
type RefreshFlagInputs struct {
	AccountID string // flag input from refresh
	RoleName  string // flag input from refresh
	StartURL  string // flag input from refresh
	Region    string // flag input from refresh
	Profile   string // flag input from refresh
	Persist   bool   // flag associated with type of credentials
}

// Credentials is used to refresh credentials
func Credentials(ctx context.Context, o *OIDCClientAPI, s *Client, inputs RefreshFlagInputs) {
	logger := zerolog.Ctx(ctx)

	destination := clientInfoFileDestination(inputs.StartURL)
	clientInformation, err := readClientInformation(ctx, destination)
	if err != nil || clientInformation.StartURL != inputs.StartURL {
		clientInformation, _ = o.processClientInformation(ctx, destination)
	}
	writeStructToFile(ctx, &clientInformation, destination)
	logger.Printf("Using Start URL %s", clientInformation.StartURL)

	if len(inputs.AccountID) == 0 && len(inputs.RoleName) == 0 {
		logger.Info().Msg("No account-id or role-name provided.")
		logger.Info().Msgf("Refreshing credentials from access to profile: %s", inputs.Profile)

		accountsOutput, err := s.listAccounts(ctx, clientInformation.AccessToken)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in listAccounts: %v", err)
		}
		accountInfo, err := terminal.SelectAccount(accountsOutput)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in selectAccount: %v", err)
		}
		listRolesOutput, err := s.listAvailableRoles(
			ctx,
			*accountInfo.AccountId,
			clientInformation.AccessToken,
		)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in listAvailableRoles: %v", err)
		}
		roleInfo, err := terminal.SelectRole(listRolesOutput)
		if err != nil {
			logger.Fatal().Msgf("Encountered error in selectRole: %v", err)
		}
		inputs.RoleName = *roleInfo.RoleName
		inputs.AccountID = *accountInfo.AccountId
	}
	logger.Info().Msgf("Refreshing for account %s with permission set role %s", inputs.AccountID, inputs.RoleName)

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

	if inputs.Persist {
		template := getPersistedCredentials(roleCredentials, inputs.Region)
		writeAWSCredentialsFile(ctx, &template, inputs.Profile)
	} else {
		template := getCredentialProcess(inputs.AccountID, inputs.RoleName, inputs.Region, inputs.StartURL)
		writeAWSCredentialsFile(ctx, &template, inputs.Profile)
	}

	logger.Info().Msgf("Successful retrieved credentials for account: %s", inputs.AccountID)
	logger.Info().Msgf("Assumed role: %s", inputs.RoleName)
	logger.Info().Msgf(
		"Credentials expire at: %s",
		time.Unix(roleCredentials.RoleCredentials.Expiration/1000, 0).Format(time.RFC3339),
	)
}
