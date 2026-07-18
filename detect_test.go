package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectProjectContextGo(t *testing.T) {
	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "go.mod", `module example.com/service

go 1.25

require (
	github.com/gin-gonic/gin v1.10.0
	gorm.io/gorm v1.30.0
)
`)
	writeProjectFile(t, projectDir, "Dockerfile", "FROM scratch\n")

	context, err := detectProjectContext(projectDir)
	if err != nil {
		t.Fatalf("detectProjectContext() error = %v", err)
	}
	if want := "Go · Gin · GORM · Docker"; context.stack != want {
		t.Errorf("stack = %q, want %q", context.stack, want)
	}
	for _, want := range []string{"`go build ./...`", "`go test ./...`", "`go vet ./...`", "`docker build .`"} {
		if !strings.Contains(context.commands, want) {
			t.Errorf("commands = %q, want %q", context.commands, want)
		}
	}
}

func TestDetectProjectContextNode(t *testing.T) {
	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "package.json", `{
  "scripts": {"dev": "next dev", "build": "next build", "test": "vitest", "lint": "next lint"},
  "dependencies": {"next": "latest", "react": "latest", "@prisma/client": "latest"},
  "devDependencies": {"typescript": "latest"}
}`)
	writeProjectFile(t, projectDir, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")

	context, err := detectProjectContext(projectDir)
	if err != nil {
		t.Fatalf("detectProjectContext() error = %v", err)
	}
	if want := "TypeScript · Next.js · Prisma"; context.stack != want {
		t.Errorf("stack = %q, want %q", context.stack, want)
	}
	for _, want := range []string{"`pnpm run dev`", "`pnpm run build`", "`pnpm run test`", "`pnpm run lint`"} {
		if !strings.Contains(context.commands, want) {
			t.Errorf("commands = %q, want %q", context.commands, want)
		}
	}
}

func TestDetectProjectContextUnknown(t *testing.T) {
	context, err := detectProjectContext(t.TempDir())
	if err != nil {
		t.Fatalf("detectProjectContext() error = %v", err)
	}
	if !strings.HasPrefix(context.stack, "Unknown") {
		t.Fatalf("stack = %q, want unknown marker", context.stack)
	}
	if !strings.Contains(context.commands, "No common commands") {
		t.Fatalf("commands = %q, want no-command guidance", context.commands)
	}
}

func TestRunInitWithDetectionPopulatesTemplate(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)
	writeProjectFile(t, templateDir, "AGENTS.md", "stack={{STACK}}\n{{COMMANDS}}\n")
	writeProjectFile(t, projectDir, "Cargo.toml", "[package]\nname = \"demo\"\nversion = \"0.1.0\"\n")

	var output bytes.Buffer
	if err := runInit(projectDir, templateDir, &output, initOptions{detect: true}); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "stack=Rust") || !strings.Contains(content, "`cargo test`") {
		t.Fatalf("generated AGENTS.md = %q", content)
	}
	if !strings.Contains(output.String(), "detected stack: Rust") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestEmbeddedTemplateWithoutDetectionContainsEditingGuidance(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, false); err != nil {
		t.Fatalf("runSetup() error = %v", err)
	}
	if err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{}); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "{{") {
		t.Fatalf("generated AGENTS.md contains unresolved placeholder: %q", content)
	}
	if !strings.Contains(content, "[Language") || !strings.Contains(content, "Add one row") {
		t.Fatalf("generated AGENTS.md is missing editing guidance: %q", content)
	}
}

func TestDetectProjectContextRejectsInvalidPackageJSON(t *testing.T) {
	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "package.json", "not-json\n")
	if _, err := detectProjectContext(projectDir); err == nil || !strings.Contains(err.Error(), "cannot parse package.json") {
		t.Fatalf("detectProjectContext() error = %v, want parse error", err)
	}
}

func writeProjectFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
