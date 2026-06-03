---
name: pr-review
description: Multi-agent PR reviewer for onit (Go, hexagonal). Use ONLY when explicitly asked to review a pull request, or when invoked by CI on a PR: "review PR #N", "review this PR", "code review", "check this pull request". Do NOT trigger automatically during coding, feature implementation, or general questions.
license: CC-BY-4.0
metadata:
  author: onit
  version: 1.0.0
---

# PR Review — Orchestration Protocol (onit)

Coordinates 6 specialized subagents (via the Task tool) then consolidates findings into a
unified summary. Each subagent loads the relevant **onit** docs and `golang-*` skills as its
rule source — this skill does not duplicate them. The contract reviewed against is, in priority
order: `docs/prd.md` (source of truth) → `.specs/` → `docs/decisions.md` →
`docs/bounded-contexts.md` → `CLAUDE.md` invariants.

## Step 1: Initialize

1. Get PR number from context or ask the user.
2. Identify repo: `gh repo view --json nameWithOwner -q .nameWithOwner`
3. Fetch diff: `gh pr diff {PR_NUMBER}`
4. Load existing inline comments: `gh api repos/{REPO}/pulls/{PR_NUMBER}/comments` — build a set of `{path, line}` pairs to avoid reposting.
5. Read PR intent: `gh pr view {PR_NUMBER} --json title,body,headRefName`
6. From the PR title/body/branch, identify the **feature under review**: match against `.specs/features/*/` (fuzzy match on the directory stem, e.g. `foundation`) and note the spec path.

## Step 2: Launch Subagents in Parallel

Send **one message** with **six Task tool calls** — all launched simultaneously. Pass REPO,
PR_NUMBER, the diff, existing comment locations, the PR intent, and the matched spec path to each
subagent prompt. After all complete, run Step 3.

---

## Severity Labels (all subagents use these)

- 🚨 Critical — bugs, logic errors, or **invariant violations** that will cause failures or break the architecture
- 🔒 Security — security vulnerabilities or data exposure
- ⚡ Performance — significant performance or concurrency concerns
- ⚠️ Warning — code smells or maintainability issues
- 💡 Suggestion — optional improvements

---

## Universal Rules (every subagent must follow)

1. **Comment allowlist:** Only post inline comments on lines in the diff starting with `+` (excluding `+++`). The quoted evidence must be the **current `+` line content as it exists at HEAD**, not the surrounding diff-hunk context and never a removed (`-`) line.
2. **Skip duplicates:** If `{path, line}` within ±3 lines already has a comment, skip.
3. **Mark resolved:** Reply `[RESOLVED] This appears resolved by the recent changes.` on existing comments where the issue is fixed.
4. **False positive guard:** Only report findings with ≥80% confidence. Skip when uncertain.
5. **Verify against HEAD before posting.** Open the actual file at the current checkout and confirm the problem is *literally present on the cited line*; quote that present line as evidence. If it does not reproduce in the working tree, **drop it**. NEVER derive a finding from a removed (`-`) diff line, from an existing review comment, or from prose in docs/specs (e.g. `STATE.md` lessons-learned). A diff that *fixes* a bug still contains the old broken code on `-` lines — that is not a finding.
6. **Positive highlight:** Include at least one well-done aspect of the change before listing issues.
7. **Tone:** Specific, actionable, collegial, **in English** (project convention). Explain WHY something is a problem.
8. **Never** approve, request-changes, or modify files. Use `--comment` only.
9. **Marker:** Start every inline comment body with `<!-- onit-review:{type} -->` (invisible in rendered view, used by the consolidation subagent).
10. **Cite the source of the rule** (PRD §, ADR, spec requirement ID, or `golang-*` skill) so the author can verify — onit's principle is "review/understand every line, no vibe coding".
11. **Severity gate on posting:** Inline comments are reserved for 🚨 Critical and **confirmed** 🔒 Security / ⚡ Performance findings. ⚠️ Warning and 💡 Suggestion are aggregated into the summary comment only — **never posted inline**.
12. **Linter-first:** Do not report anything a formatter/linter/compiler already enforces (`gofmt`/`golangci-lint`: trailing newline, import order, `t.Parallel()` suggestions, formatting/style). `make lint` runs separately. Focus on logic, invariants, security, requirements, and concurrency.

---

## Subagent 1: Hexagonal Invariants & Architecture

**Marker:** `<!-- onit-review:architecture -->`

### Phase 0 — Load the rule sources

Load before touching the diff: `CLAUDE.md` (§ Architecture invariants), `docs/decisions.md`
(locked ADRs), `docs/bounded-contexts.md` (the `/internal/core` domain map). Also consult the
`golang-project-layout`, `golang-structs-interfaces`, and `golang-design-patterns` skills.

### Phase 1 — Build the invariant checklist

Extract every hard rule into a numbered checklist. The non-negotiable invariants (from `CLAUDE.md`)
are the spine — at minimum:

1. **Dependencies point inward.** Adapters import the core; `/internal/core` imports **nobody**
   (no `/internal/adapters/...`, no driver, no SDK). Flag any import in a `core` file that points outward.
2. **Core is pure & deterministic.** The core never touches env, network, clock, or a concrete
   framework/driver directly — only through a **port interface** declared in the core.
3. **Agent loop lives in the core**, provider-agnostic. `PortLLM` is single-turn
   (`Complete(req) -> text | tool_calls`); `anthropic-sdk-go` is transport only — its toolrunner must
   NOT be used (ADR-003).
4. **Multi-tenant-ready always:** `user_id` on every row; per-`User` credentials; secrets encrypted
   at rest with AES-GCM, key via env (PRD §8, ADR-004). Flag new tables/queries missing `user_id`.
5. **Understand before Act:** free text → structured `Task` first; the agent never invents a
   provider/price/time — mismatches escalate (PRD §4).
6. **Layout:** new code lands in the right place — `/cmd/onit` (host), `/internal/core` (core +
   ports), `/internal/adapters/{postgres,anthropic,places,gcal,memory}` (adapters).
7. **No ORM:** SQL is plain, reviewed `pgx` + `sqlc`; migrations via `goose` (ADR / stack).

Also extract any rule stated in `docs/decisions.md` and `docs/bounded-contexts.md` (package
boundaries, where a subdomain lives). Number them sequentially. Do not invent rules not in the docs.

### Phase 2 — Evaluate the matrix

Work the diff **one file at a time**. For each changed file and each rule: **PASS / VIOLATION / N/A**
(N/A only when structurally inapplicable). For every VIOLATION, post an inline comment on the exact
`+` line that is the evidence, with the rule number and source.

**Comment format:**
```
<!-- onit-review:architecture -->
[🚨/⚠️/💡] — [Short title]
Invariant: [Rule number + source, e.g. "Rule 1 — CLAUDE.md: core imports nobody"]
[What in the diff violates it — quote the offending line]
**Recommendation:** [Exact fix, code snippet if < 6 lines]
```

---

## Subagent 2: Requirements & Definition of Done (vs spec)

**Marker:** `<!-- onit-review:requirements -->`
**Posts:** One PR-level summary comment only — no inline comments.

Requirements live in the repo (no Jira). Resolve them in priority order:

1. **Matched feature spec:** read `.specs/features/{feature}/spec.md`, `design.md`, and `tasks.md`.
   Extract acceptance criteria, requirement IDs, the task checklist, and stated non-goals.
2. **PRD:** read the relevant section of `docs/prd.md` (the source of truth — wins on any conflict)
   for the feature under review. Cross-check the spec against it.
3. **State:** read `.specs/project/STATE.md` and `ROADMAP.md` to confirm the PR is in scope for the
   current milestone (M0) and not pulling in deferred work.

If no spec or PRD section maps to the PR, post:
`⚠️ No matching spec or PRD section found — requirements verification skipped.` and stop.

Compare the merged requirements against the PR diff. Post the summary with
`gh pr comment {PR_NUMBER} --body '...'`.

**Summary format:**
```markdown
<!-- onit-review:requirements -->
## 📋 Requirements Review

**Source:** {e.g. "Spec: .specs/features/foundation/spec.md + PRD §9"}
**Milestone scope:** {in scope for M0 | ⚠️ pulls deferred work}

### ✅ Implemented
### ❌ Missing or Incomplete
### 🔲 Definition of Done
- [x] covered  - [ ] not covered
### 💬 Notes
```

---

## Subagent 3: Security

**Marker:** `<!-- onit-review:security -->`

Use the `golang-security` skill as the rule source. Review the diff for: hardcoded secrets or API
keys, **AES-GCM misuse** (reused nonce, key not from env, secret stored plaintext — PRD §8 / ADR-004),
missing per-`User` credential scoping, PII in logs, SQL built by string concatenation instead of
parameterized `pgx`/`sqlc` queries, missing `user_id` filter exposing other tenants' rows, unvalidated
external input (LLM tool args, Places/gcal responses), and OAuth token mishandling.

**Comment format:**
```
<!-- onit-review:security -->
🔒 Security — [Short title]
[What the issue is and why it matters — cite PRD §/ADR or golang-security]
**Recommendation:** [Specific fix]
```

---

## Subagent 4: Test Coverage

**Marker:** `<!-- onit-review:tests -->`

Use the `golang-testing` (and `golang-stretchr-testify`) skill as the reference for what correct tests
look like. The onit core is pure and deterministic — it **must** be unit-testable without infrastructure
(ports are interfaces; use fakes/mocks, not a real DB/LLM). Review the diff for: new core logic with no
table-driven unit test (🚨 Critical), adapters with no test against their port contract, tests that hit
real network/DB instead of a fake (architecture smell), missing error-path coverage, and anti-patterns
(no `t.Parallel()` where safe, hardcoded values, assertions that don't assert behavior).

**Comment format:**
```
<!-- onit-review:tests -->
[🚨/⚠️/💡] — [Short title]
[The coverage gap or test anti-pattern]
**Recommendation:** [Pattern to follow per golang-testing]
```

---

## Subagent 5: Performance & Concurrency

**Marker:** `<!-- onit-review:performance -->`

Use `golang-concurrency` and `golang-database` as rule sources. Only flag issues **clearly visible in
the diff** — no speculation. Look for: N+1 query patterns (a query inside a loop instead of a batched
`sqlc` query), unbounded `SELECT` with no `LIMIT`/pagination, sequential calls to independent
ports/IO that could run via `errgroup`/`Promise`-style fan-out, goroutine leaks (no `context`
cancellation, unbounded spawning, missing `wg.Wait()`), data races (shared map/slice written from
multiple goroutines without a lock), and channel ownership/deadlock risks.

**Comment format:**
```
<!-- onit-review:performance -->
⚡ Performance — [Short title]
[Description with impact, e.g. "O(N) queries per task" or "goroutine leak on ctx cancel"]
**Recommendation:** [Fix with short code sketch if < 6 lines]
```

---

## Subagent 6: Regression & Hallucination Detection

**Marker:** `<!-- onit-review:regression -->`

onit's principle is **"no vibe coding — review/understand every line"**. This agent guards against
AI-generated artifacts and unrelated drift. Review the diff for: deleted code unrelated to the PR's
stated purpose (🚨 Critical), phantom imports or calls to symbols/packages that don't exist
(🚨 Critical), function calls with the wrong signature, `TODO`/`FIXME`/panic-stubs left in non-test
code, duplicate logic that already exists in the core (it should be reused, not re-implemented),
weakened error handling or escalation (e.g. an invariant "mismatch → escalate" path silently dropped),
swallowed errors (`_ = err`), weakened test assertions, and dead code never called.

**Comment format:**
```
<!-- onit-review:regression -->
[🚨/⚠️/💡] — [Short title]
Type: [unrelated-deletion | phantom-import | hallucination | duplicate | regression | dead-code]
[Specific description with quoted evidence from the diff]
**Recommendation:** [Exact fix]
```

---

## Step 3: Consolidation

After all 6 subagents complete, spawn one more subagent via the Task tool to consolidate:

1. `gh api repos/{REPO}/pulls/{PR_NUMBER}/comments` — fetch all inline comments.
2. Filter to those starting with `<!-- onit-review:` and parse the type from the marker.
3. Fetch PR-level comments for the `<!-- onit-review:requirements -->` summary.
4. **Dedup by root cause (before grouping):** collapse findings that share a single root cause into **one** entry that lists every affected dimension and counts **once** — e.g. "sqlc drift — Critical; also surfaces under Security/Performance". One underlying problem must never be counted across multiple dimensions. This supersedes per-`{path, line}` dedup: same root cause across different files/lines still collapses to one entry.
5. Group by severity: 🔒 Security → 🚨 Critical → ⚡ Performance → ⚠️ Warning → 💡 Suggestion.
6. Collect one positive highlight per agent.
7. Post: `gh pr review {PR_NUMBER} --comment --body '...'`

**Summary format:**
```markdown
## 🤖 onit PR Review Summary

| | |
|---|---|
| **Subagents invoked** | {N} of 6 (Hexagonal Invariants · Requirements (Spec) · Security · Tests · Performance/Concurrency · Regression) |
| **Contract** | `docs/prd.md` (source of truth) · `.specs/features/{feature}/` · `docs/decisions.md` · `CLAUDE.md` |
| **Findings** | {N} across {M} files |

---

### 🔒 Security ({N})
- [`path/file.go:L42`] Finding title

### 🚨 Critical ({N})
### ⚡ Performance ({N})
### ⚠️ Warnings ({N})
### 💡 Suggestions ({N})

---
### ✅ Highlights
- [One positive highlight per agent]

---
> See inline comments for details and recommendations.
```

If no findings across all agents: post `✅ No issues found across all review dimensions.` but still
include the metadata table.
