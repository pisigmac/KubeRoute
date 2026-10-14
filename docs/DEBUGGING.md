# Kuberoute AI — Debugging Guide & Error Resolution Log

> **Purpose**: This living document serves as a centralized troubleshooting knowledge base. Whenever an AI agent or engineer encounters a runtime bug, configuration issue, deployment failure, or networking error, they MUST document the error symptoms, root cause analysis, diagnostic commands, and exact resolution steps here so future agents can instantly resolve similar incidents.

---

## Standard Incident Log Template

When logging an issue, follow this structure:
```markdown
### [DEBUG-XXX] <Concise Incident Title>
- **Date**: YYYY-MM-DD
- **Component**: `<service/file/subsystem>`
- **Symptoms & Error Messages**: Exact stdout/stderr or crash log.
- **Root Cause**: Deep technical explanation of why the failure occurred.
- **Diagnostic Steps**: Commands executed to identify the issue.
- **Resolution & Fix**: Exact code or configuration changes applied.
- **Verification**: How the fix was tested and confirmed.
```

---

## Logged Incidents & Resolutions

### [DEBUG-001] Docker Container Port Collision: `Bind for 0.0.0.0:8082 failed: port is already allocated`

- **Date**: 2026-10-06
- **Component**: `scripts/start_all.sh` / Docker Compose
- **Symptoms & Error Messages**:
  ```text
  Error response from daemon: failed to set up container networking: driver failed programming 
  external connectivity on endpoint kuberoute-orchestrator-1: Bind for 0.0.0.0:8082 failed: port is already allocated
  ```
- **Root Cause**:
  In `start_all.sh`, the port finding utility was called via command substitution:
  `export ORCH_HTTP_PORT=$(find_open_port "Orchestrator HTTP" 8080)`
  Because `$(...)` spawns a bash subshell, mutations to the `USED_PORTS` reservation array were isolated to that subshell and discarded upon return. Consequently, when `AI_ENGINE_PORT` checked for an open port, `8082` was not yet listening on the OS and was missing from the parent `USED_PORTS` array, causing both services to be assigned host port `8082`.
- **Diagnostic Steps**:
  1. Ran `ss -tulpn | grep -E ":(8080|8081|8082)\b"` to identify listening ports.
  2. Inspected `.pids/ports.env` and found `ORCH_HTTP_PORT=8082` and `AI_ENGINE_PORT=8082` concurrently assigned.
  3. Checked bash subshell execution semantics.
- **Resolution & Fix**:
  Refactored `find_open_port` in `scripts/start_all.sh` to take a variable name by reference using `printf -v "$var_name" '%s' "$port"` and directly export it in the parent shell context:
  ```bash
  find_open_port() {
      local var_name="$1"
      local service_name="$2"
      local port="$3"
      while is_port_in_use "$port" || is_port_reserved "$port"; do
          port=$((port + 1))
      done
      USED_PORTS+=("$port")
      printf -v "$var_name" '%s' "$port"
      export "$var_name"
  }
  find_open_port ORCH_HTTP_PORT "Orchestrator HTTP" "${ORCH_HTTP_PORT:-8080}"
  find_open_port AI_ENGINE_PORT "AI Engine" "${AI_ENGINE_PORT:-8081}"
  ```
- **Verification**:
  Ran `./scripts/start_all.sh`. `ORCH_HTTP_PORT` resolved to `8081` and `AI_ENGINE_PORT` to `8082`. All containers started cleanly with zero collisions.

---

### [DEBUG-002] Invalid UUID Format Syntax in Database Seed Script

- **Date**: 2026-10-06
- **Component**: `scripts/seed_data.sh` / PostgreSQL TimescaleDB
- **Symptoms & Error Messages**:
  ```text
  ERROR: invalid input syntax for type uuid: "n0011111-1111-1111-1111-111111111111"
  LINE 3: ('n0011111-1111-1111-1111-111111111111', 'c0011111-1111-...
  ```
- **Root Cause**:
  The seed script generated mock node and workload IDs starting with prefix letters `'n'` and `'w'`. PostgreSQL UUIDs strictly require hexadecimal digits (`[0-9a-f]`).
- **Diagnostic Steps**:
  Examined PostgreSQL execution logs during `./scripts/seed_data.sh`.
- **Resolution & Fix**:
  Replaced non-hexadecimal prefixes with valid hex characters (`a001...` for nodes, `b001...` for workloads, `c001...` for clusters, `d001...` for decisions, `e001...` for reports).
- **Verification**:
  Executed `./scripts/seed_data.sh` and confirmed `INSERT 0 3` successful status without syntax errors.

---

### [DEBUG-003] Missing Asyncio Runner in Python SDK Pytest Environment

- **Date**: 2026-10-06
- **Component**: `packages/sdk-python/tests/test_client.py`
- **Symptoms & Error Messages**:
  ```text
  FAILED test_async_client - Failed: async def functions are not natively supported.
  You need to install a suitable plugin for your async framework.
  ```
- **Root Cause**:
  The test was decorated with `@pytest.mark.asyncio`, but `pytest-asyncio` was not installed in the active virtual environment.
- **Diagnostic Steps**:
  Ran `pytest` under `uv run --directory services/ai-engine`.
- **Resolution & Fix**:
  Rewrote `test_async_client` to invoke `asyncio.run(_test())` directly inside a standard synchronous test function, making the test suite self-contained without mandatory external pytest plugins.
- **Verification**:
  Ran `uv run pytest ../../packages/sdk-python/tests/` and confirmed 8/8 tests passing.

---

## Rapid Troubleshooting Playbook

| Symptom | Diagnostic Command | Quick Remediation |
|---------|-------------------|-------------------|
| Service unreachable | `make status` | Check if container is running or port incremented in `.pids/ports.env`. |
| Port already in use | `make doctor` | Run `make restart` to auto-resolve open ports. |
| Container crashing on boot | `make logs orchestrator` (or `ai-engine`) | Inspect stack trace in logs. |
| DB connection rejected | `docker ps -f name=postgres` | Verify postgres health status via `pg_isready -U kuberoute`. |
| SDK 401 Unauthorized | `client.login_dev()` | Generate fresh JWT in local dev mode via `/auth/dev?format=json`. |
