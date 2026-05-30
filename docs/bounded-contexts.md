# Domain model — onit M0 (DDD strategic design)

> Strategic-DDD analysis of the domain, derived from `prd.md` (§4, §6–§7, §10–§11, §15).
> Output of the `domain-analysis` skill, adapted to greenfield: the input is the PRD
> and the ports of §9, not an existing codebase. It defines how `/internal/core` is
> organized **internally** — the Foundation spec keeps the data model flat; this fills
> that gap. Tactical patterns (rich aggregates) are applied per-aggregate at
> implementation time via the `tactical-ddd` skill. Date: 2026-05-29.
>
> Scope note: this is **light** strategic DDD, calibrated for an M0 single-user CLI.
> Bounded contexts here are **Go packages, not separate services/modules**. The
> boundaries are drawn so the system can *grow into* M1/M2 (separate persistence,
> ACLs, event bus) without paying that cost now. On any conflict, `prd.md` wins.

## Subdomain classification

| Subdomain | Type | Rationale (from the PRD) |
| --- | --- | --- |
| **Understanding** (Task intake) | 🟥 Core | Principles #1/#2: free text → `Task`, "ambiguity is a legitimate state", ask-vs-assume. The differentiator. |
| **Negotiation** | 🟥 Core | "The auditable heart of the system; must be well modeled before the agents" (§7). |
| **Agent Orchestration** (decision/mismatch policy) | 🟥 Core (application layer) | Principle #3 + the loop in the core (ADR-003). Not a context — the *policy* that crosses contexts. |
| **Discovery** | 🟧 Supporting | Essential but feeds the core; value is in the *use*, not the search. Provenance adds a differentiating edge. |
| **Scheduling** | 🟧 Supporting | Needed to close the task; domain logic of cross-referencing the calendar. |
| **Memory / Learning** | 🟧 Supporting | Makes the agent smarter over time; business-specific (§15.4). |
| **Identity & Credentials** | 🟨 Generic (Shared Kernel) | `user_id` everywhere + OAuth token (AES-GCM). Standard, but sensitive. |
| LLM / Calendar API / Channel | 🟨 Generic (transport) | Live in *adapters*, not domain. |

> The **Calendar API** (Google mechanics) is generic and lives in the adapter; **Scheduling**
> (the decision to fit a slot) is supporting and lives in the core. Keep them separate.

## Bounded contexts and integration

```
                    ┌─────────────────────────────────────────────┐
                    │   Agent Orchestration (core app / policy)    │
                    │   one-step loop · when-to-ask · mismatch      │
                    └──────┬───────┬───────┬───────┬───────┬───────┘
            creates/reads  │       │       │       │       │  consults
            ┌──────────────▼─┐ ┌───▼────┐ ┌▼──────┐ ┌▼───────┐ ┌▼────────┐
            │ Understanding  │ │Negotia-│ │Disco- │ │Schedu- │ │ Memory  │
            │  (Task) 🟥     │ │tion 🟥 │ │very🟧 │ │ling 🟧 │ │   🟧    │
            └────────────────┘ └────────┘ └───────┘ └────────┘ └─────────┘
                    └──────────── Identity / Tenancy (Shared Kernel 🟨) ──┘
```

| From → To | Integration pattern | Concrete rule |
| --- | --- | --- |
| Understanding → Negotiation | Customer/Supplier | A `Negotiation` is born from a complete `Task`; references **`TaskID`**, never embeds `Task`. |
| Discovery → Negotiation | Published Language | Discovery publishes the canonical `Provider` (with provenance); Negotiation references **`ProviderID`**. |
| Negotiation → Scheduling | Domain Event | `NegotiationConfirmed` triggers the booking. M0 synchronous = direct call from orchestration; M1 = real event. |
| Memory → (Understanding, Discovery, Agent) | Open Host (port) | Consumed **read-mostly** via `PortMemory`; behavioral memory enters the prompt on every turn. |
| Identity → all | Shared Kernel | `UserID` + credential access. Keep it **minimal** (identity/tenancy only). |

## Cohesion cautions (onit-specific)

1. **Do not spread the decision policy** (when-to-ask, mismatch) across contexts — it lives
   **only** in `agent`. Directly mitigates PRD risk #1 ("erosion of the core boundary"). Each
   context exposes data/transitions; *who decides* is the orchestration.
2. **Memory can become a god-context** (it touches everything). Guard: three typed layers behind
   the port, no business logic inside — a knowledge store, not a decider.
3. **`Provider` has a dual meaning** (Discovery candidate vs Negotiation counterparty). Resolve by
   **reference by ID**; Discovery is the owner/publisher.
4. **`Task` is the orchestration spine.** Owner = Understanding; others reference it by ID.
   `Task.state` transitions are driven by the orchestration, but the enum + guards live with `Task`
   (typed spine, §15.3).

## `/internal/core` layout (contexts = packages, not services)

```
/cmd/onit                       # CLI host (cobra) — thin; only orchestrates ports and UI
/internal/core
├── identity/                   # 🟨 Shared Kernel: User, UserID, tenancy, credential types
├── understanding/              # 🟥 Task (aggregate: typed spine + attributes), ask-vs-assume policy
├── negotiation/                # 🟥 Negotiation (aggregate + FSM named-type), Message
├── discovery/                  # 🟧 Provider (+provenance), DiscoverySource contract + registry
├── scheduling/                 # 🟧 availability, booking decision
├── memory/                     # 🟧 factual/behavioral/procedural layers, playbooks
└── agent/                      # 🟥 orchestration: one-step loop, tool dispatch, mismatch policy
/internal/adapters
├── postgres/   anthropic/   places/   gcal/   memory/      # 🟨 implement the ports
```

### Where ports live (ADR-008)

Following ADR-006 (interface near its consumer, no separate `ports/` package):

- **Context-specific ports** live in their package: `discovery.Source`/`discovery.Port`,
  `memory.Port`, `scheduling.CalendarPort`.
- **Cross-cutting ports** (`PortLLM`, `PortPersistence`, `PortClock`), used by several core
  packages, are declared in a thin root `core` package — avoids import cycles between contexts
  and a duplicated interface per consumer. Locked as **ADR-008** in `docs/decisions.md`.

## How this feeds the workflow

- **Design phase** (tlc-spec-driven, Foundation): this map is the input for organizing `/internal/core`.
- **Layout:** `golang-project-layout` materializes the tree above.
- **Implementation:** `tactical-ddd` applies per-aggregate when coding `Task`, `Negotiation`,
  `Provider` — keep them rich (`Task.Clarify()`, `Negotiation.RecordResponse()`), not data bags
  with setters.
