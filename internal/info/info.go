// Package info contains client info
package info

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// ClientInformation is used to store client information
type ClientInformation struct {
	AccessTokenExpiresAt    time.Time
	AccessToken             string
	ClientID                string
	ClientSecret            string
	ClientSecretExpiresAt   string
	DeviceCode              string
	VerificationURIComplete string
	StartURL                string
}

// IsExpired is used to tell if AccessToken is expired in client information
func (ati ClientInformation) IsExpired(ctx context.Context) bool {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msgf("ClientInformation time: %v", ati.AccessTokenExpiresAt)
	logger.Debug().Msgf("Time now: %v", time.Now())
	return ati.AccessTokenExpiresAt.Before(time.Now())
}
