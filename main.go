package main

import (
	"aws-sso-util/cmd"
	"fmt"
	"runtime/debug"
)

func main() {
	var commit string
	buildInfo, _ := debug.ReadBuildInfo()
	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.revision" {
			commit = setting.Value
		}
	}
	fmt.Printf("Commit: %s\n", commit)
	cmd.Execute()
}
