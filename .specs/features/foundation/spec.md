# Foundation (hexagonal wiring) — Specification

## Problem Statement

Before any agent logic, we need to prove that onit's hexagonal architecture is correctly wired: a pure Go core depending **only on interfaces**, concrete adapters at the edge, and state in Postgres. Without this foundation, business logic leaks into the hosts (the "top risk" of section 13 of the PRD) and M0 turns into a rewrite. This slice has a minimal surface and **zero agent logic** — it exists to lock in the skeleton and the pattern.

## Goals

- [ ] The core compiles and tests **without network**, depending only on ports (interfaces), with in-memory fakes.
- [ ] Persistable data model: `User`, `Task`, `Negotiation`, `Provider`, `Message` — all with `user_id`, a typed spine + open `jsonb` attributes, the `Negotiation` FSM as a named type (invalid states hard to represent).
- [ ] `onit tasks` runs end-to-end (CLI → core → `PortPersistence` → real Postgres) and renders the list (empty state is a valid result).

## Out of Scope

| Feature | Reason |
| --- | --- |
| Agent logic (Understand/Act, LLM loop) | Following slices; this one is just wiring |
| **Real** `Task` creation, negotiation, discovery, calendar, memory | Here, only interfaces + fakes; real adapters in their own slices |
| Real calls to LLM / Places / Calendar | Concrete adapters come in their own slices |
| `User` creation/bootstrap | ⚙️ **Decision B** — lives in the "Understand" slice; the foundation uses empty state |
| Auth / multi-user | M2 |
| AES-GCM encryption of credentials (ADR-004) | Only relevant once a calendar token comes in (Calendar slice) |

---

## User Stories

### P1: `onit tasks` lists tasks end-to-end ⭐ MVP

**User Story**: As a CLI user, I want to run `onit tasks` and see my tasks (or an empty notice) to confirm that the CLI → core → Postgres wiring works.

**Why P1**: It is the thin vertical proof of the architecture — the only executable path of the slice.

**Acceptance Criteria**:
1. WHEN I run `onit tasks` with the database migrated and no tasks THEN the system SHALL display a clear empty state (e.g., "no tasks yet").
2. WHEN there are tasks for the current user in the database THEN the system SHALL list them (id, service type, state) **filtered by `user_id`**.
3. WHEN the database is unavailable THEN the system SHALL fail with a clear message and exit code != 0, without a raw stacktrace.

**Independent Test**: bring up Postgres via docker, migrate, run `onit tasks` → empty state; insert a row via SQL → it reappears in the listing.

---

### P1: Core testable with in-memory fakes ⭐ MVP

**User Story**: As a developer, I want the core depending only on ports with in-memory fakes so that all logic runs in `go test` without network.

**Why P1**: It is the structural guarantee of the core/host boundary (PRD 9.1); without it, the boundary erodes.

**Acceptance Criteria**:
1. WHEN I run `go test ./internal/core/...` THEN it SHALL pass without network or Docker, using in-memory fakes.
2. WHEN any adapter package is removed from the build THEN the core SHALL keep compiling (the core **does not import** adapters).
3. WHEN I list tasks through the core via the `PortPersistence` fake THEN it SHALL filter by `user_id` (never reads another owner's data).

**Independent Test**: `go test ./internal/core/...` green offline; an import-boundary test fails if the core imports an adapter.

---

### P2: Data model + migrations

**User Story**: As a developer, I want the data model and migrations in place so that state persists with the right shape from the start.

**Why P2**: It enables the P1s, but it is the modeling that section 14 of the PRD says to do first.

**Acceptance Criteria**:
1. WHEN I apply the `up` migration (goose) THEN it SHALL create the tables for `User`, `Task`, `Negotiation`, `Provider`, `Message` with `user_id` and an `attributes jsonb` column on `Task`.
2. WHEN I apply the `down` migration THEN it SHALL revert cleanly.
3. WHEN an unknown `Negotiation` state comes from the database THEN the system SHALL fail explicitly (not assume a silent default).

**Independent Test**: `make migrate` creates the tables; `goose down` reverts; `go test` covers the invalid-state parse.

---

## Edge Cases

- WHEN there is no current `User` configured THEN the system SHALL warn clearly (no panic).
- WHEN `attributes jsonb` is null/absent THEN the system SHALL treat it as empty, without error.
- WHEN the `Negotiation` state read from the database does not map to a known state THEN it SHALL error explicitly.
- WHEN the database is unavailable at startup THEN it SHALL give a clear message + exit != 0.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| FND-01 | P1: `onit tasks` end-to-end | Execute | ✅ Done (T13–T15) |
| FND-02 | P1: empty state and filter by `user_id` | Execute | ✅ Done (T14, T15) |
| FND-03 | P1: clear error with DB unavailable | Execute | ✅ Done (T14, T15) |
| FND-04 | P1: core testable with fakes offline | Execute | ✅ Done (T7, T9) |
| FND-05 | P1: core does not import adapters (import-boundary) | Execute | ✅ Done (T10) |
| FND-06 | P2: schema + migrations up/down | Execute | ✅ Done (T12) |
| FND-07 | P2: `Negotiation` FSM as a named type | Execute | ✅ Done (T5, +T14 unknown-state guard) |
| FND-08 | P2: `Task` model typed spine + `jsonb` | Execute | ✅ Done (T4) |

**Coverage:** 8 total, 8 implemented & verified (offline core tests + real-Postgres e2e).

---

## Success Criteria

- [ ] `onit tasks` runs end-to-end against real Postgres (empty state or with data).
- [ ] `go test ./internal/core/...` green without network.
- [ ] The core does not import any adapter/host package (verifiable by test).
- [ ] All ports (`PortLLM`, `PortPersistence`, `PortDiscovery` registry, `PortCalendar`, `PortMemory`) declared as interfaces with an in-memory fake, even if only `PortPersistence` is exercised by a command in this slice.

---

## Embedded scope decisions (to validate)

- ⚙️ **Decision A — `onit tasks` connects to REAL Postgres, not to the fake.** The honest proof of the wiring exercises migrations + `sqlc` + `pgx` + the persistence adapter early. The fakes serve the core tests (P1 #2), not the command. *(Alternative: connect to the fake and defer Postgres — a weaker proof.)*
- ⚙️ **Decision B — `User` creation stays outside the foundation.** `onit tasks` resolves the "current user" via config and uses an empty state when there is no data; the real `User` bootstrap is born in the "Understand" slice. Keeps the foundation minimal. *(Alternative: include an `onit init`/seed here — adds a 2nd command and scope.)*
