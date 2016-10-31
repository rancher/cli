package cmd

import (
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func listAllFlag() cli.BoolFlag {
	return cli.BoolFlag{
		Name:  "all,a",
		Usage: "Show stop/inactive and recently removed resources",
	}
}

func listSystemFlag() cli.BoolFlag {
	return cli.BoolFlag{
		Name:  "system,s",
		Usage: "Show system resources",
	}
}

func baseListOpts() *client.ListOpts {
	return &client.ListOpts{
		Filters: map[string]interface{}{
			"limit": -2,
			"all":   true,
		},
	}
}

func defaultListOpts(ctx *cli.Context) *client.ListOpts {
	listOpts := baseListOpts()
	if ctx != nil && !ctx.Bool("all") {
		listOpts.Filters["removed_null"] = "1"
		listOpts.Filters["state_ne"] = []string{
			"inactive",
			"stopped",
			"removing",
		}
		delete(listOpts.Filters, "all")
	}
	if ctx != nil && ctx.Bool("system") {
		delete(listOpts.Filters, "system")
	} else {
		listOpts.Filters["system"] = "false"
	}
	return listOpts
}
