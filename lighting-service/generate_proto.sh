#!/bin/bash

PROTO_DIR="../proto"
OUT_DIR="./proto"

mkdir -p $OUT_DIR

protoc --go_out=$OUT_DIR \
       --go_opt=paths=source_relative \
       --go-grpc_out=$OUT_DIR \
       --go-grpc_opt=paths=source_relative \
       --proto_path=$PROTO_DIR \
       $PROTO_DIR/tunnel.proto

echo "Proto files generated successfully"
