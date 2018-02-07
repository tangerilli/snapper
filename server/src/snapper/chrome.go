package snapper

import (
	"os/exec"
	"runtime"
)

// TODO: add some utility functions for starting and stopping chrome
// TODO: Use this from standalone.go depending on CLI options

func LaunchChrome(path *string) *exec.Cmd {
	var chromePath string
	if path == nil || *path == "" {
		// TODO: Have a different default path depending on linux or mac
		if runtime.GOOS == "darwin" {
			chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		} else {
			chromePath = "./headless-chrome/headless_shell"
		}
	} else {
		chromePath = *path
	}
	cmd := exec.Command(chromePath, "--headless", "--remote-debugging-port=9222")
	cmd.Start()
	return cmd
}
