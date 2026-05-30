# onit

**Vision:** Personal AI agent that handles everyday local-services tasks — finds providers, gets quotes, cross-checks the calendar and schedules the service, escalating to the human only on the essentials.
**For:** A busy, urban person who puts off local-services tasks. In M0, the developer themselves (Florianópolis/Palhoça, SC).
**Solves:** The friction of operational tasks is cognitive (who to call, what to say, when to fit it in), not physical. onit takes on that load.

> Source of truth: `prd.md` (full vision + Section 15 addendum) and `docs/decisions.md` (stack ADRs). This file is the guiding summary; in case of conflict, `prd.md` wins.

## Goals

- The flow completes a real task end to end (e.g., scheduling a sofa cleaning).
- On the happy path, the user is interrupted at most 2x (an initial question + final approval).
- The agent never invents a provider, price, or time — mismatches always escalate.

## Tech Stack

**Core:**
- Language: Go (current stable version)
- Database: Postgres + pgvector
- LLM: `anthropic-sdk-go` (transport only) behind `PortLLM`; loop in the core

**Key dependencies:** `spf13/cobra` (CLI), `pgx` + `sqlc` (database), `goose` (migrations), `anthropic-sdk-go` (LLM transport). See `docs/decisions.md`.

**Architecture:** Hexagonal (ports & adapters). Pure, deterministic Go core; hosts (CLI in M0) and concrete adapters at the boundary. Invariant rule: dependencies point inward; decision logic only in the core.

## Scope

**M0 (current) includes:**
- git-style CLI host (short sessions over persisted state)
- Understand/Act separation; structured `Task`
- Provider discovery (Places, behind a multi-source registry)
- Wizard of Oz negotiation (explicit FSM)
- Calendar reading; question and approval points
- 3-layer memory + `onit feedback`

**Explicitly out of M0:**
- Automatic WhatsApp sending (M1)
- Voice/calls, the business side, public signup, billing, in-app payment
- Multi-provider LLM (the architecture allows it; not in scope)
- Open agentic browsing and self-authored tools/prompts (human-gated, post-M0)

## Constraints

- Technical: the core does not import a concrete framework/CLI/driver, does not read env nor access the network/clock directly — everything via a port. Single-tenant now, multi-tenant-ready always (`user_id` on everything, per-User credentials).
- Resources: 1 developer; prioritizes learning Go deeply without stalling delivery.
- Process: deliberate — small tasks, review/understand each one, atomic commit, no vibe coding.
