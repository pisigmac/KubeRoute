# API Documentation

## REST Endpoints
Base: `https://api.kuberoute.ai/v1`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/auth/github` | GitHub OAuth |
| GET | `/auth/github/callback` | GitHub callback |
| GET | `/auth/google` | Google OAuth |
| GET | `/auth/google/callback` | Google callback |
| GET | `/api/v1/clusters` | List clusters |
| POST | `/api/v1/clusters` | Register cluster |
| GET | `/api/v1/clusters/{id}` | Cluster details |
| DELETE | `/api/v1/clusters/{id}` | Deregister |
| GET | `/api/v1/clusters/{id}/decisions` | List decisions |
| POST | `/api/v1/clusters/{id}/decisions/{id}/apply` | Apply decision |
| POST | `/api/v1/clusters/{id}/decisions/{id}/reject` | Reject decision |

## gRPC Services
- `AgentService.StreamMetrics` — Bidirectional streaming
- `AgentService.Heartbeat` — Agent health ping
- `AgentService.GetConfig` — Fetch agent config
- `AgentService.ReportEvent` — Report events

Auth: Bearer token (JWT) or API key header
