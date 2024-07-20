package amazon

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

// mockAWSError implements the smithy.APIError interface
type mockAWSError struct {
	code    string
	message string
	fault   smithy.ErrorFault
}

func (e mockAWSError) ErrorCode() string {
	return e.code
}

func (e mockAWSError) ErrorMessage() string {
	return e.message
}

func (e mockAWSError) ErrorFault() smithy.ErrorFault {
	return e.fault
}

func (e mockAWSError) Error() string {
	return e.message
}

func TestGetAWSErrorCode(t *testing.T) {
	// Create a logger that writes to a buffer
	logger := zerolog.New(zerolog.NewTestWriter(t))

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "AWS Error",
			err: mockAWSError{
				code:    "AccessDenied",
				message: "Access Denied",
				fault:   smithy.FaultClient,
			},
			expected: "AccessDenied",
		},
		{
			name:     "Non-AWS Error",
			err:      errors.New("some other error"),
			expected: "",
		},
		{
			name:     "Nil Error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := logger.WithContext(context.Background())
			result := GetAWSErrorCode(ctx, tt.err)
			if result != tt.expected {
				t.Errorf("GetAWSErrorCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}
