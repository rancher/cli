//go:build windows

package cmd

import (
	"testing"

	"golang.org/x/sys/windows"
)

// TestOSOpenURLDoesNotShellInterpret guards the fix for AADSTS900144: it stubs
// the ShellExecute call and asserts osOpenURL hands the URL to the API
// opaquely — the full URL, metacharacters and all, as the target, with no
// command-line arguments — so there is no shell for a server-supplied
// authorize URL to be truncated by. The old `cmd /c start "" <url>` path built
// a command line that cmd.exe parsed, dropping everything after the first
// `&` (e.g. the OAuth `scope` parameter); this proves the current path does
// not.
func TestOSOpenURLDoesNotShellInterpret(t *testing.T) {
	const u = `https://host/authorize?client_id=x&scope=y&echo pwned> C:\Temp\pwned.txt`

	var gotVerb, gotFile, gotArgs *uint16
	orig := shellExecute
	shellExecute = func(_ windows.Handle, verb, file, args, _ *uint16, _ int32) error {
		gotVerb, gotFile, gotArgs = verb, file, args
		return nil
	}
	t.Cleanup(func() { shellExecute = orig })

	if err := osOpenURL(u); err != nil {
		t.Fatalf("osOpenURL: %v", err)
	}
	if got := windows.UTF16PtrToString(gotFile); got != u {
		t.Fatalf("URL not passed opaquely:\n got  %q\n want %q", got, u)
	}
	if gotArgs != nil {
		t.Fatalf("ShellExecute got command arguments %q — must be nil (no command line, no shell)",
			windows.UTF16PtrToString(gotArgs))
	}
	if v := windows.UTF16PtrToString(gotVerb); v != "open" {
		t.Fatalf("verb = %q, want \"open\"", v)
	}
}
