package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const testDBConnString = "postgres://postgres:postgres@localhost:5432/auth_test?sslmode=disable"

// TestMain выполняется перед всеми тестами пакета
func TestMain(m *testing.M) {
	// Запускаем тесты
	code := m.Run()

	// Завершаем процесс с кодом выполнения тестов
	os.Exit(code)
}

// prepareTestDB подготавливает БД для конкретного теста
func prepareTestDB(t *testing.T) *sql.DB {
	// Инициализируем подключение к тестовой БД
	db, err := sql.Open("postgres", testDBConnString)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Регистрируем функцию очистки, которая закроет соединение после теста
	t.Cleanup(func() {
		db.Close()
	})

	// Очищаем таблицы перед каждым тестом
	tables := []string{"refresh_tokens", "users"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}

	return db
}
