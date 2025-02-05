package gateway

import (
	"fmt"
	"os"
	"strings"
)

var staticHash string

func init() {
	// Читаем хеш из файла при инициализации
	hashBytes, err := os.ReadFile("bin/statics/hash.txt")
	if err == nil {
		staticHash = strings.TrimSpace(string(hashBytes))
	}
}

// StaticWithHash добавляет хеш к пути статического файла
func StaticWithHash(path string) string {
	if staticHash == "" {
		return path
	}
	return fmt.Sprintf("%s?hash=%s", path, staticHash)
}
