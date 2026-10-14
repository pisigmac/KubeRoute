# KubeRoute AI

<p align="center">
  <strong>The AI Control Plane for Kubernetes</strong><br>
  <em>AI that knows where your traffic and workloads should go before they get there.</em>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white" alt="Go 1.22+"></a>
  <a href="https://python.org"><img src="https://img.shields.io/badge/Python-3.11+-3776AB.svg?logo=python&logoColor=white" alt="Python 3.11+"></a>
  <a href="https://nextjs.org"><img src="https://img.shields.io/badge/Next.js-15-black.svg?logo=next.js&logoColor=white" alt="Next.js 15"></a>
  <a href="https://kubernetes.io"><img src="https://img.shields.io/badge/Kubernetes-1.28+-326CE5.svg?logo=kubernetes&logoColor=white" alt="Kubernetes"></a>
  <a href="https://pnpm.io"><img src="https://img.shields.io/badge/pnpm-latest-F69220.svg?logo=pnpm&logoColor=white" alt="pnpm"></a>
</p>

---

## Overview

Modern Kubernetes autoscaling and traffic routing are fundamentally **reactive**. Metrics like CPU, memory, and HTTP error rates trigger autoscalers (HPA) and ingress rewrites *after* latency degrades, workloads crash, or cloud budgets blow up.

**KubeRoute AI** transforms Kubernetes into a **proactive, self-optimizing control plane**. By running ML predictive models directly adjacent to your cluster telemetry, KubeRoute forecasts traffic surges, predicts node degradations, and optimizes workload placement before requests even arrive.

---

## Key Features

- **Smart Route (Predictive Placement)**: ML models evaluate real-time telemetry, pod resource requirements, latency profiles, and cost topology to schedule pods into optimal cluster positions with >85% accuracy.
- **Traffic Mind (Surge Forecasting)**: Detects traffic trends and predicts load spikes 15+ minutes ahead of time, preemptively tuning ingress routes and scaling HPA thresholds.
- **Cost Route (Intelligent Workload Balancing)**: Dynamically balances pods across Spot instances, ARM/Graviton architectures, and on-demand nodes without compromising SLA or availability.
- **Failover Route (Zero-Downtime Migration)**: Early-warning node health analysis flags impending failures, executing graceful workload drain and failover prior to node eviction.
- **Regression Shield (Continuous Deployment Verification)**: Automatically benchmarks newly released deployments, isolating performance anomalies and regressions within 5 minutes of rollout.

---

## Architecture

```
User & Developer Layer
  Web Dashboard (Next.js 15) │ CLI (Go / Cobra) │ Python & Go SDKs
              │                        │                 │
              ▼                        ▼                 ▼
Control Plane (Cloud / Hybrid)
  API Gateway ─────────► Orchestrator (Go) ────────► AI Engine (Python / FastAPI)
                               │                            │
                               │ mTLS gRPC Stream           │ ML Inference & PuLP Solvers
                               ▼                            ▼
In-Cluster Agent (Kubernetes DaemonSet)
  Metrics Collector ───► Router Engine ───► Policy Enforcer (OPA) ───► K8s API
```

### Components

| Component | Stack | Purpose |
| :--- | :--- | :--- |
| **Orchestrator** | Go 1.22, gRPC, REST, PostgreSQL, NATS, Redis | Central controller managing clusters, decisions, task credit quotas, and agent telemetry streams. |
| **AI Engine** | Python 3.11+, FastAPI, scikit-learn, XGBoost, PuLP | Dedicated predictive inference microservice calculating placement, cost optimization, and spike models. |
| **In-Cluster Agent** | Go 1.22, `client-go`, Open Policy Agent (OPA) | Lightweight DaemonSet deployed on cluster nodes scraping low-overhead metrics and executing routing mutations. |
| **Dashboard** | Next.js 15, React, Tailwind CSS, D3, Three.js | Real-time web UI for 3D cluster topology, live metric streaming, cost analysis, and decision approval. |
| **CLI (`kr`)** | Go, Cobra | Command-line management tool for registering clusters, applying decisions, and viewing costs. |
| **Python SDK** | Python (`httpx`, Pydantic V2) | Official typed SDK (`kuberoute`) supporting both synchronous and async/await orchestration flows. |

---

## Quick Start

### Prerequisites

- **Docker** & **Docker Compose**
- **Go 1.22+**
- **Python 3.11+** & **uv**
- **Node.js 20+** & **pnpm**
- **Kind** & **kubectl** (for local Kubernetes testing)

### 1. Environment Verification

Run the built-in pre-flight diagnostic utility:

```bash
make doctor
# Or run directly: ./scripts/doctor.sh
```

### 2. Start Services

KubeRoute features dynamic port resolution to avoid port collisions with existing services.

```bash
# Start infrastructure and services in Docker containers (default)
make start

# Or start in local development process mode (Go + Python + Next.js)
./scripts/start_all.sh --local

# Rebuild images/binaries before starting
./scripts/start_all.sh -b --docker
```

### 3. Check System Health

Inspect TCP listeners, HTTP health probes, and active ports:

```bash
make status
# Or run directly: ./scripts/status.sh
```

### 4. Seed Demo Cluster Data

Populate multi-cloud clusters, nodes, simulated workloads, and decision fixtures:

```bash
make seed
# Or run directly: ./scripts/seed_data.sh
```

### 5. Access Interfaces

- **Web Dashboard**: `http://localhost:3000` (or dynamically allocated dashboard port shown by `make status`)
- **Orchestrator REST API**: `http://localhost:8080/health`
- **AI Engine Health**: `http://localhost:8081/health`

### 6. Stop Services

```bash
make stop
# Or run directly: ./scripts/stop_all.sh
```

---

## Python SDK Usage

Install the official Python SDK or import it directly:

```python
from kuberoute import KubeRoute

# Initialize client
client = KubeRoute(
    base_url="http://localhost:8080",
    api_key="your-api-key"
)

# List managed clusters
clusters = client.clusters.list()
for cluster in clusters:
    print(f"Cluster: {cluster.name} | Status: {cluster.status}")

# Retrieve pending optimization decisions
decisions = client.decisions.list(cluster_id=clusters[0].id)
for decision in decisions:
    print(f"Recommendation: {decision.action} for pod {decision.target_pod}")
    
    # Preemptively apply decision
    client.decisions.apply(cluster_id=clusters[0].id, decision_id=decision.id)
```

Async support is natively available via `AsyncKubeRoute`:

```python
from kuberoute import AsyncKubeRoute

async with AsyncKubeRoute(base_url="http://localhost:8080") as client:
    summary = await client.costs.get_summary(cluster_id="cluster-prod-1")
    print(f"Projected Monthly Savings: ${summary.projected_savings:.2f}")
```

---

## CLI (`kr`) Commands

```bash
# Build the CLI
make build

# Log in to control plane
./bin/kr login --api-key <YOUR_KEY>

# Inspect connected clusters
./bin/kr cluster list

# View pending routing and placement decisions
./bin/kr decisions list --cluster-id <CLUSTER_ID>

# Approve and apply a decision
./bin/kr decisions apply <DECISION_ID> --cluster-id <CLUSTER_ID>

# Review cluster cost breakdowns
./bin/kr costs summary --cluster-id <CLUSTER_ID>
```

---

## Local Kubernetes Cluster (Kind)

Test the in-cluster agent in an isolated local Kubernetes environment:

```bash
# Provision local Kind cluster and deploy agent manifests
make kind-up

# Tear down Kind cluster
make kind-down
```

---

## Repository Codemap

```
.
├── apps/
│   ├── dashboard/          # Next.js 15 web UI (App Router, Tailwind, Three.js, D3)
│   └── landing/            # Standalone marketing site
├── services/
│   ├── orchestrator/       # Go control plane service (REST, gRPC, PostgreSQL, NATS)
│   ├── ai-engine/          # Python FastAPI service (placement, traffic, cost models)
│   └── agent/              # Go in-cluster DaemonSet agent (metrics collector, executor)
├── packages/
│   ├── sdk-python/         # Official typed Python SDK (sync & async)
│   └── shared-types/       # Protocol Buffers and generated types
├── cli/                    # Go Cobra CLI tool (`kr`)
├── infra/
│   ├── k8s/                # Kubernetes manifests and Kind configurations
│   ├── terraform/          # Multi-cloud infrastructure (AWS, Azure, GCP)
│   └── policies/           # Open Policy Agent (OPA) Rego security rules
├── scripts/                # Operational tooling (start, stop, status, logs, doctor, seed)
└── docs/                   # Complete architectural, operational, and API documentation
```

---

## Task Credits

KubeRoute uses a deterministic **Task Credit** model rather than unpredictable token billing:
- Each executed AI routing, placement, or failover decision consumes **1 task credit**.
- **Free Tier**: 100 credits / month.
- **Pro Tier**: 1,000 credits / month per cluster.
- **Enterprise**: High-frequency real-time stream optimization with custom limits.

---

## Operations & Diagnostics

All operational scripts reside in `scripts/`:

| Command | Utility Script | Function |
| :--- | :--- | :--- |
| `make start` | `./scripts/start_all.sh` | Launches full stack with automatic non-colliding port resolution. |
| `make stop` | `./scripts/stop_all.sh` | Gracefully terminates all containers and local processes. |
| `make restart` | `./scripts/restart_all.sh` | Performs zero-downtime or clean restart passing execution flags. |
| `make status` | `./scripts/status.sh` | Real-time TCP/HTTP health inspector and port audit. |
| `make logs` | `./scripts/logs.sh [service]` | Unified log tailer for Docker containers and local log files. |
| `make doctor` | `./scripts/doctor.sh` | Pre-flight environment diagnostics and dependency checker. |
| `make seed` | `./scripts/seed_data.sh` | Seeds realistic multi-cluster and AI decision fixtures. |

---

## Documentation

Full documentation is available in the [`docs/`](docs/) directory:
- [Architecture & Design](docs/ARCHITECTURE.md)
- [Code Map & Navigation](docs/CODEMAP.md)
- [Tech Stack Selection](docs/TECH-STACK.md)
- [API Specification](docs/API.md) & [OpenAPI Spec](docs/OPENAPI.yaml)
- [Database Schema](docs/DB-SCHEMA.md)
- [Troubleshooting & Runbooks](docs/DEBUGGING.md)
- [Features & Acceptance Criteria](docs/FEATURES.md)
- [Agent Audit Trail & ADRs](docs/AGENT-AUDIT.md)

---

## License

This project is licensed under the [MIT License](LICENSE).
