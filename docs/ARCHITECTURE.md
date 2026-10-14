# Architecture

```
User Layer
  Web UI (Next.js)  CLI (Go)  API Keys
       |                |           |
       v                v           v
  Control Plane (Cloud)
    API Gateway (Kong)  Orchestrator (Go)  AI Engine (Python)
         |                    |                  |
         v                    v                  v
    Model Switchboard    Memory Core       Cost Optimizer
         |
         | mTLS + gRPC
         v
  In-Cluster Agent (DaemonSet)
    Metrics Collector  Router Engine  Enforcer (OPA)
         |                  |              |
         v                  v              v
         Kubernetes API + CNI
```

## Communication
- Dashboard/CLI -> API Gateway: HTTPS + gRPC
- API Gateway -> Orchestrator: gRPC internal
- Orchestrator -> AI Engine: gRPC
- Control Plane -> Agent: Bidirectional gRPC stream (mTLS)
- Agent -> K8s API: In-cluster service account
