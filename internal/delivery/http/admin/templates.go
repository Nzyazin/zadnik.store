package admin

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

type ProductsIndexParams struct {
	BaseParams
	Products []Product
}

// Parsed templates
var (
	authTemplate          = parse("templates/pages/auth.html")
	productsIndexTemplate = parse("templates/pages/products-index.html")
)

// Helper function to parse templates with layout
func parse(file string) *template.Template {
	return template.Must(
		template.New("base.html").
			Funcs(templateFuncs).
			ParseFS(files, "templates/layout/base.html", file),
	)
}

var staticHash string

func init() {
	// Читаем хеш из файла при инициализации
	hashBytes, err := os.ReadFile("bin/static/hash.txt")
	if err == nil {
		staticHash = strings.TrimSpace(string(hashBytes))
	}
}

const (
	staticBasePath    = "bin/static"
	staticUrlPrefix   = "/static/"
)

// StaticWithHash добавляет хеш к пути статического файла
func StaticWithHash(path string) string {
	dir := staticBasePath
	if strings.HasPrefix(path, staticUrlPrefix) {
		dir = "bin" + path[:strings.LastIndex(path, "/")]
	}

	// Получаем базовое имя файла без расширения
	base := path[strings.LastIndex(path, "/")+1:]
	ext := path[strings.LastIndex(path, "."):]
	name := base[:len(base)-len(ext)]

	// Читаем содержимое директории
	files, err := os.ReadDir(dir)
	if err != nil {
		return path
	}

	// Ищем файл с нужным префиксом
	prefix := name + "-"
	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), ext) {
			// Возвращаем путь с найденным файлом
			return path[:strings.LastIndex(path, "/")+1] + file.Name()
		}
	}

	return path
}

// Template functions
var templateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"staticWithHash": StaticWithHash,
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
