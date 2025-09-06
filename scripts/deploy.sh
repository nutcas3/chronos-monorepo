#!/bin/bash
set -e

# Deployment script for Project Chronos
# This script deploys all microservices using Docker Compose

# Print colored output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Deploying Project Chronos...${NC}"

# Root directory of the monorepo
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  echo -e "${RED}Docker is not running. Please start Docker and try again.${NC}"
  exit 1
fi

# Build Docker images
echo -e "${YELLOW}Building Docker images...${NC}"
docker-compose build
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to build Docker images${NC}"
  exit 1
fi

# Start the services
echo -e "${YELLOW}Starting services...${NC}"
docker-compose up -d
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to start services${NC}"
  exit 1
fi

echo -e "${GREEN}Services started successfully!${NC}"
echo -e "${YELLOW}Service endpoints:${NC}"
echo -e "  - Scheduler API: http://localhost:8080"
echo -e "  - Executor API: http://localhost:8081"
echo -e "  - Worker Pool API: http://localhost:8082"
echo -e "  - Observatory API: http://localhost:8083"
echo -e "  - Prometheus: http://localhost:9091"
echo -e "  - Grafana: http://localhost:3000 (admin/admin)"
echo -e "  - Jaeger UI: http://localhost:16686"

echo -e "\n${YELLOW}To view logs:${NC}"
echo -e "  docker-compose logs -f [service_name]"

echo -e "\n${YELLOW}To stop services:${NC}"
echo -e "  docker-compose down"
