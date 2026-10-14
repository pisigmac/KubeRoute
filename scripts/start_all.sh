#!/usr/bin/env bash
# ==============================================================================
# Kuberoute AI — Start All Services Utility
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_DIR="$REPO_ROOT/.pids"
LOG_DIR="$REPO_ROOT/logs"

# Text styling
BOLD=$'\033[1m'
GREEN=$'\033[32m'
CYAN=$'\033[36m'
YELLOW=$'\033[33m'
RED=$'\033[31m'
RESET=$'\033[0m'

# Default configuration
MODE="docker"          # 'docker' or 'local'
START_DASHBOARD=true
BUILD_FIRST=false

show_help() {
    cat <<EOF
${BOLD}Usage:${RESET} ./scripts/start_all.sh [OPTIONS]

${BOLD}Options:${RESET}
  -d, --docker        Start infrastructure and backend services in Docker containers (default)
  -l, --local         Start infra in Docker, but run orchestrator, ai-engine, and dashboard locally
  -b, --build         Rebuild containers/binaries before starting
  --no-dashboard      Skip starting the Next.js Dashboard
  -h, --help          Show this help message

${BOLD}Port Auto-Resolution:${RESET}
  If any required port (5432, 6379, 4222, 8080, 50051, 8081, 3000) is in use,
  the script will automatically increment the port by 1 until an open port is found.

${BOLD}Examples:${RESET}
  ./scripts/start_all.sh               # Default Docker compose + local Dashboard
  ./scripts/start_all.sh --local       # Local dev mode (Go + Python + Next.js processes)
  ./scripts/start_all.sh -b --docker   # Rebuild containers and launch full stack
EOF
}

# Parse flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        -d|--docker)
            MODE="docker"
            shift
            ;;
        -l|--local)
            MODE="local"
            shift
            ;;
        -b|--build)
            BUILD_FIRST=true
            shift
            ;;
        --no-dashboard)
            START_DASHBOARD=false
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${RESET}"
            show_help
            exit 1
            ;;
    esac
done

# Ensure logs and pids directories exist
mkdir -p "$PID_DIR" "$LOG_DIR"

# Load .env if present
if [[ -f "$REPO_ROOT/.env" ]]; then
    echo -e "${CYAN}Loading environment variables from .env file...${RESET}"
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

# Detect Docker Compose command
if docker compose version &>/dev/null; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    echo -e "${RED}Error: Neither 'docker compose' nor 'docker-compose' command was found.${RESET}"
    exit 1
fi

# Check Docker daemon connectivity
if ! docker info &>/dev/null; then
    echo -e "${RED}Error: Docker daemon is not running. Please start Docker and try again.${RESET}"
    exit 1
fi

# Helper: check if a port is currently occupied / listening
is_port_in_use() {
    local port="$1"
    if command -v ss &>/dev/null; then
        ss -tuln | grep -q -E "(:|\])${port}\b"
        return $?
    elif (timeout 1 bash -c "echo > /dev/tcp/127.0.0.1/$port") &>/dev/null; then
        return 0
    elif command -v nc &>/dev/null; then
        nc -z 127.0.0.1 "$port" &>/dev/null
        return $?
    elif command -v lsof &>/dev/null; then
        lsof -i :"$port" -sTCP:LISTEN &>/dev/null
        return $?
    else
        return 1
    fi
}

USED_PORTS=()

is_port_reserved() {
    local check_port="$1"
    for p in "${USED_PORTS[@]}"; do
        if [[ "$p" == "$check_port" ]]; then
            return 0
        fi
    done
    return 1
}

# Automatically increment port by 1 until an open port is found
find_open_port() {
    local var_name="$1"
    local service_name="$2"
    local port="$3"

    while is_port_in_use "$port" || is_port_reserved "$port"; do
        echo -e "${YELLOW}Port $port is in use or reserved for $service_name. Incrementing to $((port + 1))...${RESET}"
        port=$((port + 1))
    done

    USED_PORTS+=("$port")
    printf -v "$var_name" '%s' "$port"
    export "$var_name"
}

echo -e "${BOLD}${CYAN}====================================================${RESET}"
echo -e "${BOLD}${CYAN}         Starting Kuberoute AI Stack                ${RESET}"
echo -e "${BOLD}${CYAN}====================================================${RESET}"
echo -e "${BOLD}Mode:${RESET} ${MODE}"
echo -e "${BOLD}Dashboard:${RESET} ${START_DASHBOARD}"
echo ""

# Dynamic Port Resolution
echo -e "${CYAN}Resolving open ports for all services...${RESET}"
find_open_port POSTGRES_PORT "PostgreSQL" "${POSTGRES_PORT:-5432}"
find_open_port REDIS_PORT "Redis" "${REDIS_PORT:-6379}"
find_open_port NATS_PORT "NATS Core" "${NATS_PORT:-4222}"
find_open_port NATS_MGMT_PORT "NATS Management" "${NATS_MGMT_PORT:-8222}"
find_open_port ORCH_HTTP_PORT "Orchestrator HTTP" "${ORCH_HTTP_PORT:-8080}"
find_open_port ORCH_GRPC_PORT "Orchestrator gRPC" "${ORCH_GRPC_PORT:-50051}"
find_open_port AI_ENGINE_PORT "AI Engine" "${AI_ENGINE_PORT:-8081}"

if [[ "$START_DASHBOARD" == "true" ]]; then
    find_open_port DASHBOARD_PORT "Dashboard" "${DASHBOARD_PORT:-3000}"
fi

# Save resolved ports to state file for reference and shutdown checks
cat <<ENV > "$PID_DIR/ports.env"
POSTGRES_PORT=$POSTGRES_PORT
REDIS_PORT=$REDIS_PORT
NATS_PORT=$NATS_PORT
NATS_MGMT_PORT=$NATS_MGMT_PORT
ORCH_HTTP_PORT=$ORCH_HTTP_PORT
ORCH_GRPC_PORT=$ORCH_GRPC_PORT
AI_ENGINE_PORT=$AI_ENGINE_PORT
DASHBOARD_PORT=${DASHBOARD_PORT:-3000}
ENV

# Connection strings using resolved ports
export DATABASE_URL="${DATABASE_URL:-postgres://kuberoute:kuberoute_dev@localhost:${POSTGRES_PORT}/kuberoute?sslmode=disable}"
export REDIS_URL="${REDIS_URL:-localhost:${REDIS_PORT}}"
export NATS_URL="${NATS_URL:-nats://localhost:${NATS_PORT}}"
export HTTP_ADDR=":${ORCH_HTTP_PORT}"
export GRPC_ADDR=":${ORCH_GRPC_PORT}"
export AI_ENGINE_URL="${AI_ENGINE_URL:-http://localhost:${AI_ENGINE_PORT}}"
export KUBEROUTE_ENV="${KUBEROUTE_ENV:-local}"

# Helper to wait for HTTP health endpoint
wait_for_http() {
    local name="$1"
    local url="$2"
    local max_retries="${3:-30}"
    local count=0

    echo -n "Waiting for $name to become ready at $url..."
    while [[ $count -lt $max_retries ]]; do
        if curl -s -f "$url" &>/dev/null; then
            echo -e " ${GREEN}READY!${RESET}"
            return 0
        fi
        sleep 1
        count=$((count + 1))
        echo -n "."
    done
    echo -e " ${YELLOW}TIMED OUT (will check again later)${RESET}"
    return 1
}

# 1. Start Services based on Mode
if [[ "$MODE" == "docker" ]]; then
    echo -e "${CYAN}Starting stack via Docker Compose with resolved ports...${RESET}"
    COMPOSE_SERVICES=("postgres" "redis" "nats" "orchestrator" "ai-engine")
    if [[ "$START_DASHBOARD" == "true" ]]; then
        COMPOSE_SERVICES+=("dashboard")
    fi

    COMPOSE_ARGS=("-f" "$REPO_ROOT/docker-compose.dev.yml" "up" "-d" "--build")
    COMPOSE_ARGS+=("${COMPOSE_SERVICES[@]}")

    $DOCKER_COMPOSE "${COMPOSE_ARGS[@]}"

    echo -e "${CYAN}Checking Docker infrastructure health...${RESET}"
    wait_for_http "Orchestrator HTTP" "http://localhost:${ORCH_HTTP_PORT}/health" 15 || true
    wait_for_http "AI Engine" "http://localhost:${AI_ENGINE_PORT}/health" 15 || true
    if [[ "$START_DASHBOARD" == "true" ]]; then
        wait_for_http "Dashboard" "http://localhost:${DASHBOARD_PORT}" 20 || true
    fi

else
    echo -e "${CYAN}Starting infrastructure containers (PostgreSQL, Redis, NATS)...${RESET}"
    $DOCKER_COMPOSE -f "$REPO_ROOT/docker-compose.dev.yml" up -d postgres redis nats

    echo -e "${CYAN}Waiting for infrastructure readiness...${RESET}"
    sleep 3

    # Start Orchestrator locally
    echo -e "${CYAN}Launching Orchestrator service locally on port ${ORCH_HTTP_PORT} / ${ORCH_GRPC_PORT}...${RESET}"
    if [[ "$BUILD_FIRST" == "true" ]]; then
        (cd "$REPO_ROOT/services/orchestrator" && go build -o "$REPO_ROOT/bin/orchestrator" ./cmd)
        ORCH_CMD="$REPO_ROOT/bin/orchestrator"
    else
        ORCH_CMD="go run ./cmd/main.go"
    fi

    (
        cd "$REPO_ROOT/services/orchestrator"
        exec $ORCH_CMD > "$LOG_DIR/orchestrator.log" 2>&1
    ) &
    ORCH_PID=$!
    echo $ORCH_PID > "$PID_DIR/orchestrator.pid"
    echo -e "  Orchestrator running with PID: ${GREEN}$ORCH_PID${RESET}"

    # Start AI Engine locally
    echo -e "${CYAN}Launching AI Engine service locally on port ${AI_ENGINE_PORT}...${RESET}"
    (
        cd "$REPO_ROOT/services/ai-engine"
        if command -v uv &>/dev/null; then
            exec uv run uvicorn ai_engine.main:app --reload --port "${AI_ENGINE_PORT}" > "$LOG_DIR/ai-engine.log" 2>&1
        else
            exec uvicorn ai_engine.main:app --reload --port "${AI_ENGINE_PORT}" > "$LOG_DIR/ai-engine.log" 2>&1
        fi
    ) &
    AI_PID=$!
    echo $AI_PID > "$PID_DIR/ai-engine.pid"
    echo -e "  AI Engine running with PID: ${GREEN}$AI_PID${RESET}"

    # Start Dashboard locally
    if [[ "$START_DASHBOARD" == "true" ]]; then
        echo -e "${CYAN}Launching Dashboard (Next.js) locally on port ${DASHBOARD_PORT}...${RESET}"
        DASHBOARD_DIR="$REPO_ROOT/apps/dashboard"

        if [[ -d "$DASHBOARD_DIR" ]]; then
            (
                cd "$DASHBOARD_DIR"
                export PORT="${DASHBOARD_PORT}"
                exec pnpm dev -p "${DASHBOARD_PORT}" > "$LOG_DIR/dashboard.log" 2>&1
            ) &
            DASH_PID=$!
            echo $DASH_PID > "$PID_DIR/dashboard.pid"
            echo -e "  Dashboard running with PID: ${GREEN}$DASH_PID${RESET}"
        fi
    fi

    wait_for_http "Orchestrator HTTP" "http://localhost:${ORCH_HTTP_PORT}/health" 15 || true
    wait_for_http "AI Engine" "http://localhost:${AI_ENGINE_PORT}/health" 15 || true
    if [[ "$START_DASHBOARD" == "true" ]]; then
        wait_for_http "Dashboard" "http://localhost:${DASHBOARD_PORT}" 15 || true
    fi
fi

echo ""
echo -e "${BOLD}${GREEN}====================================================${RESET}"
echo -e "${BOLD}${GREEN}        Kuberoute AI Stack Started Successfully      ${RESET}"
echo -e "${BOLD}${GREEN}====================================================${RESET}"
printf "${BOLD}%-18s %-28s %-12s${RESET}\n" "SERVICE" "URL / ADDRESS" "STATUS"
printf "%-18s %-28s %-12s\n" "---------------" "----------------------------" "------------"

check_status() {
    local name="$1"
    local url="$2"
    local port="$3"
    if is_port_in_use "$port"; then
        printf "%-18s %-28s ${GREEN}%-12s${RESET}\n" "$name" "$url" "RUNNING"
    else
        printf "%-18s %-28s ${RED}%-12s${RESET}\n" "$name" "$url" "STOPPED"
    fi
}

check_status "PostgreSQL" "localhost:${POSTGRES_PORT}" "${POSTGRES_PORT}"
check_status "Redis" "localhost:${REDIS_PORT}" "${REDIS_PORT}"
check_status "NATS" "localhost:${NATS_PORT}" "${NATS_PORT}"
check_status "Orchestrator HTTP" "http://localhost:${ORCH_HTTP_PORT}" "${ORCH_HTTP_PORT}"
check_status "Orchestrator gRPC" "localhost:${ORCH_GRPC_PORT}" "${ORCH_GRPC_PORT}"
check_status "AI Engine" "http://localhost:${AI_ENGINE_PORT}" "${AI_ENGINE_PORT}"
if [[ "$START_DASHBOARD" == "true" ]]; then
    check_status "Dashboard" "http://localhost:${DASHBOARD_PORT}" "${DASHBOARD_PORT}"
fi

echo ""
echo -e "${BOLD}Logs available in:${RESET} $LOG_DIR"
echo -e "${BOLD}To stop all services:${RESET} ./scripts/stop_all.sh"
echo ""
