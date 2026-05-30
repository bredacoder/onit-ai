# Foundation (hexagonal wiring) — Tasks

**Design**: `.specs/features/foundation/design.md`
**Spec**: `.specs/features/foundation/spec.md`
**Status**: Draft
**Module path**: `github.com/bredacoder/onit-ai` (binary `onit`)

---

## Test setup (greenfield — agreed with user)

**Gate commands:**
- **quick** (offline): `go vet ./... && go test ./internal/core/...`
- **full** (docker): `make test` → docker-compose Postgres(pgvector) up + `goose` migrate + `go test ./...`
- **lint** (available): `golangci-lint run`

**Test Coverage Matrix:**

| Code layer | Test type | Parallel-safe |
| --- | --- | --- |
| Core domain types **with behavior** (`Parse*`, FSM guards) | unit | yes |
| Core pure declarations (IDs, plain structs, interfaces) | none | yes |
| inmem persistence fake (user_id filtering) | unit | yes |
| inmem other fakes (trivial stubs) | none | yes |
| import-boundary test | unit | yes |
| goose migrations | integration | **no** (shared DB) |
| sqlc-generated code | none | yes |
| postgres adapter | integration | **no** (shared DB) |
| `cmd/onit` CLI | e2e | **no** (shared DB) |

> Parallelism rule applied: integration/e2e tasks share the Postgres instance → never `[P]` together.

---

## Execution Plan

### Phase 1 — Scaffolding (sequential)
```
T1 → T2
```

### Phase 2 — Domain types
```
T3 ──┬─→ T4 [P]
     ├─→ T5 [P]
     ├─→ T6 [P]
     └─→ T8 [P]
```

### Phase 3 — Ports & offline guarantees
```
T4 ─→ T7 ──┬─→ T9  [P]
           └─→ T10 [P]
```

### Phase 4 — Persistence (DB; full gate; sequential — shared DB)
```
T12 → T13 → T14
```
(T12 may overlap Phase 3 — independent code — but its integration gate needs docker.)

### Phase 5 — CLI host (e2e)
```
T14 → T15
```

---

## Task Breakdown

### T1: Initialize Go module + project layout
**What**: `go mod init github.com/bredacoder/onit-ai`; create `cmd/onit/`, `internal/core/`, `internal/adapters/`, `internal/inmem/`, `db/` dirs; `.gitignore`; `.golangci.yml`.
**Where**: repo root
**Depends on**: None
**Reuses**: `golang-project-layout`, `golang-dependency-management`, `golang-lint` skills
**Requirement**: (enabling)
**Tools**: MCP: NONE · Skill: `golang-project-layout`, `golang-lint`
**Done when**:
- [ ] `go.mod` has module `github.com/bredacoder/onit-ai` and current Go version
- [ ] Directory skeleton exists; `.gitignore` ignores binary + `.env`
- [ ] `.golangci.yml` present; `golangci-lint run` executes (no files yet = clean)
- [ ] Gate passes: `go vet ./...`
**Tests**: none · **Gate**: quick

---

### T2: Local dev infra — docker-compose + Makefile
**What**: `docker-compose.yml` (pgvector image) + `Makefile` targets `db-up`, `migrate`, `test`, `lint`.
**Where**: repo root
**Depends on**: T1
**Reuses**: ADR-007; `golang-lint`
**Requirement**: (enabling, FND-06/01 infra)
**Tools**: MCP: NONE · Skill: NONE
**Done when**:
- [ ] `make db-up` starts Postgres(pgvector) and it accepts connections
- [ ] `Makefile` defines `db-up`, `migrate`, `test`, `lint` (targets may reference tools added later)
- [ ] `.env.example` documents `DATABASE_URL` and `ONIT_USER_ID`
**Tests**: none · **Gate**: build (`go vet ./...`) + manual `make db-up`

---

### T3: Core IDs + `identity.User`
**What**: typed IDs (`UserID`, `TaskID`, `NegotiationID`, `ProviderID`, `MessageID`) in `internal/core/ids.go`; minimal `User` in `internal/core/identity/user.go`.
**Where**: `internal/core/ids.go`, `internal/core/identity/`
**Depends on**: T1
**Reuses**: `golang-naming`, `golang-structs-interfaces`
**Requirement**: (enabling)
**Tools**: MCP: NONE · Skill: `golang-naming`
**Done when**:
- [ ] ID types declared as named `string` types
- [ ] `identity.User` minimal (ID + room to grow), no behavior
- [ ] Gate passes: `go vet ./... && go test ./internal/core/...`
**Tests**: none (pure declarations) · **Gate**: quick

---

### T4: `understanding.Task` + `TaskState` + `ParseTaskState` [P]
**What**: `Task` (typed spine + `Attributes map[string]any`), `TaskState` named type, `ParseTaskState` (errors on unknown).
**Where**: `internal/core/understanding/`
**Depends on**: T3
**Reuses**: `golang-structs-interfaces`, `tactical-ddd` (keep Task behavior-first), PRD §15.3
**Requirement**: FND-08
**Tools**: MCP: NONE · Skill: `golang-structs-interfaces`, `tactical-ddd`
**Done when**:
- [ ] `Task` carries `UserID`, spine fields, `State`, `Attributes`
- [ ] `ParseTaskState` returns the provisional set; unknown → error
- [ ] nil/absent attributes → empty map (no panic)
- [ ] Unit tests: valid parse, unknown errors, empty-attributes
- [ ] Gate passes: `go vet ./... && go test ./internal/core/...`
- [ ] Test count: ≥3 tests pass (no silent deletions)
**Tests**: unit · **Gate**: quick

---

### T5: `negotiation.Negotiation` FSM + `Message` + `ParseNegotiationState` [P]
**What**: `Negotiation` referencing `TaskID`/`ProviderID` by value, `NegotiationState` named type (PRD §10.2 states), `ParseNegotiationState` (errors on unknown), `Message`.
**Where**: `internal/core/negotiation/`
**Depends on**: T3
**Reuses**: `golang-structs-interfaces`, `tactical-ddd`, PRD §10.2
**Requirement**: FND-07
**Tools**: MCP: NONE · Skill: `golang-structs-interfaces`, `tactical-ddd`
**Done when**:
- [ ] `NegotiationState` = `draft|awaiting_response|counteroffer|human_approval|confirmed|declined|expired`
- [ ] `ParseNegotiationState` unknown → explicit error
- [ ] `Negotiation` holds `UserID`, `TaskID`, `ProviderID`, `State`; `Message` ties to `NegotiationID`
- [ ] Unit tests: valid parse, unknown errors
- [ ] Gate passes: `go vet ./... && go test ./internal/core/...`
- [ ] Test count: ≥2 tests pass (no silent deletions)
**Tests**: unit · **Gate**: quick

---

### T6: `discovery.Provider` (+provenance) + `DiscoverySource` + registry `Port` [P]
**What**: `Provider` with provenance (`Source`, `Confidence`, `Evidence`), `DiscoverySource` interface, `discovery.Port` (registry) interface.
**Where**: `internal/core/discovery/`
**Depends on**: T3
**Reuses**: `golang-structs-interfaces`, PRD §15.2
**Requirement**: (enabling — port for success criteria)
**Tools**: MCP: NONE · Skill: `golang-structs-interfaces`
**Done when**:
- [ ] `Provider` carries `UserID` + provenance fields
- [ ] `DiscoverySource` and `discovery.Port` interfaces declared
- [ ] Gate passes: `go vet ./... && go test ./internal/core/...`
**Tests**: none (declarations/interfaces) · **Gate**: quick

---

### T7: Cross-cutting ports in root `core`
**What**: `Persistence` (`ListTasks(ctx, UserID) ([]understanding.Task, error)`), `LLM`, `Clock` interfaces (ADR-008).
**Where**: `internal/core/ports.go`
**Depends on**: T4
**Reuses**: `golang-structs-interfaces`, ADR-008
**Requirement**: FND-04 (enabling)
**Tools**: MCP: NONE · Skill: `golang-structs-interfaces`
**Done when**:
- [ ] `Persistence`, `LLM`, `Clock` declared in package `core`
- [ ] `Persistence.ListTasks` returns `understanding.Task`
- [ ] Gate passes: `go vet ./... && go test ./internal/core/...`
**Tests**: none (interfaces) · **Gate**: quick

---

### T8: Context ports — `scheduling.CalendarPort` + `memory.Port` [P]
**What**: declare `scheduling.CalendarPort` (PortCalendar) and `memory.Port` (3-layer) interfaces (declared this slice; not exercised).
**Where**: `internal/core/scheduling/`, `internal/core/memory/`
**Depends on**: T3
**Reuses**: `golang-structs-interfaces`, PRD §15.4
**Requirement**: (success criteria — all ports declared)
**Tools**: MCP: NONE · Skill: `golang-structs-interfaces`
**Done when**:
- [ ] `scheduling.CalendarPort` and `memory.Port` interfaces declared
- [ ] Gate passes: `go vet ./... && go test ./internal/core/...`
**Tests**: none (interfaces) · **Gate**: quick

---

### T9: in-memory fakes for all ports (+ persistence filtering test)
**What**: `internal/inmem` fakes for `core.Persistence` (filters by `user_id`), `core.LLM`, `core.Clock`, `discovery.Port`, `scheduling.CalendarPort`, `memory.Port`.
**Where**: `internal/inmem/`
**Depends on**: T4, T6, T7, T8
**Reuses**: `golang-testing`, `golang-stretchr-testify`
**Requirement**: FND-04
**Tools**: MCP: NONE · Skill: `golang-testing`, `golang-stretchr-testify`
**Done when**:
- [ ] A fake implements each port (compile-time `var _ core.Persistence = ...` checks)
- [ ] Persistence fake `ListTasks` returns ONLY rows matching the given `UserID`
- [ ] Unit test proves user_id isolation (rows of user B never returned for user A)
- [ ] Gate passes: `go vet ./... && go test ./internal/core/...`
- [ ] Test count: ≥2 tests pass (no silent deletions)
**Tests**: unit · **Gate**: quick

---

### T10: Import-boundary test (core ⊥ adapters) [P]
**What**: a test asserting no `internal/core/...` package imports `internal/adapters/...` (uses `go/packages` or `go list`).
**Where**: `internal/arch/boundary_test.go`
**Depends on**: T4, T5, T6, T7, T8
**Reuses**: `golang-testing`
**Requirement**: FND-05
**Tools**: MCP: NONE · Skill: `golang-testing`
**Done when**:
- [ ] Test enumerates core package imports and fails if any is under `internal/adapters/`
- [ ] Test is GREEN now and stays green after the postgres adapter (T14) exists
- [ ] Gate passes: `go vet ./... && go test ./internal/core/... ./internal/arch/...`
- [ ] Test count: 1 test passes
**Tests**: unit · **Gate**: quick

---

### T12: goose migrations — 5 tables
**What**: `0001_init.sql` up/down creating `users`, `tasks` (`attributes jsonb`), `negotiations`, `providers`, `messages`; all with `user_id`; state columns as `text`.
**Where**: `db/migrations/`
**Depends on**: T2, T4, T5, T6
**Reuses**: ADR-005, `golang-database`
**Requirement**: FND-06
**Tools**: MCP: NONE · Skill: `golang-database`
**Done when**:
- [ ] `make migrate` (goose up) creates all 5 tables with `user_id` and `tasks.attributes jsonb`
- [ ] `goose down` reverts cleanly
- [ ] Integration check: up then down leaves no residue
- [ ] Gate passes: `make test`
**Tests**: integration · **Gate**: full

---

### T13: sqlc setup + `ListTasksByUser` query
**What**: `sqlc.yaml` (pgx/v5 engine), `db/queries/tasks.sql` with `ListTasksByUser`, generate into `internal/adapters/postgres/gen`.
**Where**: `sqlc.yaml`, `db/queries/tasks.sql`, `internal/adapters/postgres/gen/`
**Depends on**: T1, T12
**Reuses**: ADR-002, `golang-database`
**Requirement**: FND-01 (enabling)
**Tools**: MCP: NONE · Skill: `golang-database`
**Done when**:
- [ ] `sqlc generate` produces compiling Go from the schema + query
- [ ] `ListTasksByUser` filters by `user_id`
- [ ] Generated code lives ONLY under `internal/adapters/postgres/gen`
- [ ] Gate passes: `go vet ./...` (build)
**Tests**: none (generated) · **Gate**: quick
**⚠ Verify at impl** (Context7 unavailable): exact `sqlc.yaml` v2 shape for `pgx/v5`.

---

### T14: Postgres adapter implements `core.Persistence`
**What**: `adapter.go` — pgxpool, `ListTasks` maps sqlc rows → `understanding.Task` (ACL), filters by `user_id`, decodes `attributes`, unknown state → explicit error, wraps errors with `%w`.
**Where**: `internal/adapters/postgres/adapter.go`
**Depends on**: T7, T13, T4
**Reuses**: `golang-database`, `golang-error-handling`
**Requirement**: FND-02, FND-03, FND-07
**Tools**: MCP: NONE · Skill: `golang-database`, `golang-error-handling`
**Done when**:
- [ ] `var _ core.Persistence = (*Adapter)(nil)` compiles
- [ ] Integration: insert user+task via SQL → `ListTasks` returns it; another user's row is NOT returned
- [ ] Integration: row with unknown state → explicit error (not silent default)
- [ ] DB connection error → wrapped error (no panic)
- [ ] Gate passes: `make test`
- [ ] Test count: ≥3 integration tests pass (no silent deletions)
**Tests**: integration · **Gate**: full

---

### T15: `cmd/onit` composition root + `onit tasks`
**What**: `main.go` (build pgx pool + postgres adapter, resolve current `UserID` from `ONIT_USER_ID`/config, wire cobra) + `tasks.go` (`onit tasks`: render list or empty state; map errors → message + exit≠0, no stacktrace).
**Where**: `cmd/onit/main.go`, `cmd/onit/tasks.go`
**Depends on**: T14, T7
**Reuses**: `golang-spf13-cobra`, `golang-cli`, `golang-error-handling`
**Requirement**: FND-01, FND-02, FND-03
**Tools**: MCP: NONE · Skill: `golang-spf13-cobra`, `golang-cli`
**Done when**:
- [ ] `onit tasks` with migrated DB and no rows → clear empty state ("no tasks yet")
- [ ] With rows for the current user → lists id, service type, state (filtered by user_id)
- [ ] DB unavailable → clear message + exit code ≠ 0, no raw stacktrace
- [ ] No current user configured → clear warning, no panic
- [ ] e2e tested via cobra `SetArgs`/`SetOut` + real Postgres
- [ ] Gate passes: `make test`
- [ ] Test count: ≥3 e2e tests pass (no silent deletions)
**Tests**: e2e · **Gate**: full
**Commit**: `feat(foundation): onit tasks end-to-end over hexagonal wiring`

---

## Parallel Execution Map
```
Phase 1:  T1 → T2
Phase 2:  T3 → { T4 [P], T5 [P], T6 [P], T8 [P] }
Phase 3:  T7 → { T9 [P], T10 [P] }
Phase 4:  T12 → T13 → T14          (full gate; shared DB → no [P])
Phase 5:  T14 → T15
```
Cross-phase deps: T7←T4 · T9←{T4,T6,T7,T8} · T10←{T4,T5,T6,T7,T8} · T12←{T2,T4,T5,T6} · T13←{T1,T12} · T14←{T7,T13,T4} · T15←{T14,T7}.

---

## Pre-approval Validation

### Check 1 — Granularity
| Task | Scope | Status |
| --- | --- | --- |
| T1 | module + layout | ✅ |
| T2 | compose + Makefile | ✅ |
| T3 | IDs + User (cohesive types) | ✅ |
| T4 | one aggregate package | ✅ |
| T5 | one aggregate package | ✅ |
| T6 | one context package | ✅ |
| T7 | one file (3 cohesive ports) | ✅ |
| T8 | two tiny port decls (cohesive) | ✅ |
| T9 | one fakes package | ✅ |
| T10 | one test | ✅ |
| T12 | one migration pair | ✅ |
| T13 | sqlc config + 1 query | ✅ |
| T14 | one adapter | ✅ |
| T15 | one command + root | ✅ |

### Check 2 — Diagram ↔ Definition cross-check
| Task | Depends on (body) | Diagram shows | Status |
| --- | --- | --- | --- |
| T2 | T1 | T1→T2 | ✅ |
| T3 | T1 | T1→T3 | ✅ |
| T4 | T3 | T3→T4 | ✅ |
| T5 | T3 | T3→T5 | ✅ |
| T6 | T3 | T3→T6 | ✅ |
| T7 | T4 | T4→T7 | ✅ |
| T8 | T3 | T3→T8 | ✅ |
| T9 | T4,T6,T7,T8 | T7→T9 (+cross-phase noted) | ✅ |
| T10 | T4,T5,T6,T7,T8 | T7→T10 (+cross-phase noted) | ✅ |
| T12 | T2,T4,T5,T6 | T12 head of Phase 4 (cross-phase noted) | ✅ |
| T13 | T1,T12 | T12→T13 | ✅ |
| T14 | T7,T13,T4 | T13→T14 | ✅ |
| T15 | T14,T7 | T14→T15 | ✅ |

`[P]` tasks don't depend on each other: {T4,T5,T6,T8} mutually independent ✅; {T9,T10} mutually independent ✅.

### Check 3 — Test co-location
| Task | Layer created | Matrix requires | Task says | Status |
| --- | --- | --- | --- | --- |
| T1 | scaffold | none | none | ✅ |
| T2 | infra | none | none | ✅ |
| T3 | pure types | none | none | ✅ |
| T4 | domain w/ behavior | unit | unit | ✅ |
| T5 | domain w/ behavior (FSM) | unit | unit | ✅ |
| T6 | declarations/interfaces | none | none | ✅ |
| T7 | interfaces | none | none | ✅ |
| T8 | interfaces | none | none | ✅ |
| T9 | inmem persistence fake | unit | unit | ✅ |
| T10 | import-boundary | unit | unit | ✅ |
| T12 | migrations | integration | integration | ✅ |
| T13 | generated | none | none | ✅ |
| T14 | postgres adapter | integration | integration | ✅ |
| T15 | CLI | e2e | e2e | ✅ |

All three checks pass — no ❌.

---

## Traceability (spec ↔ tasks)
| Requirement | Tasks |
| --- | --- |
| FND-01 `onit tasks` e2e | T13, T14, T15 |
| FND-02 empty state + user_id filter | T14, T15 |
| FND-03 clear error, DB down | T14, T15 |
| FND-04 core testable offline w/ fakes | T7, T9 |
| FND-05 core ⊥ adapters | T10 |
| FND-06 schema + migrations | T12 |
| FND-07 Negotiation FSM named type | T5 (+T14 unknown-state guard) |
| FND-08 Task spine + jsonb | T4 |
