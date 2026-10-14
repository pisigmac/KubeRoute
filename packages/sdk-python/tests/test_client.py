"""Unit tests for Kuberoute Python SDK using built-in httpx.MockTransport."""

import json
import pytest
import httpx
from kuberoute import (
    KubeRoute,
    AsyncKubeRoute,
    AuthenticationError,
    NotFoundError,
    ValidationError,
    ServerError,
)


def create_mock_transport():
    def handler(request: httpx.Request) -> httpx.Response:
        url_path = request.url.path
        method = request.method

        if url_path == "/health":
            return httpx.Response(200, json={"status": "ok"})

        if url_path == "/auth/dev":
            return httpx.Response(
                200,
                json={
                    "user": {
                        "id": "11111111-1111-1111-1111-111111111111",
                        "email": "dev@kuberoute.local",
                        "name": "Dev User",
                        "role": "admin",
                        "tenant_id": "22222222-2222-2222-2222-222222222222",
                    },
                    "token": "jwt-token-12345",
                },
            )

        if url_path == "/api/v1/clusters":
            if method == "GET":
                return httpx.Response(
                    200,
                    json={
                        "clusters": [
                            {
                                "id": "c111",
                                "name": "prod-us-east",
                                "provider": "aws",
                                "region": "us-east-1",
                                "status": "active",
                                "node_count": 5,
                            }
                        ],
                        "total": 1,
                        "page": 1,
                        "limit": 20,
                    },
                )
            elif method == "POST":
                body = json.loads(request.content)
                return httpx.Response(
                    201,
                    json={
                        "id": "c222",
                        "name": body.get("name"),
                        "provider": body.get("provider"),
                        "region": body.get("region"),
                        "status": "active",
                        "node_count": 3,
                    },
                )

        if url_path == "/api/v1/clusters/c222":
            if method == "GET":
                return httpx.Response(
                    200,
                    json={
                        "id": "c222",
                        "name": "staging-eu",
                        "provider": "gcp",
                        "region": "europe-west1",
                        "status": "active",
                        "node_count": 3,
                    },
                )
            elif method == "DELETE":
                return httpx.Response(204)

        if url_path == "/api/v1/clusters/c111/decisions":
            return httpx.Response(
                200,
                json={
                    "decisions": [
                        {
                            "id": "d111",
                            "cluster_id": "c111",
                            "workload_id": "w1",
                            "decision_type": "placement",
                            "target_node_id": "node-spot-1",
                            "confidence_score": 0.95,
                            "predicted_cost_savings": 42.50,
                            "predicted_latency_ms": 12,
                            "status": "pending",
                        }
                    ],
                    "total": 1,
                    "page": 1,
                    "limit": 20,
                },
            )

        if url_path == "/api/v1/clusters/c111/decisions/d111/apply":
            return httpx.Response(200, json={"status": "applied"})

        if url_path == "/api/v1/clusters/c111/decisions/d111/reject":
            return httpx.Response(200, json={"status": "rejected"})

        if url_path == "/api/v1/costs":
            return httpx.Response(
                200,
                json={
                    "total": 1500.0,
                    "optimized": 1050.0,
                    "savings": 450.0,
                    "breakdown": {"compute": 300.0, "storage": 150.0},
                },
            )

        if url_path == "/api/v1/clusters/nonexistent":
            return httpx.Response(404, json={"error": "not found"})

        if url_path == "/api/v1/me":
            auth = request.headers.get("Authorization")
            if not auth or "invalid" in auth:
                return httpx.Response(401, json={"error": "unauthorized"})
            return httpx.Response(
                200,
                json={
                    "user": {"id": "u1", "email": "me@kuberoute.local"},
                    "tenant": {"id": "t1", "plan": "pro", "task_credits": 100, "task_credits_used": 10},
                },
            )

        return httpx.Response(404, json={"error": f"route not found: {url_path}"})

    return httpx.MockTransport(handler)


@pytest.fixture
def sync_client():
    client = KubeRoute(base_url="http://localhost:8080", token="test-token")
    client._client = httpx.Client(transport=create_mock_transport(), timeout=client.timeout)
    yield client
    client.close()


@pytest.fixture
def async_client():
    client = AsyncKubeRoute(base_url="http://localhost:8080", token="test-token")
    client._client = httpx.AsyncClient(transport=create_mock_transport(), timeout=client.timeout)
    yield client


def test_health_check(sync_client):
    resp = sync_client.health()
    assert resp.status == "ok"


def test_dev_login(sync_client):
    res = sync_client.login_dev()
    assert res.token == "jwt-token-12345"
    assert sync_client.token == "jwt-token-12345"
    assert res.user.email == "dev@kuberoute.local"


def test_clusters_crud(sync_client):
    # List
    clusters_resp = sync_client.clusters.list()
    assert clusters_resp.total == 1
    assert len(clusters_resp.clusters) == 1
    assert clusters_resp.clusters[0].name == "prod-us-east"

    # Create
    new_cluster = sync_client.clusters.create(name="staging-eu", provider="gcp", region="europe-west1")
    assert new_cluster.id == "c222"
    assert new_cluster.name == "staging-eu"

    # Get
    fetched = sync_client.clusters.get("c222")
    assert fetched.id == "c222"

    # Delete
    sync_client.clusters.delete("c222")


def test_decisions_workflow(sync_client):
    cluster_id = "c111"
    decision_id = "d111"

    # List decisions
    decisions = sync_client.decisions.list(cluster_id)
    assert len(decisions.decisions) == 1
    assert decisions.decisions[0].predicted_cost_savings == 42.50

    # Apply decision
    assert sync_client.decisions.apply(cluster_id, decision_id) is True

    # Reject decision
    assert sync_client.decisions.reject(cluster_id, decision_id) is True


def test_costs(sync_client):
    costs = sync_client.costs.get_total()
    assert costs.total == 1500.0
    assert costs.savings == 450.0
    assert costs.breakdown["compute"] == 300.0


def test_me(sync_client):
    me = sync_client.me()
    assert me.user.email == "me@kuberoute.local"
    assert me.tenant.plan == "pro"


def test_error_handling(sync_client):
    with pytest.raises(NotFoundError) as exc_info:
        sync_client.clusters.get("nonexistent")
    assert exc_info.value.status_code == 404

    sync_client.token = "invalid"
    with pytest.raises(AuthenticationError) as exc_info:
        sync_client.me()
    assert exc_info.value.status_code == 401


def test_async_client():
    import asyncio

    async def _test():
        client = AsyncKubeRoute(base_url="http://localhost:8080", token="test-token")
        client._client = httpx.AsyncClient(transport=create_mock_transport(), timeout=client.timeout)

        resp = await client.health()
        assert resp.status == "ok"

        clusters = await client.clusters.list()
        assert len(clusters.clusters) == 1
        assert clusters.clusters[0].name == "prod-us-east"
        await client.close()

    asyncio.run(_test())
