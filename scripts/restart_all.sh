#!/usr/bin/env bash
# ==============================================================================
# Kuberoute AI — Restart All Services Utility
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Text styling
BOLD=$'\033[1m'
CYAN=$'\033[36m'
RESET=$'\033[0m'

echo -e "${BOLD}${CYAN}====================================================${RESET}"
echo -e "${BOLD}${CYAN}         Restarting Kuberoute AI Stack              ${RESET}"
echo -e "${BOLD}${CYAN}====================================================${RESET}"
echo ""

# Stop all running services
"$SCRIPT_DIR/stop_all.sh"

echo ""
echo -e "${CYAN}Starting stack back up...${RESET}"
echo ""

# Start with any passed arguments
"$SCRIPT_DIR/start_all.sh" "$@"
