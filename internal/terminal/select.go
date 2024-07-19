package terminal

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/charmbracelet/huh"
)

// NewSelectForm creates a new select form
func NewSelectForm[T comparable](options []huh.Option[T], title string) (T, error) {
	var output T

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[T]().
				Title(title).
				Options(options...).
				Value(&output),
		),
	)

	err := form.Run()
	if err != nil {
		return output, err
	}
	return output, nil
}

// GenerateGenericOptions generates options dynamically
func GenerateGenericOptions[T comparable](items []T) []huh.Option[T] {
	options := make([]huh.Option[T], len(items))
	for i, item := range items {
		options[i] = huh.Option[T]{
			Key:   fmt.Sprintf("%v", item),
			Value: item,
		}
	}
	return options
}

// generateAccountInfoOptions generates AccountInfo options
func generateAccountInfoOptions(items []types.AccountInfo) []huh.Option[types.AccountInfo] {
	options := make([]huh.Option[types.AccountInfo], len(items))
	for i, item := range items {
		options[i] = huh.Option[types.AccountInfo]{
			Key:   fmt.Sprintf("%-30s %s", *item.AccountName, *item.AccountId),
			Value: item,
		}
	}
	return options
}

// generateRoleInfoOptions generates RoleInfo options
func generateRoleInfoOptions(items []types.RoleInfo) []huh.Option[types.RoleInfo] {
	options := make([]huh.Option[types.RoleInfo], len(items))
	for i, item := range items {
		options[i] = huh.Option[types.RoleInfo]{
			Key:   fmt.Sprint(*item.RoleName),
			Value: item,
		}
	}
	return options
}
