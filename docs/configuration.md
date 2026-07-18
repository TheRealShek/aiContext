# Configuration and templates

`aiContext setup` installs editable copies of the embedded templates, profiles, and language packs into your user configuration directory.

## Configuration layout

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

The default configuration root is:

- Linux: `$XDG_CONFIG_HOME/aiContext`, or `~/.config/aiContext`
- macOS: `~/Library/Application Support/aiContext`
- Windows: `%AppData%\aiContext`

Edit these local files to change what future projects receive.

## Setup and upgrades

Run setup after manual installation and after upgrading when a release adds new configuration assets:

```sh
aiContext setup
```

Existing files trigger an overwrite prompt and are preserved unless you explicitly accept. To restore every installed asset to its embedded default:

```sh
aiContext setup --force
```

`--force` replaces customized templates, profiles, and guideline packs, so use it only when discarding those local changes is intentional.

## Custom configuration directories

Use `--template-dir` for isolated testing or a team-owned configuration checked into another location:

```sh
aiContext setup --template-dir ./aicontext/templates
aiContext init --template-dir ./aicontext/templates
```

The supplied path is specifically the templates directory. Profiles and guidelines are loaded from sibling directories:

```text
aicontext/
├── templates/
├── profiles/
└── guidelines/
```

Lifecycle commands that render adapters must use the same template directory when it differs from the user default:

```sh
aiContext diff --template-dir ./aicontext/templates
aiContext sync --template-dir ./aicontext/templates
aiContext update --template-dir ./aicontext/templates --tools codex,gemini
```

## Template placeholders

The canonical `AGENTS.md` template supports these placeholders:

| Placeholder | Replaced with |
| --- | --- |
| `{{PROJECT_NAME}}` | Target directory name |
| `{{STACK}}` | Detected stack, or an editing prompt |
| `{{COMMANDS}}` | Detected Markdown command rows, or an editing prompt |
| `{{PROFILE_GUIDELINES}}` | Selected profile contents |
| `{{LANGUAGE_GUIDELINES}}` | Selected language-pack contents |

Adapter templates may also use `{{PROJECT_NAME}}`. All other text is copied unchanged.

Unresolved known placeholders are reported by `doctor` and `check` so a customized template cannot silently produce incomplete instructions.

## Designing useful project instructions

Treat the template as a starting structure, not a place for every general coding preference. The most useful project instructions usually contain:

- Commands that are not obvious from standard manifests
- Architecture or ownership boundaries that code alone does not explain
- Repository-specific invariants and failure-prone workflows
- Real gotchas that caused previous mistakes

Keep broad language practices in guideline packs and working expectations in profiles. Keep facts unique to one repository in its `AGENTS.md`.
