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

//go:embed templates/*
var defaultTemplates embed.FS

const configSubdir = "aiContext/templates"

var version = "dev"

var errUsage = errors.New("invalid command")

type templateSpec struct {
	source      string
	destination string
}

var projectTemplates = []templateSpec{
	{source: "AGENTS.md", destination: "AGENTS.md"},
	{source: "CLAUDE.md", destination: "CLAUDE.md"},
	{source: "cursor.mdc", destination: ".cursor/rules/aicontext.mdc"},
	{source: "copilot-instructions.md", destination: ".github/copilot-instructions.md"},
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
	case "version", "--version":
		if len(args) != 1 {
			return fmt.Errorf("version does not accept arguments")
		}
		fmt.Fprintf(stdout, "aiContext %s\n", version)
		return nil
	case "help", "--help", "-h":
		if len(args) != 1 {
			return fmt.Errorf("help does not accept arguments")
		}
		printUsage(stdout)
		return nil
	default:
		printUsage(stdout)
		return errUsage
	}
}

func runInitCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stdout)
	target := flags.String("target", "", "project directory (default: current directory)")
	templateDir := flags.String("template-dir", "", "template directory (default: user config directory)")
	dryRun := flags.Bool("dry-run", false, "validate and show files without writing them")
	detect := flags.Bool("detect", false, "detect the project stack and common commands")
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
	resolvedTemplates, err := resolveTemplateDir(*templateDir)
	if err != nil {
		return err
	}
	return runInit(projectDir, resolvedTemplates, stdout, initOptions{dryRun: *dryRun, detect: *detect})
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
	content []byte
}

type initOptions struct {
	dryRun bool
	detect bool
}

// runInit validates every template and destination before creating project files.
// If any write fails, files created by this invocation are removed.
func runInit(cwd, templateDir string, stdout io.Writer, options initOptions) (err error) {
	projectName := filepath.Base(filepath.Clean(cwd))
	files := make([]pendingFile, 0, len(projectTemplates))
	templateStack := "[Language · Framework · DB · infra — only non-obvious choices]"
	templateCommands := "<!-- Add one row per non-obvious project command. -->"
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
	)

	for _, spec := range projectTemplates {
		srcPath := filepath.Join(templateDir, spec.source)
		raw, readErr := os.ReadFile(srcPath)
		if readErr != nil {
			return fmt.Errorf("cannot read template %s (run: aiContext setup): %w", spec.source, readErr)
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
			content: []byte(content),
		})
	}
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

// runSetup copies embedded defaults into the user's template directory.
// It prompts before overwriting an existing template.
func runSetup(templateDir string, stdin io.Reader, stdout io.Writer, force bool) error {
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	input := bufio.NewReader(stdin)
	for _, spec := range projectTemplates {
		dest := filepath.Join(templateDir, spec.source)

		if _, err := os.Lstat(dest); err == nil && !force {
			fmt.Fprintf(stdout, "? %s exists — overwrite? [y/N]: ", spec.source)
			answer, readErr := input.ReadString('\n')
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return fmt.Errorf("cannot read overwrite response: %w", readErr)
			}
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Fprintln(stdout, "skip", spec.source)
				continue
			}
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("cannot inspect %s: %w", dest, err)
		}

		data, err := defaultTemplates.ReadFile("templates/" + spec.source)
		if err != nil {
			return fmt.Errorf("cannot read embedded template %s: %w", spec.source, err)
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

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "aiContext bootstraps AI agent instruction files into a project.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aiContext setup [options]")
	fmt.Fprintln(w, "  aiContext init [options]")
	fmt.Fprintln(w, "  aiContext version")
	fmt.Fprintln(w, "  aiContext help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run 'aiContext <command> --help' for command options.")
}

func printInitUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: aiContext init [--detect] [--dry-run] [--target DIR] [--template-dir DIR]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --detect            Detect the project stack and common commands")
	fmt.Fprintln(w, "  --dry-run           Validate and show outputs without writing files")
	fmt.Fprintln(w, "  --target DIR         Project directory (default: current directory)")
	fmt.Fprintln(w, "  --template-dir DIR   Template directory (default: user config directory)")
}

func printSetupUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: aiContext setup [--force] [--template-dir DIR]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --force              Overwrite existing templates without prompting")
	fmt.Fprintln(w, "  --template-dir DIR   Template directory (default: user config directory)")
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
