package terminal

import "github.com/charmbracelet/huh/spinner"

func NewSpinner(title string, action func()) {
	spinner.New().Title(title).Action(action).Run()
}
