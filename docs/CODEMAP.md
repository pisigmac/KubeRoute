# Code Map

```
kuberoute/
├── apps/
│   ├── dashboard/          # Next.js 15 web UI
│   └── landing/            # Marketing site (separate)
├── services/
│   ├── orchestrator/       # Go, gRPC + REST API
│   │   ├── cmd/main.go     # Entry point
│   │   ├── internal/
│   │   │   ├── api/        # REST + gRPC handlers
│   │   │   ├── auth/       # OAuth2 + JWT
│   │   │   ├── clusters/   # Cluster CRUD
│   │   │   ├── decisions/  # Decision CRUD
│   │   │   └── db/         # Migrations
│   │   └── proto/          # Protobuf definitions
│   ├── ai-engine/          # Python, FastAPI
│   │   ├── ai_engine/
│   │   │   ├── main.py     # FastAPI app
│   │   │   ├── models/     # ML models
│   │   │   └── pipelines/  # Decision pipeline
│   │   └── config/         # Model configs
│   └── agent/              # Go, K8s client-go
│       ├── cmd/main.go     # Agent entry
│       └── internal/
│           ├── collector/  # Metrics scraping
│           ├── executor/   # Decision application
│           ├── enforcer/   # OPA policies
│           └── buffer/     # Offline storage
├── cli/                    # Go, Cobra
├── infra/
│   ├── terraform/          # AWS, GCP, Azure
│   ├── k8s/                # K8s manifests
│   └── policies/           # OPA rego
├── scripts/                # Operational utilities (start, stop, status, logs, doctor, seed)
├── packages/
│   ├── shared-types/       # Protobuf + types
│   ├── sdk-python/         # Official Python SDK (sync & async)
│   └── sdk-go/             # Go SDK
├── tests/
│   ├── e2e/                # End-to-end
│   ├── integration/        # Integration
│   └── load/               # Load tests (k6)
└── docs/                   # 32 deliverables (incl. AGENT-AUDIT.md, DEBUGGING.md)
```
