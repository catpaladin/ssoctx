package terminal

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/charmbracelet/huh"
)

func TestGenerateGenericOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []huh.Option[int]
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []huh.Option[int]{},
		},
		{
			name:  "Single item",
			input: []int{1},
			expected: []huh.Option[int]{
				{Key: "1", Value: 1},
			},
		},
		{
			name:  "Multiple items",
			input: []int{1, 2, 3},
			expected: []huh.Option[int]{
				{Key: "1", Value: 1},
				{Key: "2", Value: 2},
				{Key: "3", Value: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenerateGenericOptions(tt.input)
		})
	}
}

func TestGenerateAccountInfoOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    []types.AccountInfo
		expected []huh.Option[types.AccountInfo]
	}{
		{
			name:     "Empty slice",
			input:    []types.AccountInfo{},
			expected: []huh.Option[types.AccountInfo]{},
		},
		{
			name: "Single item",
			input: []types.AccountInfo{
				{AccountName: ptr("Test Account"), AccountId: ptr("123456789012")},
			},
			expected: []huh.Option[types.AccountInfo]{
				{
					Key:   "Test Account                    123456789012",
					Value: types.AccountInfo{AccountName: ptr("Test Account"), AccountId: ptr("123456789012")},
				},
			},
		},
		{
			name: "Multiple items",
			input: []types.AccountInfo{
				{AccountName: ptr("Test Account 1"), AccountId: ptr("123456789012")},
				{AccountName: ptr("Test Account 2"), AccountId: ptr("210987654321")},
			},
			expected: []huh.Option[types.AccountInfo]{
				{
					Key:   "Test Account 1                 123456789012",
					Value: types.AccountInfo{AccountName: ptr("Test Account 1"), AccountId: ptr("123456789012")},
				},
				{
					Key:   "Test Account 2                 210987654321",
					Value: types.AccountInfo{AccountName: ptr("Test Account 2"), AccountId: ptr("210987654321")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generateAccountInfoOptions(tt.input)
		})
	}
}

func TestGenerateRoleInfoOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    []types.RoleInfo
		expected []huh.Option[types.RoleInfo]
	}{
		{
			name:     "Empty slice",
			input:    []types.RoleInfo{},
			expected: []huh.Option[types.RoleInfo]{},
		},
		{
			name: "Single item",
			input: []types.RoleInfo{
				{RoleName: ptr("TestRole")},
			},
			expected: []huh.Option[types.RoleInfo]{
				{
					Key:   "TestRole",
					Value: types.RoleInfo{RoleName: ptr("TestRole")},
				},
			},
		},
		{
			name: "Multiple items",
			input: []types.RoleInfo{
				{RoleName: ptr("TestRole1")},
				{RoleName: ptr("TestRole2")},
			},
			expected: []huh.Option[types.RoleInfo]{
				{
					Key:   "TestRole1",
					Value: types.RoleInfo{RoleName: ptr("TestRole1")},
				},
				{
					Key:   "TestRole2",
					Value: types.RoleInfo{RoleName: ptr("TestRole2")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generateRoleInfoOptions(tt.input)
		})
	}
}

// Helper function to create string pointers
func ptr(s string) *string {
	return &s
}
