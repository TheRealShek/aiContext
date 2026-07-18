### Swift

- Follow Swift API design guidelines and use value semantics unless reference identity is required.
- Handle optionals explicitly; avoid force unwraps and implicitly unwrapped optionals outside proven lifecycle constraints.
- Keep actor isolation and `Sendable` requirements correct across concurrency boundaries.
- Use structured concurrency, propagate cancellation, and avoid detached tasks without explicit ownership.
- Model domain states with enums and typed errors instead of loosely related flags and strings.
- Keep UI updates on the appropriate actor and separate side effects from view rendering.
