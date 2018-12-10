package ui

import (
	"html/template"
	"strings"
)

func RenderIndex(remotes []Remote) (string, error) {
	tmpl, err := template.New("index").Parse(rawIndexTemplate)
	if err != nil {
		return "", err
	}
	buf := strings.Builder{}
	if err := tmpl.Execute(&buf, remotes); err != nil {
		return "", err
	}
	return buf.String(), nil
}
