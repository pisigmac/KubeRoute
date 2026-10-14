# Environment Variables

## Orchestrator
- `DATABASE_URL` — PostgreSQL connection string
- `REDIS_URL` — Redis address
- `NATS_URL` — NATS server URL
- `JWT_SECRET` — JWT signing key
- `GITHUB_CLIENT_ID` — GitHub OAuth app ID
- `GITHUB_CLIENT_SECRET` — GitHub OAuth secret
- `GOOGLE_CLIENT_ID` — Google OAuth client ID
- `GOOGLE_CLIENT_SECRET` — Google OAuth secret
- `HTTP_ADDR` — HTTP listen address (default :8080)
- `GRPC_ADDR` — gRPC listen address (default :50051)

## AI Engine
- `DATABASE_URL` — Same as orchestrator
- `REDIS_URL` — Same as orchestrator
- `NATS_URL` — Same as orchestrator
- `MODEL_CONFIG_PATH` — Path to ai-engine.config.yaml

## Agent
- `KUBEROUTE_CLUSTER_ID` — UUID of registered cluster
- `KUBEROUTE_API_ENDPOINT` — gRPC endpoint of control plane
- `KUBEROUTE_METRICS_INTERVAL` — Scrape interval (default 15s)
- `KUBEROUTE_INSECURE` — Skip TLS verification (dev only)
