package terminal

import "github.com/charmbracelet/huh"

func NewInputForm(title string) (string, error) {
	var output string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				Prompt("? ").
				Value(&output),
		),
	)

	err := form.Run()
	if err != nil {
		return output, err
	}
	return output, nil
}
