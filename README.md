[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/golang-migrate/migrate/ci.yaml?branch=master)](https://github.com/golang-migrate/migrate/actions/workflows/ci.yaml?query=branch%3Amaster)
[![GoDoc](https://pkg.go.dev/badge/github.com/golang-migrate/migrate)](https://pkg.go.dev/github.com/golang-migrate/migrate/v4)
[![Coverage Status](https://img.shields.io/coveralls/github/golang-migrate/migrate/master.svg)](https://coveralls.io/github/golang-migrate/migrate?branch=master)
[![packagecloud.io](https://img.shields.io/badge/deb-packagecloud.io-844fec.svg)](https://packagecloud.io/golang-migrate/migrate?filter=debs)
[![Docker Pulls](https://img.shields.io/docker/pulls/migrate/migrate.svg)](https://hub.docker.com/r/migrate/migrate/)
![Supported Go Versions](https://img.shields.io/badge/Go-1.21%2C%201.22-lightgrey.svg)
[![GitHub Release](https://img.shields.io/github/release/golang-migrate/migrate.svg)](https://github.com/golang-migrate/migrate/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-migrate/migrate/v4)](https://goreportcard.com/report/github.com/golang-migrate/migrate/v4)


Современная микросервисная система управления интернет-магазином с административной панелью.
>>>>>>> 912fe5004ac42482192cf2bfa8b3878caba4172e

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
>>>>>>> 912fe5004ac42482192cf2bfa8b3878caba4172e

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
>>>>>>> 912fe5004ac42482192cf2bfa8b3878caba4172e

* [PostgreSQL](database/postgres)
* [PGX v4](database/pgx)
* [PGX v5](database/pgx/v5)
* [Redshift](database/redshift)
* [Ql](database/ql)
* [Cassandra / ScyllaDB](database/cassandra)
* [SQLite](database/sqlite)
* [SQLite3](database/sqlite3) ([todo #165](https://github.com/mattes/migrate/issues/165))
* [SQLCipher](database/sqlcipher)
* [MySQL / MariaDB](database/mysql)
* [Neo4j](database/neo4j)
* [MongoDB](database/mongodb)
* [CrateDB](database/crate) ([todo #170](https://github.com/mattes/migrate/issues/170))
* [Shell](database/shell) ([todo #171](https://github.com/mattes/migrate/issues/171))
* [Google Cloud Spanner](database/spanner)
* [CockroachDB](database/cockroachdb)
* [YugabyteDB](database/yugabytedb)
* [ClickHouse](database/clickhouse)
* [Firebird](database/firebird)
* [MS SQL Server](database/sqlserver)
* [rqlite](database/rqlite)
* [Add a new source?](source/driver.go)

* [Filesystem](source/file) - read from filesystem
* [io/fs](source/iofs) - read from a Go [io/fs](https://pkg.go.dev/io/fs#FS)
* [Go-Bindata](source/go_bindata) - read from embedded binary data ([jteeuwen/go-bindata](https://github.com/jteeuwen/go-bindata))
* [pkger](source/pkger) - read from embedded binary data ([markbates/pkger](https://github.com/markbates/pkger))
* [GitHub](source/github) - read from remote GitHub repositories
* [GitHub Enterprise](source/github_ee) - read from remote GitHub Enterprise repositories
* [Bitbucket](source/bitbucket) - read from remote Bitbucket repositories
* [Gitlab](source/gitlab) - read from remote Gitlab repositories
* [AWS S3](source/aws_s3) - read from Amazon Web Services S3
* [Google Cloud Storage](source/google_cloud_storage) - read from Google Cloud Platform Storage

## CLI usage

* Simple wrapper around this library.
* Handles ctrl+c (SIGINT) gracefully.
* No config search paths, no config files, no magic ENV var injections.

__[CLI Documentation](cmd/migrate)__

### Basic usage

```bash
$ migrate -source file://path/to/migrations -database postgres://localhost:5432/database up 2
======
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
>>>>>>> 912fe5004ac42482192cf2bfa8b3878caba4172e
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
>>>>>>> 912fe5004ac42482192cf2bfa8b3878caba4172e

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

MIT License © 2024 Zadnik.Store
>>>>>>> 912fe5004ac42482192cf2bfa8b3878caba4172e
