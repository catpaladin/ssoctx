// Package info contains client info
package info

import "time"

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
func (ati ClientInformation) IsExpired() bool {
	if ati.AccessTokenExpiresAt.Before(time.Now()) {
		return true
	}
	return false
}
