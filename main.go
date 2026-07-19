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

type templateSpec struct {
	source      string
	destination string
}

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
	case "doctor":
		return runHealthCommand(args[1:], stdout, false)
	case "check":
		return runHealthCommand(args[1:], stdout, true)
	case "diff":
		return runDiffCommand(args[1:], stdout)
	case "sync":
		return runLifecycleCommand(args[1:], stdout, false)
	case "update":
		return runLifecycleCommand(args[1:], stdout, true)
	case "tools":
		if len(args) == 2 && isHelpFlag(args[1]) {
			printToolsUsage(stdout)
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("tools does not accept arguments")
		}
		printTools(stdout)
		return nil
	case "profiles":
		return runProfilesCommand(args[1:], stdout)
	case "clean":
		return runCleanCommand(args[1:], stdout)
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

func runHealthCommand(args []string, stdout io.Writer, check bool) error {
	name := "doctor"
	if check {
		name = "check"
	}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stdout)
	target := flags.String("target", "", "project directory (default: current directory)")
	strict := false
	if check {
		flags.BoolVar(&strict, "strict", false, "treat warnings as failures")
	}
	flags.Usage = func() { printHealthUsage(stdout, check) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("%s does not accept positional arguments", name)
	}
	projectDir, err := resolveTargetDir(*target)
	if err != nil {
		return err
	}
	report, err := inspectProjectHealth(projectDir)
	if err != nil {
		return err
	}
	printHealthReport(stdout, report)
	if check {
		return healthCheckError(report, strict)
	}
	return nil
}

func runDiffCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("diff", flag.ContinueOnError)
	flags.SetOutput(stdout)
	target := flags.String("target", "", "project directory (default: current directory)")
	templateDir := flags.String("template-dir", "", "template directory (default: user config directory)")
	flags.Usage = func() { printDiffUsage(stdout) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("diff does not accept positional arguments")
	}
	projectDir, resolvedTemplates, err := resolveLifecycleDirs(*target, *templateDir)
	if err != nil {
		return err
	}
	plan, err := buildLifecyclePlan(projectDir, resolvedTemplates, nil)
	if err != nil {
		return err
	}
	printLifecyclePlan(stdout, plan)
	if len(plan.conflicts) > 0 {
		return fmt.Errorf("found %d lifecycle conflict(s)", len(plan.conflicts))
	}
	return nil
}

func runLifecycleCommand(args []string, stdout io.Writer, allowToolChange bool) error {
	name := "sync"
	if allowToolChange {
		name = "update"
	}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stdout)
	target := flags.String("target", "", "project directory (default: current directory)")
	templateDir := flags.String("template-dir", "", "template directory (default: user config directory)")
	dryRun := flags.Bool("dry-run", false, "show the lifecycle plan without changing files")
	toolsFlag := ""
	if allowToolChange {
		flags.StringVar(&toolsFlag, "tools", "", "replace the selected tool set")
	}
	flags.Usage = func() { printLifecycleUsage(stdout, allowToolChange) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("%s does not accept positional arguments", name)
	}
	projectDir, resolvedTemplates, err := resolveLifecycleDirs(*target, *templateDir)
	if err != nil {
		return err
	}
	var tools []string
	if toolsFlag != "" {
		tools, err = parseToolSelection(toolsFlag)
		if err != nil {
			return err
		}
	}
	plan, err := buildLifecyclePlan(projectDir, resolvedTemplates, tools)
	if err != nil {
		return err
	}
	printLifecyclePlan(stdout, plan)
	if len(plan.conflicts) > 0 {
		return fmt.Errorf("found %d lifecycle conflict(s)", len(plan.conflicts))
	}
	if *dryRun {
		return nil
	}
	return applyLifecyclePlan(projectDir, plan, stdout)
}

func resolveLifecycleDirs(target, templateDir string) (string, string, error) {
	projectDir, err := resolveTargetDir(target)
	if err != nil {
		return "", "", err
	}
	resolvedTemplates, err := resolveTemplateDir(templateDir)
	if err != nil {
		return "", "", err
	}
	return projectDir, resolvedTemplates, nil
}

func printTools(w io.Writer) {
	fmt.Fprintln(w, "Supported tools (AGENTS.md is always generated as the canonical file):")
	for _, tool := range toolDefinitions {
		fmt.Fprintf(w, "  %-8s %s\n", tool.name, tool.description)
	}
}

func runInitCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stdout)
	target := flags.String("target", "", "project directory (default: current directory)")
	templateDir := flags.String("template-dir", "", "template directory (default: user config directory)")
	dryRun := flags.Bool("dry-run", false, "validate and show files without writing them")
	detect := flags.Bool("detect", false, "detect the project stack and common commands")
	adopt := flags.Bool("adopt", false, "create lifecycle state for existing instruction files")
	toolsFlag := flags.String("tools", strings.Join(defaultTools, ","), "comma-separated tools or 'all'")
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
	tools, err := parseToolSelection(*toolsFlag)
	if err != nil {
		return err
	}
	profileName := *profileFlag
	languageSelection := *languagesFlag
	if *adopt {
		if profileName == "" {
			profileName = "adopted"
		}
		if languageSelection == "" {
			languageSelection = "none"
		}
	} else {
		if profileName == "" {
			profileName = "standard"
		}
		if languageSelection == "" {
			languageSelection = "auto"
		}
	}
	languages, err := resolveLanguageSelection(projectDir, languageSelection)
	if err != nil {
		return err
	}
	if *adopt {
		return runAdopt(projectDir, tools, *detect, profileName, languages, stdout, *dryRun)
	}
	resolvedTemplates, err := resolveTemplateDir(*templateDir)
	if err != nil {
		return err
	}
	return runInit(projectDir, resolvedTemplates, stdout, initOptions{dryRun: *dryRun, detect: *detect, tools: tools, profile: profileName, languages: languages})
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

type pendingFile struct {
	path    string
	display string
	source  string
	content []byte
}

type initOptions struct {
	dryRun    bool
	detect    bool
	tools     []string
	profile   string
	languages []string
}

// runInit validates every template and destination before creating project files.
// If any write fails, files created by this invocation are removed.
func runInit(cwd, templateDir string, stdout io.Writer, options initOptions) (err error) {
	if len(options.tools) == 0 {
		options.tools = append([]string(nil), defaultTools...)
	}
	projectName := filepath.Base(filepath.Clean(cwd))
	specs := templateSpecsForTools(options.tools)
	files := make([]pendingFile, 0, len(specs)+1)
	templateStack := "[Language · Framework · DB · infra — only non-obvious choices]"
	templateCommands := "<!-- Add one row per non-obvious project command. -->"
	profileGuidelines := "<!-- Add the working agreement for this project. -->"
	languageGuidelines := "<!-- Add language-specific guidance when it prevents real mistakes. -->"
	if options.profile != "" && options.profile != "adopted" {
		loadedProfile, loadedLanguages, loadErr := loadProfileGuidance(templateDir, options.profile, options.languages)
		if loadErr != nil {
			return loadErr
		}
		profileGuidelines = loadedProfile
		languageGuidelines = loadedLanguages
		fmt.Fprintln(stdout, "profile:", options.profile)
		if len(options.languages) > 0 {
			fmt.Fprintln(stdout, "language guidelines:", strings.Join(options.languages, ", "))
		}
	}
	if options.detect {
		context, detectErr := detectProjectContext(cwd)
		if detectErr != nil {
			return detectErr
		}
		templateStack = context.stack
		templateCommands = context.commands
		fmt.Fprintln(stdout, "detected stack:", context.stack)
	}
	replacer := strings.NewReplacer(
		"{{PROJECT_NAME}}", projectName,
		"{{STACK}}", templateStack,
		"{{COMMANDS}}", templateCommands,
		"{{PROFILE_GUIDELINES}}", profileGuidelines,
		"{{LANGUAGE_GUIDELINES}}", languageGuidelines,
	)

	for _, spec := range specs {
		srcPath := filepath.Join(templateDir, spec.source)
		raw, readErr := os.ReadFile(srcPath)
		if readErr != nil {
			return fmt.Errorf("cannot read template %s (run: aiContext setup): %w", spec.source, readErr)
		}
		if spec.destination == "AGENTS.md" {
			if validationErr := validateCanonicalTemplate(raw, options); validationErr != nil {
				return validationErr
			}
		}

		destPath := filepath.Join(cwd, filepath.FromSlash(spec.destination))
		if _, statErr := os.Lstat(destPath); statErr == nil {
			return fmt.Errorf("%s already exists — aborting", spec.destination)
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("cannot inspect %s: %w", spec.destination, statErr)
		}

		content := replacer.Replace(string(raw))
		files = append(files, pendingFile{
			path:    destPath,
			display: spec.destination,
			source:  spec.source,
			content: []byte(content),
		})
	}

	manifest := newProjectManifest(options.tools, options.detect, options.profile, options.languages, files)
	manifestData, manifestErr := marshalManifest(manifest)
	if manifestErr != nil {
		return manifestErr
	}
	manifestPath := filepath.Join(cwd, manifestFilename)
	if _, statErr := os.Lstat(manifestPath); statErr == nil {
		return fmt.Errorf("%s already exists — aborting", manifestFilename)
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("cannot inspect %s: %w", manifestFilename, statErr)
	}
	files = append(files, pendingFile{path: manifestPath, display: manifestFilename, content: manifestData})
	if options.dryRun {
		for _, file := range files {
			fmt.Fprintln(stdout, "would create", file.display)
		}
		return nil
	}

	for _, file := range files {
		if mkdirErr := os.MkdirAll(filepath.Dir(file.path), 0o755); mkdirErr != nil {
			return fmt.Errorf("cannot create directory for %s: %w", file.display, mkdirErr)
		}
	}

	created := make([]string, 0, len(files))
	defer func() {
		if err == nil {
			return
		}
		for i := len(created) - 1; i >= 0; i-- {
			if removeErr := os.Remove(created[i]); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("cannot roll back %s: %w", created[i], removeErr))
			}
		}
	}()

	for _, file := range files {
		output, openErr := os.OpenFile(file.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if openErr != nil {
			return fmt.Errorf("cannot create %s: %w", file.display, openErr)
		}
		created = append(created, file.path)

		if _, writeErr := output.Write(file.content); writeErr != nil {
			_ = output.Close()
			return fmt.Errorf("cannot write %s: %w", file.display, writeErr)
		}
		if closeErr := output.Close(); closeErr != nil {
			return fmt.Errorf("cannot close %s: %w", file.display, closeErr)
		}
	}

	for _, file := range files {
		fmt.Fprintln(stdout, "✓", file.display)
	}
	return nil
}

func validateCanonicalTemplate(content []byte, options initOptions) error {
	template := string(content)
	required := make([]string, 0, 4)
	if options.detect {
		required = append(required, "{{STACK}}", "{{COMMANDS}}")
	}
	if options.profile != "" && options.profile != "adopted" {
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

func runAdopt(cwd string, tools []string, detect bool, profile string, languages []string, stdout io.Writer, dryRun bool) error {
	manifestPath := filepath.Join(cwd, manifestFilename)
	if _, err := os.Lstat(manifestPath); err == nil {
		return fmt.Errorf("%s already exists — project is already managed", manifestFilename)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot inspect %s: %w", manifestFilename, err)
	}

	files := make([]pendingFile, 0, len(templateSpecsForTools(tools)))
	for _, spec := range templateSpecsForTools(tools) {
		path := filepath.Join(cwd, filepath.FromSlash(spec.destination))
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot adopt %s: %w", spec.destination, err)
		}
		files = append(files, pendingFile{path: path, display: spec.destination, source: spec.source, content: content})
	}
	manifest := newProjectManifest(tools, detect, profile, languages, files)
	data, err := marshalManifest(manifest)
	if err != nil {
		return err
	}
	if dryRun {
		for _, file := range files {
			fmt.Fprintln(stdout, "would adopt", file.display)
		}
		fmt.Fprintln(stdout, "would create", manifestFilename)
		return nil
	}
	output, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", manifestFilename, err)
	}
	if _, err := output.Write(data); err != nil {
		_ = output.Close()
		_ = os.Remove(manifestPath)
		return fmt.Errorf("cannot write %s: %w", manifestFilename, err)
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(manifestPath)
		return fmt.Errorf("cannot close %s: %w", manifestFilename, err)
	}
	for _, file := range files {
		fmt.Fprintln(stdout, "✓ adopted", file.display)
	}
	fmt.Fprintln(stdout, "✓", manifestFilename)
	return nil
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
