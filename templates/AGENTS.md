# AGENTS.md

> **Project:** [One sentence: what it does, for whom.]
> **Core constraints:** [Non-obvious constraints that change how agents reason. e.g. "offline-first", "ACID on all writes", "zero external deps without discussion"]

---

## Toolchain

| Action      | Command              | Authority                        |
|-------------|----------------------|----------------------------------|
| Install     | `[command]`          | –                                |
| Build       | `[command]`          | outputs to `[dir]`               |
| Dev server  | `[command]`          | → `http://localhost:[port]`      |
| Test        | `[command]`          | `[flags worth knowing]`          |
| Test single | `[command]`          | –                                |
| Lint        | `[command]`          | see `[config file]`              |
| Format      | `[command]`          | –                                |
| Type check  | `[command]`          | –                                |
| ⚠️ Migrate  | `[command]`          | mutates DB                       |

> List commands only. Do not describe what tools enforce — that lives in their config.

---

## Judgment Boundaries

**NEVER**
- Commit secrets, `.env` files, or tokens
- Add external dependencies without discussion
- Swallow or suppress errors silently
- Guess on ambiguous specs — stop and ask

**ASK BEFORE**
- Running migrations
- Deleting or renaming files
- Touching auth, payments, or billing code

**ALWAYS**
- State plan before writing code
- Run lint + tests before marking done
- Handle errors explicitly

---

## Architecture

<!-- Non-obvious decisions ONLY. Format: "chose X over Y — reason"
     Skip anything inferable from framework or directory names.
     4–6 bullets max. Delete stale entries. -->

- [e.g.] Raw SQL over ORM — query patterns too complex for abstraction
- [e.g.] Auth delegated to [service] — do not implement locally
- [e.g.] `module-a` and `module-b` intentionally decoupled — shared types via `shared/types` only
- [e.g.] Event-driven between services — no direct cross-service calls; broker: [name]

---

## Context Map

<!-- Only for non-obvious or monorepo layouts. Delete if structure is standard for the framework. -->

```yaml
# Only list paths that would surprise someone who knows the framework.
services:
  [name]: [what it does, entry point]
notable:
  [path/]: [why it's notable — e.g. "all DB queries live here, no inline SQL elsewhere"]
  [path/]: [e.g. "ops scripts, not imported by app"]
  [path/]: [e.g. "pending deletion — do not import"]
```

---

## Gotchas

<!-- Verified pitfalls that have burned a session. Delete once fixed. -->

- [e.g.] Integration tests require Docker running — `[skip command]` avoids them
- [e.g.] `[module]` reads env at import time — import order matters in tests
- [e.g.] Migration `[N]` was rolled back in prod — `schema.sql` is authoritative, not history

---

## Verified

Last verified: `YYYY-MM-DD` · by: [name or commit hash]
Staleness: <4 weeks trust fully · 1–6 months verify commands · >6 months orientation only
