package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli"
)

func KubectlCommand() cli.Command {
	return cli.Command{
		Name:            "kubectl",
		Usage:           "Run kubectl commands",
		Description:     "Use the current kubectl context to run commands",
		Action:          runKubectl,
		SkipFlagParsing: true,
	}
}

func runKubectl(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "kubectl")
	}

	path, err := exec.LookPath("kubectl")
	if nil != err {
		return fmt.Errorf("kubectl is required to use this command: %s", err.Error())
	}

	cmd := exec.Command(path, ctx.Args()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if nil != err {
		return err
	}
	return nil
}
