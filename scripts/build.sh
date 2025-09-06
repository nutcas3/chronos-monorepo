#!/bin/bash
set -e

# Build script for Project Chronos
# This script builds all microservices in the monorepo

# Print colored output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Building Project Chronos...${NC}"

# Root directory of the monorepo
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# Build Go services
build_go_service() {
  local service=$1
  echo -e "${YELLOW}Building $service...${NC}"
  cd "$ROOT_DIR/$service"
  go build -o bin/$service
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}Successfully built $service${NC}"
  else
    echo -e "${RED}Failed to build $service${NC}"
    exit 1
  fi
}

# Build Rust services
build_rust_service() {
  local service=$1
  echo -e "${YELLOW}Building $service...${NC}"
  cd "$ROOT_DIR/$service"
  cargo build --release
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}Successfully built $service${NC}"
  else
    echo -e "${RED}Failed to build $service${NC}"
    exit 1
  fi
}

# Generate code from proto files
echo -e "${YELLOW}Generating code from proto files...${NC}"
cd "$ROOT_DIR"
./scripts/gen-protos.sh
if [ $? -eq 0 ]; then
  echo -e "${GREEN}Successfully generated code from proto files${NC}"
else
  echo -e "${RED}Failed to generate code from proto files${NC}"
  exit 1
fi

# Build Go services
build_go_service "scheduler"
build_go_service "executor"
build_go_service "worker-pool"
build_go_service "observatory"

# Build Rust services
build_rust_service "durable-engine"

echo -e "${GREEN}All services built successfully!${NC}"
