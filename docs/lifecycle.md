# Project lifecycle

aiContext manages generated adapter files without taking ownership of a project's evolving `AGENTS.md`.

## The manifest

Initialization creates `.aicontext.json`. It records:

- The manifest schema and aiContext generator version
- The selected coding-agent tools
- Profile and language-pack metadata
- Whether stack detection was requested
- Each managed file's path, role, source template, and SHA-256 hash

The hashes let lifecycle commands tell the difference between unchanged generated content and a human edit.

Do not hand-edit the manifest. Use `update` to change tool selection, or run `clean` before starting over.

## Initialize a new project

```sh
aiContext init --detect --dry-run
aiContext init --detect
```

Initialization preflights every source template and destination. If any destination already exists, it creates nothing. If a write fails partway through, it removes the files created by that invocation.

Detection reads known manifests and configuration files. It never executes package managers, build scripts, or project code.

## Adopt existing files

Use adoption when the repository already contains the instruction files implied by the chosen tools:

```sh
aiContext init --adopt --tools codex,claude
```

Every selected file must exist. Adoption reads and hashes those files, writes only `.aicontext.json`, and does not alter instruction content.

For example, `codex,claude` expects `AGENTS.md` and `CLAUDE.md`; `codex` expects only `AGENTS.md`.

## Inspect project health

```sh
aiContext doctor
aiContext check
aiContext check --strict
```

`doctor` prints findings for a person and does not fail because it found project-health problems.

`check` returns a failing exit status for errors, such as missing managed files or invalid manifest state. `check --strict` also fails for warnings, such as a modified adapter. This makes strict mode useful in CI:

```yaml
- name: Check AI instructions
  run: aiContext check --strict
```

A modified `AGENTS.md` is expected and appears as information rather than a warning.

## Preview lifecycle changes

```sh
aiContext diff
aiContext sync --dry-run
```

The plan can include adapter creation, template updates, or removal after a tool is deselected. Modified adapters are conflicts. Lifecycle commands stop instead of overwriting or deleting them.

`diff` and `--dry-run` do not write project files.

## Synchronize adapters

```sh
aiContext sync
```

`sync` uses the tool set already recorded in the manifest. It can recreate a missing adapter or update an unchanged adapter from its installed template.

It does not regenerate `AGENTS.md`, even when that file differs from its creation-time hash. The canonical instructions remain project-owned.

After upgrading aiContext:

1. Run `aiContext setup` and decide whether to update customized local assets.
2. Run `aiContext diff` inside a managed project.
3. Run `aiContext sync` after reviewing the plan.

## Change selected tools

```sh
aiContext update --dry-run --tools codex,cursor,gemini
aiContext update --tools codex,cursor,gemini
```

`update` follows the same safety rules as `sync`, while allowing adapters to be added or removed. An adapter is removed only when its current hash matches the manifest. An edited adapter produces a conflict and is preserved.

Changing profiles or language packs does not rewrite an existing `AGENTS.md`. Edit that project-owned file directly.

## Stop managing a project

```sh
aiContext clean --dry-run
aiContext clean
```

Cleanup classifies every manifest-recorded file:

- Unchanged files are removed.
- Modified or symlinked paths are preserved.
- Missing files are treated as already absent.

After processing managed files, clean removes `.aicontext.json`. It removes `.cursor/rules`, `.cursor`, and `.github` only when those generated directories are empty. It never recursively deletes them, so unrelated workflows, settings, and other project files remain intact.

There is intentionally no force-clean option. Review and remove preserved content manually if it is no longer needed.

## Conflict and rollback guarantees

- Lifecycle writes do not proceed while the plan contains conflicts.
- Apply operations back up affected content and restore it if a later write fails.
- Managed paths are validated as project-relative before use.
- Cleanup preserves paths containing symlinks rather than traversing them.
- Commands accept `--target DIR` to operate on a project other than the current directory.

For custom adapter locations and placeholders, see [Configuration and templates](configuration.md).
