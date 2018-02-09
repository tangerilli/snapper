package snapper

import (
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
)

func LaunchChrome(path *string) (*exec.Cmd, io.ReadCloser, error) {
	var chromePath string
	args := []string{"--headless", "--remote-debugging-port=9222"}
	if path == nil || *path == "" {
		if runtime.GOOS == "darwin" {
			chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		} else {
			var err error
			chromePath, err = filepath.Abs("./headless-chrome/headless_shell")
			if err != nil {
				log.Printf("Could not resolve chrome path: %s\n", err)
				return nil, nil, err
			}
			args = append(args, "--window-size=1280x1696", "--no-sandbox", "--user-data-dir=/tmp/user-data",
				"--homedir=/tmp", "--disk-cache-dir=/tmp/cache-dir", "--data-path=/tmp/data-path", "--single-process",
				"--disable-gpu", "--enable-logging")
		}
	} else {
		chromePath = *path
	}
	log.Printf("Launching %s %s\n", chromePath, args)
	cmd := exec.Command(chromePath, args...)

	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		log.Printf("Error starting chrome: %s\n", err)
		return nil, nil, err
	}

	return cmd, stdout, nil
}
