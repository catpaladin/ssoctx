package terminal

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/charmbracelet/huh"
)

func TestSelectRegion(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     string
		mockError      error
		expectedResult string
	}{
		{
			name:           "Successful selection",
			mockReturn:     "us-west-2",
			mockError:      nil,
			expectedResult: "us-west-2",
		},
		{
			name:           "Error in selection",
			mockReturn:     "",
			mockError:      errors.New("mock error"),
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSelector := func(options []huh.Option[string], title string) (string, error) {
				return tt.mockReturn, tt.mockError
			}

			result := SelectRegion(mockSelector)
			if result != tt.expectedResult {
				t.Errorf("SelectRegion() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestSelectRole(t *testing.T) {
	tests := []struct {
		name           string
		roles          *sso.ListAccountRolesOutput
		mockReturn     types.RoleInfo
		mockError      error
		expectedResult *types.RoleInfo
		expectedError  bool
	}{
		{
			name: "Single role",
			roles: &sso.ListAccountRolesOutput{
				RoleList: []types.RoleInfo{
					{RoleName: ptr("SingleRole")},
				},
			},
			expectedResult: &types.RoleInfo{RoleName: ptr("SingleRole")},
			expectedError:  false,
		},
		{
			name: "Multiple roles, successful selection",
			roles: &sso.ListAccountRolesOutput{
				RoleList: []types.RoleInfo{
					{RoleName: ptr("Role1")},
					{RoleName: ptr("Role2")},
				},
			},
			mockReturn:     types.RoleInfo{RoleName: ptr("Role2")},
			mockError:      nil,
			expectedResult: &types.RoleInfo{RoleName: ptr("Role2")},
			expectedError:  false,
		},
		{
			name: "Multiple roles, error in selection",
			roles: &sso.ListAccountRolesOutput{
				RoleList: []types.RoleInfo{
					{RoleName: ptr("Role1")},
					{RoleName: ptr("Role2")},
				},
			},
			mockReturn:     types.RoleInfo{},
			mockError:      errors.New("mock error"),
			expectedResult: &types.RoleInfo{},
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSelector := func(options []huh.Option[types.RoleInfo], title string) (types.RoleInfo, error) {
				return tt.mockReturn, tt.mockError
			}

			_, err := SelectRole(tt.roles, mockSelector)
			if (err != nil) != tt.expectedError {
				t.Errorf("SelectRole() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
		})
	}
}

func TestSelectAccount(t *testing.T) {
	tests := []struct {
		name           string
		accounts       *sso.ListAccountsOutput
		mockReturn     types.AccountInfo
		mockError      error
		expectedResult *types.AccountInfo
		expectedError  bool
	}{
		{
			name: "Successful selection",
			accounts: &sso.ListAccountsOutput{
				AccountList: []types.AccountInfo{
					{AccountName: ptr("Account1"), AccountId: ptr("123")},
					{AccountName: ptr("Account2"), AccountId: ptr("456")},
				},
			},
			mockReturn:     types.AccountInfo{AccountName: ptr("Account2"), AccountId: ptr("456")},
			mockError:      nil,
			expectedResult: &types.AccountInfo{AccountName: ptr("Account2"), AccountId: ptr("456")},
			expectedError:  false,
		},
		{
			name: "Error in selection",
			accounts: &sso.ListAccountsOutput{
				AccountList: []types.AccountInfo{
					{AccountName: ptr("Account1"), AccountId: ptr("123")},
					{AccountName: ptr("Account2"), AccountId: ptr("456")},
				},
			},
			mockReturn:     types.AccountInfo{},
			mockError:      errors.New("mock error"),
			expectedResult: &types.AccountInfo{},
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSelector := func(options []huh.Option[types.AccountInfo], title string) (types.AccountInfo, error) {
				return tt.mockReturn, tt.mockError
			}

			_, err := SelectAccount(tt.accounts, mockSelector)
			if (err != nil) != tt.expectedError {
				t.Errorf("SelectAccount() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
		})
	}
}

func TestSortAccounts(t *testing.T) {
	tests := []struct {
		name     string
		input    []types.AccountInfo
		expected []types.AccountInfo
	}{
		{
			name: "Sort accounts",
			input: []types.AccountInfo{
				{AccountName: ptr("C Account"), AccountId: ptr("3")},
				{AccountName: ptr("A Account"), AccountId: ptr("1")},
				{AccountName: ptr("B Account"), AccountId: ptr("2")},
			},
			expected: []types.AccountInfo{
				{AccountName: ptr("A Account"), AccountId: ptr("1")},
				{AccountName: ptr("B Account"), AccountId: ptr("2")},
				{AccountName: ptr("C Account"), AccountId: ptr("3")},
			},
		},
		{
			name:     "Empty list",
			input:    []types.AccountInfo{},
			expected: []types.AccountInfo{},
		},
		{
			name: "Single account",
			input: []types.AccountInfo{
				{AccountName: ptr("A Account"), AccountId: ptr("1")},
			},
			expected: []types.AccountInfo{
				{AccountName: ptr("A Account"), AccountId: ptr("1")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortAccounts(tt.input)
		})
	}
}
