# PRD — Onit (personal services agent) — Go edition

> **How to use this document.** Full-vision PRD, written to feed a Spec Driven
> Development skill. It describes the entire system, divided into
> **milestones (M0–M3)**. The skill should generate specs **only for the current
> milestone** — **M0** — never for the whole system at once. Anything marked M1+
> is directional context, not immediate scope.
>
> This edition replaces the previous PRD. The difference from it is the stack:
> the system is built in **Go**. The design principles, the milestones and the
> core/host architecture are language-independent and remain.

---

## 1. Executive summary

Onit is a **personal AI agent that solves day-to-day tasks** of local services.
The user describes a demand in free language ("I need to clean the sofa at home")
and the agent does the boring work: it finds providers, asks for quotes, checks
availability, cross-references it with the user's calendar and schedules the
service — escalating to the human only at two moments (an initial question if
critical information is missing, and the final approval) and in a third exception
case (when reality does not match the request).

The core problem: busy people procrastinate simple operational tasks because the
friction is not executing the task, it is **thinking** about it — who to call,
what to say, when to fit it in. Onit takes on that cognitive load.

The long-term vision is a **multi-user platform** and, eventually, a
**two-sided marketplace** where consumer agents and business agents negotiate
with each other. But the product is built from the inside out: first the personal
agent working for a single user, via CLI.

## 2. Verbal branding

Scope here is **verbal branding only** — name, meaning, tone of voice, tagline.
Visual branding (logo, palette, typography) is a future milestone and is out of
M0 on purpose: M0 is a CLI with no aesthetic surface, and visual identity is a
time sink that should not be opened now. What is in this section is a product
decision: it guides how Onit *talks* to the user and how the agent drafts
messages — needed already in M0.

### 2.1 Name

**Onit.** From "I'm on it" — the reply of a competent assistant who takes on the
task: "leave it with me, I'm already handling it". The name carries the product's
promise (competence and action, not postponement) and works as a tool name in
English, neutral enough for a Portuguese speaker.

### 2.2 Spelling

- Official writing: `onit`, all lowercase, in any context.
- The CLI binary is `onit`.
- Do not use stylized variations (OnIt, ONIT) — one spelling only, always.

### 2.3 Tagline

Primary: **"Consider it done."** Support: "I'm on it." The tagline closes the
meaning of the name for those who did not immediately catch the "on it" pun.

### 2.4 Tone of voice

Onit speaks like a competent, calm assistant — not like a robot, not like an
effusive friend:
- **Direct and calm.** Short sentences. It says what it did and what it needs,
  without embellishment.
- **Confident, never sycophantic.** "I found 3 providers, the highest-rated one
  charges R$180" — not "Great choice! I was happy to help!".
- **Honest about limits.** Faced with a mismatch, it states the problem clearly
  and offers options; it never pretends everything is fine.
- **No jargon.** The user does not see the words "task", "negotiation",
  "mismatch" — they see natural language.
- This applies to the CLI messages and, above all, to the text the agent drafts
  to send to providers: polite, objective, human.

## 3. Problem and user

### 3.1 Target user (M0)

A busy, urban person who postpones local service tasks (cleaning, locksmith, pest
control, small repairs). In M0 the only user is the developer themselves, in
Florianópolis/Palhoça (SC, Brazil).

### 3.2 Pains

- The friction of the task is cognitive, not physical — deciding who, what, when.
- Calls go to voicemail; the real response channel in Brazil is WhatsApp.
- Comparing prices and cross-referencing schedules manually is tedious and is
  therefore postponed.

### 3.3 What Onit delivers

The user describes the demand and the agent works. There are no "modes" to
choose: the same flow serves "list gardening services near here", "quote the
cleaning of my sofa" and "schedule a pest control" — the agent advances in the
flow as far as the request allows.

## 4. Design principles (non-negotiable)

These eight principles must be reflected in the architecture of any generated
spec. They are language-independent.

1. **Separate Understand from Act.** The free-text request is first converted
   into a structured object (`Task`). Only at this stage is the agent allowed to
   ask the user questions. Act (search, quote, check calendar) only begins when
   the `Task` is complete enough. This boundary makes the system auditable.

2. **Ambiguity is a legitimate state, not an error.** If an essential field is
   missing or has a double reading, the agent does NOT guess. It asks **one**
   question — the highest-impact one — and always offers a visible default. Rule:
   only ask what changes the action; what does not change, assume with a declared
   default.

3. **A mismatch always escalates to the human, it is never silenced.** When
   reality does not match the request (no provider in the region, all above the
   budget cap, divergent quotes, nobody replied), the agent presents the conflict
   already digested and with resolution options. It never "works it out" alone.

4. **The human is a thin decision layer.** On the happy path the user is
   interrupted only twice: a question at the start (if necessary) and the final
   approval. Everything between the two points the agent does alone.

5. **Show the work.** During Act, the host streams progress
   ("searching in Palhoça… found 4 places… asking for a quote…"). In the CLI,
   incremental text output. It turns waiting into trust.

6. **Single-tenant now, multi-tenant-ready always.** See section 8.

7. **Core decoupled from the host.** The business logic neither knows nor depends
   on where it runs or how the user talks to it. The core is a pure Go package;
   CLI, web and WhatsApp are just hosts that drive it. See section 9.

8. **Negotiation is an explicit state machine.** The `Negotiation` is modeled as
   named states and transitions, each transition triggerable by an event. This
   applies already in M0, even synchronous. See section 10.

## 5. Scope by milestone

### M0 — Personal agent, single-tenant, Wizard of Oz (CURRENT SCOPE)

Goal: prove that the flow really saves time, with the minimum of automation.

Included:
- Command CLI host (git style): create a task, list tasks in progress, inspect a
  task, record a provider's response. Each command is a short session over the
  persisted state.
- Agent loop with Understand/Act separation.
- Provider **discovery** tool via Google Places API.
- User **calendar reading** tool (Google Calendar).
- **Negotiation in Wizard of Oz mode**: the agent drafts the opening message for
  the provider and delivers it ready (text + phone) for the user to paste
  manually into WhatsApp. The user records the provider's response back (via a CLI
  command) and the agent interprets it.
- **Question** points (ambiguity) and **approval/escalation** points (mismatch).
- "Notification" in M0 = the list-tasks command makes evident what needs
  attention (e.g. "1 task awaiting you to record the response"). There is no
  active push in M0 — that depends on the persistent service of M1.
- Persistence of `Task`, `Negotiation` and messages in a database.

Explicitly out of M0:
- Automatic WhatsApp sending (enters in M1).
- Voice / phone calls (distant vision, perhaps never).
- Anything on the business side / provider registration.
- Public signup, billing, admin panel.
- Payment — in M0 and M1 payment happens in person between user and provider; the
  system only records the scheduling.
- Support for multiple LLM providers — M0 uses only Anthropic models (see section
  9.4). Switching providers is possible by architecture, but it is not M0 scope.

### M1 — Channel automation (asynchronous negotiation)

- Automatic sending and receiving via a WhatsApp API (e.g. Evolution API).
- The agent conducts the negotiation end to end without copy-paste from the user.
- The negotiation becomes **asynchronous**: the agent sends the message and the
  response comes back minutes or hours later. It requires the event-oriented
  model of section 10 — there is no "live" process; the state sleeps in the
  database and is woken by triggers.
- Requires the **persistent service** (section 9) up: a stable address to receive
  webhooks and a reliable scheduler for the follow-up jobs.

### M2 — Multi-user platform (consumer side)

- Open signup, authentication, per-user data isolation.
- Each user connects their own Google (OAuth) and their own WhatsApp.
- Trust handling: what happens when the agent acts on behalf of a user and
  something goes wrong.

### M3 — Two-sided marketplace

- Business registration, business agents, agent-to-agent negotiation protocol
  intermediated by a central broker.
- PLG loop: a business publishes a scheduling link on Google Maps; visible demand
  attracts new businesses.

> The SDD skill generates specs **only for M0**. M1–M3 exist here to ensure that
> M0 architecture decisions do not block the future.

## 6. Functional flow of M0

1. **Input.** The user creates a task describing the demand in free text, via a
   CLI command.
2. **Understand.** The agent extracts a structured `Task`. If an essential field
   is missing or ambiguous, it asks one question (with a default) and goes back to
   step 2.
3. **Act.** With the `Task` complete: it searches providers (Places API),
   assembles the quote messages, cross-references availability with the calendar.
   The CLI streams the progress in text.
4. **Review.** The agent compares the result with the `Task`. If there is a
   mismatch (region, price, time, silence), it escalates to the user with options.
5. **Approve.** The user approves via a command (or adjusts).
6. **Confirm.** The scheduling is recorded and added to the calendar.

Between the steps, the user closes the terminal freely: the state persists in the
database. List and inspect commands give the "snapshot" of the state at any
moment.

## 7. Data model (high level — the spec must detail)

- **`User`** — owner of everything. In M0 there is a single record. NEVER
  hardcode the user's data (calendar, phone, search radius, preferences);
  everything is a `User` attribute.
- **`Task`** — the structured demand. Minimum fields to specify: service type,
  location, acceptable time window, budget cap, constraints, and the task state.
  Each field must have defined: whether it is required, what the default is, and
  the "ask vs. assume" rule.
- **`Negotiation`** — state machine (section 10). Sequence of typed messages
  exchanged with a provider. The auditable heart of the system; it must be well
  modeled before the agents.
- **`Provider`** — a found provider (in M0, derived from the Places API result;
  with no calendar nor agent of their own).
- **`Message`** — individual messages, always tied to a `Negotiation`.

Every transactional entity (`Task`, `Negotiation`, `Message`) carries `user_id`.
No query reads data without filtering by owner.

In Go, model states and variants with explicit types (named types,
constants/enum, structs per variant) so that invalid states are as hard as
possible to represent.

## 8. Multi-tenant discipline (mandatory already in M0)

M0 is built for one user, with three disciplines that avoid rewrites:

1. **Nothing hardcoded.** Calendar, phone, search radius, preferences —
   everything is a `User` attribute, loaded by `userID`.
2. **Owner in everything.** Every task/negotiation/message has `user_id`; every
   query filters by it.
3. **Per-user credentials, not the system's.** The Google Calendar token is kept
   in a table tied to the `User`, not in a configuration file — even with a single
   user. This way the M2 "connect my Google account" is just repeating an existing
   flow.

## 9. Architecture: core, host and persistent service

Three layers with explicit boundaries. This separation is the most important
architecture decision of the project, and in Go it is expressed naturally with
interfaces and packages.

### 9.1 The core

A pure Go package with **all** the business logic: understanding the `Task`, the
agent loop, the `Negotiation` state machine, model routing, the decisions, the
mismatch handling.

Inviolable rules of the core:
- It imports nothing from a web framework, a CLI, a concrete database driver, nor
  a telephony/WhatsApp SDK.
- It does not read environment variables, does not access the clock nor the
  network directly.
- Everything that is "outside world" (database, LLM, Places, calendar, message
  channel, clock) enters the core as an **interface (port)** that the host
  implements and injects. In Go: the core declares the `interface`; the adapter is
  the concrete type that satisfies it. Ports & adapters / hexagonal architecture
  pattern.
- The core is deterministic and testable without the network: with fake in-memory
  adapters, all the logic runs in unit tests. In Go this is idiomatic — small
  interfaces, fakes trivial to write.
- The dependency rule: all dependencies point inward, toward the core. The core
  does not import any host or adapter package.

### 9.2 The hosts

A host gives the core a concrete environment: it implements the ports, connects
user input and output, decides where it runs. The same core serves several hosts:
- **CLI host** — the M0 host. Command CLI in git style (`onit new`,
  `onit tasks`, `onit task <id>`, `onit reply <id>`), not a conversational loop.
  Each command is a short session that reads or nudges the state in the database
  and ends. Chosen for M0 because it keeps the focus on the logic, has no visual
  surface to polish, and is the most honest host to validate that the core is
  decoupled. Distributable as a single compiled binary.
- **Web host** (later milestone) — not M0 scope.
- **WhatsApp host** (M1+) — receives and sends messages over the channel.

Each CLI command works as an **event adapter**: it triggers a transition in the
core's state machine (section 10). It is the same role that a webhook and a
scheduled job will have in M1 — the command CLI is the system's first event
adapter.

Hosts are thin: they orchestrate ports and UI, they do not contain business
logic. If a decision rule appears in a host, it is in the wrong place — it belongs
to the core. "No UI" does not apply even to the CLI: the set of commands, what
each one displays and how errors and pending items appear are interaction design —
finite, guided by conventions (git, docker), and part of M0 scope.

### 9.3 The persistent service

The central, managed home of the core: it receives webhooks, runs scheduled jobs,
keeps the state, runs the sweeper (section 10). It is **managed by the team**,
never self-hosted by a lay user — asynchronous negotiation requires a stable
address and a reliable scheduler always up, which a personal machine does not
offer. Self-hosting remains possible only as a CLI host mode for the technical
audience that wants it; here the single Go binary is a real distribution
advantage.

In M0 the persistent service is minimal (basically the local backend of the CLI
host). It grows in M1, when it starts hosting the asynchronous negotiation. Go's
concurrency model (goroutines) is suitable for the persistent service to handle
many negotiations in parallel, but that is an M1+ concern, not M0.

### 9.4 Reference stack (Go)

The language decision is **Go**, end to end. Justification: it is the language the
developer wants to deepen; the Go AI ecosystem in 2026 is mature enough for Onit's
scope (there is no real technical limitation); and Go is strong for CLIs and
concurrent network services. There is no stack fragmentation — one runtime, one
language.

- **Language:** Go (current stable version). Core, CLI host and persistent
  service, all in Go.
- **LLM (M0):** the official Anthropic SDK for Go (`anthropic-sdk-go`). It is
  complete and stable, and already includes a ready agent loop (toolrunner). Known
  and accepted limitation: it talks only to Anthropic models — aligned with M0,
  which routes only among Anthropic models (Haiku for classification and cheap
  tasks, Sonnet for negotiation reasoning).
- **Multi-provider — explicit decision:** do NOT adopt a third-party
  multi-provider framework in M0. The capacity to switch models is provided by the
  architecture itself: the core talks to the `PortLLM` interface; the Anthropic
  SDK is an adapter of that interface. Switching or adding providers in the future
  is writing another adapter, without touching the core. Frameworks like Eino
  (ByteDance/CloudWeGo, the most solid) or LangChainGo remain a future adapter
  option, if and when the need is concrete — not as an M0 dependency.
- **Agent loop:** use the official SDK's toolrunner OR implement the loop in the
  core. Implementing the loop is a valid learning goal for the developer; if it is
  done, it lives in the core, behind `PortLLM`, and does not depend on a
  framework. Do not reimplement the HTTP transport SDK — use the official one.
- **Database:** Postgres. Accessed by the core only via `PortPersistence`. RAG,
  when it enters, is over Postgres (pgvector) — more own code over a Postgres
  client than a dependency on a RAG framework.
- **Discovery:** Google Places API, behind `PortDiscovery`.
- **Calendar:** Google Calendar API (OAuth), behind `PortCalendar`.
- **Channel (M1+):** WhatsApp API (e.g. Evolution API), behind `PortChannel`.
- **CLI:** an idiomatic Go command-parsing library; a single compiled binary.
- **Dependency principle:** prefer the standard library and few well-maintained
  dependencies. An SDK that concentrates credentials is an attack surface; fewer
  transitive dependencies is better for security. Evaluate each dependency by
  active maintenance and backing (prefer projects with a team/company behind them
  over single-maintainer projects on the critical path).

## 10. The `Negotiation` as an event-oriented state machine

The `Negotiation` is modeled as **named states and named transitions**, already
in M0. Each transition is triggered by an event. In M0 all events arrive
synchronously (Wizard of Oz, user in the conversation); in M1 the same events
start arriving from a webhook and a scheduled job — without rewriting the core.

### 10.1 Conceptual model

- The **state** of a negotiation lives in the database. Between events, nothing
  "runs" — the state sleeps, at zero cost.
- Two types of **trigger** wake a negotiation:
  - **Reactive event** — the provider replied (in M0: the user records the
    response via the CLI; in M1: WhatsApp webhook).
  - **Scheduled job** — the expected time has passed (e.g. follow-up if nobody
    replied).
- On each trigger, the agent: loads the negotiation → reads the state → decides
  **one** step → writes the new state → goes back to sleep. The agent is never
  "live".
- A **sweeper** (periodic job) is the safety net against a lost webhook: it sweeps
  negotiations stalled too long without an event and reconciles. The only thing
  that runs on a fixed interval; it does not conduct a negotiation, it only
  recovers what escaped.

### 10.2 Reference states (the spec must refine)

`draft` → `awaiting_response` → (`counteroffer` ⇄ `awaiting_response`)
→ `human_approval` → `confirmed`. Additional terminals: `declined`,
`expired`. Each arrow is a named transition with an event that triggers it.

### 10.3 Implication for M0

Even though M0 is synchronous, the `Negotiation` must be implemented as explicit
states and transitions — not as a procedural flow. If procedural, M1 is a rewrite;
if a state machine, M1 only adds the trigger adapters (webhook and scheduler). In
Go, representing the states as a named type and the transitions as explicit
functions makes the set auditable.

## 11. Agent tools (M0)

The spec must define the input/output contract of each one (in Go, explicit
parameter and return types):

- `find_providers` — queries the Places API by category + geolocation + user's
  radius; returns providers with name, phone, rating, hours.
- `draft_message` — generates the opening/quote text for a provider.
- `record_response` — receives the pasted text of the provider's response and
  interprets it (price, availability, conditions).
- `read_calendar` — checks the user's availability in a window.
- `ask_decision` — escalates to the user: an ambiguity question or a mismatch
  conflict, always with options/defaults.

## 12. Success metrics (M0)

- The flow completes a real task end to end (e.g. scheduling a sofa cleaning).
- The user is interrupted at most twice on the happy path.
- The agent never "invents" a nonexistent provider, price or time — mismatches are
  always escalated.
- Subjective perception: Onit really saved time and mental load.

## 13. Risks and open questions

- **Uneven service coverage.** Categories well mapped on Google (cleaning,
  locksmith, pest control) will work well; niches will return little. Decision:
  start with a few high-quality categories instead of promising "any service" and
  disappointing.
- **Calibration of "when to ask".** Asking too much annoys; too little errs. The
  rule (only what changes the action) needs to be validated in practice.
- **Erosion of the core boundary.** The most likely long-term risk: business logic
  leaking into the hosts for convenience. Every spec treats the core boundary as
  an invariant; code review refuses decision logic outside the core.
- **Maturity of Go dependencies.** The Go AI ecosystem is viable, but it mixes
  solid projects (official Anthropic SDK, Eino) and single-maintainer projects
  (LangGraph/RAG ports). Decision: on the critical path, only well-backed
  dependencies; fragile projects, if used, stay behind a port and are swappable.
- **Learning vs. delivery.** The desire to deepen Go and the AI machinery is
  legitimate, but it must not stall the product. Study reimplementations (e.g. an
  own SDK) stay in a separate project, behind a port; the `onit` of M0 uses the
  official tools.
- **Trust when acting on behalf of third parties.** A real platform problem (M2+),
  not M0 — do not solve it too early.

## 14. Final instruction for the SDD skill

Generate specs **only for M0**. For each spec:
- Respect the eight design principles of section 4 as architecture invariants.
- Treat the Understand/Act separation as a structural boundary, not a detail.
- Treat the **core / host** boundary (section 9) as an invariant: make clear what
  is core (pure logic) and what is host (concrete adapter). Every access to the
  outside world — database, LLM, Places, calendar, channel, clock — is a port (Go
  interface) with an explicit contract.
- Model the `Negotiation` as an explicit state machine (section 10), even in
  synchronous M0.
- Start with the data modeling (`User`, `Task`, `Negotiation`, `Provider`,
  `Message`) and with the port contracts (Go interfaces) before any agent spec.
- Include `user_id` and per-user credentials from the first spec.
- Assume Go as the language; the M0 LLM is the official Anthropic SDK behind
  `PortLLM`; do not specify multi-provider support, automatic WhatsApp sending,
  voice, the business side, public signup or billing — future milestones.

---

## 15. Addendum — Extensible discovery, schema with an open boundary and memory

> **Status.** Addendum that amends sections 7, 9.4, 11 and 13. It does not change
> the eight design principles (section 4), the core/host boundary (section 9), the
> `Negotiation` state machine (section 10) nor the scope of the milestones
> (section 5): everything here is **port contract** and **data modeling**
> decisions, which section 14 already mandates fixing before any agent spec. The
> effect on M0 scope is small — what increases is the precision of the boundaries,
> not the work surface.

### 15.1 The principle that stitches the addendum together: data evolves on its own, code does not

Onit must be able to improve with use, but the product acts **outward**, over real
providers and money. Hence the rule:

- **Online self-improvement is only data** — preferences, memories, playbooks,
  learned attributes. It is safe, runs without supervision, touches neither the
  code nor the DDL.
- **Harness evolution is offline and human-gated** — the agent may *propose* a new
  tool, a prompt adjustment or a playbook (in the style of a PR), but the change
  only lands after human review. An agent that rewrites its own code or prompt at
  runtime, acting over third parties, is **out of M0** and remains gated after it.

This rule keeps alive principles #1 (separate Understand/Act, auditable),
#3 (a mismatch always escalates) and the metric of section 12 (the agent never
invents a provider, price or time).

### 15.2 Multi-source discovery (amends 9.4 and 11)

Discovery stops being tied to Google Places. `PortDiscovery` becomes a
**registry of sources**, not a single call:

- The core declares a `DiscoverySource` interface that each source satisfies
  (Google Places in M0; web search, social networks or a sub-agent that navigates
  as future adapters). The core talks to **one** port and does not know how many
  sources exist.
- The registry fans out to the relevant sources, **normalizes** the results into a
  canonical `Provider` and **deduplicates/merges** the same provider found in
  different sources.
- **Mandatory provenance.** Every `Provider` carries `source`, a `confidence` and
  the raw evidence of where it came from. This is what keeps the "never invents" of
  section 12 when discovery moves beyond Places: the agent can say "I found it on
  Instagram, but I did not confirm the phone" instead of treating everything as
  truth.

The `find_providers` tool (section 11) starts talking to the registry —
transparently for the core.

**M0 scope:** start **only with the Google Places adapter**. The interface already
admits the other sources; the M0 metric is to close a real task end to end, not to
cover five sources. Open agentic browsing is a future source, behind the same
port.

### 15.3 Schema with a typed spine and an open boundary (amends 7)

The data model is deliberately split into two zones, to reconcile "invalid states
hard to represent" (section 7) with extensibility:

- **Typed and rigid spine (immutable at runtime).** `user_id`, the `Task`'s
  `state`, the `Negotiation` FSM, service type and location. It is what the
  queries, the state machine and the multi-tenant isolation depend on. In Go:
  named types, enums and structs per variant. The DDL of this zone does **not**
  change at runtime; the agent never runs `ALTER TABLE`.
- **Open boundary (data, not DDL).** The long tail of descriptive attributes that
  each service requires (sofa fabric, square meters, has pets, floor without an
  elevator). It lives in a semi-structured bag (`attributes jsonb` on the `Task`,
  or a `TaskAttribute` table). The agent adds attributes here as it learns, as
  data — without a change to the physical schema. It keeps carrying `user_id`.

The `Provider` gains the provenance fields of section 15.2 (`source`,
`confidence`, evidence).

### 15.4 Three-layer memory (new port `PortMemory`, amends 9.4 and 11)

The core gains a new port, `PortMemory`, symmetric to the others and implemented
over Postgres + pgvector (already foreseen in 9.4). A trivial in-memory fake for
the core tests. Three layers, each with `user_id`:

- **Factual** — how the domain works and what has already been observed: the typed
  model, plus episodic and provider memory (prices seen, who replied fast, who was
  reliable). It composes over time and makes the agent smarter about the local
  reality.
- **Behavioral** — editable instructions (per-`User` and global) loaded into the
  system prompt **on every turn** of the agent. It is the layer the user corrects
  as they would correct an employee.
- **Procedural** — playbooks per service category, stored **as parameterized data**
  (not as compiled code): "for pest control, ask about pets"; "for sofa cleaning,
  capture fabric and square meters". Evolving them is editing data; creating a
  truly new playbook goes through the human gate of section 15.1.

**New tool and command:** `record_feedback`, exposed in the CLI as
`onit feedback "<text>"`. The agent distills the user's correction and appends it
to the behavioral memory, which enters the next runs. This is the only
self-improvement muscle that enters M0 — and it is safe because it is data driven
by human feedback, aligned with the "nothing hardcoded" of section 8.

### 15.5 Amendment to the risks (section 13)

- Multi-source discovery (15.2) directly attacks the risk **"uneven service
  coverage"**: a niche poorly mapped on Google stops being a hole when other
  sources enter as adapters.
- New guard, anchored in the risk **"Learning vs. delivery"**: M0
  self-improvement is limited to data (memory, playbooks, attributes).
  Self-authored tools and prompts by the agent stay out of M0 and, when they enter,
  require human approval. This prevents the desire for an "agent that evolves
  itself" from stalling delivery or eroding auditability when acting over third
  parties.

### 15.6 Summary of the architecture delta

The hexagon gains **one new port and one richer port**, without piercing the
core/host boundary:

- `PortDiscovery`: from a single call → a registry with fan-out, dedup and
  normalization over `DiscoverySource[]`.
- `PortMemory` (new): three layers (factual/behavioral/procedural) over pgvector.
- `Task`: typed spine + open attributes (`jsonb`).
- `Provider`: + provenance (`source`, `confidence`, evidence).

Unchanged: the pure and deterministic core (testable with fakes, including of the
two new ports), the Understand/Act separation, the `Negotiation` FSM, the
multi-tenant discipline and the single binary.

### 15.7 Additional instruction for the SDD skill

- Model `PortDiscovery` as a registry of `DiscoverySource` and `PortMemory` with
  the three layers, both with an explicit Go contract and an in-memory fake,
  **before** the agent specs.
- Specify the `Task` with the two zones (typed spine vs. open attributes) and the
  `Provider` with the provenance fields.
- Include the `record_feedback` tool and the `onit feedback` command.
- Do **not** specify: open agentic browsing as an active source, an agent that
  writes/compiles its own tools, or an agent that rewrites its own prompt — future
  milestones, always human-gated.
