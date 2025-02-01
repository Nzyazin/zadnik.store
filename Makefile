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
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down:
	@echo "==> Running migrations down..."
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

.PHONY: proto migrate-up migrate-down migrate-create