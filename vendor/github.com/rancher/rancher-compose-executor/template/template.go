package template

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/rancher/rancher-compose-executor/template/funcs"
)

type StackInfo struct {
	Version         string
	PreviousVersion string
	Name            string
}

func Apply(contents []byte, stackInfo StackInfo, variables map[string]string) ([]byte, error) {
	// Skip templating if contents begin with '# notemplating'
	trimmedContents := strings.TrimSpace(string(contents))
	if strings.HasPrefix(trimmedContents, "#notemplating") || strings.HasPrefix(trimmedContents, "# notemplating") {
		return contents, nil
	}

	templateFuncs := sprig.TxtFuncMap()
	templateFuncs["splitPreserveQuotes"] = funcs.SplitPreserveQuotes

	t, err := template.New("template").Funcs(templateFuncs).Parse(string(contents))
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	t.Execute(&buf, map[string]interface{}{
		"Values":  variables,
		"Release": stackInfo,
		"Stack":   stackInfo,
	})
	return buf.Bytes(), nil
}
