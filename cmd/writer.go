package cmd

import (
	"encoding/json"
	"io"
	"os"
	"text/tabwriter"

	"github.com/ghodss/yaml"
	"github.com/urfave/cli"
)

type TableWriter struct {
	HeaderFormat  string
	ValueFormat   string
	err           error
	headerPrinted bool
	Writer        *tabwriter.Writer
}

type TableWriterConfig struct {
	Quiet  bool
	Format string
	Writer io.Writer
}

func NewTableWriter(values [][]string, ctx *cli.Context) *TableWriter {
	cfg := &TableWriterConfig{
		Writer: os.Stdout,
		Quiet:  ctx.Bool("quiet"),
		Format: ctx.String("format"),
	}

	return NewTableWriterWithConfig(values, cfg)
}

func NewTableWriterWithConfig(values [][]string, config *TableWriterConfig) *TableWriter {
	writer := config.Writer
	if writer == nil {
		writer = os.Stdout
	}

	t := &TableWriter{
		Writer: tabwriter.NewWriter(writer, 10, 1, 3, ' ', 0),
	}
	t.HeaderFormat, t.ValueFormat = SimpleFormat(values)

	// remove headers if quiet or with a different format
	if config.Quiet || config.Format != "" {
		t.HeaderFormat = ""
	}

	// when quiet show only the ID
	if config.Quiet {
		t.ValueFormat = "{{.ID}}\n"
	}

	// check for custom formatting
	if config.Format != "" {
		customFormat := config.Format

		// add a newline for other custom formats
		if customFormat != "json" && customFormat != "yaml" {
			customFormat += "\n"
		}
		t.ValueFormat = customFormat
	}

	return t
}

func (t *TableWriter) Err() error {
	return t.err
}

func (t *TableWriter) writeHeader() {
	if t.HeaderFormat != "" && !t.headerPrinted {
		t.headerPrinted = true
		t.err = printTemplate(t.Writer, t.HeaderFormat, struct{}{})
		if t.err != nil {
			return
		}
	}
}

func (t *TableWriter) Write(obj interface{}) {
	if t.err != nil {
		return
	}

	t.writeHeader()
	if t.err != nil {
		return
	}

	if t.ValueFormat == "json" {
		content, err := json.Marshal(obj)
		t.err = err
		if t.err != nil {
			return
		}
		_, t.err = t.Writer.Write(append(content, byte('\n')))
	} else if t.ValueFormat == "yaml" {
		content, err := yaml.Marshal(obj)
		t.err = err
		if t.err != nil {
			return
		}
		_, t.err = t.Writer.Write(append(content, byte('\n')))
	} else {
		t.err = printTemplate(t.Writer, t.ValueFormat, obj)
	}
}

func (t *TableWriter) Close() error {
	if t.err != nil {
		return t.err
	}
	t.writeHeader()
	if t.err != nil {
		return t.err
	}
	return t.Writer.Flush()
}
