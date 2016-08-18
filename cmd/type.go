package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

func typesStringFlag(def []string) cli.StringSliceFlag {
	usage := "Restrict restart to specific types"
	if len(def) > 0 {
		usage = fmt.Sprintf("%s (%s)", usage, strings.Join(def, ", "))
	}
	return cli.StringSliceFlag{
		Name:  "type",
		Usage: usage,
	}
}

func getTypesStringFlag(ctx *cli.Context, def []string) []string {
	val := ctx.StringSlice("type")
	if len(val) > 0 {
		return val
	}
	return def
}
