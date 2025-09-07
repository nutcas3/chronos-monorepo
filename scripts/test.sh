#!/bin/bash
set -e

# Test script for Project Chronos
# This script runs tests for all microservices in the monorepo

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}Running tests for Project Chronos...${NC}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

test_go_service() {
  local service=$1
  echo -e "${YELLOW}Testing $service...${NC}"
  cd "$ROOT_DIR/$service"
  go test -v ./...
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}All tests passed for $service${NC}"
  else
    echo -e "${RED}Tests failed for $service${NC}"
    exit 1
  fi
}

# Test Rust services
test_rust_service() {
  local service=$1
  echo -e "${YELLOW}Testing $service...${NC}"
  cd "$ROOT_DIR/$service"
  cargo test
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}All tests passed for $service${NC}"
  else
    echo -e "${RED}Tests failed for $service${NC}"
    exit 1
  fi
}

# Run Go tests
test_go_service "scheduler"
test_go_service "executor"
test_go_service "worker-pool"
test_go_service "observatory"

# Run Rust tests
test_rust_service "durable-engine"

echo -e "${GREEN}All tests passed successfully!${NC}"
