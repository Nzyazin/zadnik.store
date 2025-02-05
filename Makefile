# Имя бинарного файла
BINARY_NAME=app

# Переменные для protobuf
PROTO_DIR=api
PROTO_FILES=$(PROTO_DIR)/*.proto

proto:
	@echo "==> Generation protobuf..."
	@cd $(PROTO_DIR) && \
	echo "Generation protobuf - product" && \
	echo "Generation protobuf - auth" && \
	protoc --go_out=. --go-grpc_out=. *.proto

# Migration commands
migrate-up:
	@echo "==> Running migrations up..."
	@if [ "$(SERVICE)" = "auth" ]; then \
		set -a && . ./internal/auth/config/.env-auth && set +a && \
		migrate -path migrations/auth -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable" up; \
	elif [ "$(SERVICE)" = "product" ]; then \
		set -a && . ./internal/product/config/.env-product && set +a && \
		migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable" up; \
	else \
		echo "Please specify SERVICE=auth or SERVICE=product"; \
		exit 1; \
	fi

migrate-down:
	@echo "==> Running migrations down..."
	@if [ "$(SERVICE)" = "auth" ]; then \
		set -a && . ./internal/auth/config/.env-auth && set +a && \
		migrate -path migrations/auth -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable" down; \
	elif [ "$(SERVICE)" = "product" ]; then \
		set -a && . ./internal/product/config/.env-product && set +a && \
		migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable" down; \
	else \
		echo "Please specify SERVICE=auth or SERVICE=product"; \
		exit 1; \
	fi

migrate-create:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=auth or SERVICE=product"; \
		exit 1; \
	fi
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations/$(SERVICE) -seq $$name

# Database commands
create-db:
	@echo "==> Creating database..."
	@if [ "$(SERVICE)" = "auth" ]; then \
		set -a && . ./internal/auth/config/.env-auth && set +a && \
		PGPASSWORD=$$DB_PASSWORD psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER -d postgres -c "CREATE DATABASE $$DB_NAME;"; \
	elif [ "$(SERVICE)" = "product" ]; then \
		set -a && . ./internal/product/config/.env-product && set +a && \
		PGPASSWORD=$$DB_PASSWORD psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER -d postgres -c "CREATE DATABASE $$DB_NAME;"; \
	fi

drop-db:
	@echo "==> Dropping database..."
	@if [ "$(SERVICE)" = "auth" ]; then \
		set -a && . ./internal/auth/config/.env-auth && set +a && \
		PGPASSWORD=$$DB_PASSWORD psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$$DB_NAME';" && \
		PGPASSWORD=$$DB_PASSWORD psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER -d postgres -c "DROP DATABASE IF EXISTS $$DB_NAME;"; \
	elif [ "$(SERVICE)" = "product" ]; then \
		set -a && . ./internal/product/config/.env-product && set +a && \
		PGPASSWORD=$$DB_PASSWORD psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$$DB_NAME';" && \
		PGPASSWORD=$$DB_PASSWORD psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER -d postgres -c "DROP DATABASE IF EXISTS $$DB_NAME;"; \
	else \
		echo "Please specify SERVICE=auth or SERVICE=product"; \
		exit 1; \
	fi

# Service commands
run-auth:
	@echo "==> Starting auth service..."
	@set -a && . ./internal/auth/config/.env-auth && set +a && \
	go run ./cmd/auth/main.go

# Frontend
.PHONY: install-frontend
install-frontend:
	cd web && npm install

.PHONY: build-frontend
build-frontend:
	cd web && npm run build

.PHONY: dev-frontend
dev-frontend:
	cd web && npm run dev

# Static files
.PHONY: setup-static
setup-static:
	@echo "==> Setting up static files..."
	@mkdir -p bin/statics/js
	@mkdir -p bin/statics/css
	@mkdir -p bin/statics/images
	@cp -r web/html-css-js-admin/assets/js/* bin/statics/js/
	@cp -r web/html-css-js-admin/assets/css/* bin/statics/css/
	@cp -r web/html-css-js-admin/assets/images/* bin/statics/images/
	@git rev-parse --short HEAD > bin/statics/hash.txt

# Build commands
.PHONY: build
build: setup-static
	@echo "==> Building gateway..."
	@go build -o bin/gateway ./cmd/gateway
	@echo "==> Building auth service..."
	@go build -o bin/auth ./cmd/auth
	@echo "==> Building product service..."
	@go build -o bin/product ./cmd/product

# Gateway
.PHONY: run-gateway
run-gateway: setup-static
	@echo "==> Starting gateway service..."
	go run ./cmd/gateway/main.go

# Combined
.PHONY: run-all
run-all:
	@echo "==> Starting all services..."
	make run-auth & make run-gateway & make dev-frontend

.PHONY: proto migrate-up migrate-down migrate-create create-db drop-db run-auth install-frontend build-frontend dev-frontend run-gateway run-all