### Ruby

- Follow the project's established Ruby style and favor clear objects and methods over metaprogramming.
- Keep Active Record callbacks and other implicit lifecycle behavior limited and documented.
- Use transactions for multi-step persistence invariants and avoid query patterns that create N+1 behavior.
- Rescue specific exceptions and preserve useful context; do not swallow unexpected failures.
- Treat symbols, constant lookup, dynamic dispatch, and deserialized input as trust boundaries where applicable.
- Add focused examples or tests around business rules and edge cases.
