# aiContext

Create one `AGENTS.md` with project-specific instructions for coding agents.

aiContext combines a working profile, language guidelines, and optional project detection into a single file. Edit that file after initialization.

## Quick start

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh | sh
cd my-project
aiContext init --detect
```

Review `AGENTS.md` and add the project facts an agent cannot infer from code. The installer supports macOS and Linux. Windows users can download the binary from [GitHub Releases](https://github.com/TheRealShek/aiContext/releases), put `aiContext.exe` on `PATH`, and run `aiContext setup` before initializing a project.

## What it creates

`aiContext init` creates only `AGENTS.md` in the target project. It refuses to overwrite an existing file. Use `--dry-run` to validate the template and preview the action.

```sh
aiContext init --detect --dry-run
aiContext init --profile strict --languages go,typescript
```

`--detect` reads project files to suggest the stack and common commands. The default working profile is `standard`; language guidelines are selected from detected project files. Run `aiContext profiles` to see available profiles and guidelines.

Existing projects only need edits to their `AGENTS.md`. Older aiContext versions may have created adapters and `.aicontext.json`; this version leaves those files untouched. Review and remove them manually if they are no longer needed.

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

- [Profiles and language guidelines](docs/profiles.md)
- [Configuration and templates](docs/configuration.md)

Run `aiContext help` or `aiContext <command> --help` for built-in help.

## Contributing

Development requires Go 1.25 or newer. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).
