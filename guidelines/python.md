### Python

- Follow idiomatic Python and the repository's formatter and lint configuration; prioritize readability over metaprogramming.
- Add useful type annotations to new and changed public interfaces without hiding runtime behavior behind casts.
- Use context managers for resources and explicit exception types for expected failure modes.
- Do not use a broad `except` unless re-raising or deliberately handling every captured failure.
- Avoid mutable default arguments, hidden module-level state, and import-time side effects.
- Validate decoded or deserialized input before relying on its shape or type.
- Keep async code non-blocking and make task ownership, cancellation, and cleanup explicit.
