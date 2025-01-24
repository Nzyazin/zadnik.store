#!/bin/bash

cd ../api/

echo "Генерация protobuf - product"
echo "Генерация protobuf - auth"
protoc --go_out=. --go-grpc_out=. *.proto