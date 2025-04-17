package client_templates

import (
	"embed"
	"text/template"
	"io"
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
	baseTemplates := []string{
		"templates/layout/base.html",
		"templates/components/header.html",
		"templates/components/footer.html",
		"templates/components/cookies.html",
		"templates/components/meta.html",
	}

	t.index = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, "templates/pages/index.html")...),
	)

	return nil
}

func (t *Templates) RenderIndex(w io.Writer, p IndexParams) error {
	p.View = "index"

	return t.index.Execute(w, p)
}
