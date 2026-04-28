package cmd

import (
	"github.com/rancher/norman/types"
	"github.com/urfave/cli/v3"
)

func baseListOpts() *types.ListOpts {
	return &types.ListOpts{
		Filters: map[string]interface{}{
			"limit": -1,
			"all":   true,
		},
	}
}

func defaultListOpts(cmd *cli.Command) *types.ListOpts {
	listOpts := baseListOpts()
	if cmd != nil && !cmd.Bool("all") {
		listOpts.Filters["removed_null"] = "1"
		listOpts.Filters["state_ne"] = []string{
			"inactive",
			"stopped",
			"removing",
		}
		delete(listOpts.Filters, "all")
	}
	if cmd != nil && cmd.Bool("system") {
		delete(listOpts.Filters, "system")
	} else {
		listOpts.Filters["system"] = "false"
	}
	return listOpts
}
