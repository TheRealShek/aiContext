### Rust

- Use idiomatic ownership and borrowing; do not add cloning or allocation merely to silence the borrow checker without understanding the tradeoff.
- Model absence and failure with `Option` and `Result`, propagate with `?`, and add context at subsystem boundaries.
- Avoid `unwrap` and `expect` in production paths unless an invariant makes failure impossible and the invariant is documented.
- Prefer expressive types, enums, and newtypes that make invalid states difficult to represent.
- Keep public APIs and lifetimes as simple as the domain allows; avoid exposing implementation-specific generic complexity.
- Treat `unsafe` as an exceptional boundary. Keep unsafe blocks minimal, include a `SAFETY:` explanation for every block, and document all caller obligations.
- Never create a safe abstraction over unsafe code unless its invariants are enforced for every safe caller.
- Be explicit about `Send`, `Sync`, locking, cancellation, and task lifetimes in concurrent or asynchronous code.
- Run `cargo fmt`, `cargo clippy`, and relevant tests; use Miri or sanitizers for unsafe or memory-sensitive changes when available.
