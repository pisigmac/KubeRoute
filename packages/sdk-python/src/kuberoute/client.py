"""Synchronous Kuberoute Client."""

import os
from typing import Optional, Dict, Any
import httpx

from .models import (
    User,
    CurrentUserResponse,
    Cluster,
    ClusterListResponse,
    Decision,
    DecisionListResponse,
    CostSummary,
    HealthResponse,
    AuthResponse,
)
from .exceptions import (
    KubeRouteError,
    AuthenticationError,
    NotFoundError,
    ValidationError,
    ServerError,
)


def _handle_response(resp: httpx.Response) -> Any:
    if resp.status_code == 204:
        return None
    try:
        data = resp.json()
    except Exception:
        data = {"raw": resp.text}

    if resp.status_code >= 400:
        err_msg = data.get("error") if isinstance(data, dict) else str(data)
        if not err_msg:
            err_msg = f"HTTP {resp.status_code}"

        if resp.status_code in (401, 403):
            raise AuthenticationError(err_msg, status_code=resp.status_code, details=data)
        elif resp.status_code == 404:
            raise NotFoundError(err_msg, status_code=resp.status_code, details=data)
        elif resp.status_code == 400:
            raise ValidationError(err_msg, status_code=resp.status_code, details=data)
        elif resp.status_code >= 500:
            raise ServerError(err_msg, status_code=resp.status_code, details=data)
        else:
            raise KubeRouteError(err_msg, status_code=resp.status_code, details=data)

    return data


class ClustersResource:
    """Operations on Kubernetes clusters."""

    def __init__(self, client: "KubeRoute") -> None:
        self._client = client

    def list(self, page: int = 1, limit: int = 20) -> ClusterListResponse:
        """List all clusters registered with Kuberoute."""
        data = self._client._get("/api/v1/clusters", params={"page": page, "limit": limit})
        return ClusterListResponse.model_validate(data)

    def get(self, cluster_id: str) -> Cluster:
        """Get details of a specific cluster by ID."""
        data = self._client._get(f"/api/v1/clusters/{cluster_id}")
        return Cluster.model_validate(data)

    def create(self, name: str, provider: Optional[str] = None, region: Optional[str] = None) -> Cluster:
        """Register a new cluster."""
        payload: Dict[str, Any] = {"name": name}
        if provider:
            payload["provider"] = provider
        if region:
            payload["region"] = region
        data = self._client._post("/api/v1/clusters", json_data=payload)
        return Cluster.model_validate(data)

    def delete(self, cluster_id: str) -> None:
        """Deregister and delete a cluster."""
        self._client._delete(f"/api/v1/clusters/{cluster_id}")


class DecisionsResource:
    """Operations on AI routing decisions."""

    def __init__(self, client: "KubeRoute") -> None:
        self._client = client

    def list(
        self,
        cluster_id: str,
        status: Optional[str] = None,
        page: int = 1,
        limit: int = 20,
    ) -> DecisionListResponse:
        """List AI placement and routing decisions for a cluster."""
        params: Dict[str, Any] = {"page": page, "limit": limit}
        if status:
            params["status"] = status
        data = self._client._get(f"/api/v1/clusters/{cluster_id}/decisions", params=params)
        return DecisionListResponse.model_validate(data)

    def get(self, cluster_id: str, decision_id: str) -> Decision:
        """Get a single decision by ID."""
        data = self._client._get(f"/api/v1/clusters/{cluster_id}/decisions/{decision_id}")
        return Decision.model_validate(data)

    def apply(self, cluster_id: str, decision_id: str) -> bool:
        """Apply an AI routing decision to the cluster."""
        data = self._client._post(f"/api/v1/clusters/{cluster_id}/decisions/{decision_id}/apply")
        return data.get("status") == "applied"

    def reject(self, cluster_id: str, decision_id: str) -> bool:
        """Reject an AI routing decision."""
        data = self._client._post(f"/api/v1/clusters/{cluster_id}/decisions/{decision_id}/reject")
        return data.get("status") == "rejected"

    def revert(self, cluster_id: str, decision_id: str) -> Decision:
        """Revert a previously applied decision."""
        data = self._client._post(f"/api/v1/clusters/{cluster_id}/decisions/{decision_id}/revert")
        return Decision.model_validate(data)


class CostsResource:
    """Operations on infrastructure cost summaries and analytics."""

    def __init__(self, client: "KubeRoute") -> None:
        self._client = client

    def get_total(self) -> CostSummary:
        """Get aggregate cost summary across all registered clusters."""
        data = self._client._get("/api/v1/costs")
        return CostSummary.model_validate(data)

    def get_cluster(self, cluster_id: str) -> CostSummary:
        """Get cost summary and estimated savings for a specific cluster."""
        data = self._client._get(f"/api/v1/clusters/{cluster_id}/costs")
        return CostSummary.model_validate(data)


class KubeRoute:
    """Synchronous client for Kuberoute AI Orchestrator API."""

    def __init__(
        self,
        base_url: Optional[str] = None,
        token: Optional[str] = None,
        timeout: float = 30.0,
    ) -> None:
        self.base_url = (
            base_url
            or os.environ.get("KUBEROUTE_API_URL")
            or "http://localhost:8080"
        ).rstrip("/")
        self.token = token or os.environ.get("KUBEROUTE_TOKEN")
        self.timeout = timeout
        self._client = httpx.Client(timeout=self.timeout)

        self.clusters = ClustersResource(self)
        self.decisions = DecisionsResource(self)
        self.costs = CostsResource(self)

    def _headers(self) -> Dict[str, str]:
        headers = {
            "Accept": "application/json",
            "Content-Type": "application/json",
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers

    def _get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        url = f"{self.base_url}{path}"
        resp = self._client.get(url, headers=self._headers(), params=params)
        return _handle_response(resp)

    def _post(self, path: str, json_data: Optional[Dict[str, Any]] = None, params: Optional[Dict[str, Any]] = None) -> Any:
        url = f"{self.base_url}{path}"
        resp = self._client.post(url, headers=self._headers(), json=json_data, params=params)
        return _handle_response(resp)

    def _delete(self, path: str) -> Any:
        url = f"{self.base_url}{path}"
        resp = self._client.delete(url, headers=self._headers())
        return _handle_response(resp)

    def health(self) -> HealthResponse:
        """Check Orchestrator health status."""
        data = self._get("/health")
        return HealthResponse.model_validate(data)

    def login_dev(self) -> AuthResponse:
        """Perform dev authentication (local dev only) and save JWT token."""
        url = f"{self.base_url}/auth/dev?format=json"
        resp = self._client.post(url, headers={"Accept": "application/json"})
        data = _handle_response(resp)
        auth_resp = AuthResponse.model_validate(data)
        self.token = auth_resp.token
        return auth_resp

    def me(self) -> CurrentUserResponse:
        """Get current authenticated user and tenant info."""
        data = self._get("/api/v1/me")
        return CurrentUserResponse.model_validate(data)

    def close(self) -> None:
        """Close the underlying HTTP client."""
        self._client.close()

    def __enter__(self) -> "KubeRoute":
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.close()
