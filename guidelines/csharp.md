### C# / .NET

- Follow the repository's nullable-reference-type policy and avoid suppressing warnings without proving the invariant.
- Prefer immutable models and explicit dependency lifetimes; avoid hidden service-locator access.
- Use `async` end to end, accept and propagate `CancellationToken`, and avoid `.Result` or `.Wait()` on asynchronous work.
- Dispose owned resources deterministically with `using` or `await using`.
- Keep LINQ expressions readable and watch for repeated enumeration or client-side database evaluation.
- Preserve thread-safety and dependency-injection lifetime rules in shared services.
