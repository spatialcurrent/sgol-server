package sgol

import (
	"bytes"
	"strings"
	"text/template"
)

func RenderTemplate(template_text string, ctx map[string]string) (string, error) {
	templateFunctions := template.FuncMap{
		"lower": func(value string) string {
			return strings.ToLower(value)
		},
		"upper": func(value string) string {
			return strings.ToUpper(value)
		},
		"replace": func(old string, new string, value string) string {
			return strings.Replace(value, old, new, -1)
		},
	}
	tmpl, err := template.New("tmpl").Funcs(templateFunctions).Parse(template_text)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, "tmpl", ctx)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
