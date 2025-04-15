package client_templates

import (
	"embed"
	"text/template"
)

//go:embed templates/*
var files embed.FS

type BaseParams struct {
	Title string
	Description string
	View string
}

type IndexParams struct {
	BaseParams
	Error string
	Products []Product
}


type TemplateFunctions struct {
	StaticWithHash func(string) string
}

type Templates struct {
	index *template.Template
	funcs template.FuncMap
}

func NewTemplates(tf TemplateFunctions) (*Templates, error) {
	t := &Templates{
		funcs: template.FuncMap{
			"staticWithHash": tf.StaticWithHash,
		},
	}

	if err := t.parseTemplates(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Templates) parseTemplates() error {
	t.index = template.Must(
		template.New("index.html").
	)
}
