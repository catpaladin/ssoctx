package aws

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// OpenURLInBrowser opens browser for supported runtimes
func OpenURLInBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = fmt.Errorf("could not open %s - unsupported platform. Please open the URL manually", url)
	}
	if err != nil {
		log.Fatal(err)
	}
}
