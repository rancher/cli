package cmd

import (
	"bytes"
	"strings"

	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func HostCommand() cli.Command {
	hostLsFlags := []cli.Flag{
		listAllFlag(),
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: '{{.ID}} {{.Host.Hostname}}'",
		},
	}

	return cli.Command{
		Name:      "hosts",
		ShortName: "host",
		Usage:     "Operations on hosts",
		Action:    defaultAction(hostLs),
		Flags:     hostLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List hosts",
				Description: "\nLists all hosts in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher hosts ls\n\t$ rancher --env 1a5 hosts ls\n",
				ArgsUsage:   "None",
				Action:      hostLs,
				Flags:       hostLsFlags,
			},
			cli.Command{
				Name:            "create",
				Usage:           "Create a host",
				Description:     "\nCreates a host in the $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher --env k8slab host create newHostName\n",
				ArgsUsage:       "[NEWHOSTNAME...]",
				SkipFlagParsing: true,
				Action:          hostCreate,
			},
		},
	}
}

type HostsData struct {
	ID             string
	Host           client.Host
	State          string
	ContainerCount int
	Labels         string
}

func getHostState(host *client.Host) string {
	state := host.State
	if state == "active" && host.AgentState != "" {
		state = host.AgentState
	}
	return state
}

func getLabels(host *client.Host) string {
	var buffer bytes.Buffer
	it := 0
	for key, value := range host.Labels {
		if strings.HasPrefix(key, "io.rancher") {
			continue
		} else if it > 0 {
			buffer.WriteString(",")
		}

		buffer.WriteString(key)
		buffer.WriteString("=")
		buffer.WriteString(value.(string))
		it++
	}
	return buffer.String()
}

func hostLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Host.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	collectiondata := collection.Data

	for {
		collection, _ = collection.Next()
		if collection == nil {
			break
		}
		collectiondata = append(collectiondata, collection.Data...)
		if !collection.Pagination.Partial {
			break
		}
	}

	writer := NewTableWriter([][]string{
		{"ID", "Host.Id"},
		{"HOSTNAME", "Host.Hostname"},
		{"STATE", "State"},
		{"CONTAINERS", "ContainerCount"},
		{"IP", "Host.AgentIpAddress"},
		{"LABELS", "Labels"},
		{"DETAIL", "Host.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collectiondata {
		writer.Write(&HostsData{
			ID:             item.Id,
			Host:           item,
			State:          getHostState(&item),
			ContainerCount: len(item.InstanceIds),
			Labels:         getLabels(&item),
		})
	}

	return writer.Err()
}
