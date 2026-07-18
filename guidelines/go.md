### Go

- Write idiomatic, straightforward Go; prefer clarity and explicit control flow over clever abstractions.
- Keep packages cohesive, exported APIs small, and interfaces defined near the code that consumes them.
- Use standard Go naming, `gofmt`, and zero-value-friendly types where practical.
- Wrap errors with operation context using `%w`; use `errors.Is` and `errors.As` instead of string matching.
- Do not use `panic` for expected runtime failures, and do not silently discard returned errors.
- Accept `context.Context` as the first parameter for cancellable work, propagate it, and do not store it in structs.
- Give every goroutine an explicit lifetime and cancellation path; avoid leaks, blocked sends, and unbounded fan-out.
- Choose channels for ownership/coordination and mutexes or atomics for shared state only when their invariants are clear.
- Prefer table-driven tests where they improve coverage; run the race detector for meaningful concurrent changes.
- Avoid reflection and `unsafe` by default. If `unsafe` is unavoidable, isolate it, document the memory and lifetime invariants, and test relevant architectures.
