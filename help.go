package main

import (
	"fmt"
	"io"
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
	case "profiles":
		printProfilesUsage(w)
	case "languages":
		printLanguagesHelp(w)
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
	fmt.Fprintln(w, "aiContext creates one AGENTS.md with project-specific AI coding instructions.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext <command> [options]")
	fmt.Fprintln(w, "  aiContext help [command]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Quick start:")
	fmt.Fprintln(w, "  aiContext setup                 Install editable templates (once per user)")
	fmt.Fprintln(w, "  cd path/to/project")
	fmt.Fprintln(w, "  aiContext init --detect         Create AGENTS.md using the detected stack")
	fmt.Fprintln(w, "  $EDITOR AGENTS.md               Add project-specific guidance")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  setup      Install or refresh the template, profiles, and guidelines")
	fmt.Fprintln(w, "  init       Create AGENTS.md in a project")
	fmt.Fprintln(w, "  profiles   List available working profiles and language packs")
	fmt.Fprintln(w, "  version    Print the installed aiContext version")
	fmt.Fprintln(w, "  help       Show this guide or detailed help for one command")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Help topics:")
	fmt.Fprintln(w, "  languages  Select, find, and add language-specific rules")
}

func printInitUsage(w io.Writer) {
	fmt.Fprintln(w, "Create one project-owned AGENTS.md.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext init [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "init refuses to replace an existing AGENTS.md. Edit that file directly.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --detect             Read project manifests to infer stack and common commands")
	fmt.Fprintln(w, "  --dry-run            Validate and show the output without writing it")
	fmt.Fprintln(w, "  --profile NAME       Working profile (default: standard)")
	fmt.Fprintln(w, "  --languages LIST     'auto', 'none', 'all', or comma-separated packs (default: auto)")
	printDirectoryOptions(w, true)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  aiContext init --detect --dry-run")
	fmt.Fprintln(w, "  aiContext init --profile strict --languages go,typescript")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Profiles and language packs are copied into AGENTS.md during init.")
	fmt.Fprintln(w, "See 'aiContext help languages' for selection and troubleshooting.")
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
	fmt.Fprintln(w, "Select language-specific rules for AGENTS.md.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  aiContext init --languages java")
	fmt.Fprintln(w, "  aiContext init --languages java,kotlin")
	fmt.Fprintln(w, "  aiContext init --languages auto       # default; detect from project files")
	fmt.Fprintln(w, "  aiContext init --languages none       # omit language rules")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Supported values:")
	fmt.Fprintf(w, "  %s\n", supportedLanguageList())
	fmt.Fprintln(w, "  'all' selects every pack. If detection misses a language, select it explicitly.")
	fmt.Fprintln(w, "  For an existing project, edit AGENTS.md directly.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "If selected rules are missing, update your local template with aiContext setup.")
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
