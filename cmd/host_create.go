package cmd

import (
	"bytes"
	"fmt"
	"os"
	"unicode"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func toEnv(name string) string {
	buf := bytes.Buffer{}
	for _, c := range name {
		if unicode.IsUpper(c) {
			buf.WriteRune('_')
			buf.WriteRune(unicode.ToLower(c))
		} else if c == '-' {
			buf.WriteRune('_')
		} else {
			buf.WriteRune(c)
		}
	}
	return strings.ToUpper(buf.String())
}

func toAPI(name string) string {
	buf := bytes.Buffer{}
	upper := false
	for _, c := range name {
		if c == '-' {
			upper = true
		} else if upper {
			upper = false
			buf.WriteRune(unicode.ToUpper(c))
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

func toArg(name string) string {
	buf := bytes.Buffer{}
	for _, c := range name {
		if unicode.IsUpper(c) {
			buf.WriteRune('-')
			buf.WriteRune(unicode.ToLower(c))
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

func buildFlag(name string, field client.Field) cli.Flag {
	var flag cli.Flag
	switch field.Type {
	case "bool":
		flag = cli.BoolFlag{
			Name:   toArg(name),
			EnvVar: toEnv(name),
			Usage:  field.Description,
		}
	case "array[string]":
		fallthrough
	case "map[string]":
		flag = cli.StringSliceFlag{
			Name:   toArg(name),
			EnvVar: toEnv(name),
			Usage:  field.Description,
		}
	default:
		sflag := cli.StringFlag{
			Name:   toArg(name),
			EnvVar: toEnv(name),
			Usage:  field.Description,
		}
		flag = sflag
		if field.Default != nil {
			sflag.Value = fmt.Sprint(field.Default)
		}
	}

	return flag
}

func buildFlags(prefix string, schema client.Schema, schemas *client.Schemas) []cli.Flag {
	flags := []cli.Flag{}
	for name, field := range schema.ResourceFields {
		if !field.Create || name == "name" {
			continue
		}

		if strings.HasSuffix(name, "Config") {
			subSchema := schemas.Schema(name)
			driver := strings.TrimSuffix(name, "Config")
			flags = append(flags, buildFlags(driver+"-", subSchema, schemas)...)
		} else {
			if prefix != "" {
				name = prefix + name
			}
			flags = append(flags, buildFlag(name, field))
		}
	}

	return flags
}

func hostCreate(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	hostSchema := c.GetSchemas().Schema("host")
	flags := buildFlags("", hostSchema, c.GetSchemas())
	drivers := []string{}

	for name := range hostSchema.ResourceFields {
		if strings.HasSuffix(name, "Config") {
			drivers = append(drivers, strings.TrimSuffix(name, "Config"))
		}
	}

	hostCommand := HostCommand()

	for i := range hostCommand.Subcommands {
		if hostCommand.Subcommands[i].Name == "create" {
			hostCommand.Subcommands[i].Flags = append(flags, cli.StringFlag{
				Name:   "driver,d",
				Usage:  "Driver to use: " + strings.Join(drivers, ", "),
				EnvVar: "MACHINE_DRIVER",
			})
			hostCommand.Subcommands[i].Action = func(ctx *cli.Context) error {
				return hostCreateRun(ctx, c, hostSchema, c.GetSchemas())
			}
			hostCommand.Subcommands[i].SkipFlagParsing = false
		}
	}

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		//TODO: remove duplication here
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug logging",
		},
		cli.StringFlag{
			Name:   "config,c",
			Usage:  "Client configuration file (default ${HOME}/.rancher/cli.json)",
			EnvVar: "RANCHER_CLIENT_CONFIG",
		},
		cli.StringFlag{
			Name:   "environment,env",
			Usage:  "Environment name or ID",
			EnvVar: "RANCHER_ENVIRONMENT",
		},
		cli.StringFlag{
			Name:   "url",
			Usage:  "Specify the Rancher API endpoint URL",
			EnvVar: "RANCHER_URL",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "Specify Rancher API access key",
			EnvVar: "RANCHER_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "Specify Rancher API secret key",
			EnvVar: "RANCHER_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "host",
			Usage:  "Host used for docker command",
			EnvVar: "RANCHER_DOCKER_HOST",
		},
		cli.StringFlag{
			Name:  "rancher-file,r",
			Usage: "Specify an alternate Rancher compose file (default: rancher-compose.yml)",
		},
		cli.StringFlag{
			Name:  "env-file,e",
			Usage: "Specify a file from which to read environment variables",
		},
		cli.StringSliceFlag{
			Name:   "file,f",
			Usage:  "Specify one or more alternate compose files (default: docker-compose.yml)",
			Value:  &cli.StringSlice{},
			EnvVar: "COMPOSE_FILE",
		},
		cli.StringFlag{
			Name:  "stack,s",
			Usage: "Specify an alternate project name (default: directory name)",
		},
		cli.BoolFlag{
			Name:  "wait,w",
			Usage: "Wait for resource to reach resting state",
		},
		cli.IntFlag{
			Name:  "wait-timeout",
			Usage: "Timeout in seconds to wait",
			Value: 600,
		},
		cli.StringFlag{
			Name:  "wait-state",
			Usage: "State to wait for (active, healthy, etc)",
		},
	}
	app.Commands = []cli.Command{
		hostCommand,
	}
	return app.Run(os.Args)
}

func hostCreateRun(ctx *cli.Context, c *client.RancherClient, machineSchema client.Schema, schemas *client.Schemas) error {
	args := map[string]interface{}{}
	driverArgs := map[string]interface{}{}
	driver := ctx.String("driver")

	if driver == "" {
		return fmt.Errorf("--driver is required")
	}

	driverSchema, ok := schemas.CheckSchema(driver + "Config")
	if !ok {
		return fmt.Errorf("Invalid driver: %s", driver)
	}

	for _, name := range ctx.FlagNames() {
		schema := machineSchema
		destArgs := args
		key := name
		value := ctx.Generic(name)

		// really dumb way to detect empty values
		if str := fmt.Sprint(value); str == "" || str == "[]" {
			continue
		}

		if strings.HasPrefix(name, driver+"-") {
			key = toAPI(strings.TrimPrefix(name, driver+"-"))
			schema = driverSchema
			destArgs = driverArgs
		}

		fieldType := schema.ResourceFields[key].Type
		if fieldType == "map[string]" {
			mapValue := map[string]string{}
			for _, val := range ctx.StringSlice(name) {
				parts := strings.SplitN(val, "=", 2)
				if len(parts) == 1 {
					mapValue[parts[0]] = ""
				} else {
					mapValue[parts[0]] = parts[1]
				}
			}
			value = mapValue
		}

		destArgs[key] = value
	}

	args[driver+"Config"] = driverArgs

	names := ctx.Args()
	if len(names) == 0 {
		names = []string{RandomName()}
	}

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		args["name"] = name
		var machine client.Machine
		if err := c.Create("machine", args, &machine); err != nil {
			lastErr = err
			logrus.Error(err)
		} else {
			w.Add(machine.Id)
		}
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait()
}
