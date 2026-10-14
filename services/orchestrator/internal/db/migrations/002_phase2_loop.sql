-- 002_phase2_loop.sql
-- Decision execution state, denormalized names for the agent/API, kind provider.

ALTER TABLE routing_decisions
    ADD COLUMN IF NOT EXISTS executed BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS executed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS params JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS workload_name TEXT,
    ADD COLUMN IF NOT EXISTS workload_namespace TEXT,
    ADD COLUMN IF NOT EXISTS workload_kind TEXT,
    ADD COLUMN IF NOT EXISTS target_node_name TEXT;

ALTER TABLE clusters DROP CONSTRAINT IF EXISTS clusters_provider_check;
ALTER TABLE clusters ADD CONSTRAINT clusters_provider_check
    CHECK (provider IN ('aws', 'gcp', 'azure', 'on-prem', 'kind'));

CREATE UNIQUE INDEX IF NOT EXISTS idx_nodes_cluster_name ON nodes (cluster_id, node_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_workloads_cluster_ns_name ON workloads (cluster_id, namespace, name);
