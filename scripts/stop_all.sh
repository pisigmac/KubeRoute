#!/usr/bin/env bash
# ==============================================================================
# Kuberoute AI — Stop All Services Utility
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

DOWN_VOLUMES=false
KEEP_CONTAINERS=false

show_help() {
    cat <<EOF
${BOLD}Usage:${RESET} ./scripts/stop_all.sh [OPTIONS]

${BOLD}Options:${RESET}
  -v, --volumes         Stop Docker containers and remove volumes (docker-compose down -v)
  --keep-containers     Stop local background processes only (keep Docker containers running)
  -h, --help            Show this help message

${BOLD}Examples:${RESET}
  ./scripts/stop_all.sh             # Stop local background processes and Docker containers
  ./scripts/stop_all.sh -v          # Stop all services and purge Docker volumes
  ./scripts/stop_all.sh --keep-containers  # Stop local processes, leave DB/Redis/NATS running
EOF
}

# Parse flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        -v|--volumes)
            DOWN_VOLUMES=true
            shift
            ;;
        --keep-containers)
            KEEP_CONTAINERS=true
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

# Detect Docker Compose command
if docker compose version &>/dev/null; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE=""
fi

# Load resolved ports if available
if [[ -f "$PID_DIR/ports.env" ]]; then
    set -a
    source "$PID_DIR/ports.env"
    set +a
fi

echo -e "${BOLD}${CYAN}====================================================${RESET}"
echo -e "${BOLD}${CYAN}         Stopping Kuberoute AI Stack                ${RESET}"
echo -e "${BOLD}${CYAN}====================================================${RESET}"
echo ""

# 1. Stop tracked local processes
if [[ -d "$PID_DIR" ]]; then
    PID_FILES=("$PID_DIR"/*.pid)
    if [[ -e "${PID_FILES[0]}" ]]; then
        echo -e "${CYAN}Stopping local background processes...${RESET}"
        for pid_file in "${PID_FILES[@]}"; do
            service_name="$(basename "$pid_file" .pid)"
            pid="$(cat "$pid_file" 2>/dev/null || true)"

            if [[ -n "$pid" ]] && kill -0 "$pid" &>/dev/null; then
                echo -n "  Stopping $service_name (PID: $pid)... "
                kill "$pid" 2>/dev/null || true

                # Wait for process to exit
                count=0
                while kill -0 "$pid" &>/dev/null && [[ $count -lt 5 ]]; do
                    sleep 1
                    count=$((count + 1))
                done

                # Force kill if still running
                if kill -0 "$pid" &>/dev/null; then
                    echo -n "(force killing)... "
                    kill -9 "$pid" 2>/dev/null || true
                fi
                echo -e "${GREEN}STOPPED${RESET}"
            else
                echo -e "  $service_name (PID: ${pid:-N/A}) is not running."
            fi
            rm -f "$pid_file"
        done
    else
        echo -e "${CYAN}No local PID files found in $PID_DIR.${RESET}"
    fi
fi

# 2. Stop Docker Compose containers
if [[ "$KEEP_CONTAINERS" == "false" ]]; then
    if [[ -n "$DOCKER_COMPOSE" ]] && [[ -f "$REPO_ROOT/docker-compose.dev.yml" ]]; then
        echo -e "${CYAN}Stopping Docker Compose containers...${RESET}"
        DOWN_ARGS=("-f" "$REPO_ROOT/docker-compose.dev.yml" "down")
        if [[ "$DOWN_VOLUMES" == "true" ]]; then
            DOWN_ARGS+=("-v")
            echo -e "${YELLOW}Purging persistent volumes...${RESET}"
        fi
        $DOCKER_COMPOSE "${DOWN_ARGS[@]}"
        echo -e "  Docker Compose containers: ${GREEN}STOPPED${RESET}"
    else
        echo -e "${YELLOW}Docker Compose not found or docker-compose.dev.yml missing. Skipping container shutdown.${RESET}"
    fi
else
    echo -e "${YELLOW}Skipping Docker container shutdown (--keep-containers specified).${RESET}"
fi

# Clean up state file
rm -f "$PID_DIR/ports.env"

echo ""
echo -e "${BOLD}${GREEN}====================================================${RESET}"
echo -e "${BOLD}${GREEN}        All Kuberoute AI Services Stopped           ${RESET}"
echo -e "${BOLD}${GREEN}====================================================${RESET}"
echo ""
