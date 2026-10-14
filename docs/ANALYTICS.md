# Analytics Plan

## Events
- `cluster_register` — New cluster added
- `cluster_connect` — Agent first heartbeat
- `decision_generate` — AI engine creates decision
- `decision_apply` — User applies decision
- `decision_reject` — User rejects decision
- `cost_report_generated` — Monthly report ready
- `agent_disconnect` — Heartbeat missed
- `task_credit_purchase` — Credits bought
- `task_credit_deduct` — Credits used

## Tools
- Segment for event routing
- Amplitude for product analytics
- Prometheus for engineering metrics
