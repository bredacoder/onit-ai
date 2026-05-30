# Roadmap

**Current Milestone:** M0 — Personal agent, single-tenant, Wizard of Oz
**Status:** In Progress

> Milestones derived from `prd.md` (section 5). The skill generates specs **only for M0**.

---

## M0 — Personal agent (single-tenant, Wizard of Oz)

**Goal:** Prove that the flow truly saves time, with the minimum of automation — one real task closed end to end via CLI.
**Target:** Full flow (create task → understand → search/quote → review → approve → confirm) working for one user.

### Features

**Foundation (hexagonal wiring)** - IN PROGRESS
- Go project skeleton + local dev (docker-compose pgvector, Makefile, goose)
- Data model: `User`, `Task`, `Negotiation` (FSM), `Provider`, `Message` — typed spine + open `jsonb` attributes; `user_id` on everything
- All port interfaces with in-memory fakes (`PortLLM`, `PortPersistence`, `PortDiscovery` registry, `PortCalendar`, `PortMemory`)
- One CLI command (`onit tasks`) wired end to end proving the wiring

**Understand** - PLANNED
- Free text → structured `Task`; ambiguity question with a default

**Discovery** - PLANNED
- Google Places adapter behind the multi-source registry; `find_providers` tool

**Negotiation (Wizard of Oz)** - PLANNED
- FSM `draft → awaiting_response → … → confirmed`; `draft_message`, `record_response`

**Calendar + Approval** - PLANNED
- `read_calendar`, scheduling, approval point/mismatch escalation

**Memory + Feedback** - PLANNED
- `PortMemory` (factual/behavioral/procedural) + `onit feedback` command

---

## M1 — Channel automation (asynchronous negotiation)

### Features
**Automatic WhatsApp (Evolution API)** - PLANNED
**Persistent service (webhooks + scheduler + sweeper)** - PLANNED

## M2 — Multi-user platform (consumer side) - PLANNED

## M3 — Two-sided marketplace - PLANNED

---

## Future Considerations

- The agent's own loop has been in the core since M0 (ADR-003); discovery sources beyond Places come in as adapters.
- Open agentic browsing and evolution of a self-authored harness (always human-gated).
- Multi-provider LLM via a new `PortLLM` adapter.
