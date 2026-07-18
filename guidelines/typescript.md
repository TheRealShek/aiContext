### TypeScript

- Keep strict type checking enabled and preserve useful inference; avoid `any`, broad casts, and non-null assertions as escape hatches.
- Use `unknown` for untrusted values and narrow them with runtime checks at API, storage, and message boundaries.
- Prefer discriminated unions and exhaustive handling for state machines and variant data.
- Keep domain types separate from transport payloads when validation or serialization rules differ.
- Make promise rejection and cancellation behavior explicit; do not leave floating promises.
- Preserve framework lifecycle and rendering rules, especially around effects, cleanup, and server/client boundaries.
- Test runtime behavior rather than relying on compile-time types to validate external data.
