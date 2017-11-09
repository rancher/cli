package cmd

import (
	"encoding/json"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"
)

type TableWriter struct {
	quite         bool
	json          bool
	HeaderFormat  string
	ValueFormat   string
	err           error
	headerPrinted bool
	Writer        *tabwriter.Writer
}

func NewTableWriter(values [][]string, ctx *cli.Context) *TableWriter {
	t := &TableWriter{
		Writer: tabwriter.NewWriter(os.Stdout, 10, 1, 3, ' ', 0),
	}
	t.HeaderFormat, t.ValueFormat = SimpleFormat(values)

	if ctx.Bool("quiet") {
		t.HeaderFormat = ""
		t.ValueFormat = "{{.ID}}\n"
	}

	t.json = ctx.Bool("json")

	customFormat := ctx.String("format")
	if customFormat == "json" {
		t.HeaderFormat = ""
		t.ValueFormat = "json"
	} else if customFormat != "" {
		t.ValueFormat = customFormat + "\n"
		t.HeaderFormat = ""
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
	} else {
		if t.json {
      jsonContent, err := json.Marshal(obj)
      t.err = err
      if t.err != nil {
        return
      }

			var jsonObj interface{}
			if t.err = json.Unmarshal(jsonContent, &jsonObj); t.err != nil {
				return
			}
			t.err = printTemplate(t.Writer, t.ValueFormat, jsonObj)
		} else {
			t.err = printTemplate(t.Writer, t.ValueFormat, obj)
		}
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
