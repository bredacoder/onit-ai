# onit — agent guide

Personal AI agent for everyday local-services tasks: finds providers, gets quotes,
cross-checks the calendar, and schedules — escalating to the human only on the
essentials. Go, single-binary CLI. Currently in **M0** (single-tenant, Wizard of Oz).

## Where the truth lives (read before acting)

| Doc | What it holds | When to read |
|-----|---------------|--------------|
| `docs/prd.md` | Full vision + §15 addendum. **Source of truth — wins on any conflict.** | Scoping a feature, doubts about intent |
| `.specs/project/PROJECT.md` | Guiding summary: vision, stack, scope, constraints | Quick orientation |
| `.specs/project/STATE.md` | Current work, recent ADRs, blockers, todos | Start of every task — **keep it updated** |
| `.specs/project/ROADMAP.md` | Milestones (M0–M3) and feature status | Picking up the next feature |
| `.specs/features/*/spec.md` | Per-feature specs | Implementing that feature |
| `docs/decisions.md` | Locked stack ADRs (Decision · Why · Rejected) | Any stack/architecture choice |
| `docs/bounded-contexts.md` | DDD domain map: subdomains, bounded contexts, `/internal/core` layout | Organizing core packages, designing a domain type |

`.specs/` is managed by the **tlc-spec-driven** skill — follow that workflow; don't
hand-edit specs ad hoc, and update `STATE.md` as work progresses.

## Architecture invariants (never break these)

Hexagonal (ports & adapters). Layout:
- `/cmd/onit` — CLI entry (host)
- `/internal/core` — pure, deterministic core; **port interfaces declared here**
- `/internal/adapters/{postgres,anthropic,places,gcal,memory}` — concrete adapters

Rules:
- **Dependencies point inward.** Adapters import the core; the **core imports nobody**.
- The **core never** touches env, network, clock, or a concrete framework/driver directly — only through a port.
- **Agent loop lives in the core** (provider-agnostic). `PortLLM` is single-turn
  (`Complete(req) -> text | tool_calls`); `anthropic-sdk-go` is **transport only** —
  do NOT use its toolrunner. (ADR-003)
- **Multi-tenant-ready always:** `user_id` on every row; per-`User` credentials;
  secrets encrypted at rest with AES-GCM, key via env. (PRD §8, ADR-004)
- **Understand before Act:** free text → structured `Task` first; only then act.
  The agent never invents a provider/price/time — **mismatches always escalate**. (PRD §4)

## Stack (see `docs/decisions.md` for the why)

Go · `cobra` (CLI) · `pgx` + `sqlc` (plain reviewed SQL, **no ORM**) · `goose` (migrations)
· Postgres + pgvector · `docker-compose` + `Makefile` for local dev.

## Conventions

- **Deliberate:** one small task at a time, review/understand every line, atomic
  commits, **no vibe coding**.
- **All docs, comments, and commit messages in English.**
- Idiomatic Go, few well-maintained deps. Every SQL query is read and understood.
- Prefer the project `golang-*` skills in `.claude/skills/` when working in their area.

## Commands

Module `github.com/bredacoder/onit-ai` (Go 1.26; binary `onit`). Dev tooling
(`golangci-lint`, `goose`, `sqlc`) is pinned via `go tool` — no global installs.

- `make db-up` — start Postgres+pgvector (compose; published on **5433**)
- `make migrate` — `goose up` (needs `DATABASE_URL`, see `.env.example`)
- `make test` — full gate: db-up + migrate + `go test ./...`
- `make lint` — `go tool golangci-lint run`
- `go test ./internal/core/...` — core suite, runs **offline** (no DB/network)
- `go run ./cmd/onit tasks` — list tasks (needs `DATABASE_URL` + `ONIT_USER_ID`)

`DATABASE_URL=postgres://onit:onit@localhost:5433/onit?sslmode=disable`. Migrations
in `db/migrations` (goose); queries in `db/queries` (sqlc → `internal/adapters/postgres/gen`).
