#!/usr/bin/env bash
# ==============================================================================
# Kuberoute AI — Stack Status & Health Check Utility
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_DIR="$REPO_ROOT/.pids"
LOG_DIR="$REPO_ROOT/logs"

# Styling
BOLD=$'\033[1m'
GREEN=$'\033[32m'
CYAN=$'\033[36m'
YELLOW=$'\033[33m'
RED=$'\033[31m'
DIM=$'\033[2m'
RESET=$'\033[0m'

# Load resolved ports if available
if [[ -f "$PID_DIR/ports.env" ]]; then
    set -a
    source "$PID_DIR/ports.env"
    set +a
fi

POSTGRES_PORT="${POSTGRES_PORT:-5432}"
REDIS_PORT="${REDIS_PORT:-6379}"
NATS_PORT="${NATS_PORT:-4222}"
NATS_MGMT_PORT="${NATS_MGMT_PORT:-8222}"
ORCH_HTTP_PORT="${ORCH_HTTP_PORT:-8080}"
ORCH_GRPC_PORT="${ORCH_GRPC_PORT:-50051}"
AI_ENGINE_PORT="${AI_ENGINE_PORT:-8081}"
DASHBOARD_PORT="${DASHBOARD_PORT:-3000}"

echo -e "${BOLD}${CYAN}====================================================================${RESET}"
echo -e "${BOLD}${CYAN}                   Kuberoute AI — Stack Status                      ${RESET}"
echo -e "${BOLD}${CYAN}====================================================================${RESET}"
echo ""

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

check_http_endpoint() {
    local url="$1"
    local status_code
    status_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 2 "$url" 2>/dev/null || echo "000")
    if [[ "$status_code" =~ ^(200|204|302|307)$ ]]; then
        echo -e "${GREEN}HEALTHY ($status_code)${RESET}"
    elif [[ "$status_code" == "000" ]]; then
        echo -e "${RED}UNREACHABLE${RESET}"
    else
        echo -e "${YELLOW}HTTP $status_code${RESET}"
    fi
}

check_tcp_port() {
    local port="$1"
    if is_port_in_use "$port"; then
        echo -e "${GREEN}LISTENING${RESET}"
    else
        echo -e "${RED}DOWN${RESET}"
    fi
}

# 1. Services Health & Endpoints
printf "${BOLD}%-20s %-24s %-12s %-18s${RESET}\n" "SERVICE" "ENDPOINT" "PORT" "HEALTH / STATUS"
printf "%-20s %-24s %-12s %-18s\n" "--------------------" "------------------------" "------------" "------------------"

printf "%-20s %-24s %-12s " "PostgreSQL (Timescale)" "localhost:${POSTGRES_PORT}" "${POSTGRES_PORT}"
check_tcp_port "${POSTGRES_PORT}"

printf "%-20s %-24s %-12s " "Redis" "localhost:${REDIS_PORT}" "${REDIS_PORT}"
check_tcp_port "${REDIS_PORT}"

printf "%-20s %-24s %-12s " "NATS Core" "localhost:${NATS_PORT}" "${NATS_PORT}"
check_tcp_port "${NATS_PORT}"

printf "%-20s %-24s %-12s " "NATS Management" "http://localhost:${NATS_MGMT_PORT}" "${NATS_MGMT_PORT}"
check_http_endpoint "http://localhost:${NATS_MGMT_PORT}"

printf "%-20s %-24s %-12s " "Orchestrator REST" "http://localhost:${ORCH_HTTP_PORT}" "${ORCH_HTTP_PORT}"
check_http_endpoint "http://localhost:${ORCH_HTTP_PORT}/health"

printf "%-20s %-24s %-12s " "Orchestrator gRPC" "localhost:${ORCH_GRPC_PORT}" "${ORCH_GRPC_PORT}"
check_tcp_port "${ORCH_GRPC_PORT}"

printf "%-20s %-24s %-12s " "AI Engine" "http://localhost:${AI_ENGINE_PORT}" "${AI_ENGINE_PORT}"
check_http_endpoint "http://localhost:${AI_ENGINE_PORT}/health"

printf "%-20s %-24s %-12s " "Dashboard (Web UI)" "http://localhost:${DASHBOARD_PORT}" "${DASHBOARD_PORT}"
check_http_endpoint "http://localhost:${DASHBOARD_PORT}"

echo ""

# 2. Process / Container Details
echo -e "${BOLD}${CYAN}Service Runtimes:${RESET}"

# Check Docker Containers
if command -v docker &>/dev/null && docker info &>/dev/null; then
    CONTAINERS=$(docker ps --filter "name=kuberoute" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || true)
    if [[ -n "$CONTAINERS" && "$(echo "$CONTAINERS" | wc -l)" -gt 1 ]]; then
        echo -e "${BOLD}Docker Containers:${RESET}"
        echo "$CONTAINERS"
        echo ""
    fi
fi

# Check Local Background Processes
if [[ -d "$PID_DIR" ]]; then
    PID_FILES=("$PID_DIR"/*.pid)
    if [[ -e "${PID_FILES[0]}" ]]; then
        echo -e "${BOLD}Local Background Processes:${RESET}"
        for pf in "${PID_FILES[@]}"; do
            sname="$(basename "$pf" .pid)"
            spid="$(cat "$pf" 2>/dev/null || true)"
            if [[ -n "$spid" ]] && kill -0 "$spid" &>/dev/null; then
                echo -e "  • ${BOLD}$sname${RESET}: PID ${GREEN}$spid${RESET} (Running)"
            else
                echo -e "  • ${BOLD}$sname${RESET}: PID ${RED}${spid:-N/A}${RESET} (Dead / Stale PID)"
            fi
        done
        echo ""
    fi
fi

# 3. Quick Actions summary
echo -e "${DIM}Commands:${RESET}"
echo -e "  • View logs:    ${BOLD}./scripts/logs.sh [service]${RESET}"
echo -e "  • Restart:      ${BOLD}./scripts/restart_all.sh${RESET}"
echo -e "  • Stop all:     ${BOLD}./scripts/stop_all.sh${RESET}"
echo ""
