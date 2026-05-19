# aiContext

CLI to bootstrap AI instruction context files (`AGENTS.md` + `CLAUDE.md`) into any project. No Go installation required.

---

## Installation

**1. Download the binary for your OS and architecture**

Go to [github.com/TheRealShek/aiContext/releases](https://github.com/TheRealShek/aiContext/releases) and download the right file:

| OS | Architecture | File |
|---|---|---|
| macOS | Apple Silicon (M1/M2/M3) | `aiContext_darwin_arm64.tar.gz` |
| macOS | Intel | `aiContext_darwin_amd64.tar.gz` |
| Linux | 64-bit | `aiContext_linux_amd64.tar.gz` |
| Linux | ARM | `aiContext_linux_arm64.tar.gz` |
| Windows | 64-bit | `aiContext_windows_amd64.zip` |

Not sure which Mac you have? Apple menu → About This Mac. M1/M2/M3 = arm64. Intel = amd64.

**2. Install**

macOS / Linux:
```bash
tar -xzf aiContext_*.tar.gz
chmod +x aiContext
mv aiContext /usr/local/bin/
```

Windows: extract the zip, move `aiContext.exe` to any folder in your `PATH` (e.g. `%USERPROFILE%\bin`).

**3. Run setup once**
```bash
aiContext setup
```

Installs default templates to `~/.config/aiContext/templates/`. Edit those files anytime to change what gets generated globally.

---

## Usage

```bash
cd myproject
aiContext init
```

Writes `AGENTS.md` + `CLAUDE.md` into current directory. Project name is inferred from the directory name. Refuses to overwrite existing files.

---

## Commands

| Command | What it does |
|---|---|
| `aiContext setup` | Copies default templates to `~/.config/aiContext/templates/`. Run once after install. Safe to re-run — prompts before overwriting. |
| `aiContext init` | Writes `AGENTS.md` + `CLAUDE.md` into current dir from your templates. |
