package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSetupInstallsProfilesAndLanguageGuidelines(t *testing.T) {
	templateDir := filepath.Join(t.TempDir(), "templates")
	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, false); err != nil {
		t.Fatalf("runSetup() error = %v", err)
	}

	for embedded, installed := range map[string]string{
		"profiles/standard.md": "profiles/standard.md",
		"guidelines/go.md":     "guidelines/go.md",
		"guidelines/rust.md":   "guidelines/rust.md",
	} {
		want, err := defaultTemplates.ReadFile(embedded)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(filepath.Dir(templateDir), filepath.FromSlash(installed)))
		if err != nil {
			t.Fatalf("cannot read installed %s: %v", installed, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("installed %s differs from embedded default", installed)
		}
	}
}

func TestInitDefaultsToStandardProfileAndDetectedGoGuidelines(t *testing.T) {
	templateDir := filepath.Join(t.TempDir(), "templates")
	projectDir := t.TempDir()
	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	writeProjectFile(t, projectDir, "go.mod", "module example.com/project\n\ngo 1.24\n")

	var output bytes.Buffer
	err := run([]string{
		"init", "--target", projectDir, "--template-dir", templateDir, "--tools", "codex",
	}, strings.NewReader(""), &output)
	if err != nil {
		t.Fatalf("init error = %v\noutput: %s", err, output.String())
	}

	agents := readProjectText(t, projectDir, "AGENTS.md")
	for _, want := range []string{
		"Read nearby code and tests before changing behavior",
		"Write idiomatic, straightforward Go",
		"Avoid reflection and `unsafe` by default",
	} {
		if !strings.Contains(agents, want) {
			t.Errorf("AGENTS.md does not contain %q", want)
		}
	}
	manifest, err := readManifest(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Profile != "standard" || !reflect.DeepEqual(manifest.Languages, []string{"go"}) {
		t.Fatalf("manifest profile/languages = %q/%v", manifest.Profile, manifest.Languages)
	}
}

func TestInitSupportsSecurityProfileAndExplicitRustGuidelines(t *testing.T) {
	templateDir := filepath.Join(t.TempDir(), "templates")
	projectDir := t.TempDir()
	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}

	err := run([]string{
		"init", "--target", projectDir, "--template-dir", templateDir, "--tools", "codex",
		"--profile", "security", "--languages", "rust",
	}, strings.NewReader(""), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}

	agents := readProjectText(t, projectDir, "AGENTS.md")
	for _, want := range []string{
		"Treat every external input",
		"Treat `unsafe` as an exceptional boundary",
		"include a `SAFETY:` explanation for every block",
	} {
		if !strings.Contains(agents, want) {
			t.Errorf("AGENTS.md does not contain %q", want)
		}
	}
}

func TestInitSupportsCustomLocalProfileAndNoLanguagePack(t *testing.T) {
	configRoot := t.TempDir()
	templateDir := filepath.Join(configRoot, "templates")
	projectDir := t.TempDir()
	if err := runSetup(templateDir, strings.NewReader(""), &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	writeProjectFile(t, configRoot, "profiles/team.md", "- Preserve our public event schema.\n")

	err := run([]string{
		"init", "--target", projectDir, "--template-dir", templateDir, "--tools", "codex",
		"--profile", "team", "--languages", "none",
	}, strings.NewReader(""), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	agents := readProjectText(t, projectDir, "AGENTS.md")
	if !strings.Contains(agents, "Preserve our public event schema") {
		t.Fatalf("custom profile missing from AGENTS.md: %s", agents)
	}
	if !strings.Contains(agents, "No language-specific guidelines selected") {
		t.Fatalf("no-language marker missing from AGENTS.md: %s", agents)
	}

	var output bytes.Buffer
	if err := run([]string{"profiles", "--template-dir", templateDir}, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "team") || !strings.Contains(output.String(), "rust") {
		t.Fatalf("profiles output = %q", output.String())
	}
}

func TestResolveLanguageSelectionDetectsMixedProject(t *testing.T) {
	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "package.json", `{"devDependencies":{"typescript":"^5.0.0"}}`)
	writeProjectFile(t, projectDir, "main.tf", "terraform {}\n")
	writeProjectFile(t, projectDir, "Gemfile", "source 'https://rubygems.org'\n")

	got, err := resolveLanguageSelection(projectDir, "auto")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"typescript", "ruby", "terraform"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("detected languages = %v, want %v", got, want)
	}
	if _, err := resolveLanguageSelection(projectDir, "brainfuck"); err == nil {
		t.Fatal("unknown language guideline was accepted")
	}
}

func readProjectText(t *testing.T, root, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
