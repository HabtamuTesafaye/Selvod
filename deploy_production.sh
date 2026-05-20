#!/bin/bash

# Selvod Production Deployment Automation
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

clear
echo -e "${BLUE}${BOLD}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}${BOLD}║              SELVOD PRODUCTION DEPLOY                ║${NC}"
echo -e "${BLUE}${BOLD}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

# 1. Environment configuration check
if [ ! -f .env ]; then
    echo -e " ${YELLOW}!${NC} No .env file found. Copying .env.example..."
    cp .env.example .env
fi

# 2. Secret Hygiene & Auto-Hardening
HAS_INSECURE_KEYS=0
if grep -E "local-dev-secret-key-12345|local-admin-key-abcde|local-playback-key-xyz|generate-a-long-random-string-for-secure-link|generate-a-long-random-admin-api-key|generate-a-long-random-playback-only-key" .env > /dev/null; then
    HAS_INSECURE_KEYS=1
fi

if [ $HAS_INSECURE_KEYS -eq 1 ]; then
    echo -e " ${YELLOW}⚠️ WARNING:${NC} Insecure or default placeholder keys detected in your .env file."
    
    AUTO_GENERATE=0
    if [ -t 0 ]; then
        echo -n " Would you like to automatically generate secure, high-entropy production keys? (y/n) "
        read -r RESPONSE
        if [[ "$RESPONSE" =~ ^[Yy]$ ]]; then
            AUTO_GENERATE=1
        fi
    else
        echo -e " ${YELLOW}!${NC} Non-interactive terminal detected. Auto-generating secure keys for protection."
        AUTO_GENERATE=1
    fi

    if [ $AUTO_GENERATE -eq 1 ]; then
        echo -e " ${BOLD}→${NC} Generating cryptographically secure random keys..."
        
        # Generate 32-byte hex keys
        STREAM_SECRET=$(openssl rand -hex 32 2>/dev/null || od -vAn -N32 -tx1 /dev/urandom | tr -d ' \n')
        API_KEY=$(openssl rand -hex 32 2>/dev/null || od -vAn -N32 -tx1 /dev/urandom | tr -d ' \n')
        PLAYBACK_KEY=$(openssl rand -hex 32 2>/dev/null || od -vAn -N32 -tx1 /dev/urandom | tr -d ' \n')

        # Replace placeholders in .env file
        # Check OS type for sed syntax differences
        SED_OPTS="-i"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            SED_OPTS="-i ''"
        fi

        sed $SED_OPTS "s/local-dev-secret-key-12345/$STREAM_SECRET/g" .env
        sed $SED_OPTS "s/local-admin-key-abcde/$API_KEY/g" .env
        sed $SED_OPTS "s/local-playback-key-xyz/$PLAYBACK_KEY/g" .env
        
        sed $SED_OPTS "s/generate-a-long-random-string-for-secure-link/$STREAM_SECRET/g" .env
        sed $SED_OPTS "s/generate-a-long-random-admin-api-key/$API_KEY/g" .env
        sed $SED_OPTS "s/generate-a-long-random-playback-only-key/$PLAYBACK_KEY/g" .env

        echo -e " ${GREEN}✔${NC} Keys updated successfully in .env."
    else
        echo -e " ${RED}✖${NC} Aborting deployment. Please manually set secure keys in your .env before deploying."
        exit 1
    fi
fi

# 3. Run Integrity Checks (Pre-deployment Gate)
echo -e " ${BOLD}→${NC} Executing Integrity & Verification tests..."
./check.sh
if [ $? -ne 0 ]; then
    echo -e " ${RED}✖${NC} Integration/integrity checks failed. Aborting deployment."
    exit 1
fi

# 4. Launch Production Stack
echo -e " ${BOLD}→${NC} Starting Production Containers..."
docker compose up -d --build
if [ $? -ne 0 ]; then
    echo -e " ${RED}✖${NC} Docker Compose failed to spin up. Aborting production deployment."
    exit 1
fi

echo ""
echo -e "${GREEN}${BOLD}✔ PRODUCTION STACK SUCCESSFULLY DEPLOYED${NC}"
echo -e "--------------------------------------------------------"
echo -e " ${BOLD}Streaming Gate (HTTP):${NC}  http://localhost"
echo -e " ${BOLD}Streaming Gate (HTTPS):${NC} https://localhost"
echo -e "--------------------------------------------------------"
echo ""
