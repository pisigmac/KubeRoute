# Agent Audit Log & Architecture Decision Records (ADR)

> **Purpose**: This living document serves as a permanent, searchable audit trail for all AI agents and engineers working on this codebase. Whenever an agent creates, modifies, or refactors architecture, code, schemas, or operational tooling, it must record the rationale, alternatives evaluated, and decision justification here so subsequent agents understand *why* the system is built this way.

---

## Log Format Standard

Every entry in this document follows the standardized ADR / Agent Audit format:
- **ID & Title**: Sequential ID (e.g., `ADR-001`) and descriptive title.
- **Timestamp & Context**: Date, author/agent, and problem statement.
- **Decision Made**: The exact solution, architectural pattern, or code structure chosen.
- **Alternatives Considered & Rejected**: What other approaches were evaluated and why they were rejected.
- **System Impact & Consequences**: Performance, reliability, maintainability, or compatibility implications.

---

## Decision Records

### ADR-001: Subshell-Free Dynamic Port Auto-Resolution in `start_all.sh`

- **Date**: 2026-10-06
- **Status**: Implemented & Verified
- **Context**: When multiple local containers or host processes were running, port collisions occurred (e.g., `Orchestrator HTTP` and `AI Engine` both attempting to bind `0.0.0.0:8082`), causing Docker daemon networking crashes (`Bind for 0.0.0.0:8082 failed: port is already allocated`).
- **Root Cause**: `find_open_port` was executed inside command substitutions (`PORT=$(find_open_port ...)`), which ran in an isolated bash subshell. Any modifications to the `USED_PORTS` reservation array were lost when the subshell exited, causing multiple services checking free ports before launch to claim the exact same open port.
- **Decision**: Refactored `find_open_port` to use `printf -v "$var_name" '%s' "$port"` and directly export in the parent shell process without command substitutions or subshells.
- **Alternatives Considered**:
  - *Hardcoded port offsets (e.g., 8080, 8090, 8100)*: Rejected because if a whole range is taken by existing developer stacks, static offsets still collide.
  - *Temporary reservation lockfiles on disk*: Rejected because stale lockfiles require complex cleanup handlers if processes exit unexpectedly.
- **Consequences**: Zero port collisions across all services; transparent dynamic port resolution persisted into `.pids/ports.env`.

---

### ADR-002: Official Python SDK (`packages/sdk-python`) Architecture

- **Date**: 2026-10-06
- **Status**: Implemented & Verified
- **Context**: The orchestrator provided REST & gRPC APIs, but external clients and Python integrations had no typed, ergonomic client library.
- **Decision**: Built `packages/sdk-python` as a standalone monorepo package featuring:
  - Synchronous (`KubeRoute`) and Asynchronous (`AsyncKubeRoute`) clients using `httpx`.
  - Strict Pydantic V2 data models (`Cluster`, `Decision`, `CostSummary`, `User`).
  - Domain-specific resource namespaces (`client.clusters`, `client.decisions`, `client.costs`).
  - Comprehensive custom error hierarchy (`AuthenticationError`, `NotFoundError`, `ValidationError`, `ServerError`).
  - Native `httpx.MockTransport` test suite without third-party mocking dependencies.
- **Alternatives Considered**:
  - *Code-generating via OpenAPI generator*: Rejected for v1 because generated clients often include boilerplate, non-idiomatic abstractions, and poor async support compared to a clean, hand-crafted Pydantic/httpx library.
  - *`requests`-only client*: Rejected because modern async services (FastAPI, asyncio event loops) require native async I/O.
- **Consequences**: Enables Python scripts, ML pipelines, and automation tools to interact with Kuberoute with full IDE autocomplete and type safety.

---

### ADR-003: Operational Tooling Suite & Makefile Unification

- **Date**: 2026-10-06
- **Status**: Implemented & Verified
- **Context**: Developers and CI/CD systems required standardized, non-interactive utilities to manage, restart, inspect, and seed local services.
- **Decision**: Implemented modular standalone bash scripts under `scripts/` paired with `Makefile` shortcuts:
  - `scripts/status.sh` (`make status`): Real-time health dashboard inspecting TCP listeners, HTTP health probes, and Docker/local PIDs.
  - `scripts/restart_all.sh` (`make restart`): Safe chained invocation of `stop_all.sh` followed by `start_all.sh`.
  - `scripts/logs.sh` (`make logs`): Centralized unified log tailing across Docker containers and local log files.
  - `scripts/doctor.sh` (`make doctor`): Pre-flight environment diagnostics checking CLI tools, versions, and Docker permissions.
  - `scripts/seed_data.sh` (`make seed`): Direct SQL population of realistic multi-cluster, node, workload, and AI decision fixtures.
- **Alternatives Considered**:
  - *Embedding all logic into a heavy single script*: Rejected because separating single-responsibility scripts allows granular CI execution and simpler debugging.
- **Consequences**: Clean project root with all operational scripts neatly isolated in `scripts/`; consistent developer experience via direct script execution or Makefile.

---

### ADR-004: Comprehensive `.gitignore` Hardening and Extraneous Directory Filtering

- **Date**: 2026-10-14
- **Status**: Implemented & Verified
- **Context**: The repository had a minimal `.gitignore` that omitted standard polyglot patterns for Go binaries/coverage profiles, Python testing caches (`.mypy_cache`, `.ruff_cache`, `.hypothesis`), Next.js/Turbopack caches (`.turbo`, `.vercel`, `out/`), secrets/certificates (`*.pem`, `*.key`, `*.crt`), IDE artifacts (`.vscode/`, `.idea/`), and unmanaged local directories (`vikasbudde-landing/`).
- **Decision**: Restructured and hardened [`.gitignore`](../.gitignore) into logical sections covering:
  - Secrets and environment files (`.env`, `.env.*`, `!.env.example`, certificates/keys).
  - Operating system metadata (`.DS_Store`, `Thumbs.db`).
  - IDE and editor configurations (`.idea/`, `.vscode/`, `*.swp`).
  - Runtime logs, PIDs, and process state (`.pids/`, `logs/`, package manager logs).
  - Frontend / Next.js builds (`node_modules/`, `.next/`, `out/`, `.turbo/`, `.vercel/`).
  - Python virtualenvs and caches (`.venv/`, `__pycache__/`, `.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/`).
  - Go binaries and profiling/coverage artifacts (`bin/`, `*.test`, `*.prof`, `coverage.out`).
  - Terraform states and lockfiles.
  - Kubernetes/Kind dev configs and external/unrelated directories (`vikasbudde-landing/`).
- **Alternatives Considered**:
  - *Deleting `vikasbudde-landing/` from the filesystem*: Rejected taking destructive action on user filesystem folders without explicit instruction; ignoring via `.gitignore` cleanly prevents accidental commits while preserving files.
- **Consequences**: Prevents secret leaks, avoids tracking build/test artifacts across polyglot microservices, and eliminates clutter from untracked git status.

---

### ADR-005: Comprehensive Root `README.md` and Documentation Unification

- **Date**: 2026-10-14
- **Status**: Implemented & Verified
- **Context**: The repository lacked a root `README.md` for prospective open-source developers, contributors, and operators. Furthermore, `*.md` was inadvertently ignored in `.gitignore`, preventing documentation discovery.
- **Decision**: 
  - Restored proper markdown tracking in `.gitignore` while keeping runtime and cache files ignored.
  - Authored a comprehensive root `README.md` providing end-to-end architecture diagrams, feature acceptance definitions, CLI and Python SDK quickstarts, operational tooling guides, Kind testing workflows, and cross-references to `docs/`.
- **Alternatives Considered**:
  - *Symlinking `docs/README.md` to root*: Rejected because `docs/README.md` was a brief internal stub, whereas an open-source root `README.md` requires badges, quickstart flows, architecture breakdowns, SDK examples, and complete repository maps.
- **Consequences**: Standardized onboarding experience across GitHub; clear technical entry point for both humans and AI agents.

