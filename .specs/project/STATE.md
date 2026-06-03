# State

**Last Updated:** 2026-06-02
**Current Work:** Foundation (hexagonal wiring) — ✅ **Execute COMPLETE.** All 14 tasks (T1–T15, no T11) implemented & committed on branch `feat/foundation` (16 atomic commits). `onit tasks` runs end-to-end over real Postgres; core tests pass offline; boundary test green; lint clean. Branch not yet merged to `main`. Next: review/merge, then the **Understand** slice.

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

### AD-005: Typed IDs in leaf pkg `internal/core/ids` (2026-05-29)

**Decision:** The five typed IDs (`UserID`, `TaskID`, `NegotiationID`, `ProviderID`, `MessageID`) live in a dependency-free leaf package `internal/core/ids` (package `ids`), **not** in the root `core` package as T3 originally drafted.
**Reason:** Root `core` ports (ADR-008) return `understanding.Task`, so root imports `understanding`; if IDs also sat in root, `understanding` would import root for the ID types → mutual import cycle. A leaf `ids` package imported by both root ports and every aggregate breaks the cycle while keeping `Persistence` in root core (ADR-008 intact).
**Trade-off:** One extra tiny package vs. an `ids.go` file in root.
**Impact:** T3 creates `internal/core/ids/`; aggregates (T4–T6, T8) and root ports (T7) import `ids`; aggregates never import root `core`. tasks.md T3 + design.md updated.

---

## Active Blockers

_(none)_

---

## Lessons Learned

- **Cross-cutting port + centralized IDs = import cycle.** Putting `Persistence` in root `core` (ADR-008) while it returns `understanding.Task` means root imports the aggregate; IDs therefore can't also live in root (the aggregate would import root back). Fix: dependency-free leaf pkg `internal/core/ids` (AD-005). Watch for the same shape when adding future root ports that return aggregate types.
- **`go tool` deps propagate their `go` directive.** Pinning `sqlc` (v1.31.1) bumped the module's `go` directive to **1.26.0**; `goose` had bumped it to 1.25.7. Tool deps raise the module's minimum Go even though they don't ship in the binary — acceptable here (toolchain auto-managed) but a real coupling to note.
- **`constraints` is a SQL keyword.** The `Task.Constraints` domain field maps to column **`task_constraints`** to stay keyword-safe across psql/sqlc.
- **Docker port 5432 was occupied** by another project; the db service publishes **5433**. `DATABASE_URL=postgres://onit:onit@localhost:5433/onit?sslmode=disable`.
- **sqlc generated code can silently drift from its SQL source.** The committed `gen/tasks.sql.go` lost `ORDER BY ... DESC` and `LIMIT 100` because `sqlc generate` was not re-run after the `.sql` was finalized — a runtime bug (oldest-first, unbounded) that compiles clean and no linter catches. Caught by the PR review. Fix: regenerate + a `sqlc-drift` CI job (`go tool sqlc generate` then `git diff --cached --exit-code` on the gen dir) so generated/source divergence fails the build.

---

## Quick Tasks Completed

- [x] **Fix PR-review CRITICAL: sqlc drift** (2026-06-02) — regenerated `gen/tasks.sql.go` to restore `ORDER BY created_at DESC LIMIT 100`; added a `sqlc-drift` guard in `.github/workflows/ci.yml` to prevent recurrence.

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
