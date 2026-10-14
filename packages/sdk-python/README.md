# Kuberoute Python SDK

Official Python SDK for [Kuberoute AI](https://kuberoute.ai) — Intelligent Kubernetes Workload Placement & Cost Optimization.

## Installation

```bash
pip install kuberoute
```

Or install in editable mode from this monorepo:

```bash
pip install -e packages/sdk-python
```

## Quickstart

### Synchronous Usage

```python
from kuberoute import KubeRoute

# Initialize the client
client = KubeRoute(
    base_url="http://localhost:8080", # or https://api.kuberoute.ai
    token="your-jwt-token"            # or KUBEROUTE_TOKEN env var
)

# In local dev mode, you can log in directly:
client.login_dev()

# Check health
health = client.health()
print(f"Server status: {health.status}")

# List registered Kubernetes clusters
clusters = client.clusters.list()
for cluster in clusters.clusters:
    print(f"Cluster: {cluster.name} ({cluster.provider}/{cluster.region})")

# Register a new cluster
new_cluster = client.clusters.create(
    name="production-eks",
    provider="aws",
    region="us-west-2"
)
print(f"Registered cluster ID: {new_cluster.id}")

# Fetch routing & placement decisions
decisions = client.decisions.list(cluster_id=new_cluster.id)
for d in decisions.decisions:
    print(f"Decision: {d.id} | Savings: ${d.predicted_cost_savings}/mo | Reason: {d.reason}")
    if d.confidence_score > 0.9:
        client.decisions.apply(cluster_id=new_cluster.id, decision_id=d.id)
        print("Applied decision!")

# Get aggregate infrastructure costs and savings
costs = client.costs.get_total()
print(f"Total: ${costs.total} | Optimized: ${costs.optimized} | Savings: ${costs.savings}")
```

### Asynchronous Usage (`AsyncKubeRoute`)

```python
import asyncio
from kuberoute import AsyncKubeRoute

async def main():
    async with AsyncKubeRoute(base_url="http://localhost:8080", token="your-token") as client:
        clusters = await client.clusters.list()
        print(f"Found {clusters.total} clusters")

asyncio.run(main())
```

## Resources

- **`client.clusters`**: `list()`, `get(id)`, `create(name, provider, region)`, `delete(id)`
- **`client.decisions`**: `list(cluster_id, status, page, limit)`, `get(cluster_id, decision_id)`, `apply(...)`, `reject(...)`, `revert(...)`
- **`client.costs`**: `get_total()`, `get_cluster(cluster_id)`
- **`client.me()`**: Get current authenticated user and tenant credit stats.
