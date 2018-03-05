package rancherprompt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Executor(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	if s == "exit" {
		os.Exit(0)
		return
	}
	//hack for rancher docker
	// docker --host 1h1 ps -> --host 1h1 docker ps
	if strings.HasPrefix(s, "docker ") {
		parts := strings.Split(s, " ")
		if len(parts) > 2 && (parts[1] == "--host" || parts[1] == "-host") {
			t := parts[0]
			parts[0] = parts[1]
			parts[1] = parts[2]
			parts[2] = t
			s = strings.Join(parts, " ")
		}
	}

	cmd := exec.Command("/bin/sh", "-c", "rancher "+s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Got error: %s\n", err.Error())
	}
	return
}
