// Package aws contains all the aws logic
package aws

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestOpenURLInBrowser(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		system  string
		url     string
		wantErr bool
	}{
		{
			name:    "TestStartDeviceAuthorizationLinux",
			ctx:     log.Logger.WithContext(context.Background()),
			system:  "linux",
			url:     "https://fakebanana.awsapp.com/start",
			wantErr: false,
		},
		{
			name:    "TestStartDeviceAuthorizationDarwin",
			ctx:     log.Logger.WithContext(context.Background()),
			system:  "darwin",
			url:     "https://fakebanana.awsapp.com/start",
			wantErr: false,
		},
		{
			name:    "TestStartDeviceAuthorizationWindows",
			ctx:     log.Logger.WithContext(context.Background()),
			system:  "windows",
			url:     "https://fakebanana.awsapp.com/start",
			wantErr: false,
		},
		{
			name:    "TestStartDeviceAuthorizationUnknown",
			ctx:     log.Logger.WithContext(context.Background()),
			system:  "potato",
			url:     "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		// set execCmd to call echo instead of open
		execCmd = fakeExecCmd
		t.Run(tt.name, func(t *testing.T) {
			if err := openURLInBrowser(tt.ctx, tt.system, tt.url); (err != nil) != tt.wantErr {
				t.Errorf("openURLInBrowser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
