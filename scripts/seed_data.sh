#!/usr/bin/env bash
# ==============================================================================
# Kuberoute AI — Database & Demo Data Seeder Utility
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Styling
BOLD=$'\033[1m'
GREEN=$'\033[32m'
CYAN=$'\033[36m'
YELLOW=$'\033[33m'
RED=$'\033[31m'
RESET=$'\033[0m'

echo -e "${BOLD}${CYAN}====================================================================${RESET}"
echo -e "${BOLD}${CYAN}                   Kuberoute AI — Demo Data Seeder                  ${RESET}"
echo -e "${BOLD}${CYAN}====================================================================${RESET}"
echo ""

# Find PostgreSQL container
PG_CONTAINER=$(docker ps --filter "name=kuberoute-postgres" -q | head -n1 || true)

if [[ -z "$PG_CONTAINER" ]]; then
    echo -e "${RED}Error: Kuberoute PostgreSQL container is not running.${RESET}"
    echo -e "Start the stack first with: ${BOLD}./scripts/start_all.sh${RESET}"
    exit 1
fi

echo -e "${CYAN}Seeding demo tenant, clusters, workloads, nodes, and AI routing decisions...${RESET}"

docker exec -i "$PG_CONTAINER" psql -U kuberoute -d kuberoute <<'EOSQL'
-- 1. Create or get Demo Tenant
INSERT INTO tenants (id, name, slug, plan, billing_email, task_credits, task_credits_used)
VALUES ('6dda641c-e664-460b-9114-640ac5019e93', 'Acme Cloud Corp', 'acme-cloud', 'pro', 'billing@acme.corp', 500, 42)
ON CONFLICT (id) DO UPDATE
SET task_credits = 500, plan = 'pro';

-- 2. Create or update Default Dev User
INSERT INTO users (id, tenant_id, provider, provider_id, email, display_name, role)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '6dda641c-e664-460b-9114-640ac5019e93',
    'github',
    'dev-user-001',
    'dev@localhost',
    'Dev Admin',
    'owner'
)
ON CONFLICT (email) DO UPDATE
SET tenant_id = '6dda641c-e664-460b-9114-640ac5019e93';

-- 3. Seed Clusters
INSERT INTO clusters (id, tenant_id, name, provider, region, k8s_version, agent_version, status, node_count, created_at)
VALUES 
    ('c0011111-1111-1111-1111-111111111111', '6dda641c-e664-460b-9114-640ac5019e93', 'aws-prod-us-west-2', 'aws', 'us-west-2', 'v1.30.2', 'v0.5.1', 'connected', 12, now() - INTERVAL '10 days'),
    ('c0022222-2222-2222-2222-222222222222', '6dda641c-e664-460b-9114-640ac5019e93', 'gcp-analytics-eu-west1', 'gcp', 'europe-west1', 'v1.29.5', 'v0.5.1', 'connected', 8, now() - INTERVAL '8 days'),
    ('c0033333-3333-3333-3333-333333333333', '6dda641c-e664-460b-9114-640ac5019e93', 'kind-local-sandbox', 'on-prem', 'local', 'v1.31.0', 'v0.5.1', 'connected', 3, now() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- 4. Seed Nodes for AWS Cluster
INSERT INTO nodes (id, cluster_id, node_name, instance_type, capacity_cpu, capacity_memory, spot_instance, zone, cost_per_hour)
VALUES
    ('a0011111-1111-1111-1111-111111111111', 'c0011111-1111-1111-1111-111111111111', 'ip-10-0-1-12.us-west-2.compute.internal', 'm5.2xlarge', 8, 34359738368, false, 'us-west-2a', 0.384000),
    ('a0012222-2222-2222-2222-222222222222', 'c0011111-1111-1111-1111-111111111111', 'ip-10-0-1-18.us-west-2.compute.internal', 'c5.2xlarge', 8, 17179869184, true, 'us-west-2b', 0.102000),
    ('a0013333-3333-3333-3333-333333333333', 'c0011111-1111-1111-1111-111111111111', 'ip-10-0-2-45.us-west-2.compute.internal', 'r5.xlarge', 4, 34359738368, true, 'us-west-2c', 0.078000)
ON CONFLICT (id) DO NOTHING;

-- 5. Seed Workloads
INSERT INTO workloads (id, cluster_id, namespace, name, kind, replica_count, created_at)
VALUES
    ('b0011111-1111-1111-1111-111111111111', 'c0011111-1111-1111-1111-111111111111', 'production', 'checkout-service', 'Deployment', 4, now() - INTERVAL '5 days'),
    ('b0012222-2222-2222-2222-222222222222', 'c0011111-1111-1111-1111-111111111111', 'production', 'recommendation-worker', 'Deployment', 8, now() - INTERVAL '5 days'),
    ('b0013333-3333-3333-3333-333333333333', 'c0011111-1111-1111-1111-111111111111', 'analytics', 'clickstream-aggregator', 'StatefulSet', 2, now() - INTERVAL '4 days')
ON CONFLICT (id) DO NOTHING;

-- 6. Seed AI Routing Decisions
INSERT INTO routing_decisions (
    id, cluster_id, workload_id, decision_type, target_node_id, 
    confidence_score, predicted_cost_savings, predicted_latency_ms, 
    reason, applied, applied_at, created_at
)
VALUES
    (
        'd0011111-1111-1111-1111-111111111111',
        'c0011111-1111-1111-1111-111111111111',
        'b0012222-2222-2222-2222-222222222222',
        'placement',
        'a0012222-2222-2222-2222-222222222222',
        0.965,
        284.50,
        14,
        'Shifted batch recommendation pods to c5.2xlarge spot instances with 99.4% SLA compatibility',
        true,
        now() - INTERVAL '2 days',
        now() - INTERVAL '3 days'
    ),
    (
        'd0012222-2222-2222-2222-222222222222',
        'c0011111-1111-1111-1111-111111111111',
        'b0011111-1111-1111-1111-111111111111',
        'cost_shift',
        'a0013333-3333-3333-3333-333333333333',
        0.942,
        175.20,
        11,
        'Bin-pack checkout service replicas onto memory-optimized r5.xlarge during off-peak hours',
        false,
        NULL,
        now() - INTERVAL '1 hour'
    ),
    (
        'd0013333-3333-3333-3333-333333333333',
        'c0022222-2222-2222-2222-222222222222',
        NULL,
        'scaling',
        NULL,
        0.980,
        420.00,
        25,
        'Consolidate underutilized analytics node group into high-density preemptible instances',
        false,
        NULL,
        now() - INTERVAL '30 minutes'
    )
ON CONFLICT (id) DO NOTHING;

-- 7. Seed Cost Reports
INSERT INTO cost_reports (id, tenant_id, cluster_id, period_start, period_end, total_cost, optimized_cost, savings, breakdown)
VALUES
    ('e0011111-1111-1111-1111-111111111111', '6dda641c-e664-460b-9114-640ac5019e93', 'c0011111-1111-1111-1111-111111111111', CURRENT_DATE - 30, CURRENT_DATE, 3450.00, 2570.30, 879.70, '{"compute": 2200.00, "spot_savings": 650.00, "storage": 250.00, "network": 120.00}'::jsonb),
    ('e0022222-2222-2222-2222-222222222222', '6dda641c-e664-460b-9114-640ac5019e93', 'c0022222-2222-2222-2222-222222222222', CURRENT_DATE - 30, CURRENT_DATE, 2180.00, 1640.00, 540.00, '{"compute": 1500.00, "spot_savings": 420.00, "storage": 180.00, "network": 80.00}'::jsonb)
ON CONFLICT (id) DO NOTHING;
EOSQL

echo ""
echo -e "${BOLD}${GREEN}✓ Demo data seeded successfully!${RESET}"
echo ""
echo -e "You can now view clusters and decisions in the Dashboard or via the Python SDK / CLI."
echo ""
