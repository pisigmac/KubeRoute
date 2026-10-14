#!/usr/bin/env bash
# ==============================================================================
# Kuberoute AI — Logs Viewer Utility
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="$REPO_ROOT/logs"

# Text styling
BOLD=$'\033[1m'
CYAN=$'\033[36m'
YELLOW=$'\033[33m'
RED=$'\033[31m'
RESET=$'\033[0m'

SERVICE="all"
FOLLOW=false
TAIL_LINES=100

show_help() {
    cat <<EOF
${BOLD}Usage:${RESET} ./scripts/logs.sh [SERVICE] [OPTIONS]

${BOLD}Services:${RESET}
  all             View logs from all services (default)
  orchestrator    Kuberoute Orchestrator (Go REST/gRPC backend)
  ai-engine       AI Engine (FastAPI Python service)
  dashboard       Next.js Web Dashboard
  postgres        TimescaleDB / PostgreSQL database
  redis           Redis cache
  nats            NATS JetStream message broker

${BOLD}Options:${RESET}
  -f, --follow        Stream/follow log output in real-time
  -n, --tail N        Number of lines to show (default: 100)
  -h, --help          Show this help message

${BOLD}Examples:${RESET}
  ./scripts/logs.sh -f                       # Stream all logs
  ./scripts/logs.sh orchestrator -f          # Stream orchestrator logs
  ./scripts/logs.sh ai-engine -n 50          # Show last 50 lines of AI Engine logs
EOF
}

# Parse positional service argument and flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        all|orchestrator|ai-engine|dashboard|postgres|redis|nats)
            SERVICE="$1"
            shift
            ;;
        -f|--follow)
            FOLLOW=true
            shift
            ;;
        -n|--tail)
            TAIL_LINES="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option or service: $1${RESET}"
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

# Check if Docker containers are running for Kuberoute
DOCKER_RUNNING=false
if [[ -n "$DOCKER_COMPOSE" ]] && docker info &>/dev/null; then
    if [[ -n "$(docker ps -q --filter "name=kuberoute" 2>/dev/null)" ]]; then
        DOCKER_RUNNING=true
    fi
fi

if [[ "$DOCKER_RUNNING" == "true" ]]; then
    COMPOSE_ARGS=("-f" "$REPO_ROOT/docker-compose.dev.yml" "logs" "--tail" "$TAIL_LINES")
    if [[ "$FOLLOW" == "true" ]]; then
        COMPOSE_ARGS+=("-f")
    fi

    if [[ "$SERVICE" != "all" ]]; then
        COMPOSE_ARGS+=("$SERVICE")
    fi

    echo -e "${CYAN}Streaming logs from Docker containers (${SERVICE})...${RESET}"
    $DOCKER_COMPOSE "${COMPOSE_ARGS[@]}"
else
    # Local log files fallback
    if [[ "$SERVICE" == "all" ]]; then
        LOG_FILES=("$LOG_DIR"/*.log)
        if [[ ! -e "${LOG_FILES[0]}" ]]; then
            echo -e "${YELLOW}No log files found in $LOG_DIR.${RESET}"
            exit 0
        fi
        echo -e "${CYAN}Reading local logs from $LOG_DIR...${RESET}"
        if [[ "$FOLLOW" == "true" ]]; then
            tail -n "$TAIL_LINES" -f "$LOG_DIR"/*.log
        else
            tail -n "$TAIL_LINES" "$LOG_DIR"/*.log
        fi
    else
        TARGET_FILE="$LOG_DIR/${SERVICE}.log"
        if [[ ! -f "$TARGET_FILE" ]]; then
            echo -e "${YELLOW}Log file $TARGET_FILE not found.${RESET}"
            exit 1
        fi
        echo -e "${CYAN}Reading logs from $TARGET_FILE...${RESET}"
        if [[ "$FOLLOW" == "true" ]]; then
            tail -n "$TAIL_LINES" -f "$TARGET_FILE"
        else
            tail -n "$TAIL_LINES" "$TARGET_FILE"
        fi
    fi
fi
