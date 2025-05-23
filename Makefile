# Имя бинарного файла
BINARY_NAME=app.exe

# Переменные для protobuf
PROTO_DIR=api
PROTO_FILES=$(PROTO_DIR)/*.proto

.PHONY: build-product
build-product:
	@echo "==> Building product service..."
	@go build -o bin/product ./cmd/product

.PHONY: build-auth
build-auth:
	@echo "==> Building auth service..."
	@go build -o bin/auth ./cmd/auth

.PHONY: build-gateway
build-gateway:
	@echo "==> Building gateway service..."
	@go build -o bin/gateway ./cmd/gateway

.PHONY: build-image
build-image:
	@echo "==> Building image service..."
	@go build -o bin/image ./cmd/image

.PHONY: build-all
build-all: build-product build-auth build-gateway build-image

proto:
	@echo "==> Generation protobuf..."
	@cd $(PROTO_DIR) && \
	echo "Generation protobuf - auth" && \
	protoc --go_out=. --go-grpc_out=. *.proto

# Migration commands
migrate-up:	
	@if [ "$(SERVICE)" = "auth" ]; then \
		echo "==> Running migrations up 'auth'" && \
		set -a && . ./internal/auth/config/.env-auth && set +a && \
		migrate -path migrations/auth -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable" up; \
	elif [ "$(SERVICE)" = "product" ]; then \
		echo "==> Running migrations up 'product'" && \
		set -a && . ./internal/product/config/.env-product && set +a && \
		migrate -path migrations/product -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable" up; \
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
		migrate -path migrations/product -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=disable" down; \
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

migrate-clean:
	@if [ "$(SERVICE)" = "auth" ]; then \
		echo "==> Cleaning migrations state for 'auth'" && \
		PGPASSWORD=postgres psql -h localhost -U postgres -d auth -c "DROP TABLE IF EXISTS schema_migrations;" && \
		PGPASSWORD=postgres psql -h localhost -U postgres -d auth -c "DROP TABLE IF EXISTS users;" && \
		make migrate-up SERVICE=auth; \
	elif [ "$(SERVICE)" = "product" ]; then \
		echo "==> Cleaning migrations state for 'product'" && \
		PGPASSWORD=postgres psql -h localhost -U postgres -d product -c "DROP TABLE IF EXISTS schema_migrations;" && \
		PGPASSWORD=postgres psql -h localhost -U postgres -d product -c "DROP TABLE IF EXISTS products;" && \
		make migrate-up SERVICE=product; \
	else \
		echo "Please specify SERVICE=auth or SERVICE=product"; \
		exit 1; \
	fi

migrate-force:
	@echo "==> Forcing migration version..."
	migrate -database "postgres://postgres:postgres@localhost:5432/$(SERVICE)?sslmode=disable" \
		-path migrations/$(SERVICE) force $(VERSION)

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
install-frontend-admin:
	cd web/frontend-admin && npm install

.PHONY: install-frontend-client
install-frontend-client:
	cd web/frontend-client && npm install

.PHONY: build-frontend
build-frontend-admin:
	cd web/frontend-admin && npm run build

.PHONY: build-frontend-client
build-frontend-client:
	cd web/frontend-client && npm run build

.PHONY: dev-frontend
dev-frontend-admin:
	cd web/frontend-admin && npm run dev

.PHONY: dev-frontend-client
dev-frontend-client:
	cd web/frontend-client && npm run dev

.PHONY: setup-static-client
setup-static-client: build-frontend-client
	@echo "==> Setting up client static files..."
	@rm -rf bin/static/client/
	@mkdir -p bin/static/client/js
	@mkdir -p bin/static/client/css
	@mkdir -p bin/static/client/images
	@mkdir -p bin/static/client/fonts
	@cp web/frontend-client/build/statics/scripts/*.js bin/static/client/js/
	@cp web/frontend-client/build/statics/styles/*.css bin/static/client/css/
	@cp web/frontend-client/build/statics/fonts/* bin/static/client/fonts/
	@cp -r web/frontend-client/build/statics/images/* bin/static/client/images/
	@git rev-parse --short HEAD > bin/static/client/hash.txt

# Static files
.PHONY: setup-static-admin
setup-static-admin: build-frontend-admin
	@echo "==> Setting up admin static files..."
	@rm -rf bin/static/admin/
	@mkdir -p bin/static/admin/js
	@mkdir -p bin/static/admin/css
	@mkdir -p bin/static/admin/images
	@mkdir -p bin/static/admin/fonts
	@cp web/frontend-admin/build/statics/scripts/*.js bin/static/admin/js/
	@cp web/frontend-admin/build/statics/styles/*.css bin/static/admin/css/
	@cp web/frontend-admin/build/statics/fonts/* bin/static/admin/fonts/
	@cp -r web/frontend-admin/build/statics/images/* bin/static/admin/images/
	@git rev-parse --short HEAD > bin/static/admin/hash.txt

# Gateway
.PHONY: run-gateway
run-gateway:
	@echo "==> Starting gateway service..."
	go run ./cmd/gateway/main.go

.PHONY: run-product
run-product:
	@echo "==> Starting product service..."
	go run ./cmd/product/main.go

.PHONY: run-image
run-image:
	@echo "==> Starting image service..."
	go run ./cmd/image/main.go

# RabbitMQ commands
.PHONY: rabbitmq-start rabbitmq-stop rabbitmq-restart rabbitmq-status

setcap:
	@echo "==> Setting capabilities for gateway binary..."
	sudo setcap cap_net_bind_service=+ep ./bin/gateway

rabbitmq-start:
	@echo "==> Starting RabbitMQ..."
	docker compose up -d rabbitmq

rabbitmq-stop:
	@echo "==> Stopping RabbitMQ..."
	docker compose stop rabbitmq
	docker compose rm -f rabbitmq

rabbitmq-restart: rabbitmq-stop rabbitmq-start

rabbitmq-status:
	@echo "==> RabbitMQ status:"
	@docker ps -f name=rabbitmq

# Combined
.PHONY: run-all
run-all:
	@echo "==> Starting all services..."
	make run-auth & make run-gateway & make run-product & make run-image

.PHONY: check-ports
check-ports:
	@echo "==> Checking service ports..."
	@echo "Auth Service (50051):"
	@-lsof -i :50051 || echo "Port available"
	@echo "\nProduct Service (50055):"
	@-lsof -i :50055 || echo "Port available"
	@echo "\nGateway Service (8082):"
	@-lsof -i :8082 || echo "Port available"
	@echo "\nImage Service (8084):"
	@-lsof -i :8084 || echo "Port available"

.PHONY: check-services
check-services:
	@echo "==> Checking running services..."
	@echo "Auth service:"
	@ps aux | grep -E "bin/auth" | grep -v grep || echo "Not running"
	@echo "\nProduct service:"
	@ps aux | grep -E "bin/product" | grep -v grep || echo "Not running"
	@echo "\nGateway service:"
	@ps aux | grep -E "bin/gateway" | grep -v grep || echo "Not running"
	@echo "\nImage service:"
	@ps aux | grep -E "bin/image" | grep -v grep || echo "Not running"
	@echo "\nListening ports:"
	@echo "Auth (50051):"
	@lsof -i :50051 || echo "Port not in use"
	@echo "\nProduct (50055):"
	@lsof -i :50055 || echo "Port not in use"
	@echo "\nGateway (HTTP):"
	@lsof -i :80 || lsof -i :8082 || echo "HTTP port not in use"
	@echo "\nGateway (HTTPS):"
	@lsof -i :443 || lsof -i :8443 || echo "HTTPS port not in use"
	@echo "\nImage service:"
	@lsof -i :8084 || echo "Port not in use"

.PHONY: start-all-bin
start-all-bin:
	@echo "==> Starting all binaries in background..."
	@./bin/auth &
	@./bin/product &
	@./bin/gateway &
	@./bin/image &

.PHONY: stop-all
stop-all:
	@echo "==> Stopping all services..."
	@-pkill -f "cmd/auth/main"
	@-pkill -f "cmd/gateway/main"
	@-pkill -f "cmd/product/main"
	@-pkill -f "cmd/image/main"
	@echo "All services stopped"


.PHONY: stop-all-bin
stop-all-bin:
	@echo "==> Stopping all binaries..."
	@ps aux | grep -E 'auth|product|gateway|image' | grep -v grep | awk '{print $$2}' | xargs kill -9 || true
	@echo "All binaries stopped"

.PHONY: generate-mocks
generate-mocks: install-mockgen
	@echo "==> Generating mocks..."
	@mockgen -source=internal/auth/domain/user.go -destination=internal/auth/mocks/mock_repositories.go -package=mocks

# Test commands
.PHONY: test test-auth test-auth-unit test-auth-integration test-gateway test-image test-product

test:
	@echo "==> Running all tests..."
	go test ./... -v

test-auth: test-auth-unit test-auth-integration

test-auth-unit:
	@echo "==> Running auth unit tests..."
	go test ./internal/auth/usecase/... -v

test-auth-integration:
	@echo "==> Running auth integration tests..."
	go test ./internal/auth/repository/postgres/... -v
