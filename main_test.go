package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitCreatesOnlyAgentsMD(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)
	if err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "AGENTS.md" {
		t.Fatalf("project entries = %v, want only AGENTS.md", entries)
	}
	got, err := os.ReadFile(filepath.Join(projectDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := "project: " + filepath.Base(projectDir) + "\n" +
		"<!-- Add the working agreement for this project. -->\n" +
		"<!-- Add language-specific guidance when it prevents real mistakes. -->\n"
	if string(got) != want {
		t.Fatalf("AGENTS.md = %q, want %q", got, want)
	}
}

func TestRunInitMissingTemplateCreatesNothing(t *testing.T) {
	projectDir := t.TempDir()
	err := runInit(projectDir, t.TempDir(), &bytes.Buffer{}, initOptions{})
	if err == nil || !strings.Contains(err.Error(), "AGENTS.md") {
		t.Fatalf("runInit() error = %v, want missing template error", err)
	}
	entries, err := os.ReadDir(projectDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("project entries = %v, error = %v", entries, err)
	}
}

func TestRunInitRefusesExistingAgentsMD(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)
	existing := filepath.Join(projectDir, "AGENTS.md")
	if err := os.WriteFile(existing, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("runInit() error = %v, want existing file error", err)
	}
	data, err := os.ReadFile(existing)
	if err != nil || string(data) != "keep me\n" {
		t.Fatalf("existing AGENTS.md = %q, error = %v", data, err)
	}
}

func TestRunInitLeavesOtherInstructionFilesAlone(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)
	writeProjectFile(t, projectDir, "CLAUDE.md", "custom instructions\n")
	writeProjectFile(t, projectDir, ".aicontext.json", "legacy state\n")
	if err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"CLAUDE.md":       "custom instructions\n",
		".aicontext.json": "legacy state\n",
	} {
		got, err := os.ReadFile(filepath.Join(projectDir, path))
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, error = %v", path, got, err)
		}
	}
}

func TestRunSetupCopiesEmbeddedTemplate(t *testing.T) {
	templateDir := t.TempDir()
	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(templateDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := defaultTemplates.ReadFile("templates/AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("installed template differs from embedded template")
	}
	entries, err := os.ReadDir(templateDir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "AGENTS.md" {
		t.Fatalf("template entries = %v, error = %v", entries, err)
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
		{name: "command help", args: []string{"help", "init"}, want: "refuses to replace an existing AGENTS.md"},
		{name: "command help flag", args: []string{"init", "--help"}, want: "Create one project-owned AGENTS.md"},
		{name: "language help topic", args: []string{"help", "languages"}, want: "For an existing project"},
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

func TestHelpRejectsUnknownTopicWithGuidance(t *testing.T) {
	err := run([]string{"help", "unknown"}, strings.NewReader(""), &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "aiContext help") {
		t.Fatalf("run() error = %v, want unknown-topic guidance", err)
	}
}

func TestRunInitRejectsStaleTemplateMissingSelectedGuidancePlaceholders(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, false); err != nil {
		t.Fatalf("runSetup() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "AGENTS.md"), []byte("# old template\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{
		profile:   "standard",
		languages: []string{"java"},
	})
	if err == nil || !strings.Contains(err.Error(), "{{LANGUAGE_GUIDELINES}}") || !strings.Contains(err.Error(), "aiContext setup") {
		t.Fatalf("runInit() error = %v, want actionable missing-placeholder error", err)
	}
	if _, statErr := os.Lstat(filepath.Join(projectDir, "AGENTS.md")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("AGENTS.md created from stale template; stat error = %v", statErr)
	}
}

func writeTestTemplates(t *testing.T, dir string) {
	t.Helper()
	contents := map[string]string{
		"AGENTS.md": "project: {{PROJECT_NAME}}\n" +
			"{{PROFILE_GUIDELINES}}\n" +
			"{{LANGUAGE_GUIDELINES}}\n",
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
