// Package prompt contains all prompt functionality
package prompt

import "github.com/chzyer/readline"

type silentStdout struct{}

// SilentStdout is needed because users with terminal sounds enabled hear alerts
var SilentStdout = &silentStdout{}

// Write ensures no sounds from the terminal in stdout
func (s *silentStdout) Write(b []byte) (int, error) {
	if len(b) == 1 && b[0] == readline.CharBell {
		return 0, nil
	}
	return readline.Stdout.Write(b)
}

// Close closes the stdout
func (s *silentStdout) Close() error {
	return readline.Stdout.Close()
}
