#!/bin/bash
set -e

# Proto generation script for Project Chronos
# This script generates code from proto files for all languages used in the project

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' 

echo -e "${YELLOW}Generating code from proto files...${NC}"

# Root directory of the monorepo
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
  echo -e "${RED}protoc is not installed. Please install Protocol Buffers compiler.${NC}"
  exit 1
fi

# Check if required plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
  echo -e "${RED}protoc-gen-go is not installed. Please install it with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest${NC}"
  exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
  echo -e "${RED}protoc-gen-go-grpc is not installed. Please install it with: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest${NC}"
  exit 1
fi

# Create output directories if they don't exist
mkdir -p "$ROOT_DIR/scheduler/proto"
mkdir -p "$ROOT_DIR/executor/proto"
mkdir -p "$ROOT_DIR/worker-pool/proto"
mkdir -p "$ROOT_DIR/observatory/proto"
mkdir -p "$ROOT_DIR/durable-engine/src/proto"
mkdir -p "$ROOT_DIR/clients/go/chronos-client/proto"
mkdir -p "$ROOT_DIR/clients/rust/chronos-client/src/proto"

# Generate Go code
echo -e "${YELLOW}Generating Go code...${NC}"
for proto_file in "$ROOT_DIR"/proto/*.proto; do
  filename=$(basename "$proto_file")
  echo -e "  Processing $filename..."
  
  # Generate Go code for each service
  protoc -I="$ROOT_DIR/proto" \
    --go_out="$ROOT_DIR/scheduler/proto" \
    --go-grpc_out="$ROOT_DIR/scheduler/proto" \
    "$proto_file"
  
  protoc -I="$ROOT_DIR/proto" \
    --go_out="$ROOT_DIR/executor/proto" \
    --go-grpc_out="$ROOT_DIR/executor/proto" \
    "$proto_file"
  
  protoc -I="$ROOT_DIR/proto" \
    --go_out="$ROOT_DIR/worker-pool/proto" \
    --go-grpc_out="$ROOT_DIR/worker-pool/proto" \
    "$proto_file"
  
  protoc -I="$ROOT_DIR/proto" \
    --go_out="$ROOT_DIR/observatory/proto" \
    --go-grpc_out="$ROOT_DIR/observatory/proto" \
    "$proto_file"
  
  # Generate Go client code
  protoc -I="$ROOT_DIR/proto" \
    --go_out="$ROOT_DIR/clients/go/chronos-client/proto" \
    --go-grpc_out="$ROOT_DIR/clients/go/chronos-client/proto" \
    "$proto_file"
done

# Generate Rust code
echo -e "${YELLOW}Generating Rust code...${NC}"
for proto_file in "$ROOT_DIR"/proto/*.proto; do
  filename=$(basename "$proto_file")
  echo -e "  Processing $filename..."
  
  # Generate Rust code for durable-engine
  protoc -I="$ROOT_DIR/proto" \
    --rust_out="$ROOT_DIR/durable-engine/src/proto" \
    --tonic_out="$ROOT_DIR/durable-engine/src/proto" \
    "$proto_file"
  
  # Generate Rust client code
  protoc -I="$ROOT_DIR/proto" \
    --rust_out="$ROOT_DIR/clients/rust/chronos-client/src/proto" \
    --tonic_out="$ROOT_DIR/clients/rust/chronos-client/src/proto" \
    "$proto_file"
done

echo -e "${GREEN}Successfully generated code from proto files!${NC}"
