# AI Assistant Context — Kuberoute AI

This is a production-grade Kubernetes AI control plane.

## Conventions
- Go: Standard layout, internal packages, protobuf-generated types
- Python: FastAPI, async/await, type hints
- Next.js: App Router, server components, Tailwind
- Auth: Custom OAuth2 (GitHub + Google), JWT tokens
- Billing: Task credits, not API tokens

## Key Files
- `services/orchestrator/cmd/main.go` — Entry point
- `services/orchestrator/internal/api/rest.go` — REST handlers
- `services/orchestrator/internal/api/grpc.go` — gRPC agent server
- `services/ai-engine/ai_engine/main.py` — FastAPI app
- `services/ai-engine/ai_engine/models/` — ML models
- `services/agent/cmd/main.go` — Agent entry point

## Testing
- Go: `go test -race -coverprofile=coverage.out ./...`
- Python: `pytest --cov=ai_engine tests/`
- E2E: `make kind-up` then run integration tests
