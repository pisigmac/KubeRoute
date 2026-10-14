# Error Codes

## KBR-001 — Auth Failed
Invalid or expired JWT token. Re-authenticate via GitHub or Google.

## KBR-002 — Cluster Not Found
The cluster ID does not exist or you do not have access.

## KBR-003 — Decision Not Found
The decision ID does not exist or was already processed.

## KBR-004 — Insufficient Task Credits
Your tenant has run out of task credits. Upgrade plan or purchase more.

## KBR-005 — Agent Disconnected
Agent heartbeat missed for > 5 minutes. Check cluster health.

## KBR-006 — Model Error
AI engine failed to generate prediction. Check model config and logs.

## KBR-007 — OPA Policy Violation
Decision violates safety policy. Review decision details.

## KBR-008 — Rate Limit Exceeded
Too many requests. Limit: 1000 req/min per API key.
