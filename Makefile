# Имя бинарного файла
BINARY_NAME=app

# Переменные для protobuf
PROTO_DIR=api
PROTO_FILES=$(PROTO_DIR)/*.proto

proto:
	@echo "==> Генерация protobuf..."
	@cd $(PROTO_DIR) && \
	echo "Генерация protobuf - product" && \
	echo "Генерация protobuf - auth" && \
	protoc --go_out=. --go-grpc_out=. *.proto