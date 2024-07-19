//go:build darwin
// +build darwin

package amazon

import "fmt"

func openURLInBrowser(url string) error {
	if err := execCmd("open", url).Start(); err != nil {
		return fmt.Errorf("unable to open browser: %w", err)
	}
	return nil
}
