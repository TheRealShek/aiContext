<!--
  AGENTS.md — Compressed repository memory for AI coding agents.
  Purpose : fast orientation, reduced re-exploration, fewer hallucinated assumptions.
  Contract: contains only verified facts. Prefer file pointers over inline prose.
  Scope   : NOT a behavioral handbook. NOT a replacement for linters or CI.
  Size    : keep under 150 lines. If it grows, move detail to agent-docs/*.
  Updates : only with verified information. If state is uncertain, ask the user
            before editing. Bump the VERIFIED date after any confirmed change.
-->

# AGENTS.md

## PROJECT

**[Name]** — [One sentence: what it does, for whom, at what scale.]
Stack: [language/runtime] · [primary framework] · [data store(s)] · [deploy target]
Shape: [monorepo yes/no] · [services: list them, or "single service"] · [API style: REST / GraphQL / tRPC / event-driven]


## COMMANDS

<!-- Exact commands with real flags. Do not paraphrase or generalize.
     An agent that guesses commands wastes turns and can corrupt state.
     Mark anything that mutates persistent state with ⚠️. -->

```sh
# Bootstrap
[e.g.]  bun install
[e.g.]  cp .env.example .env          # required vars: VAR_A, VAR_B

# Build
[e.g.]  bun build

# Dev server
[e.g.]  bun dev                      # → http://localhost:3000

# Run all tests
[e.g.]  bun test

# Run one test file
[e.g.]  bun test -- src/foo.test.ts

# Run tests matching a name pattern
[e.g.]  bun test -- -t "pattern"

# Lint + format check
[e.g.]  bun lint

# Type check (if separate from build)
[e.g.]  bun typecheck

# ⚠️ Database migrations (mutates state)
[e.g.]  bun db:migrate
```

Required env vars: `VAR_A`, `VAR_B`. Full list: `[.env.example]`.


## ARCHITECTURE

<!-- Non-obvious decisions ONLY.
     Format: "chose X over Y — [reason or → pointer]"
     Skip anything obvious from the framework or inferable from directory names.
     Target: 4–8 bullets. Delete any that have become obvious or stale. -->

- [e.g.] Raw SQL via `[lib]` over an ORM — query patterns: `→ agent-docs/db.md`
- [e.g.] Auth delegated to `[external service]` — do not implement auth locally
- [e.g.] `[module-a]` and `[module-b]` are intentionally decoupled — no cross-imports; shared types via `[shared/types]`
- [e.g.] Client SDK is codegen'd at build time from `[schema path]` — do not edit generated files
- [e.g.] Event-driven between services instead of direct calls — broker: `[name]`
- [add or remove lines; empty section → delete it]


## LAYOUT

<!-- Non-standard paths and canonical file pointers ONLY.
     Omit anything standard for the framework (src/, tests/, public/, etc.).
     Useful entries: unusual locations, "copy this" exemplars, off-limits dirs. -->

| Path | Notes |
|---|---|
| `[config/]` | [e.g. runtime config only — build config is in vite.config.ts] |
| `[scripts/]` | [e.g. one-off ops scripts, not imported by app] |
| `[src/lib/]` | [e.g. shared utilities — no business logic here] |
| `[path/to/canonical.ts]` | [e.g. canonical pattern for new API routes — copy this] |
| `[legacy/]` | [e.g. pending deletion — do not import] |

<!-- If every directory follows framework conventions, delete this section. -->


## CONVENTIONS

<!-- Deviations from stack defaults ONLY.
     If the project follows standard idioms: delete this section.
     Pair every prohibition with a concrete alternative.
     A list of "don'ts" without "dos" causes agents to over-explore. -->

- **[Error handling]**: [e.g. don't throw in service layer — return `Result<T,E>` from `src/lib/result.ts`]
- **[State]**: [e.g. React Query for server state; Zustand only for pure UI state — see `agent-docs/state.md`]
- **[Domain term]** means [precise definition] — maps to `[ClassName / module path]`
- **[Pattern]**: [e.g. all DB queries in `src/db/queries/` — no inline SQL elsewhere]
- Formatter: `[tool]` config at `[path]`. Auto-fix: `[command]`.
- [Remove entries that are obvious from the stack alone]


## GOTCHAS

<!-- Verified pitfalls that have actually burned a session.
     Concrete and recoverable — not a general safety checklist.
     Delete entries once the underlying issue is resolved. -->

- **Setup**: [e.g. requires Node ≥ 20; `.nvmrc` exists but `nvm use` does not auto-activate]
- **Tests**: [e.g. integration tests require Docker running — `[test:unit command]` skips them]
- **Tests**: [e.g. `TestFoo` in `foo_test.go` is flaky under load — skip locally with `-run '^(?!TestFoo)'`]
- **[Module]**: [e.g. `config` package reads env at import time — import order matters in tests]
- **Migrations**: [e.g. migration 0042 was rolled back in prod; `schema.sql` is authoritative, not history]
- **[External service]**: [e.g. Stripe webhooks need `stripe listen --forward-to localhost:3000/webhooks`]
- [Stale entries are worse than no entries — delete aggressively]


## AGENT-DOCS

<!-- Optional. Pointers to deeper reference files agents load on demand.
     Create a file only when content is too large or too task-specific for here.
     Format: path — what's in it · when to read it -->

<!--
- `agent-docs/db.md`     — query patterns, transactions, index conventions · when touching DB layer
- `agent-docs/auth.md`   — token lifecycle, session edge cases · when touching auth
- `agent-docs/state.md`  — React Query vs Zustand decision table · when adding state
- `agent-docs/deploy.md` — deploy steps, rollback procedure · when shipping to production
-->

<!-- If no reference files exist yet, delete this section. -->


## VERIFIED

<!--
  Records when this file was last confirmed against actual repo state.
  Agents should trust this file as accurate as of this date.
  If a discrepancy is found: correct the specific entry, bump the date.
  Do NOT rewrite the file without confirming current state.
  If state is uncertain, ask the user before making changes.
-->

Last verified : `YYYY-MM-DD`
Verified by   : [human name / agent session / `git log --oneline -1` hash]
Environment   : [OS, runtime version, package manager version used to test commands]

<!--
  Staleness guide for future agents:
  < 4 weeks  → trust commands and gotchas directly
  1–6 months → verify commands before running; architecture decisions likely still hold
  > 6 months → treat as orientation only; re-verify before any significant work
-->
