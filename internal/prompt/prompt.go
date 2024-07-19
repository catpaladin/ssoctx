// Package prompt contains all prompt functionality
package prompt

import (
	"log"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
)

// Input is used to interface with promptui
type Input interface {
	Select(label string, toSelect []string, searcher func(input string, index int) bool) (index int, value string)
	Enter(label string, dfault string) string
}

// Prompter is used to interface promptui
type Prompter struct{}

// Select is used to select from the prompt
func (p Prompter) Select(label string, toSelect []string, searcher func(input string, index int) bool) (int, string) {
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
		log.Fatalf("Error in prompt: %q", err)
	}
	return index, value
}

// Enter is used to pop open a selectable prompt
func (p Prompter) Enter(label string, dfault string) string {
	prompt := promptui.Prompt{
		Label:     label,
		Default:   dfault,
		AllowEdit: false,
	}
	val, err := prompt.Run()
	if err != nil {
		log.Fatalf("Error in prompt: %q", err)
	}
	return val
}

// AwsRegions contains selectable regions
var AwsRegions = []string{
	"us-east-2",
	"us-east-1",
	"us-west-1",
	"us-west-2",
	"af-south-1",
	"ap-east-1",
	"ap-south-1",
	"ap-northeast-3",
	"ap-northeast-2",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-northeast-1",
	"ca-central-1",
	"eu-central-1",
	"eu-west-1",
	"eu-west-2",
	"eu-south-1",
	"eu-west-3",
	"eu-north-1",
	"me-south-1",
	"sa-east-1",
}

// GetRegion is used to setup a prompt for region
func GetRegion(prompt Input) string {
	_, region := prompt.Select("Select your AWS Region", AwsRegions, func(input string, index int) bool {
		target := AwsRegions[index]
		return fuzzy.MatchFold(input, target)
	})
	return region
}
