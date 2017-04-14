package template

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

type ReleaseInfo struct {
	Version         string
	PreviousVersion string
}

func Apply(contents []byte, releaseInfo ReleaseInfo, variables map[string]string) ([]byte, error) {
	// Skip templating if contents begin with '# notemplating'
	trimmedContents := strings.TrimSpace(string(contents))
	if strings.HasPrefix(trimmedContents, "#notemplating") || strings.HasPrefix(trimmedContents, "# notemplating") {
		return contents, nil
	}

	t, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(string(contents))
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	t.Execute(&buf, map[string]interface{}{
		"Values":  variables,
		"Release": releaseInfo,
	})
	return buf.Bytes(), nil
}
