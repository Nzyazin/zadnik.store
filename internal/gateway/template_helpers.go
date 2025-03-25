package gateway

import (
	"os"
	"strings"
)

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

func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, nil
	}

	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, nil
		}
		dict[key] = values[i+1]
	}

	return dict, nil
} 
