package picker

import (
	"bytes"
	"text/template"
)

const pluginTemplate = `package {{.Package}}

import (
{{- range .Import }}
    "{{.}}"
{{- end }}
)

func Fn{{.Name}}() interface{} {
	{{ .Define }}

	return {{.Function}}
}

func Register_(ct host.IContainer) error {
	return ct.Register("{{.Name}}", Fn{{.Name}}())
}
`

var (
	tplt = template.Must(template.New("plugin.template").Parse(pluginTemplate))
)

func buildCode(args *pluginTpltArgs) (string, error) {
	buf := bytes.Buffer{}
	if err := tplt.Execute(&buf, args); err != nil {
		return "", err
	}
	return buf.String(), nil
}
