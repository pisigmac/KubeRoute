# Rate Limits

| Scope | Limit |
|-------|-------|
| Per API key | 1000 req/min |
| Per user | 100 req/min |
| Per cluster (agent) | 60 heartbeats/min |

Enforced via Redis token bucket. Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`.
