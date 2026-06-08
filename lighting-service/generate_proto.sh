#!/bin/bash

PROTO_DIR="../proto"
OUT_DIR="."

mkdir -p proto

protoc --go_out=$OUT_DIR \
       --go_opt=module=lighting-service \
       --go-grpc_out=$OUT_DIR \
       --go-grpc_opt=module=lighting-service \
       --proto_path=$PROTO_DIR \
       $PROTO_DIR/tunnel.proto

echo "Proto files generated successfully"
