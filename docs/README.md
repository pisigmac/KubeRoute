# Kuberoute AI

AI control plane for Kubernetes. AI that knows where your traffic should go before it gets there.

## Quick Start

```bash
# Start all services (Docker Compose + Dashboard)
./start_all.sh

# Or start in local dev process mode
./start_all.sh --local

# Stop all services
./stop_all.sh

# Create kind cluster for testing
make kind-up
```

## Architecture

Three layers:
1. **In-cluster Agent** (DaemonSet) — collects metrics, executes routing decisions
2. **Control Plane** (cloud-hosted) — AI engine, model switchboard, cost optimization
3. **Dashboard + CLI** — human interface for review, override, and configuration

## Task Credits

Instead of API tokens, Kuberoute uses task credits:
- Each routing decision costs 1 task credit
- Free tier: 100 credits/month
- Pro: 1000 credits/month per cluster
- Enterprise: custom

## Auth

GitHub OAuth and Google OAuth. No Clerk.

## Tech Stack

- Orchestrator: Go 1.22, gRPC, PostgreSQL, NATS, Redis
- AI Engine: Python 3.11, FastAPI, scikit-learn, XGBoost, PuLP
- Agent: Go 1.22, client-go
- Dashboard: Next.js 15, Tailwind CSS, TypeScript
- CLI: Go, Cobra

## License

MIT
