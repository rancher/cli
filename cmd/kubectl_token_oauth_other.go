//go:build !windows

package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
)

// osOpenURL launches the platform's default URL handler with openURL as a
// single argv element (no shell), so a server-supplied authorize URL cannot
// inject shell metacharacters. openBrowser has already validated the scheme.
func osOpenURL(openURL string) error {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return exec.Command(cmd, openURL).Start()
}
