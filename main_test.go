package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitCreatesAllFiles(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)

	var output bytes.Buffer
	if err := runInit(projectDir, templateDir, &output, initOptions{}); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	wantFiles := map[string]string{
		"AGENTS.md":                       "project: " + filepath.Base(projectDir) + "\n",
		"CLAUDE.md":                       "@AGENTS.md\n",
		".cursor/rules/aicontext.mdc":     "@AGENTS.md\n",
		".github/copilot-instructions.md": "@AGENTS.md\n",
	}
	for name, want := range wantFiles {
		data, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", name, err)
		}
		if got := string(data); got != want {
			t.Errorf("%s content = %q, want %q", name, got, want)
		}
	}
	if _, err := readManifest(projectDir); err != nil {
		t.Fatalf("readManifest() error = %v", err)
	}
}

func TestRunInitMissingTemplateCreatesNothing(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)
	if err := os.Remove(filepath.Join(templateDir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}

	err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{})
	if err == nil || !strings.Contains(err.Error(), "CLAUDE.md") {
		t.Fatalf("runInit() error = %v, want missing CLAUDE.md error", err)
	}

	for _, spec := range templateSpecsForTools(defaultTools) {
		_, statErr := os.Lstat(filepath.Join(projectDir, filepath.FromSlash(spec.destination)))
		if !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("destination %s exists after failed init; stat error = %v", spec.destination, statErr)
		}
	}
	if _, statErr := os.Lstat(filepath.Join(projectDir, manifestFilename)); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("%s exists after failed init; stat error = %v", manifestFilename, statErr)
	}
}

func TestRunInitRefusesExistingDestination(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)
	existing := filepath.Join(projectDir, "CLAUDE.md")
	if err := os.WriteFile(existing, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("runInit() error = %v, want existing destination error", err)
	}
	data, readErr := os.ReadFile(existing)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if got := string(data); got != "keep me\n" {
		t.Fatalf("existing file content = %q, want unchanged", got)
	}
	if _, statErr := os.Lstat(filepath.Join(projectDir, "AGENTS.md")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("AGENTS.md created despite preflight failure; stat error = %v", statErr)
	}
}

func TestRunSetupCopiesEmbeddedTemplates(t *testing.T) {
	templateDir := t.TempDir()
	var output bytes.Buffer
	if err := runSetup(templateDir, strings.NewReader(""), &output, false); err != nil {
		t.Fatalf("runSetup() error = %v", err)
	}

	for _, spec := range setupTemplateSpecs() {
		got, err := os.ReadFile(filepath.Join(templateDir, spec.source))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", spec.source, err)
		}
		want, err := defaultTemplates.ReadFile("templates/" + spec.source)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("setup template %s differs from embedded default", spec.source)
		}
	}
}

func TestRunSetupPreservesExistingTemplateByDefault(t *testing.T) {
	templateDir := t.TempDir()
	existing := filepath.Join(templateDir, "AGENTS.md")
	if err := os.WriteFile(existing, []byte("custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runSetup(templateDir, strings.NewReader("\n"), &bytes.Buffer{}, false); err != nil {
		t.Fatalf("runSetup() error = %v", err)
	}
	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "custom\n" {
		t.Fatalf("existing template content = %q, want unchanged", got)
	}
}

func TestRunInitDryRunDoesNotWrite(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)

	var output bytes.Buffer
	if err := runInit(projectDir, templateDir, &output, initOptions{dryRun: true}); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}
	if !strings.Contains(output.String(), "would create AGENTS.md") {
		t.Fatalf("dry-run output = %q", output.String())
	}
	if _, err := os.Lstat(filepath.Join(projectDir, "AGENTS.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run created AGENTS.md; stat error = %v", err)
	}
}

func TestRunSetupForceOverwritesExistingTemplate(t *testing.T) {
	templateDir := t.TempDir()
	existing := filepath.Join(templateDir, "AGENTS.md")
	if err := os.WriteFile(existing, []byte("custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, true); err != nil {
		t.Fatalf("runSetup() error = %v", err)
	}
	want, err := defaultTemplates.ReadFile("templates/AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("--force did not replace existing template")
	}
}

func TestRunSupportsHelpVersionAndCustomDirectories(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "help", args: []string{"help"}, want: "Usage:"},
		{name: "version", args: []string{"version"}, want: "aiContext " + version},
		{
			name: "custom init dry run",
			args: []string{"init", "--dry-run", "--target", projectDir, "--template-dir", templateDir},
			want: "would create AGENTS.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := run(tt.args, strings.NewReader(""), &output); err != nil {
				t.Fatalf("run() error = %v", err)
			}
			if !strings.Contains(output.String(), tt.want) {
				t.Fatalf("run() output = %q, want substring %q", output.String(), tt.want)
			}
		})
	}
}

func writeTestTemplates(t *testing.T, dir string) {
	t.Helper()
	contents := map[string]string{
		"AGENTS.md":               "project: {{PROJECT_NAME}}\n",
		"CLAUDE.md":               "@AGENTS.md\n",
		"cursor.mdc":              "@AGENTS.md\n",
		"copilot-instructions.md": "@AGENTS.md\n",
		"GEMINI.md":               "@AGENTS.md\n",
	}
	for name, content := range contents {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	profileDir := filepath.Join(filepath.Dir(dir), "profiles")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profileDir, "standard.md"), []byte("Use the test working agreement.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
