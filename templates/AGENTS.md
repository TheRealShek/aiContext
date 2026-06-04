# AGENTS.md

> **Project:** [One sentence. What it does, for whom.]
> **Stack:** [Language · Framework · DB · infra — only non-obvious choices]

---

## Commands

<!-- Add only non-standard or non-obvious commands. Standard ones (npm install, go test) → skip. -->

| Action     | Command | Note                  |
| ---------- | ------- | --------------------- |
| [e.g. Dev] | `[cmd]` | [only if non-obvious] |

---

## Gotchas

<!-- Each line = a real mistake that happened. Delete once fixed in code. -->

- [e.g.] `internal/x/` uses PascalCase filenames — `Add.go` not `add.go`. Wrong casing = shadow file = broken build.
- [e.g.] Module imports as `MyProject/internal/...` not `myproject/` — Go module name is case-sensitive.

---

## Rules

<!-- Project-specific only. If it applies to every project everywhere, delete it. -->

- [e.g.] Never use X directly in Y — always go through Z package.
- [e.g.] Errorf = execution stops. Warningf = execution continues. Pick correctly.

---

## Verified

Last verified: `YYYY-MM-DD`
