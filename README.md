# aiContext

`aiContext` is a small CLI that bootstraps shared AI coding-agent instructions into a project. It keeps `AGENTS.md` as the main source of truth and creates compatibility files for Claude Code, Cursor, and GitHub Copilot.

The released binaries are self-contained; Go is not required to use the CLI.

## Installation

### macOS and Linux

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh | sh
```

The installer detects the platform, downloads the latest release, verifies its SHA-256 checksum, installs the binary, and runs `aiContext setup`. It uses `/usr/local/bin` when possible and falls back to `$HOME/.local/bin` when `sudo` is unavailable.

Set `AICONTEXT_INSTALL_DIR` to choose another installation directory:

```sh
curl -fsSL https://raw.githubusercontent.com/TheRealShek/aiContext/main/install.sh | AICONTEXT_INSTALL_DIR="$HOME/bin" sh
```

### Manual installation

Download the appropriate archive from [GitHub Releases](https://github.com/TheRealShek/aiContext/releases). Release asset names include the version:

| OS | Architecture | Asset pattern |
|---|---|---|
| macOS | Apple Silicon | `aiContext_<version>_darwin_arm64.tar.gz` |
| macOS | Intel | `aiContext_<version>_darwin_amd64.tar.gz` |
| Linux | x86-64 | `aiContext_<version>_linux_amd64.tar.gz` |
| Linux | ARM64 | `aiContext_<version>_linux_arm64.tar.gz` |
| Windows | x86-64 | `aiContext_<version>_windows_amd64.zip` |

Extract the archive, place `aiContext` (`aiContext.exe` on Windows) in a directory on `PATH`, then run:

```sh
aiContext setup
```

## Quick start

```sh
cd my-project
aiContext init
```

The command creates:

```text
AGENTS.md
CLAUDE.md
.cursor/rules/aicontext.mdc
.github/copilot-instructions.md
```

`AGENTS.md` contains the editable project instructions. The other files reference it using the supported instruction-import format. `init` validates every template and destination before writing and refuses to overwrite an existing project file.

## Commands

### `aiContext setup`

Copies the embedded defaults into the user configuration directory. Existing templates are preserved unless their overwrite prompts are accepted.

```sh
aiContext setup
aiContext setup --force
aiContext setup --template-dir ./my-templates
```

The default location follows the operating system's user configuration convention:

- Linux: `$XDG_CONFIG_HOME/aiContext/templates`, or `~/.config/aiContext/templates`
- macOS: `~/Library/Application Support/aiContext/templates`
- Windows: `%AppData%\aiContext\templates`

Edit these files to customize what future projects receive. After upgrading from a version that used `.cursorrules`, run `aiContext setup` once to install the new `cursor.mdc` template.

### `aiContext init`

Creates instruction files in the current directory:

```sh
aiContext init
aiContext init --dry-run
aiContext init --target ../another-project
aiContext init --template-dir ./team-templates
```

Options can be combined. For example:

```sh
aiContext init --dry-run --target ./service --template-dir ./templates
```

### Other commands

```sh
aiContext help
aiContext version
aiContext init --help
aiContext setup --help
```

## Template values

`{{PROJECT_NAME}}` is replaced with the target directory name during `init`. Other template text is copied unchanged.

## Optional shell alias

```sh
# zsh
echo "alias ac='aiContext'" >> ~/.zshrc

# bash
echo "alias ac='aiContext'" >> ~/.bashrc

# fish
echo "alias ac='aiContext'" >> ~/.config/fish/config.fish
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
