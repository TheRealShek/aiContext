# AGENTS.md

Guidance for AI coding agents working on **{{PROJECT_NAME}}**.

## Project Overview

<!-- What the project does, who uses it, and the domain concepts agents need. -->

**Stack:** {{STACK}}

**Priorities:** <!-- List decision-making priorities in order, e.g. correctness > reliability > performance. -->

## Development Commands

Use repository wrappers when they perform setup that direct tool commands skip.

| Action | Command | Notes |
| --- | --- | --- |
{{COMMANDS}}

## Architecture

<!-- Add the primary runtime/data flow and only the important ownership boundaries. -->

```text
[Entry point] -> [Application layer] -> [Domain logic] -> [Storage / external systems]
```

| Path | Responsibility |
| --- | --- |
| `[path]` | [What belongs here.] |

### Key Invariants

<!-- Document behavior that spans components or is easy to break. -->

- [Invariant] — [why it must remain true].

## Working Agreement

{{PROFILE_GUIDELINES}}

## Language Guidelines

{{LANGUAGE_GUIDELINES}}

## Project Conventions

<!-- Keep only repository-specific or lint-enforced rules. -->

- [Naming, structure, error-handling, dependency, or configuration rule.]

## Testing

<!-- Explain test locations, required layers, fixtures, and special dependencies. -->

- Run `[focused command]` while iterating and `[full command]` before finishing.
- Cover [success, boundary, and failure behavior required by this project].

## Files You Must Not Edit

<!-- Name generated, vendored, mirrored, or externally owned paths and how to update them. -->

- `[path]` — update via `[source or generation command]`.

## Common Gotchas

<!-- Record real failure modes; remove them when the underlying hazard is fixed. -->

- `[path or subsystem]` has [constraint] — [correct workflow].

## Common Development Tasks

<!-- Short recipes for recurring changes that cross ownership boundaries. -->

- **[Task]:** `[source]` -> `[implementation]` -> `[tests/docs]`.

## Verified

Last verified: not yet verified
