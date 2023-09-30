// Package aws contains all the aws logic
package aws

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/rs/zerolog"
)

var (
	execCmd = exec.Command
	system  = runtime.GOOS
)

// openURLInBrowser opens browser for supported runtimes
func openURLInBrowser(ctx context.Context, system, url string) error {
	var err error
	logger := zerolog.Ctx(ctx)

	switch system {
	case "linux":
		err = execCmd("xdg-open", url).Start()
	case "darwin":
		err = execCmd("open", url).Start()
	case "windows":
		err = execCmd("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		logger.Debug().Msgf("Unable to open browser on platform: %s", system)
		err = fmt.Errorf("Could not open %s on unsupported platform. Please open the URL manually", url)
	}
	if err != nil {
		return err
	}
	return nil
}
