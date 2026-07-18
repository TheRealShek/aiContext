# aiContext

Manage one `AGENTS.md` across Codex, Claude Code, Cursor, GitHub Copilot, and Gemini CLI.

aiContext gives a repository one canonical set of AI coding instructions, creates the adapters other agents need, and keeps that generated plumbing safe and synchronized.

## Quick Start

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh | sh
cd my-project
aiContext init --detect
```

Then open `AGENTS.md`, review the detected stack and commands, and add the project-specific facts an agent cannot infer from the code.

The installer supports macOS and Linux. Windows users can download the binary from [GitHub Releases](https://github.com/TheRealShek/aiContext/releases), put `aiContext.exe` on `PATH`, and run `aiContext setup` once before initializing a project.

## Why aiContext?

| Without aiContext | With aiContext |
| --- | --- |
| Repeat instructions in several tool-specific files | Edit one canonical `AGENTS.md` |
| Copies drift as project guidance changes | Small adapters continue to reference the source of truth |
| Start every repository with generic instructions | Compose a working profile with language-aware guidance |
| Manually inspect broken or stale adapters | Run `doctor`, `diff`, and `sync` |
| Risk deleting files that contain human edits | Hash-aware updates and cleanup preserve modified files |

## Features

- One project-owned `AGENTS.md` shared across five coding-agent ecosystems
- Selectable adapters for Codex, Claude Code, Cursor, GitHub Copilot, and Gemini CLI
- Automatic stack, framework, and common-command detection without executing project code
- Working profiles for minimal, standard, strict, or security-focused changes
- Detailed guideline packs for Go, Rust, TypeScript, Python, and other supported languages
- Manifest-backed health checks, diffs, synchronization, and tool-set updates
- Safe adoption and cleanup that preserve human-modified or linked files

## What gets created?

The default tool selection creates:

```text
AGENTS.md
CLAUDE.md
.cursor/rules/aicontext.mdc
.github/copilot-instructions.md
.aicontext.json
```

`AGENTS.md` is yours to edit. The adapter files are generated references for other tools. `.aicontext.json` records the selected tools and generated-file hashes so lifecycle commands can distinguish generated content from human changes.

Gemini CLI is optional:

```sh
aiContext init --tools codex,claude,gemini
```

Run `aiContext tools` to list supported values, or use `--tools all`.

## Everyday commands

```sh
# Preview initialization
aiContext init --detect --dry-run

# Check an initialized project
aiContext doctor
aiContext check --strict

# Preview and apply adapter updates
aiContext diff
aiContext sync

# Change the selected coding agents
aiContext update --tools codex,cursor,gemini

# Preview safe removal
aiContext clean --dry-run
```

aiContext refuses to overwrite existing destinations during initialization. To manage instruction files already present in a repository, use adoption instead:

```sh
aiContext init --adopt --tools codex,claude
```

## Installation

### macOS and Linux

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh | sh
```

The installer downloads the latest release, verifies its SHA-256 checksum, installs the binary, and initializes the editable user configuration. It prefers `/usr/local/bin` and falls back to `$HOME/.local/bin` when elevated access is unavailable.

Choose another install directory with:

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh \
  | AICONTEXT_INSTALL_DIR="$HOME/bin" sh
```

### Manual installation

Download the archive for your platform from [GitHub Releases](https://github.com/TheRealShek/aiContext/releases), verify it against `checksums.txt`, and put the binary on `PATH`. Release archives support macOS and Linux on AMD64 and ARM64, and Windows on AMD64.

Then run:

```sh
aiContext setup
```

## Documentation

- [Profiles and language guidelines](docs/profiles.md) — working profiles, automatic language selection, and custom team guidance
- [Configuration and templates](docs/configuration.md) — config locations, custom template directories, and placeholders
- [Project lifecycle](docs/lifecycle.md) — manifests, health checks, synchronization, adoption, tool updates, and safe cleanup

Run `aiContext help` or `aiContext <command> --help` for the built-in command reference.

## Safety at a glance

- Initialization validates all destinations before writing and rolls back its own files on failure.
- Detection reads known manifests; it never invokes project commands or package managers.
- Synchronization treats `AGENTS.md` as project-owned and never regenerates it.
- Generated adapters are changed only when their hashes prove they are still unmodified.
- Cleanup does not follow symlinked managed paths or recursively delete directories.

The [lifecycle documentation](docs/lifecycle.md) describes these guarantees and conflict behavior in detail.

## Contributing

Development requires Go 1.25 or newer. See [CONTRIBUTING.md](CONTRIBUTING.md) for local checks and the release workflow.

## License

MIT — see [LICENSE](LICENSE).
