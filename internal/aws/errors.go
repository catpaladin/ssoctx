// Package aws contains all the aws logic
package aws

import (
	"context"
	"errors"

	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

// GetAWSErrorCode is used to get AWS error code
func GetAWSErrorCode(ctx context.Context, err error) string {
	logger := zerolog.Ctx(ctx)

	var awsErr smithy.APIError
	if errors.As(err, &awsErr) {
		logger.Debug().Msgf("%s: %s", awsErr.ErrorCode(), awsErr.ErrorMessage())
		logger.Debug().Msgf("%s", awsErr.ErrorFault().String())
		return awsErr.ErrorCode()
	}
	logger.Debug().Msg("Unable to get aws error code from error")
	return ""
}
