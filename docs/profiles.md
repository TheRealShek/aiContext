# Profiles and language guidelines

aiContext composes two kinds of focused guidance into a new project's `AGENTS.md`:

- A profile defines the overall working agreement.
- Language packs add concrete rules for the languages in the repository.

This keeps the generated instructions useful without filling them with generic advice.

## Working profiles

| Profile | Intended use |
| --- | --- |
| `minimal` | Small, focused changes with only essential checks |
| `standard` | Balanced defaults for everyday repository work |
| `strict` | Stronger compatibility, validation, and documentation expectations |
| `security` | Trust boundaries, least privilege, secret handling, and negative tests |

New projects use `standard` by default:

```sh
aiContext init --profile strict
aiContext init --profile security --languages go,rust
```

Run `aiContext profiles` to list the profiles installed in your local configuration.

## Language selection

`--languages` accepts one of these values:

- `auto` detects relevant packs and is the default.
- `none` omits language-specific guidance.
- `all` includes every installed built-in pack.
- A comma-separated list such as `go,terraform` selects explicit packs.

The supported packs are:

```text
go, rust, typescript, javascript, python, ruby,
java, kotlin, csharp, swift, php, terraform
```

For example:

```sh
aiContext init --profile standard --languages typescript,terraform
```

Language detection and `--detect` do different jobs:

- Automatic language selection chooses guideline packs from project manifests.
- `--detect` additionally fills the stack and common-command sections in `AGENTS.md`.

Neither mechanism runs project code.

## What the packs contain

Language packs focus on errors that broad instructions tend to miss.

The Go pack asks for idiomatic Go, contextual error handling, explicit goroutine lifetimes, correct `context.Context` propagation, race testing for concurrent changes, and caution around reflection and `unsafe`.

The Rust pack covers ownership and borrowing, `Result` and `Option`, production uses of `unwrap`, concurrency traits and lifetimes, and requires a documented `SAFETY:` invariant around exceptional uses of `unsafe`.

Other packs cover equivalent language-specific concerns such as strict TypeScript boundaries, Python resource management, Java concurrency, Swift structured concurrency, and Terraform state safety.

## Add a custom profile

Profiles are Markdown files in the configuration root's `profiles/` directory. Add a file such as:

```text
profiles/backend.md
```

Then select it by filename without the extension:

```sh
aiContext init --profile backend --languages go,terraform
```

The profile name may contain lowercase letters, numbers, hyphens, and underscores.

Profile and language text is composed into `AGENTS.md` during initialization. Because `AGENTS.md` becomes project-owned, later edits to the source profile do not rewrite existing projects. Update an existing project's guidance directly in its `AGENTS.md`.

See [Configuration and templates](configuration.md) for configuration locations and isolated team-owned setups.
