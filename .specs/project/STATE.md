# State

**Last Updated:** 2026-05-29
**Current Work:** Foundation (hexagonal wiring) — Tasks (tasks.md drafted, 14 tasks, awaiting approval)

---

## Recent Decisions (Last 60 days)

### AD-001: M0 tech stack locked (2026-05-29)

**Decision:** Go; `cobra` (CLI); `pgx`+`sqlc` (database, type-safe pure SQL, no ORM); `goose` (migrations); app-level AES-GCM with a key via env (credentials); interfaces-in-the-core layout (`/cmd/onit`, `/internal/core`, `/internal/adapters/*`); docker-compose (pgvector) + Makefile. Full detail in `docs/decisions.md`.
**Reason:** Few well-maintained deps, no heavy frameworks; you read/understand every query; reproducible.
**Trade-off:** More custom code than an ORM/framework would give out of the box.
**Impact:** Defines the skeleton and the foundation's adapters.

### AD-002: Agent loop in the core, SDK as transport only (2026-05-29)

**Decision:** `PortLLM` is a single-turn primitive (`Complete(req) -> text | tool_calls`); the loop (`while there is tool_use`) lives once in the core, provider-agnostic. The Anthropic SDK only for the `Messages.New` call; do **not** use the toolrunner.
**Reason:** Keeps orchestration in the core; future multi-provider = a thin adapter, loop untouched; learning objective.
**Trade-off:** ~40 extra lines of loop in M0 vs. using the ready-made toolrunner.
**Impact:** The `PortLLM` contract is born thin; review ADR-003 in `docs/decisions.md`.

### AD-003: PRD amended — Section 15 (2026-05-29)

**Decision:** Multi-source discovery (`PortDiscovery` as a registry of `DiscoverySource`, `Provider` with provenance); schema with typed spine + open `jsonb` attributes; 3-layer `PortMemory` (factual/behavioral/procedural) + `onit feedback` command. Self-improvement = data only; harness evolution = human-gated, out of M0.
**Reason:** Extensible discovery and an agent that learns without breaking auditability or the core's boundary.
**Trade-off:** Two richer ports and a schema split into two zones; no self-writing of code in M0.
**Impact:** `PortDiscovery` and `PortMemory` come in already at the foundation; `Task` modeled in two zones.

### AD-004: Bounded contexts + port location (2026-05-29)

**Decision:** `/internal/core` is split into bounded-context packages (`identity`, `understanding`, `negotiation`, `discovery`, `scheduling`, `memory`, `agent`); cross-cutting ports (`Persistence`, `LLM`, `Clock`) live in the root `core` package, context-specific ports in their package (ADR-008). Contexts are Go packages, not services. Full map in `docs/bounded-contexts.md`.
**Reason:** Aligns the core's internal seams with the domain; avoids import cycles; light strategic DDD calibrated for M0.
**Trade-off:** More packages up front vs a flat core; deliberately skips heavy DDD machinery (per-context persistence, ACLs, event bus) until M1+.
**Impact:** Shapes the Foundation design; `golang-project-layout` materializes the tree.

---

## Active Blockers

_(none)_

---

## Lessons Learned

_(none yet)_

---

## Quick Tasks Completed

_(none)_

---

## Deferred Ideas

- [ ] The agent's own loop is already decided (in the core) — not deferred, it is M0.
- [ ] Discovery sources beyond Places (web, social, sub-agent) — Captured during: PRD §15
- [ ] Open agentic browsing + self-authored harness (human-gated) — Captured during: PRD §15

---

## Todos

- [x] Decide the exact granularity of the `Negotiation` FSM in the Foundation's Design phase. → Foundation uses the PRD §10.2 states verbatim (`draft → awaiting_response → counteroffer → human_approval → confirmed`; `declined`/`expired`); refinement of transitions deferred to the Negotiation slice.
- [ ] `TaskState` value set is PROVISIONAL in the Foundation design — finalize in the Understand slice.
- [x] Ratify Go module path before `go mod init` → `github.com/bredacoder/onit-ai` (binary stays `onit`).

---

## Preferences

**Model Guidance Shown:** never
**Working style:** deliberate — one task at a time, review/understand, atomic commit, no vibe coding (see memory `working-style`).
