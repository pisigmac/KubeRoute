-- 001_initial_schema.sql
-- Kuberoute AI — Initial Database Schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Tenants (organizations)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    plan TEXT NOT NULL CHECK (plan IN ('free', 'pro', 'enterprise')),
    billing_email TEXT NOT NULL,
    task_credits INTEGER DEFAULT 0,
    task_credits_used INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    provider TEXT NOT NULL CHECK (provider IN ('github', 'google')),
    provider_id TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    display_name TEXT,
    avatar_url TEXT,
    role TEXT DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(provider, provider_id)
);

-- OAuth sessions
CREATE TABLE oauth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Clusters (registered K8s clusters)
CREATE TABLE clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    name TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN ('aws', 'gcp', 'azure', 'on-prem')),
    region TEXT NOT NULL,
    k8s_version TEXT,
    agent_version TEXT,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'connected', 'degraded', 'disconnected')),
    last_heartbeat TIMESTAMPTZ,
    node_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tenant_id, name)
);

-- Nodes (within clusters)
CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
    node_name TEXT NOT NULL,
    instance_type TEXT,
    capacity_cpu NUMERIC,
    capacity_memory BIGINT,
    labels JSONB DEFAULT '{}',
    taints JSONB DEFAULT '[]',
    spot_instance BOOLEAN DEFAULT false,
    zone TEXT,
    cost_per_hour NUMERIC(10,6),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Workloads (deployments, statefulsets)
CREATE TABLE workloads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
    namespace TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('Deployment', 'StatefulSet', 'DaemonSet', 'Job')),
    replica_count INTEGER DEFAULT 1,
    labels JSONB DEFAULT '{}',
    resource_requests JSONB DEFAULT '{}',
    resource_limits JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Routing Decisions (AI-generated)
CREATE TABLE routing_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
    workload_id UUID REFERENCES workloads(id) ON DELETE SET NULL,
    decision_type TEXT NOT NULL CHECK (decision_type IN ('placement', 'scaling', 'ingress', 'cost_shift', 'failover')),
    target_node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    confidence_score NUMERIC(4,3),
    predicted_cost_savings NUMERIC(10,2),
    predicted_latency_ms INTEGER,
    reason TEXT,
    applied BOOLEAN DEFAULT false,
    applied_at TIMESTAMPTZ,
    applied_by UUID REFERENCES users(id),
    reverted BOOLEAN DEFAULT false,
    reverted_at TIMESTAMPTZ,
    reverted_by UUID REFERENCES users(id),
    task_cost INTEGER DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Metrics (TimescaleDB hypertable)
CREATE TABLE metrics (
    time TIMESTAMPTZ NOT NULL,
    cluster_id UUID NOT NULL,
    node_id UUID,
    workload_id UUID,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    labels JSONB DEFAULT '{}'
);
SELECT create_hypertable('metrics', 'time', chunk_time_interval => INTERVAL '1 hour');

CREATE INDEX idx_metrics_cluster_time ON metrics(cluster_id, time DESC);
CREATE INDEX idx_metrics_name ON metrics(metric_name, time DESC);

-- Cost Reports
CREATE TABLE cost_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_cost NUMERIC(12,2),
    optimized_cost NUMERIC(12,2),
    savings NUMERIC(12,2),
    breakdown JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    user_id UUID REFERENCES users(id),
    cluster_id UUID REFERENCES clusters(id),
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Model Configs (for AI Engine)
CREATE TABLE model_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    model_name TEXT NOT NULL,
    model_type TEXT NOT NULL CHECK (model_type IN ('traffic_prediction', 'cost_optimization', 'failure_prediction', 'placement')),
    model_source TEXT NOT NULL CHECK (model_source IN ('kuberoute', 'byom')),
    model_path TEXT,
    hyperparameters JSONB DEFAULT '{}',
    training_data_window TEXT DEFAULT '7d',
    is_active BOOLEAN DEFAULT true,
    last_trained_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Task Credits History
CREATE TABLE task_credit_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    amount INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('purchase', 'usage', 'refund', 'bonus')),
    description TEXT,
    reference_id UUID,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- API Keys (for service accounts)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    name TEXT NOT NULL,
    key_hash TEXT UNIQUE NOT NULL,
    scopes JSONB DEFAULT '[]',
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);
