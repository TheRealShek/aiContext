### Terraform

- Keep plans deterministic and avoid values that cause perpetual diffs.
- Pin provider constraints deliberately and review provider upgrade notes before changing lock files.
- Treat state, variable files, and plan output as potentially sensitive; never commit secrets.
- Use modules for stable reusable boundaries, not as wrappers around single resources without a clear contract.
- Prefer validation, preconditions, and explicit types for module inputs and outputs.
- Review destroy/replace actions in the plan and document migrations for renamed or moved resources.
