// Package aws contains all the aws logic
package aws

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

var execCmd = exec.Command
var system = runtime.GOOS

// OpenURLInBrowser opens browser for supported runtimes
func OpenURLInBrowser(system, url string) {
	var err error

	switch system {
	case "linux":
		err = execCmd("xdg-open", url).Start()
	case "darwin":
		err = execCmd("open", url).Start()
	case "windows":
		err = execCmd("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = fmt.Errorf("could not open %s - unsupported platform. Please open the URL manually", url)
	}
	if err != nil {
		log.Fatal(err)
	}
}
