"""Pydantic Models for Kuberoute SDK."""

from datetime import datetime
from typing import Dict, List, Optional, Any
from pydantic import BaseModel, Field


class User(BaseModel):
    id: str
    email: Optional[str] = None
    name: Optional[str] = None
    avatar_url: Optional[str] = None
    role: Optional[str] = None
    tenant_id: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None


class TenantInfo(BaseModel):
    id: str
    plan: str = "free"
    task_credits: int = 0
    task_credits_used: int = 0


class CurrentUserResponse(BaseModel):
    user: User
    tenant: TenantInfo


class Cluster(BaseModel):
    id: str
    tenant_id: Optional[str] = None
    name: str
    provider: Optional[str] = None
    region: Optional[str] = None
    k8s_version: Optional[str] = None
    agent_version: Optional[str] = None
    status: Optional[str] = "active"
    node_count: Optional[int] = 0
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None


class ClusterListResponse(BaseModel):
    clusters: List[Cluster] = Field(default_factory=list)
    total: int = 0
    page: int = 1
    limit: int = 20


class Decision(BaseModel):
    id: str
    cluster_id: str
    workload_id: Optional[str] = None
    decision_type: Optional[str] = None
    target_node_id: Optional[str] = None
    confidence_score: Optional[float] = None
    predicted_cost_savings: Optional[float] = None
    predicted_latency_ms: Optional[int] = None
    reason: Optional[str] = None
    status: Optional[str] = "pending"
    applied: Optional[bool] = False
    applied_at: Optional[datetime] = None
    applied_by: Optional[str] = None
    revert_of: Optional[str] = None
    created_at: Optional[datetime] = None


class DecisionListResponse(BaseModel):
    decisions: List[Decision] = Field(default_factory=list)
    total: int = 0
    page: int = 1
    limit: int = 20


class CostSummary(BaseModel):
    cluster_id: Optional[str] = None
    total: float = 0.0
    optimized: float = 0.0
    savings: float = 0.0
    breakdown: Dict[str, float] = Field(default_factory=dict)


class HealthResponse(BaseModel):
    status: str = "ok"


class AuthResponse(BaseModel):
    user: User
    token: str
