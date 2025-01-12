#!/bin/bash

cd ../api/

echo "Генерация protobuf - product"
protoc --go_out=. --go-grpc_out=. *.proto