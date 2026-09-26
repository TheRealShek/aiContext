package main

import (
	"bufio"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/* profiles/* guidelines/*
var defaultTemplates embed.FS

const configSubdir = "aiContext/templates"

var version = "dev"

var errUsage = errors.New("invalid command")

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		if !errors.Is(err, errUsage) {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return errUsage
	}

	switch args[0] {
	case "init":
		return runInitCommand(args[1:], stdout)
	case "setup":
		return runSetupCommand(args[1:], stdin, stdout)
	case "profiles":
		return runProfilesCommand(args[1:], stdout)
	case "version", "--version":
		if args[0] == "version" && len(args) == 2 && isHelpFlag(args[1]) {
			printVersionUsage(stdout)
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("version does not accept arguments")
		}
		fmt.Fprintf(stdout, "aiContext %s\n", version)
		return nil
	case "help", "--help", "-h":
		if args[0] == "help" {
			return runHelpCommand(args[1:], stdout)
		}
		if len(args) != 1 {
			return fmt.Errorf("%s does not accept arguments", args[0])
		}
		printUsage(stdout)
		return nil
	default:
		printUsage(stdout)
		return errUsage
	}
}

func isHelpFlag(arg string) bool {
	return arg == "--help" || arg == "-h"
}

func runProfilesCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("profiles", flag.ContinueOnError)
	flags.SetOutput(stdout)
	templateDir := flags.String("template-dir", "", "template directory (default: user config directory)")
	flags.Usage = func() { printProfilesUsage(stdout) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("profiles does not accept positional arguments")
	}
	resolvedTemplates, err := resolveTemplateDir(*templateDir)
	if err != nil {
		return err
	}
	return printProfiles(stdout, resolvedTemplates)
}

func runInitCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stdout)
	target := flags.String("target", "", "project directory (default: current directory)")
	templateDir := flags.String("template-dir", "", "template directory (default: user config directory)")
	dryRun := flags.Bool("dry-run", false, "validate and show AGENTS.md without writing it")
	detect := flags.Bool("detect", false, "detect the project stack and common commands")
	profileFlag := flags.String("profile", "", "working profile (default: standard)")
	languagesFlag := flags.String("languages", "", "language guideline packs (default: auto)")
	flags.Usage = func() { printInitUsage(stdout) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("init does not accept positional arguments")
	}

	projectDir, err := resolveTargetDir(*target)
	if err != nil {
		return err
	}
	profileName := *profileFlag
	languageSelection := *languagesFlag
	if profileName == "" {
		profileName = "standard"
	}
	if languageSelection == "" {
		languageSelection = "auto"
	}
	languages, err := resolveLanguageSelection(projectDir, languageSelection)
	if err != nil {
		return err
	}
	resolvedTemplates, err := resolveTemplateDir(*templateDir)
	if err != nil {
		return err
	}
	return runInit(projectDir, resolvedTemplates, stdout, initOptions{dryRun: *dryRun, detect: *detect, profile: profileName, languages: languages})
}

func runSetupCommand(args []string, stdin io.Reader, stdout io.Writer) error {
	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	flags.SetOutput(stdout)
	templateDir := flags.String("template-dir", "", "template directory (default: user config directory)")
	force := flags.Bool("force", false, "overwrite existing templates without prompting")
	flags.Usage = func() { printSetupUsage(stdout) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("setup does not accept positional arguments")
	}

	resolvedTemplates, err := resolveTemplateDir(*templateDir)
	if err != nil {
		return err
	}
	return runSetup(resolvedTemplates, stdin, stdout, *force)
}

type initOptions struct {
	dryRun    bool
	detect    bool
	profile   string
	languages []string
}

// runInit creates one project-owned AGENTS.md and never replaces an existing file.
func runInit(cwd, templateDir string, stdout io.Writer, options initOptions) error {
	projectName := filepath.Base(filepath.Clean(cwd))
	templateStack := "[Language · Framework · DB · infra — only non-obvious choices]"
	templateCommands := "<!-- Add one row per non-obvious project command. -->"
	profileGuidelines := "<!-- Add the working agreement for this project. -->"
	languageGuidelines := "<!-- Add language-specific guidance when it prevents real mistakes. -->"
	if options.profile != "" {
		loadedProfile, loadedLanguages, err := loadProfileGuidance(templateDir, options.profile, options.languages)
		if err != nil {
			return err
		}
		profileGuidelines = loadedProfile
		languageGuidelines = loadedLanguages
		fmt.Fprintln(stdout, "profile:", options.profile)
		if len(options.languages) > 0 {
			fmt.Fprintln(stdout, "language guidelines:", strings.Join(options.languages, ", "))
		}
	}
	if options.detect {
		context, err := detectProjectContext(cwd)
		if err != nil {
			return err
		}
		templateStack = context.stack
		templateCommands = context.commands
		fmt.Fprintln(stdout, "detected stack:", context.stack)
	}
	raw, err := os.ReadFile(filepath.Join(templateDir, "AGENTS.md"))
	if err != nil {
		return fmt.Errorf("cannot read template AGENTS.md (run: aiContext setup): %w", err)
	}
	if err := validateCanonicalTemplate(raw, options); err != nil {
		return err
	}
	content := strings.NewReplacer(
		"{{PROJECT_NAME}}", projectName,
		"{{STACK}}", templateStack,
		"{{COMMANDS}}", templateCommands,
		"{{PROFILE_GUIDELINES}}", profileGuidelines,
		"{{LANGUAGE_GUIDELINES}}", languageGuidelines,
	).Replace(string(raw))
	path := filepath.Join(cwd, "AGENTS.md")
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("AGENTS.md already exists — aborting")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot inspect AGENTS.md: %w", err)
	}
	if options.dryRun {
		fmt.Fprintln(stdout, "would create AGENTS.md")
		return nil
	}
	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("cannot create AGENTS.md: %w", err)
	}
	if _, err := io.WriteString(output, content); err != nil {
		_ = output.Close()
		_ = os.Remove(path)
		return fmt.Errorf("cannot write AGENTS.md: %w", err)
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("cannot close AGENTS.md: %w", err)
	}
	fmt.Fprintln(stdout, "✓ AGENTS.md")
	return nil
}

func validateCanonicalTemplate(content []byte, options initOptions) error {
	template := string(content)
	required := make([]string, 0, 4)
	if options.detect {
		required = append(required, "{{STACK}}", "{{COMMANDS}}")
	}
	if options.profile != "" {
		required = append(required, "{{PROFILE_GUIDELINES}}")
	}
	if len(options.languages) > 0 {
		required = append(required, "{{LANGUAGE_GUIDELINES}}")
	}

	missing := make([]string, 0, len(required))
	for _, placeholder := range required {
		if !strings.Contains(template, placeholder) {
			missing = append(missing, placeholder)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf(
		"AGENTS.md template is missing required placeholder(s) %s; run 'aiContext setup' and accept the updated AGENTS.md template, or add the placeholder(s) manually",
		strings.Join(missing, ", "),
	)
}

// runSetup copies embedded defaults into the user's template directory.
// It prompts before overwriting an existing template.
func runSetup(templateDir string, stdin io.Reader, stdout io.Writer, force bool) error {
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	input := bufio.NewReader(stdin)
	for _, asset := range setupAssets(templateDir) {
		dest := asset.destination

		if _, err := os.Lstat(dest); err == nil && !force {
			fmt.Fprintf(stdout, "? %s exists — overwrite? [y/N]: ", asset.label)
			answer, readErr := input.ReadString('\n')
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return fmt.Errorf("cannot read overwrite response: %w", readErr)
			}
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Fprintln(stdout, "skip", asset.label)
				continue
			}
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("cannot inspect %s: %w", dest, err)
		}

		data, err := defaultTemplates.ReadFile(asset.embedded)
		if err != nil {
			return fmt.Errorf("cannot read embedded asset %s: %w", asset.label, err)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("cannot create directory for %s: %w", asset.label, err)
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("cannot write %s: %w", dest, err)
		}
		fmt.Fprintln(stdout, "✓", dest)
	}
	return nil
}

func userTemplateDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve user config dir: %w", err)
	}
	return filepath.Join(base, filepath.FromSlash(configSubdir)), nil
}

func resolveTargetDir(target string) (string, error) {
	if target == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot read current dir: %w", err)
		}
		target = cwd
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("cannot resolve target directory: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("cannot inspect target directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("target is not a directory: %s", abs)
	}
	return abs, nil
}

func resolveTemplateDir(dir string) (string, error) {
	if dir == "" {
		return userTemplateDir()
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve template directory: %w", err)
	}
	return abs, nil
}
