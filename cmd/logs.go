package cmd

import "github.com/codegangsta/cli"

func LogCommand() cli.Command {
	return cli.Command{
		Name:            "logs",
		Usage:           "Fetch the logs of a container",
		SkipFlagParsing: true,
		HideHelp:        true,
		Action:          logsCommand,
	}
}

func logsCommand(ctx *cli.Context) error {
	if isHelp(ctx.Args()) {
		return runDockerHelp("logs")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	args, hostId, _, err := selectContainer(c, ctx.Args())
	if err != nil {
		return err
	}

	return runDockerCommand(hostId, c, "exec", args)
}
