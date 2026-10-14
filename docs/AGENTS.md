# Agent Development Guide

## Responsibilities
1. Metrics Collection — Scrape node-level metrics every 15s
2. Decision Execution — Apply routing decisions via K8s API
3. Local Enforcement — OPA sidecar validates decisions
4. Offline Buffer — SQLite on emptyDir for up to 1 hour
5. Self-Update — Rolling update via DaemonSet

## Metrics
- cpu_usage, memory_usage, request_rate, error_rate, latency_p99
- Streamed via gRPC bidirectional stream

## Safety Policies (OPA)
- No evict system pods
- No drain > 30% cluster capacity
- No cross-namespace moves without label

## Local Development
```bash
cd services/agent
go run ./cmd/main.go
# Set KUBEROUTE_CLUSTER_ID and KUBEROUTE_API_ENDPOINT
```
