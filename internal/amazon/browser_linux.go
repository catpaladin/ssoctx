//go:build linux
// +build linux

package amazon

import "fmt"

func openURLInBrowser(url string) error {
	if err := execCmd("xdg-open", url).Start(); err != nil {
		return fmt.Errorf("unable to open browser: %w", err)
	}
	return nil
}
