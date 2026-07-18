package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanRemovesUnchangedManagedFilesAndEmptyDirectories(t *testing.T) {
	_, projectDir := initializeManagedProject(t)
	plan, err := buildCleanPlan(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.remove) != 4 || len(plan.preserve) != 0 {
		t.Fatalf("clean plan remove/preserve = %d/%d", len(plan.remove), len(plan.preserve))
	}
	if err := applyCleanPlan(projectDir, plan, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".cursor/rules/aicontext.mdc", ".github/copilot-instructions.md", manifestFilename, ".cursor", ".github"} {
		if _, err := os.Lstat(filepath.Join(projectDir, filepath.FromSlash(path))); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s remains after clean; stat error = %v", path, err)
		}
	}
}

func TestCleanPreservesModifiedFilesAndStopsManaging(t *testing.T) {
	_, projectDir := initializeManagedProject(t)
	writeProjectFile(t, projectDir, "AGENTS.md", "custom canonical instructions\n")
	writeProjectFile(t, projectDir, "CLAUDE.md", "custom Claude instructions\n")

	plan, err := buildCleanPlan(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.preserve) != 2 || len(plan.remove) != 2 {
		t.Fatalf("clean plan preserve/remove = %d/%d", len(plan.preserve), len(plan.remove))
	}
	var output bytes.Buffer
	if err := applyCleanPlan(projectDir, plan, &output); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"AGENTS.md": "custom canonical instructions\n",
		"CLAUDE.md": "custom Claude instructions\n",
	} {
		if got := readProjectText(t, projectDir, path); got != want {
			t.Errorf("%s = %q, want preserved %q", path, got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(projectDir, manifestFilename)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("manifest remains after clean; stat error = %v", err)
	}
	if !strings.Contains(output.String(), "preserved AGENTS.md") {
		t.Fatalf("clean output = %q", output.String())
	}
}

func TestCleanDryRunChangesNothing(t *testing.T) {
	_, projectDir := initializeManagedProject(t)
	var output bytes.Buffer
	if err := run([]string{"clean", "--dry-run", "--target", projectDir}, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "managed and unchanged") || !strings.Contains(output.String(), "stop managing") {
		t.Fatalf("clean dry-run output = %q", output.String())
	}
	for _, path := range []string{"AGENTS.md", "CLAUDE.md", manifestFilename} {
		if _, err := os.Stat(filepath.Join(projectDir, path)); err != nil {
			t.Errorf("dry-run removed %s: %v", path, err)
		}
	}
}

func TestCleanTreatsMissingFileAsAlreadyAbsent(t *testing.T) {
	_, projectDir := initializeManagedProject(t)
	if err := os.Remove(filepath.Join(projectDir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	plan, err := buildCleanPlan(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.missing) != 1 || plan.missing[0].Path != "CLAUDE.md" {
		t.Fatalf("missing files = %+v", plan.missing)
	}
	if err := applyCleanPlan(projectDir, plan, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
}

func TestCleanKeepsNonEmptyProjectDirectories(t *testing.T) {
	_, projectDir := initializeManagedProject(t)
	writeProjectFile(t, projectDir, ".github/workflows/test.yml", "name: test\n")
	writeProjectFile(t, projectDir, ".cursor/settings.json", "{}\n")

	plan, err := buildCleanPlan(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyCleanPlan(projectDir, plan, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		".github/workflows/test.yml": "name: test\n",
		".cursor/settings.json":      "{}\n",
	} {
		if got := readProjectText(t, projectDir, path); got != want {
			t.Errorf("unrelated file %s = %q, want %q", path, got, want)
		}
	}
}

func TestCleanNeverTraversesManagedPathSymlinks(t *testing.T) {
	_, projectDir := initializeManagedProject(t)
	externalDir := t.TempDir()
	externalRules := filepath.Join(externalDir, "rules")
	if err := os.MkdirAll(externalRules, 0o755); err != nil {
		t.Fatal(err)
	}
	externalFile := filepath.Join(externalRules, "aicontext.mdc")
	if err := os.WriteFile(externalFile, []byte("@AGENTS.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(projectDir, ".cursor/rules/aicontext.mdc")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(projectDir, ".cursor/rules")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(projectDir, ".cursor")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(externalDir, filepath.Join(projectDir, ".cursor")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	plan, err := buildCleanPlan(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, file := range plan.preserve {
		found = found || file.Path == ".cursor/rules/aicontext.mdc"
	}
	if !found {
		t.Fatalf("symlinked managed path was not preserved: %+v", plan)
	}
	if err := applyCleanPlan(projectDir, plan, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(externalFile); err != nil || string(got) != "@AGENTS.md\n" {
		t.Fatalf("external file changed: %q, %v", got, err)
	}
}
