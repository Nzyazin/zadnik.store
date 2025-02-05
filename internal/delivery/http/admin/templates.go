package admin

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*
var files embed.FS

// Template parameters structs
type BaseParams struct {
	Title      string
	View       string
	HideHeader bool
}

type AuthParams struct {
	BaseParams
	Error string
}

type ProductsIndexParams struct {
	BaseParams
	Products []Product
}

// Parsed templates
var (
	authTemplate          = parse("pages/auth.html")
	productsIndexTemplate = parse("pages/products-index.html")
)

// Helper function to parse templates with layout
func parse(file string) *template.Template {
	return template.Must(
		template.New("base.html").
			Funcs(templateFuncs).
			ParseFS(files, "layout/base.html", file),
	)
}

// Template functions
var templateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
}

// Template render functions
func RenderAuth(w io.Writer, p AuthParams) error {
	// Установим базовые параметры
	p.Title = "Вход в систему"
	p.View = "auth"
	p.HideHeader = true
	
	return authTemplate.Execute(w, p)
}

func RenderProductsIndex(w io.Writer, p ProductsIndexParams) error {
	// Установим базовые параметры
	p.Title = "Товары"
	p.View = "products-index"
	
	return productsIndexTemplate.Execute(w, p)
}
