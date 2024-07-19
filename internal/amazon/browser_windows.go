//go:build windows
// +build windows

package amazon

import "fmt"

func openURLInBrowser(url string) error {
	if err := execCmd("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
		return fmt.Errorf("unable to open browser: %w", err)
	}
	return nil
}
