//go:build windows

package cmd

import "golang.org/x/sys/windows"

// shellExecute is windows.ShellExecute, indirected so a test can capture the
// arguments — proving the URL is passed opaquely with no command line — without
// actually launching a browser.
var shellExecute = windows.ShellExecute

// osOpenURL opens openURL through the Windows ShellExecute API. ShellExecute
// takes the target as a string argument to the shell API: it builds no command
// line and spawns no cmd.exe, so a server-supplied authorize URL (which
// legitimately contains `&` between query parameters) cannot be truncated or
// have its metacharacters interpreted the way `cmd /c start` did. openBrowser
// has already validated the scheme.
func osOpenURL(openURL string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := windows.UTF16PtrFromString(openURL)
	if err != nil {
		return err
	}
	return shellExecute(0, verb, target, nil, nil, windows.SW_SHOWNORMAL)
}
