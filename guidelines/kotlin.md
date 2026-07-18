### Kotlin

- Use Kotlin's null-safety and type system rather than Java-style sentinel values or pervasive `!!` assertions.
- Prefer immutable data, expression-oriented code, sealed hierarchies, and exhaustive `when` handling.
- Keep coroutine scopes structured; propagate cancellation and avoid unowned global tasks.
- Do not block coroutine threads with synchronous I/O unless execution is moved to an appropriate dispatcher.
- Keep Java interop boundaries explicit, especially around platform types and nullability.
- Follow the repository's formatter, detekt, and test conventions.
