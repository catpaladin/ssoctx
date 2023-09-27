// Package prompt contains functionality for terminal prompt and search
package prompt

import (
	"context"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog"
)

// Prompt is used to interface with promptui
type Prompt interface {
	Select(label string, toSelect []string, searcher func(input string, index int) bool) (index int, value string)
	Prompt(label string, dfault string) string
}

// Prompter is used to interface promptui
type Prompter struct {
	Ctx context.Context // context to pass in logger
}

// Select is used to select from the prompt
func (p Prompter) Select(label string, toSelect []string, searcher func(input string, index int) bool) (int, string) {
	logger := zerolog.Ctx(p.Ctx)
	prompt := promptui.Select{
		Label:             label,
		Items:             toSelect,
		Size:              20,
		Searcher:          searcher,
		StartInSearchMode: true,
		Stdout:            SilentStdout,
	}
	index, value, err := prompt.Run()
	if err != nil {
		logger.Fatal().Msgf("Error in prompt: %q", err)
	}
	return index, value
}

// Prompt is used to pop open a selectable prompt
func (p Prompter) Prompt(label string, dfault string) string {
	logger := zerolog.Ctx(p.Ctx)
	prompt := promptui.Prompt{
		Label:     label,
		Default:   dfault,
		AllowEdit: false,
	}
	val, err := prompt.Run()
	if err != nil {
		logger.Fatal().Msgf("Error in prompt: %q", err)
	}
	return val
}

func fuzzySearchWithPrefixAnchor(itemsToSelect []string, linePrefix string) func(input string, index int) bool {
	return func(input string, index int) bool {
		role := itemsToSelect[index]

		if strings.HasPrefix(input, linePrefix) {
			return strings.HasPrefix(role, input)
		}

		if fuzzy.MatchFold(input, role) {
			return true
		}
		return false
	}
}
