# Testing Strategy

## Unit Tests
- Go: `go test -race -coverprofile=coverage.out ./...`
- Python: `pytest --cov=ai_engine tests/`
- Target: 80% minimum coverage

## Integration Tests
- Requires: PostgreSQL, Redis, NATS
- Run: `docker-compose -f docker-compose.dev.yml up` then `go test ./tests/integration/`

## E2E Tests
- Requires: kind cluster
- Run: `make kind-up` then `go test ./tests/e2e/`

## Load Tests
- Tool: k6
- Target: 1000 clusters, 10K req/s
- Run: `k6 run tests/load/k6_script.js`

## Chaos Tests
- Random pod kills, network partitions
- Tool: Litmus
