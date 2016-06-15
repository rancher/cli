package cmd

import (
	"github.com/codegangsta/cli"
	ui "github.com/gizak/termui"
	"github.com/rancher/cli/monitor"
	"github.com/rancher/go-rancher/client"
)

func TuiCommand() cli.Command {
	return cli.Command{
		Name:   "dashboard",
		Usage:  "TUI Dashboard",
		Action: dashboard,
	}
}

type Model struct {
	c *client.RancherClient
	m *monitor.Monitor
}

func NewModel(c *client.RancherClient) *Model {
	return nil
}

func dashboard(ctx *cli.Context) error {
	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	ls := ui.NewList()
	ls.Border = true
	ls.Items = []string{
		"[1] Downloading File 1",
		"", // == \newline
		"[2] Downloading File 2",
		"",
		"[3] Uploading File 3",
	}
	ls.Height = 1000

	// build layout
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, ls)))

	// calculate layout
	ui.Body.Align()

	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/timer/1s", func(e ui.Event) {
		ls.Items = append(ls.Items, "hi")
		ui.Render(ui.Body)
	})

	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	ui.Loop()
	return nil
}
