//go:build linux
// +build linux

package amazon

import (
	"errors"
	"os/exec"
	"testing"
)

func TestOpenURLInBrowser(t *testing.T) {
	tests := []struct {
		name        string
		mockFunc    func(string, ...string) *exec.Cmd
		expectedErr error
	}{
		{
			name: "successful command execution",
			mockFunc: func(command string, args ...string) *exec.Cmd {
				return exec.Command("echo", "test")
			},
			expectedErr: nil,
		},
		{
			name: "command execution failure",
			mockFunc: func(command string, args ...string) *exec.Cmd {
				return &exec.Cmd{Err: errors.New("mock error")}
			},
			expectedErr: errors.New("unable to open browser: mock error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCmd = tt.mockFunc
			err := openURLInBrowser("http://example.com")
			if err == nil && tt.expectedErr != nil || err != nil && tt.expectedErr == nil {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			if err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			// reset
			execCmd = exec.Command
		})
	}
}
