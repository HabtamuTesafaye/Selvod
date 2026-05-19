#!/bin/bash

# Selvod - High Fidelity Integrity Pipeline (ABSOLUTE VERACITY v27)
# Colors & Formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

# Spinner function
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Header
clear
echo -e "${BLUE}${BOLD}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}${BOLD}║              SELVOD INTEGRITY PIPELINE               ║${NC}"
echo -e "${BLUE}${BOLD}║             (ABSOLUTE VERACITY EDITION)              ║${NC}"
echo -e "${BLUE}${BOLD}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

run_step() {
    local label=$1
    local cmd=$2
    echo -ne " ${BOLD}→${NC} ${label}..."
    
    # Run in a subshell to prevent CWD side-effects
    (eval "$cmd") > /tmp/selvod_check.log 2>&1 &
    local pid=$!
    spinner $pid
    wait $pid
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo -e " ${GREEN}${BOLD}[SUCCESS]${NC}"
        return 0
    else
        echo -e " ${RED}${BOLD}[FAILED]${NC}"
        echo -e "${YELLOW}Error Log:${NC}"
        tail -n 15 /tmp/selvod_check.log
        echo ""
        exit 1
    fi
}

# --- PIPELINE START ---

run_step "Backend: Unit Testing (Includes Internal Nginx Simulator)" "cd backend && go test -buildvcs=false ./..."
run_step "Backend: Build Compilation" "cd backend && go build -buildvcs=false -o /dev/null ./cmd/selvod/..."

# Mandatory Security Logic Validation
run_step "System: Security Logic Validation" "cd backend && go run -buildvcs=false ./cmd/security-audit/main.go"

# MANDATORY: Live Edge Perimeter Verification
echo -e " ${BOLD}→${NC} Preparing Live Perimeter Test Environment..."
# 1. Ensure certs and test_data permissions
./setup_certs.sh > /dev/null 2>&1
mkdir -p test_data/certs
cp -r data/certs/* test_data/certs/ > /dev/null 2>&1
chmod -R 777 test_data > /dev/null 2>&1
rm -f test_data/selvod.db*

# 2. Seed database
(cd backend && SV_DATA_DIR=../test_data go run -buildvcs=false ./cmd/seed-audit/main.go) > /dev/null 2>&1
chmod -R 777 test_data > /dev/null 2>&1

# 3. Spin up test stack
(docker compose -f docker-compose.test.yml up -d) > /dev/null 2>&1
echo -e " ${BOLD}→${NC} Waiting for stack to initialize..."
sleep 8

# 4. Run the test
# We use absolute path to be safe, or just ensure we are in the right spot
BASE_DIR=$(pwd)
run_step "System: Live Perimeter Integration Test" "$BASE_DIR/test_delivery.sh"

# 5. Spin down and clean up
echo -e " ${BOLD}→${NC} Tearing down Test Environment..."
(docker compose -f docker-compose.test.yml down) > /dev/null 2>&1
rm -rf test_data > /dev/null 2>&1

# Vulnerability Scanning
if command -v govulncheck >/dev/null 2>&1; then
    run_step "Backend: Vulnerability Scan (govulncheck)" "cd backend && govulncheck ./..."
else
    echo -e " ${YELLOW}! Skipped govulncheck (not installed)${NC}"
fi

# Project-Wide Secret Hygiene Audit
CHECK_CMD="! grep -rE 'changeme|selvod-admin-secret' . --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=data --exclude-dir=test_data --exclude-dir=dist --exclude=check.sh --exclude=.env.example --exclude=*.md"
run_step "System: Global Secret Hygiene Audit" "$CHECK_CMD"

run_step "Frontend: Yarn Build Check" "cd frontend && yarn build"

# --- PIPELINE END ---

echo ""
echo -e "${GREEN}${BOLD}✔ ALL SYSTEMS STABLE & ABSOLUTE VERACITY HARDENED${NC}"
echo -e "${BLUE}Selvod is ready for production deployment.${NC}"
echo "--------------------------------------------------------"
