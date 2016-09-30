package cmd

import (
	"bytes"
	"strings"

	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func HostCommand() cli.Command {
	return cli.Command{
		Name:      "hosts",
		ShortName: "host",
		Usage:     "Operations on hosts",
		Action:    defaultAction(hostLs),
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List hosts",
				Description: "\nLists all hosts in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher hosts ls\n\t$ rancher --env 1a5 hosts ls\n",
				ArgsUsage:   "None",
				Action:      hostLs,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "quiet,q",
						Usage: "Only display IDs",
					},
					cli.StringFlag{
						Name:  "format",
						Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
					},
				},
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
	ID          string
	Host        client.Host
	State       string
	IPAddresses []client.IpAddress
	Labels      string
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

	collection, err := c.Host.List(nil)
	if err != nil {
		return err
	}

	knownMachines := map[string]bool{}
	for _, host := range collection.Data {
		knownMachines[host.PhysicalHostId] = true
	}

	machineCollection, err := c.Machine.List(nil)
	if err != nil {
		return err
	}

	for _, machine := range machineCollection.Data {
		if knownMachines[machine.Id] {
			continue
		}
		host := client.Host{
			Resource: client.Resource{
				Id: machine.Id,
			},
			Hostname:             machine.Name,
			State:                machine.State,
			TransitioningMessage: machine.TransitioningMessage,
		}
		if machine.State == "active" {
			host.State = "waiting"
			host.TransitioningMessage = "Almost there... Waiting for agent connection"
		}
		collection.Data = append(collection.Data, host)
	}

	writer := NewTableWriter([][]string{
		{"ID", "Host.Id"},
		{"HOSTNAME", "Host.Hostname"},
		{"STATE", "State"},
		{"IP", "{{ips .IPAddresses}}"},
		{"LABELS", "Labels"},
		{"DETAIL", "Host.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		ips := client.IpAddressCollection{}
		// ignore errors getting IPs, machines don't have them
		c.GetLink(item.Resource, "ipAddresses", &ips)

		writer.Write(&HostsData{
			ID:          item.Id,
			Host:        item,
			State:       getHostState(&item),
			Labels:      getLabels(&item),
			IPAddresses: ips.Data,
		})
	}

	return writer.Err()
}
