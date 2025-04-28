# Zadnik Store

### Backend

- Go 1.22.1
- gRPC
- PostgreSQL
- RabbitMQ
- Docker

### Frontend

- HTML/CSS/JavaScript
- Go template

### Structure of project

```
zadnik.store/
├── api/                    # gRPC API определения и сгенерированный код
├── bin/                    # Скомпилированные бинарные файлы
├── cmd/                    # Точки входа приложения
│   ├── auth/               # Микросервис авторизации
│   ├── gateway/            # API Gateway
│   ├── image/              # Микросервис изображений
│   └── product/            # Микросервис товаров
├── deployments/            # Конфигурации для развертывания
├── internal/               # Внутренняя логика приложения
├── static/                 # Собранная статика
│   ├── admin/             # Статика админ-панели
│   │   ├── css/          # Стили
│   │   ├── js/           # Скрипты
│   │   ├── images/       # Изображения
│   │   ├── fonts/        # Шрифты
│   │   └── templates/    # HTML шаблоны
│   └── client/           # Статика клиентской части
│       ├── css/
│       ├── js/
│       ├── images/
│       ├── fonts/
│       └── templates/
├── public/                # Публичная статика
│   ├── admin/             # Статика админ-панели
│   │   ├── css/          # Стили
│   │   ├── js/           # Скрипты
│   │   ├── images/       # Изображения
│   │   ├── fonts/        # Шрифты
│   │   └── templates/    # HTML шаблоны
│   └── client/           # Статика клиентской части
│       ├── css/
│       ├── js/
│       ├── images/
│       ├── fonts/
│       └── templates/
├── logs/                  # Логи приложения
│   └── app.log           # Файл логов
└── web/                   # Исходники фронтенда
    ├── html-css-js-admin/ # Исходники админ-панели
    │   ├── assets/       # Исходные файлы
    │   │   ├── fonts/
    │   │   ├── images/
    │   │   ├── scripts/
    │   │   ├── styles/
    │   │   └── views/
    │   └── tasks/        # Gulp задачи
    └── html-css-js-client/ # Исходники клиентской части
        ├── assets/
        │   ├── fonts/
        │   ├── images/
        │   ├── scripts/
        │   ├── styles/
        │   └── views/
        └── tasks/
```

### Basic usage

```bash
$ migrate -source file://path/to/migrations -database postgres://localhost:5432/database up 2

### Requirements

- Go 1.22.1+
- Docker
- RabbitMQ
- PostgreSQL

### Frontend build
```bash
# Build admin panel
cd web/html-css-js-admin
npm install
npm run build  # Соберет в static/admin/

# Сборка клиентской части
cd web/html-css-js-client
npm install
npm run build  # Соберет в static/client/
```

### Admin panel usage
- 👤 Авторизация пользователей и администраторов
- 📦 Управление товарами
  - Просмотр списка товаров
  - Добавление новых товаров
  - Редактирование существующих товаров
- 🖼️ Управление изображениями
- Обработка заявок
- 🎨 Современный адаптивный дизайн
- 🚀 Микросервисная архитектура для масштабируемости

## 🧪 Testing

В проекте используются два типа тестов (в настоящее время реализованы только для микросервиса авторизации):

### Unit Tests
Проверяют отдельные компоненты в изоляции, используя моки для имитации зависимостей.
Находятся в директориях `internal/auth/usecase` и проверяют бизнес-логику сервисов.

### Integration Tests
Проверяют взаимодействие с реальными внешними системами (PostgreSQL).
Находятся в директориях `internal/auth/repository/postgres` и требуют настройки тестовой базы данных `auth_test`.

```bash
# Запуск всех тестов
make test

# Запуск тестов для конкретного микросервиса
make test-auth
make test-auth-unit      # Только модульные тесты авторизации
make test-auth-integration # Только интеграционные тесты авторизации


## 📊 Monitoring

Проект интегрирован с Prometheus для мониторинга производительности и состояния микросервисов.

## 📜 Migrations

Для управления миграциями базы данных используется утилита [golang-migrate](https://github.com/golang-migrate/migrate). Каждая миграция состоит из двух файлов: прямой (up) и обратной (down) миграции.

```bash
# Формат файлов миграций
1481574547_create_users_table.up.sql   
1481574547_create_users_table.down.sql
```

### Commands for working with migrations

```bash
# Create database
make create-db SERVICE=auth    
make create-db SERVICE=product

# Удаление базы данных
make drop-db SERVICE=auth     
make drop-db SERVICE=product 

# Создание новой миграции
make migrate-create SERVICE=auth   
make migrate-create SERVICE=product

# Применение миграций
make migrate-up SERVICE=auth   
make migrate-up SERVICE=product

# Откат миграций
make migrate-down SERVICE=auth   
make migrate-down SERVICE=product

# Сброс состояния миграций
make migrate-clean SERVICE=auth   
make migrate-clean SERVICE=product 

# Установка конкретной версии миграций
make migrate-force SERVICE=auth VERSION=1
```

---
