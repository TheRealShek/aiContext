package main

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var defaultTemplates embed.FS

const configDir = ".config/aiContext/templates"

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
	if len(args) != 1 {
		printUsage(stdout)
		return errUsage
	}

	switch args[0] {
	case "init":
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot read current dir: %w", err)
		}
		templateDir, err := userTemplateDir()
		if err != nil {
			return err
		}
		return runInit(cwd, templateDir, stdout)
	case "setup":
		templateDir, err := userTemplateDir()
		if err != nil {
			return err
		}
		return runSetup(templateDir, stdin, stdout)
	default:
		printUsage(stdout)
		return errUsage
	}
}

type pendingFile struct {
	path    string
	display string
	content []byte
}

// runInit validates every template and destination before creating project files.
// If any write fails, files created by this invocation are removed.
func runInit(cwd, templateDir string, stdout io.Writer) (err error) {
	projectName := filepath.Base(filepath.Clean(cwd))
	files := make([]pendingFile, 0, len(projectTemplates))

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

		content := strings.ReplaceAll(string(raw), "{{PROJECT_NAME}}", projectName)
		files = append(files, pendingFile{
			path:    destPath,
			display: spec.destination,
			content: []byte(content),
		})
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
func runSetup(templateDir string, stdin io.Reader, stdout io.Writer) error {
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	input := bufio.NewReader(stdin)
	for _, spec := range projectTemplates {
		dest := filepath.Join(templateDir, spec.source)

		if _, err := os.Lstat(dest); err == nil {
			fmt.Fprintf(stdout, "? %s exists — overwrite? [y/N]: ", spec.source)
			answer, readErr := input.ReadString('\n')
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return fmt.Errorf("cannot read overwrite response: %w", readErr)
			}
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Fprintln(stdout, "skip", spec.source)
				continue
			}
		} else if !errors.Is(err, os.ErrNotExist) {
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
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home dir: %w", err)
	}
	return filepath.Join(home, configDir), nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  aiContext setup   — install default templates to ~/.config/aiContext/templates/")
	fmt.Fprintln(w, "  aiContext init    — write AI agent instruction files into the current directory")
}
