-- 003_placement_pr.sql
ALTER TABLE routing_decisions
    ADD COLUMN IF NOT EXISTS solver_trace JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS proposed_patch JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS revert_of UUID REFERENCES routing_decisions(id) ON DELETE SET NULL;
