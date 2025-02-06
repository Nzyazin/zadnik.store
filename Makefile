# Имя бинарного файла
BINARY_NAME=app.exe

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
	cd web/frontend-admin && npm install

.PHONY: build-frontend
build-frontend:
	cd web/frontend-admin && npm run build

.PHONY: dev-frontend
dev-frontend:
	cd web/frontend-admin && npm run dev

# Static files
.PHONY: setup-static
setup-static: build-frontend
	@echo "==> Setting up static files..."
	@rm -rf bin/static
	@mkdir -p bin/static/js
	@mkdir -p bin/static/css
	@mkdir -p bin/static/images
	@mkdir -p bin/static/fonts
	@cp web/frontend-admin/build/statics/scripts/*.js bin/static/js/
	@cp web/frontend-admin/build/statics/styles/*.css bin/static/css/
	@cp web/frontend-admin/build/statics/fonts/* bin/static/fonts/
	@cp -r web/frontend-admin/build/statics/images/* bin/static/images/
	@git rev-parse --short HEAD > bin/static/hash.txt

# Build commands
.PHONY: build-auth
build-auth:
	@echo "==> Building auth service..."
	@go build -o bin/auth.exe ./cmd/auth

.PHONY: build-gateway
build-gateway: setup-static
	@echo "==> Building gateway service..."
	@go build -o bin/gateway.exe ./cmd/gateway

.PHONY: build
build: build-auth build-gateway

# Gateway
.PHONY: run-gateway
run-gateway:
	@echo "==> Starting gateway service..."
	go run ./cmd/gateway/main.go

# Run commands
.PHONY: run-services
run-services:
	@echo "==> Starting auth service..."
	@./bin/auth.exe &
	@echo "==> Waiting for auth service to start..."
	@sleep 2
	@echo "==> Starting gateway service..."
	@./bin/gateway.exe

# Combined
.PHONY: run-all
run-all:
	@echo "==> Starting all services..."
	make run-auth & make run-gateway

.PHONY: proto migrate-up migrate-down migrate-create create-db drop-db run-auth install-frontend build-frontend dev-frontend run-gateway run-all run-services