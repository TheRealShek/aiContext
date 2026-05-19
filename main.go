package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var defaultTemplates embed.FS

const configDir = ".config/aiContext/templates"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		runInit()
	case "setup":
		runSetup()
	default:
		printUsage()
		os.Exit(1)
	}
}

// runInit writes AGENTS.md + CLAUDE.md into cwd from user's config templates.
func runInit() {
	cwd, err := os.Getwd()
	if err != nil {
		fatal("cannot read current dir:", err)
	}
	projectName := filepath.Base(cwd)

	// refuse to overwrite
	for _, f := range []string{"AGENTS.md", "CLAUDE.md"} {
		if _, err := os.Stat(f); err == nil {
			fatalf("%s already exists — aborting", f)
		}
	}

	templateDir := userTemplateDir()

	for _, f := range []string{"AGENTS.md", "CLAUDE.md"} {
		src := filepath.Join(templateDir, f)
		raw, err := os.ReadFile(src)
		if err != nil {
			fatalf("template %s missing — run: aiContext setup", f)
		}

		content := strings.ReplaceAll(string(raw), "{{PROJECT_NAME}}", projectName)

		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			fatalf("cannot write %s: %v", f, err)
		}
		fmt.Println("✓", f)
	}
}

// runSetup copies embedded defaults into ~/.config/aiContext/templates/.
// Safe to re-run — prompts before overwriting existing files.
func runSetup() {
	templateDir := userTemplateDir()

	if err := os.MkdirAll(templateDir, 0755); err != nil {
		fatal("cannot create config dir:", err)
	}

	entries, err := defaultTemplates.ReadDir("templates")
	if err != nil {
		fatal("cannot read embedded templates:", err)
	}

	for _, entry := range entries {
		dest := filepath.Join(templateDir, entry.Name())

		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("? %s exists — overwrite? [y/N]: ", entry.Name())
			var ans string
			fmt.Scanln(&ans)
			if strings.ToLower(ans) != "y" {
				fmt.Println("skip", entry.Name())
				continue
			}
		}

		data, _ := defaultTemplates.ReadFile("templates/" + entry.Name())
		if err := os.WriteFile(dest, data, 0644); err != nil {
			fatalf("cannot write %s: %v", dest, err)
		}
		fmt.Println("✓", dest)
	}
}

func userTemplateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fatal("cannot resolve home dir:", err)
	}
	return filepath.Join(home, configDir)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  aiContext setup   — install default templates to ~/.config/aiContext/templates/")
	fmt.Fprintln(os.Stderr, "  aiContext init    — write AGENTS.md + CLAUDE.md into current dir")
}

func fatal(msg string, args ...any) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(append([]any{msg}, args...)...))
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
