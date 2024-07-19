package amazon

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func TestClientInformation_IsExpired(t *testing.T) {
	dur, _ := time.ParseDuration("8h")
	type fields struct {
		AccessTokenExpiresAt    time.Time
		AccessToken             string
		ClientID                string
		ClientSecret            string
		ClientSecretExpiresAt   string
		DeviceCode              string
		VerificationURIComplete string
		StartURL                string
	}
	tests := []struct {
		name   string
		fields fields
		ctx    context.Context
		want   bool
	}{
		{
			name: "Is Expired",
			fields: fields{
				AccessTokenExpiresAt: time.Now().Add(-dur),
			},
			ctx:  log.Logger.WithContext(context.Background()),
			want: true,
		},
		{
			name: "Is Not Expired",
			fields: fields{
				AccessTokenExpiresAt: time.Now().Add(dur),
			},
			ctx:  log.Logger.WithContext(context.Background()),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ati := ClientInformation{
				AccessTokenExpiresAt:    tt.fields.AccessTokenExpiresAt,
				AccessToken:             tt.fields.AccessToken,
				ClientID:                tt.fields.ClientID,
				ClientSecret:            tt.fields.ClientSecret,
				ClientSecretExpiresAt:   tt.fields.ClientSecretExpiresAt,
				DeviceCode:              tt.fields.DeviceCode,
				VerificationURIComplete: tt.fields.VerificationURIComplete,
				StartURL:                tt.fields.StartURL,
			}
			if got := ati.isExpired(tt.ctx); got != tt.want {
				t.Errorf("ClientInformation.isExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
