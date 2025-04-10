package picker

import (
	"bytes"
	"text/template"
)

const pluginTemplate = `package {{.Package}}

import (
    "host"
{{- range .Import }}
    "{{.}}"
{{- end }}
)

func Fn{{.Name}}() func(ctx context.Context, args interface{}) error {
	{{- range .Define }}
	{{.}}
	{{- end }}
	return {{.Function}}
}

func Register(ct host.IContainer) error {
	return ct.Register("{{.Name}}", Fn{{.Name}}())
}
`

var (
	tplt = template.Must(template.New("plugin.template").Parse(pluginTemplate))
)

func buildPluginCode(args *pluginTpltArgs) (string, error) {
	buf := bytes.Buffer{}
	if err := tplt.Execute(&buf, args); err != nil {
		return "", err
	}
	return buf.String(), nil
}
