#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}Building Selvod Backend...${NC}"
cd backend
go build -o ../bin/selvod ./cmd/selvod
cd ..

echo -e "${GREEN}Building Docker Images...${NC}"
docker compose build

echo -e "${GREEN}Build Complete! Binary located in ./bin/selvod${NC}"
