### Java

- Follow the project's Java version, build conventions, and existing package boundaries.
- Prefer immutable data and constructor-enforced invariants; avoid nullable state when an explicit type can represent absence.
- Keep interfaces focused and use composition before inheritance.
- Do not catch broad exceptions unless adding context and rethrowing or handling the entire failure class deliberately.
- Preserve thread-safety contracts and avoid sharing mutable state without a documented synchronization strategy.
- Keep persistence transactions and resource lifetimes explicit, using try-with-resources where appropriate.
