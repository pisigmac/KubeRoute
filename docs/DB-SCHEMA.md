# Database Schema

## tenants
Organizations. Plans: free, pro, enterprise.

## users
OAuth users (GitHub, Google). Roles: owner, admin, member.

## clusters
Registered K8s clusters. Status: pending, connected, degraded, disconnected.

## nodes
Cluster nodes with capacity, cost, spot flag.

## workloads
Deployments, StatefulSets, DaemonSets, Jobs.

## routing_decisions
AI-generated decisions. Types: placement, scaling, ingress, cost_shift, failover.

## metrics
TimescaleDB hypertable. Chunk interval: 1 hour.

## cost_reports
Monthly cost breakdown with savings.

## audit_logs
Immutable audit trail.

## model_configs
Per-tenant model configs. Source: kuberoute or byom.

## task_credit_history
Task credit transactions.
