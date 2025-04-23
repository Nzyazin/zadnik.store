Современная микросервисная система управления интернет-магазином с административной панелью.

__Database migrations written in Go. Use as [CLI](#cli-usage) or import as [library](#use-in-your-go-project).__

### Бэкенд

- Go 1.22.1
- gRPC
- PostgreSQL
- RabbitMQ
- Docker

### Фронтенд

- HTML/CSS/JavaScript
- Шаблонизация Go

## Databases

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

## CLI usage

* Simple wrapper around this library.
* Handles ctrl+c (SIGINT) gracefully.
* No config search paths, no config files, no magic ENV var injections.

__[CLI Documentation](cmd/migrate)__

### Basic usage

```bash
$ migrate -source file://path/to/migrations -database postgres://localhost:5432/database up 2

### Требования

- Go 1.22.1+
- Docker
- Docker Compose

### Запуск инфраструктуры

```bash
# Запуск PostgreSQL и RabbitMQ
docker-compose up -d
```

### Сборка фронтенда
```bash
# Сборка админ-панели
cd web/html-css-js-admin
npm install
npm run build  # Соберет в static/admin/

# Сборка клиентской части
cd web/html-css-js-client
npm install
npm run build  # Соберет в static/client/
```

### Docker usage
- 👤 Авторизация пользователей и администраторов
- 📦 Управление товарами
  - Просмотр списка товаров
  - Добавление новых товаров
  - Редактирование существующих товаров
- 🖼️ Управление изображениями
- 🛒 Оформление заказов
- 🎨 Современный адаптивный дизайн
- 🚀 Микросервисная архитектура для масштабируемости

## 🧪 Тестирование

```bash
# Запуск всех тестов
make test

# Запуск тестов для конкретного микросервиса
make test-auth
make test-gateway
make test-image
make test-product
```

## 📊 Мониторинг

Проект интегрирован с Prometheus для мониторинга производительности и состояния микросервисов.

## Getting started

Go to [getting started](GETTING_STARTED.md)

## Tutorials

* [CockroachDB](database/cockroachdb/TUTORIAL.md)
* [PostgreSQL](database/postgres/TUTORIAL.md)

(more tutorials to come)

## Migration files

Each migration has an up and down migration. [Why?](FAQ.md#why-two-separate-files-up-and-down-for-a-migration)

```bash
1481574547_create_users_table.up.sql
1481574547_create_users_table.down.sql
```

[Best practices: How to write migrations.](MIGRATIONS.md)

## Coming from another db migration tool?

Check out [migradaptor](https://github.com/musinit/migradaptor/).
*Note: migradaptor is not affiliated or supported by this project*

## Versions

Version | Supported? | Import | Notes
--------|------------|--------|------
**master** | :white_check_mark: | `import "github.com/golang-migrate/migrate/v4"` | New features and bug fixes arrive here first |
**v4** | :white_check_mark: | `import "github.com/golang-migrate/migrate/v4"` | Used for stable releases |
**v3** | :x: | `import "github.com/golang-migrate/migrate"` (with package manager) or `import "gopkg.in/golang-migrate/migrate.v3"` (not recommended) | **DO NOT USE** - No longer supported |

## Development and Contributing

Yes, please! [`Makefile`](Makefile) is your friend,
read the [development guide](CONTRIBUTING.md).

Also have a look at the [FAQ](FAQ.md).

---

MIT License © 2025 Zadnik
