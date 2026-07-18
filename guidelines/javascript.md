### JavaScript

- Use modern, readable JavaScript with consistent module boundaries and no implicit globals.
- Validate untrusted values at API, storage, and message boundaries; static editor hints are not runtime validation.
- Prefer immutable bindings and focused pure functions when they make state transitions easier to understand.
- Handle promise rejection and cancellation deliberately; avoid floating promises and callback/promise mixtures.
- Use JSDoc for public or structurally complex APIs when the project does not use TypeScript.
- Preserve framework lifecycle rules and clean up subscriptions, timers, listeners, and effects.
