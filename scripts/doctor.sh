#!/usr/bin/env bash
# ==============================================================================
# Kuberoute AI — Environment Doctor & System Diagnostic Utility
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Text styling
BOLD=$'\033[1m'
GREEN=$'\033[32m'
CYAN=$'\033[36m'
YELLOW=$'\033[33m'
RED=$'\033[31m'
DIM=$'\033[2m'
RESET=$'\033[0m'

ERRORS=0
WARNINGS=0

echo -e "${BOLD}${CYAN}====================================================================${RESET}"
echo -e "${BOLD}${CYAN}                   Kuberoute AI — System Doctor                     ${RESET}"
echo -e "${BOLD}${CYAN}====================================================================${RESET}"
echo ""

report_ok() {
    echo -e "  [${GREEN}✓${RESET}] $1"
}

report_warn() {
    echo -e "  [${YELLOW}!${RESET}] $1"
    WARNINGS=$((WARNINGS + 1))
}

report_err() {
    echo -e "  [${RED}✗${RESET}] $1"
    ERRORS=$((ERRORS + 1))
}

# 1. System Tooling Checks
echo -e "${BOLD}1. Checking Required Development Tools:${RESET}"

# Docker
if command -v docker &>/dev/null; then
    DOCKER_VER=$(docker --version | awk '{print $3}' | tr -d ',')
    report_ok "Docker CLI installed (${DOCKER_VER})"
    if docker info &>/dev/null; then
        report_ok "Docker daemon is running and accessible"
    else
        report_err "Docker daemon is NOT running or user lacks permissions"
    fi
else
    report_err "Docker is not installed. Required to run database, redis, and containers."
fi

# Docker Compose
if docker compose version &>/dev/null; then
    COMPOSE_VER=$(docker compose version --short 2>/dev/null || docker compose version)
    report_ok "Docker Compose v2 plugin installed (${COMPOSE_VER})"
elif command -v docker-compose &>/dev/null; then
    COMPOSE_VER=$(docker-compose --version)
    report_ok "Docker Compose standalone installed (${COMPOSE_VER})"
else
    report_err "Neither 'docker compose' nor 'docker-compose' found"
fi

# Go
if command -v go &>/dev/null; then
    GO_VER=$(go version | awk '{print $3}')
    report_ok "Go compiler installed (${GO_VER})"
else
    report_warn "Go is not installed (needed only for local Go builds / agent development)"
fi

# Python & UV
if command -v python3 &>/dev/null; then
    PY_VER=$(python3 --version | awk '{print $2}')
    report_ok "Python 3 installed (${PY_VER})"
else
    report_err "Python 3 is not installed"
fi

if command -v uv &>/dev/null; then
    UV_VER=$(uv --version | awk '{print $2}')
    report_ok "uv package manager installed (${UV_VER})"
else
    report_warn "uv is not installed. Recommended for fast Python dependency management."
fi

# Node.js & PNPM
if command -v node &>/dev/null; then
    NODE_VER=$(node -v)
    report_ok "Node.js installed (${NODE_VER})"
else
    report_warn "Node.js not found (needed for local Next.js Dashboard development)"
fi

if command -v pnpm &>/dev/null; then
    PNPM_VER=$(pnpm -v)
    report_ok "pnpm package manager installed (${PNPM_VER})"
else
    report_warn "pnpm not found (used by apps/dashboard)"
fi

# Kubernetes tools (optional)
if command -v kubectl &>/dev/null; then
    KCTL_VER=$(kubectl version --client -o json 2>/dev/null | grep gitVersion | head -n1 | awk -F'"' '{print $4}' || echo "detected")
    report_ok "kubectl installed (${KCTL_VER})"
else
    report_warn "kubectl not found (needed for testing agent on real clusters)"
fi

if command -v kind &>/dev/null; then
    KIND_VER=$(kind version | awk '{print $2}')
    report_ok "kind (Kubernetes in Docker) installed (${KIND_VER})"
else
    report_warn "kind not found (optional for local multi-node cluster emulation)"
fi

echo ""

# 2. Port Availability Diagnostics
echo -e "${BOLD}2. Checking Port Availability:${RESET}"

is_port_occupied() {
    local port="$1"
    if command -v ss &>/dev/null; then
        ss -tuln | grep -q -E "(:|\])${port}\b"
        return $?
    elif (timeout 1 bash -c "echo > /dev/tcp/127.0.0.1/$port") &>/dev/null; then
        return 0
    else
        return 1
    fi
}

check_port_status() {
    local name="$1"
    local default_port="$2"
    if is_port_occupied "$default_port"; then
        report_warn "Port ${default_port} (${name}) is currently in use (start_all.sh will auto-increment)"
    else
        report_ok "Port ${default_port} (${name}) is free"
    fi
}

check_port_status "PostgreSQL default" 5432
check_port_status "Redis default" 6379
check_port_status "NATS Core default" 4222
check_port_status "Orchestrator HTTP default" 8080
check_port_status "Orchestrator gRPC default" 50051
check_port_status "AI Engine default" 8081
check_port_status "Dashboard default" 3000

echo ""

# 3. Environment & Configuration
echo -e "${BOLD}3. Checking Project Configuration:${RESET}"

if [[ -f "$REPO_ROOT/.env" ]]; then
    report_ok ".env file exists"
else
    report_warn ".env file is missing. Using default local environment values."
fi

if [[ -f "$REPO_ROOT/docker-compose.dev.yml" ]]; then
    report_ok "docker-compose.dev.yml found"
else
    report_err "docker-compose.dev.yml is missing"
fi

if [[ -d "$REPO_ROOT/packages/sdk-python" ]]; then
    report_ok "Python SDK package found (packages/sdk-python)"
else
    report_err "packages/sdk-python missing"
fi

echo ""

# 4. Final Verdict
echo -e "${BOLD}${CYAN}====================================================================${RESET}"
if [[ $ERRORS -eq 0 && $WARNINGS -eq 0 ]]; then
    echo -e "${BOLD}${GREEN}  Everything looks great! System is 100% ready for Kuberoute.        ${RESET}"
elif [[ $ERRORS -eq 0 ]]; then
    echo -e "${BOLD}${YELLOW}  System is ready with ${WARNINGS} non-blocking warning(s).         ${RESET}"
else
    echo -e "${BOLD}${RED}  Found ${ERRORS} error(s) and ${WARNINGS} warning(s). Please review above. ${RESET}"
fi
echo -e "${BOLD}${CYAN}====================================================================${RESET}"
echo ""
