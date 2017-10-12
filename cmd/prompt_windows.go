package cmd

import (
	"fmt"
	"github.com/urfave/cli"
)

func promptWrapper(ctx *cli.Context) error {
	fmt.Println("Prompt mode is not supported in Windows currently")
	return nil
}
