package main

import (
	"fmt"
	"io"
	"strings"
)

func runHelpCommand(args []string, w io.Writer) error {
	if len(args) == 0 {
		printUsage(w)
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("usage: aiContext help [command]")
	}

	switch args[0] {
	case "init":
		printInitUsage(w)
	case "setup":
		printSetupUsage(w)
	case "doctor":
		printHealthUsage(w, false)
	case "check":
		printHealthUsage(w, true)
	case "diff":
		printDiffUsage(w)
	case "sync":
		printLifecycleUsage(w, false)
	case "update":
		printLifecycleUsage(w, true)
	case "tools":
		printToolsUsage(w)
	case "profiles":
		printProfilesUsage(w)
	case "languages":
		printLanguagesHelp(w)
	case "clean":
		printCleanUsage(w)
	case "version":
		printVersionUsage(w)
	case "help", "--help", "-h":
		printHelpUsage(w)
	default:
		return fmt.Errorf("unknown help topic %q; run 'aiContext help' to list commands", args[0])
	}
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "aiContext — keep one set of AI coding instructions in sync across tools")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "aiContext creates a project-owned AGENTS.md, generates small adapters for")
	fmt.Fprintln(w, "your coding agents, and safely tracks those files in .aicontext.json.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext <command> [options]")
	fmt.Fprintln(w, "  aiContext help [command]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Quick start:")
	fmt.Fprintln(w, "  aiContext setup                 Install editable templates (once per user)")
	fmt.Fprintln(w, "  cd path/to/project")
	fmt.Fprintln(w, "  aiContext init --detect         Create instructions using the detected stack")
	fmt.Fprintln(w, "  $EDITOR AGENTS.md               Add project-specific guidance")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  setup      Install or refresh user templates and guideline packs")
	fmt.Fprintln(w, "  init       Start managing a project, or adopt files already there")
	fmt.Fprintln(w, "  doctor     Explain the health of a managed project")
	fmt.Fprintln(w, "  check      Validate project health; designed for scripts and CI")
	fmt.Fprintln(w, "  diff       Preview changes needed to match installed templates")
	fmt.Fprintln(w, "  sync       Apply safe adapter updates using the current tool set")
	fmt.Fprintln(w, "  update     Change the selected tools and synchronize their adapters")
	fmt.Fprintln(w, "  tools      List supported coding agents and tool-selection values")
	fmt.Fprintln(w, "  profiles   List available working profiles and language packs")
	fmt.Fprintln(w, "  clean      Stop managing a project and remove unchanged generated files")
	fmt.Fprintln(w, "  version    Print the installed aiContext version")
	fmt.Fprintln(w, "  help       Show this guide or detailed help for one command")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Help topics:")
	fmt.Fprintln(w, "  languages  Select, find, and add language-specific rules")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Common workflows:")
	fmt.Fprintln(w, "  Preview first:          aiContext init --detect --dry-run")
	fmt.Fprintln(w, "  Use existing files:     aiContext init --adopt --tools codex,claude")
	fmt.Fprintln(w, "  Check and synchronize:  aiContext doctor && aiContext diff && aiContext sync")
	fmt.Fprintln(w, "  Change integrations:    aiContext update --tools codex,cursor,gemini")
	fmt.Fprintln(w, "  Remove safely:          aiContext clean --dry-run")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use 'aiContext help <command>' or 'aiContext <command> --help' for details.")
}

func printInitUsage(w io.Writer) {
	fmt.Fprintln(w, "Create AI instruction files and begin managing the project.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext init [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "By default, init creates AGENTS.md, selected tool adapters, and")
	fmt.Fprintln(w, ".aicontext.json. It preflights every destination and will not overwrite")
	fmt.Fprintln(w, "an existing file. Use --adopt when the instruction files already exist.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --adopt              Track existing instruction files without changing them")
	fmt.Fprintln(w, "  --detect             Read project manifests to infer stack and common commands")
	fmt.Fprintln(w, "  --dry-run            Validate and show outputs without writing files")
	fmt.Fprintf(w, "  --tools LIST         Comma-separated tools or 'all' (default: %s)\n", strings.Join(defaultTools, ","))
	fmt.Fprintln(w, "  --profile NAME       Working profile: minimal, standard, strict, or security")
	fmt.Fprintln(w, "                       (default: standard; run 'aiContext profiles')")
	fmt.Fprintln(w, "  --languages LIST     'auto', 'none', 'all', or comma-separated guideline packs")
	fmt.Fprintln(w, "                       (default: auto)")
	printDirectoryOptions(w, true)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  aiContext init --detect --dry-run")
	fmt.Fprintln(w, "  aiContext init --tools codex,claude,gemini --profile strict")
	fmt.Fprintln(w, "  aiContext init --languages go,typescript --target ../service")
	fmt.Fprintln(w, "  aiContext init --adopt --tools codex,claude")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Where the generated guidance goes:")
	fmt.Fprintln(w, "  Profiles and language packs are composed into AGENTS.md during init.")
	fmt.Fprintln(w, "  Tool adapter files only reference that canonical file. After init, edit")
	fmt.Fprintln(w, "  AGENTS.md directly; init cannot be rerun over an already managed project.")
	fmt.Fprintln(w, "  See 'aiContext help languages' for selection and troubleshooting.")
}

func printSetupUsage(w io.Writer) {
	fmt.Fprintln(w, "Install aiContext's editable templates, profiles, and language guidelines.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext setup [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run setup once after a manual installation, and again after upgrading to")
	fmt.Fprintln(w, "review new bundled assets. Existing files prompt before replacement unless")
	fmt.Fprintln(w, "--force is used.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --force              Replace existing local assets without prompting")
	fmt.Fprintln(w, "  --template-dir DIR   Use this template directory instead of the user config")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  aiContext setup")
	fmt.Fprintln(w, "  aiContext setup --template-dir ./team-templates")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Troubleshooting:")
	fmt.Fprintln(w, "  Templates are local, editable assets and can become stale after an upgrade.")
	fmt.Fprintln(w, "  If init reports selected guidance but the generated AGENTS.md lacks it, run")
	fmt.Fprintln(w, "  setup and accept the updated template. --force replaces every local asset,")
	fmt.Fprintln(w, "  including customizations.")
}

func printLanguagesHelp(w io.Writer) {
	fmt.Fprintln(w, "Select and manage language-specific rules in AGENTS.md.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "For a new project:")
	fmt.Fprintln(w, "  aiContext init --languages java")
	fmt.Fprintln(w, "  aiContext init --languages java,kotlin")
	fmt.Fprintln(w, "  aiContext init --languages auto       # default; detect from manifests")
	fmt.Fprintln(w, "  aiContext init --languages none       # omit language rules")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Supported values:")
	fmt.Fprintf(w, "  %s\n", supportedLanguageList())
	fmt.Fprintln(w, "  'all' selects every pack; 'auto' detects packs such as Java from pom.xml")
	fmt.Fprintln(w, "  or build.gradle. Run 'aiContext profiles' to inspect installed packs.")
	fmt.Fprintln(w, "  If the manifest is below the repository root or detection misses it, select")
	fmt.Fprintln(w, "  the pack explicitly with '--languages java'.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Where rules appear:")
	fmt.Fprintln(w, "  Selected packs are copied into the 'Language Guidelines' section of the")
	fmt.Fprintln(w, "  new AGENTS.md. CLAUDE.md and other adapters only reference AGENTS.md.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "For an already initialized project:")
	fmt.Fprintln(w, "  AGENTS.md is project-owned and lifecycle commands never regenerate it.")
	fmt.Fprintln(w, "  Edit its 'Language Guidelines' section directly. Use 'aiContext profiles'")
	fmt.Fprintln(w, "  to find the config root, then copy from guidelines/<language>.md if useful.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "If selected rules are missing after init:")
	fmt.Fprintln(w, "  Your local AGENTS.md template may predate language placeholders. Run")
	fmt.Fprintln(w, "  'aiContext setup' and accept the updated template. Use --force only if")
	fmt.Fprintln(w, "  replacing every customized template, profile, and guideline is acceptable.")
}

func printHealthUsage(w io.Writer, check bool) {
	name := "doctor"
	description := "Inspect a managed project and explain every health finding."
	if check {
		name = "check"
		description = "Validate a managed project and return a CI-friendly exit status."
	}
	fmt.Fprintln(w, description)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  aiContext %s [options]\n", name)
	fmt.Fprintln(w)
	if check {
		fmt.Fprintln(w, "check fails for errors such as missing managed files or invalid manifest")
		fmt.Fprintln(w, "state. With --strict, warnings such as modified adapters also fail.")
	} else {
		fmt.Fprintln(w, "doctor reports missing, modified, and healthy managed files for a person")
		fmt.Fprintln(w, "to review; health findings themselves do not produce a failing exit status.")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --target DIR         Inspect this project (default: current directory)")
	if check {
		fmt.Fprintln(w, "  --strict             Treat warnings as failures as well as errors")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  aiContext %s\n", name)
	if check {
		fmt.Fprintln(w, "  aiContext check --strict")
	}
}

func printDiffUsage(w io.Writer) {
	fmt.Fprintln(w, "Preview how managed adapters differ from the installed templates.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext diff [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "diff never writes files. It shows adapters that would be created, updated,")
	fmt.Fprintln(w, "or removed. Human-modified adapters are conflicts and cause a failing exit.")
	fmt.Fprintln(w, "AGENTS.md is project-owned and is never regenerated by lifecycle commands.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	printDirectoryOptions(w, true)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Example:")
	fmt.Fprintln(w, "  aiContext diff --target ../service")
}

func printLifecycleUsage(w io.Writer, update bool) {
	name := "sync"
	description := "Safely synchronize generated adapters with the installed templates."
	if update {
		name = "update"
		description = "Change the selected coding tools and synchronize their adapters."
	}
	fmt.Fprintln(w, description)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  aiContext %s [options]\n", name)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Unchanged generated adapters may be created, updated, or removed. Modified")
	fmt.Fprintln(w, "adapters are reported as conflicts and are never overwritten. AGENTS.md is")
	fmt.Fprintln(w, "project-owned and is not regenerated.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --dry-run            Show the plan without changing files")
	if update {
		fmt.Fprintln(w, "  --tools LIST         Replace the selected tool set (comma-separated or 'all')")
	}
	printDirectoryOptions(w, true)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  aiContext %s --dry-run\n", name)
	if update {
		fmt.Fprintln(w, "  aiContext update --tools codex,cursor,gemini")
	} else {
		fmt.Fprintln(w, "  aiContext sync")
	}
}

func printProfilesUsage(w io.Writer) {
	fmt.Fprintln(w, "List working profiles and language guideline packs available to init.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext profiles [--template-dir DIR]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Profiles define the project's working agreement. Language packs add focused")
	fmt.Fprintln(w, "guidance for detected or explicitly selected languages.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Option:")
	fmt.Fprintln(w, "  --template-dir DIR   Read profiles from this directory's asset collection")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Example:")
	fmt.Fprintln(w, "  aiContext profiles")
}

func printToolsUsage(w io.Writer) {
	fmt.Fprintln(w, "List values accepted by init and update's --tools option.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext tools")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "AGENTS.md is always the canonical file. Selecting another tool adds its")
	fmt.Fprintln(w, "adapter; use the special value 'all' to select every supported tool.")
	fmt.Fprintln(w)
	printTools(w)
}

func printCleanUsage(w io.Writer) {
	fmt.Fprintln(w, "Stop managing a project and remove unchanged generated files safely.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext clean [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "clean removes files only when their content still matches the manifest.")
	fmt.Fprintln(w, "Modified or symlinked files are preserved. The manifest is then removed,")
	fmt.Fprintln(w, "so aiContext no longer manages the project.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --dry-run            Show what would be removed or preserved")
	fmt.Fprintln(w, "  --target DIR         Clean this project (default: current directory)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Recommended:")
	fmt.Fprintln(w, "  aiContext clean --dry-run")
	fmt.Fprintln(w, "  aiContext clean")
}

func printVersionUsage(w io.Writer) {
	fmt.Fprintln(w, "Print the installed aiContext version.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext version")
	fmt.Fprintln(w, "  aiContext --version")
}

func printHelpUsage(w io.Writer) {
	fmt.Fprintln(w, "Show the command overview or detailed help for one command or topic.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext help")
	fmt.Fprintln(w, "  aiContext help <command-or-topic>")
	fmt.Fprintln(w, "  aiContext <command> --help")
}

func printDirectoryOptions(w io.Writer, templates bool) {
	fmt.Fprintln(w, "  --target DIR         Project directory (default: current directory)")
	if templates {
		fmt.Fprintln(w, "  --template-dir DIR   Template directory (default: user config directory)")
	}
}
