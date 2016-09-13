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

func defaultListOpts(ctx *cli.Context) *client.ListOpts {
	listOpts := &client.ListOpts{
		Filters: map[string]interface{}{
			"limit": -2,
			"all":   true,
		},
	}
	if ctx != nil && !ctx.Bool("all") {
		listOpts.Filters["removed_null"] = "1"
		listOpts.Filters["state_ne"] = []string{
			"inactive",
			"stopped",
			"removing",
		}
		delete(listOpts.Filters, "all")
	}
	return listOpts
}
