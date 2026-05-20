#!/bin/bash

# Selvod Local Deployment Automation
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

clear
echo -e "${BLUE}${BOLD}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}${BOLD}║             SELVOD LOCAL DEVELOPMENT DEPLOY          ║${NC}"
echo -e "${BLUE}${BOLD}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

# 1. Environment configuration
if [ ! -f .env ]; then
    echo -e " ${YELLOW}!${NC} No .env file found. Copying .env.example..."
    cp .env.example .env
else
    echo -e " ${GREEN}✔${NC} Environment file .env detected."
fi

# 2. SSL Certificate Generation
echo -e " ${BOLD}→${NC} Verification of SSL local certificates..."
./setup_certs.sh
if [ $? -ne 0 ]; then
    echo -e " ${RED}✖${NC} Failed to set up SSL certificates. Aborting local deployment."
    exit 1
fi

# 3. Spin up Docker Stack
echo -e " ${BOLD}→${NC} Launching Docker containers..."
docker compose -f docker-compose.local.yml up -d --build
if [ $? -ne 0 ]; then
    echo -e " ${RED}✖${NC} Docker Compose failed to spin up. Aborting local deployment."
    exit 1
fi

echo ""
echo -e "${GREEN}${BOLD}✔ LOCAL STACK SUCCESSFULLY RUNNING${NC}"
echo -e "--------------------------------------------------------"
echo -e " ${BOLD}API Backend:${NC}  http://localhost:8081"
echo -e " ${BOLD}Nginx Edge:${NC}   https://localhost:18443 (HTTP: http://localhost:18080)"
echo -e " ${BOLD}Data Mount:${NC}   ./data/"
echo -e "--------------------------------------------------------"
echo ""
echo -e "${BLUE}${BOLD}Next steps to run frontend dashboard:${NC}"
echo -e " 1. Open a new terminal tab/window"
echo -e " 2. Run: ${BOLD}cd frontend && yarn install && yarn dev${NC}"
echo -e " 3. Navigate to: ${BOLD}http://localhost:5173${NC}"
echo ""
