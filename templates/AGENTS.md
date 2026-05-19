# AGENTS.md

## Project overview

{{PROJECT_NAME}} — (fill in description)

## Package structure

(fill in after initial scaffolding)

## Build and test commands

(fill in for this project)

## Code conventions

- No global state. Inject dependencies via config structs.
- No mocks unless interface demands it.
- Table-driven tests alongside each package.
- Comments explain why, not what.
- Code readable over clever.
- One package per session unless told otherwise.

## Critical design rules

- No hardcoded config values. Read from config structs.
- Fail early, fail loud. Never silently swallow errors.
- No nil checks or error suppression as first response to a problem.

## Things to avoid

- No global variables or package-level init().
- Do not invent API surface not in the design doc.
- Do not add dependencies without a real problem demanding it.

## Output constraints

- Concise: only essential, high-signal information.
- No explanations or extra context unless asked.
- Prefer short statements or minimal bullets.
