package admin_templates

import (
	"embed"
	"html/template"
	"io"
	"os"
	"strings"
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

type ProductEditParams struct {
	BaseParams
	Product Product
	Error string
}

type ProductsIndexParams struct {
	BaseParams
	Products []Product
	Error string
}

// TemplateFunctions содержит функции для использования в шаблонах
type TemplateFunctions struct {
	StaticWithHash func(string) string
	Add func(int, int) int
	Dict func(...interface{}) (map[string]interface{}, error)
}

// Templates хранит все шаблоны и их функции
type Templates struct {
	auth     *template.Template
	products *template.Template
	productEdit *template.Template
	funcs    template.FuncMap
}

// NewTemplates создает новый экземпляр Templates с переданными функциями
func NewTemplates(tf TemplateFunctions) (*Templates, error) {
	t := &Templates{
		funcs: template.FuncMap{
			"add":           tf.Add,
			"staticWithHash": tf.StaticWithHash,
			"dict":          tf.Dict,
		},
	}

	// Парсим шаблоны
	if err := t.parseTemplates(); err != nil {
		return nil, err
	}

	return t, nil
}

// parseTemplates парсит все шаблоны
func (t *Templates) parseTemplates() error {
	
	t.auth = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, "templates/layout/base.html", "templates/pages/auth.html"),
	)

	t.products = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, 
				"templates/layout/base.html", 
				"templates/pages/products-index.html",
			),
	)

	t.productEdit = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, 
				"templates/layout/base.html", 
				"templates/pages/product-edit.html",
				"templates/components/product-header.html",
				"templates/components/product-form.html",
			),
	)

	return nil
}

// Template render functions
func (t *Templates) RenderAuth(w io.Writer, p AuthParams) error {
	// Установим базовые параметры
	p.Title = "Вход в систему"
	p.View = "auth"
	p.HideHeader = true
	
	return t.auth.Execute(w, p)
}

func (t *Templates) RenderProductEdit(w io.Writer, p ProductEditParams) error {
	p.View = "product-edit"
	
	return t.productEdit.Execute(w, p)
}

func (t *Templates) RenderProductsIndex(w io.Writer, p ProductsIndexParams) error {
	// Установим базовые параметры
	p.View = "products-index"
	
	return t.products.Execute(w, p)
}

var staticHash string

func init() {
	// Читаем хеш из файла при инициализации
	hashBytes, err := os.ReadFile("bin/static/hash.txt")
	if err == nil {
		staticHash = strings.TrimSpace(string(hashBytes))
	}
}
