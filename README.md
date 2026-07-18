# aiContext

One project should not need five separately maintained sets of AI coding instructions.

aiContext creates one canonical `AGENTS.md` for the repository, adds the small adapter files needed by other coding agents, and records enough state to keep those adapters healthy without overwriting project-owned guidance.

It is a dependency-free Go CLI. Release binaries do not require Go.

## The idea

Most coding agents look for a different instruction file. Copying the same rules into each one works initially, but the copies drift. aiContext uses this model instead:

```text
                         edit this
                            │
                            ▼
                       AGENTS.md
                   canonical instructions
                      ▲      ▲      ▲
                      │      │      │
                 CLAUDE.md   │   GEMINI.md
                             │
              .cursor/rules/aicontext.mdc
              .github/copilot-instructions.md

                       .aicontext.json
              selected tools + generated-file hashes
```

`AGENTS.md` belongs to the project and is expected to evolve. The adapter files are generated plumbing. The manifest lets aiContext distinguish an unchanged generated file from one a person edited.

## Start a project

Install aiContext, then run this from a repository:

```sh
cd my-project
aiContext init --detect
```

The default setup creates:

```text
AGENTS.md
CLAUDE.md
.cursor/rules/aicontext.mdc
.github/copilot-instructions.md
.aicontext.json
```

`--detect` fills in likely stack details and common project commands by reading manifests such as `go.mod`, `package.json`, and `Cargo.toml`. It never executes project code. Without `--detect`, the generated file contains short prompts for the details a maintainer should add.

Open `AGENTS.md`, correct anything detection got wrong, and add the few repository-specific facts that an agent cannot infer from the code. That file is the source of truth from then on.

aiContext refuses to overwrite any destination that already exists. Preview a new project without writing anything with:

```sh
aiContext init --detect --dry-run
```

## Choose the agents you use

Pass a comma-separated tool list during initialization:

```sh
aiContext init --tools codex,claude,gemini
```

| Tool | Generated instruction file | Default |
| --- | --- | :---: |
| Codex | `AGENTS.md` only | yes |
| Claude Code | `CLAUDE.md` | yes |
| Cursor | `.cursor/rules/aicontext.mdc` | yes |
| GitHub Copilot | `.github/copilot-instructions.md` | yes |
| Gemini CLI | `GEMINI.md` | no |

Run `aiContext tools` to see the supported values. Use `--tools all` to enable every adapter.

Tool selection can change later without touching `AGENTS.md`:

```sh
aiContext update --tools codex,claude,gemini
```

An adapter is added, refreshed, or removed only when aiContext can do so safely. A manually edited adapter is reported as a conflict instead of being overwritten.

## Profiles and language guidelines

An instruction file should say more than “write good code,” but it should not bury agents in generic policy either. aiContext composes two kinds of focused guidance into a new `AGENTS.md`:

- A profile defines the overall working agreement.
- Language packs add concrete rules for the languages in the repository.

The built-in profiles are:

| Profile | Intended use |
| --- | --- |
| `minimal` | Small, focused changes with only essential checks |
| `standard` | Balanced defaults for everyday repository work |
| `strict` | Stronger compatibility, validation, and documentation expectations |
| `security` | Trust boundaries, least privilege, secret handling, and negative tests |

New projects use `standard` unless another profile is selected:

```sh
aiContext init --profile strict
aiContext init --profile security --languages go,rust
```

Language selection accepts `auto`, `none`, `all`, or a comma-separated list. `auto` is the default and recognizes:

```text
go, rust, typescript, javascript, python, ruby,
java, kotlin, csharp, swift, php, terraform
```

The packs are intentionally language-specific. The Go pack asks for idiomatic Go, contextual error handling, explicit goroutine lifetimes, and caution around reflection and `unsafe`. The Rust pack covers ownership, `Result`, production `unwrap`, concurrency traits, and requires documented `SAFETY:` invariants around exceptional uses of `unsafe`.

Language detection and `--detect` are separate:

- Language detection runs by default to select guideline packs.
- `--detect` additionally fills the stack and command sections in `AGENTS.md`.

Profiles and language packs are copied to your user configuration by `aiContext setup`, so teams can edit them or add a profile such as `profiles/backend.md`. List locally available profiles and supported language packs with:

```sh
aiContext profiles
```

Then select a custom profile by filename:

```sh
aiContext init --profile backend --languages go,terraform
```

Profile text is composed into `AGENTS.md` at initialization. Because `AGENTS.md` becomes project-owned, later profile edits do not rewrite an existing project's instructions; make those changes directly in its `AGENTS.md`.

## Keep an initialized project healthy

The `.aicontext.json` manifest records selected tools, profile metadata, and the SHA-256 hash of every generated file. Lifecycle commands use it to protect human changes.

### Inspect

```sh
aiContext doctor
aiContext check
aiContext check --strict
```

`doctor` prints findings for a person. `check` also returns a failing exit status for errors, which makes it suitable for CI. `check --strict` treats warnings—such as an edited adapter—as failures too. Editing `AGENTS.md` is expected and is reported as information, not a warning.

### Preview and synchronize

```sh
aiContext diff
aiContext sync --dry-run
aiContext sync
```

`diff` shows pending adapter changes. `sync` applies safe changes using the currently selected tools. Neither command regenerates or replaces the canonical `AGENTS.md`.

After upgrading aiContext, run `aiContext setup` to install any new default templates, review the prompts before replacing customized local files, and then run `aiContext diff` inside a managed project.

### Change tool selection

```sh
aiContext update --dry-run --tools codex,cursor,gemini
aiContext update --tools codex,cursor,gemini
```

`update` has the same safety rules as `sync`, with the additional ability to add or remove adapters.

## Adopt existing instruction files

If a repository already has the files you want aiContext to manage, adopt them instead of recreating them:

```sh
aiContext init --adopt --tools codex,claude
```

Every file implied by the selected tools must already exist. Adoption reads and hashes those files, creates only `.aicontext.json`, and does not alter their contents.

Choose the tool list carefully. For example, `codex,claude` expects both `AGENTS.md` and `CLAUDE.md`; `codex` expects only `AGENTS.md`.

## Stop using aiContext in a project

Preview cleanup first:

```sh
aiContext clean --dry-run
aiContext clean
```

`clean` removes only manifest-recorded files whose contents still match their recorded hashes. Modified files and linked paths are preserved. It removes generated directories only when they are empty, leaves unrelated `.github` and `.cursor` content alone, and finally removes `.aicontext.json` so the project is no longer managed.

There is intentionally no force option. If you also want to delete a preserved file, review and remove it yourself.

## Install

### macOS and Linux

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh | sh
```

The installer detects the platform, downloads the latest release, verifies its SHA-256 checksum, installs the binary, and runs `aiContext setup`. It prefers `/usr/local/bin` and falls back to `$HOME/.local/bin` when elevated access is unavailable.

Choose a different install directory with:

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh \
  | AICONTEXT_INSTALL_DIR="$HOME/bin" sh
```

### Manual installation and Windows

Download the archive for your platform from [GitHub Releases](https://github.com/TheRealShek/aiContext/releases), verify it against `checksums.txt`, and put `aiContext` or `aiContext.exe` on `PATH`. Then initialize the editable local configuration:

```sh
aiContext setup
```

Release archives are available for macOS and Linux on AMD64 and ARM64, and Windows on AMD64.

## Local configuration

`aiContext setup` installs the embedded defaults into the operating system's user configuration directory:

```text
aiContext/
├── templates/
│   ├── AGENTS.md
│   ├── CLAUDE.md
│   ├── GEMINI.md
│   ├── copilot-instructions.md
│   └── cursor.mdc
├── profiles/
│   ├── minimal.md
│   ├── standard.md
│   ├── strict.md
│   └── security.md
└── guidelines/
    ├── go.md
    ├── rust.md
    └── ...
```

The configuration root is normally:

- Linux: `$XDG_CONFIG_HOME/aiContext`, or `~/.config/aiContext`
- macOS: `~/Library/Application Support/aiContext`
- Windows: `%AppData%\aiContext`

Setup preserves existing files unless you approve an overwrite. `aiContext setup --force` replaces all installed templates, profiles, and guideline packs with the embedded version, so use it only when that is what you intend.

For isolated testing or team-owned configuration, point commands at a different template directory:

```sh
aiContext setup --template-dir ./aicontext/templates
aiContext init --template-dir ./aicontext/templates
```

Profiles and guidelines are resolved from sibling `profiles/` and `guidelines/` directories beside that `templates/` directory.

### Template placeholders

| Placeholder | Replaced with |
| --- | --- |
| `{{PROJECT_NAME}}` | Target directory name |
| `{{STACK}}` | Detected stack, or an editing prompt |
| `{{COMMANDS}}` | Detected Markdown command rows, or an editing prompt |
| `{{PROFILE_GUIDELINES}}` | Selected profile contents |
| `{{LANGUAGE_GUIDELINES}}` | Selected language-pack contents |

All other template text is copied unchanged.

## Command reference

| Command | Purpose |
| --- | --- |
| `setup` | Install editable templates, profiles, and language packs |
| `init` | Initialize or adopt project instruction files |
| `tools` | List supported coding agents |
| `profiles` | List installed profiles and supported language packs |
| `doctor` | Explain project instruction health |
| `check` | Validate health with CI-friendly exit behavior |
| `diff` | Preview adapter changes and conflicts |
| `sync` | Apply safe adapter template changes |
| `update` | Safely change the selected tool set |
| `clean` | Stop managing a project and remove unchanged generated files |
| `version` | Print the installed version |
| `help` | Show command usage |

Every command that acts on a project accepts `--target DIR` where relevant. Run `aiContext <command> --help` for its complete options.

## Safety model

- Initialization validates all inputs and destinations before writing, refuses existing files, and rolls back files it created if a write fails.
- Detection reads known project manifests and never runs package managers, build scripts, or project code.
- `AGENTS.md` is treated as user-owned after creation; synchronization does not regenerate it.
- Generated adapters are updated or deleted only when their current hash matches the manifest.
- Conflicts stop lifecycle writes instead of choosing one version automatically.
- Cleanup does not traverse symlinked managed paths and does not recursively delete directories.

## Contributing

The project requires Go 1.25 or newer for development. See [CONTRIBUTING.md](CONTRIBUTING.md) for the local checks and release workflow.

## License

MIT — see [LICENSE](LICENSE).
