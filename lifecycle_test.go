package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseToolSelection(t *testing.T) {
	tests := []struct {
		value string
		want  []string
	}{
		{value: "gemini,claude,gemini", want: []string{"claude", "gemini"}},
		{value: "all", want: []string{"codex", "claude", "cursor", "copilot", "gemini"}},
	}
	for _, tt := range tests {
		got, err := parseToolSelection(tt.value)
		if err != nil {
			t.Fatalf("parseToolSelection(%q) error = %v", tt.value, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseToolSelection(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
	if _, err := parseToolSelection("unknown"); err == nil {
		t.Fatal("parseToolSelection(unknown) succeeded")
	}
}

func TestRunInitSelectsToolsAndWritesManifest(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)

	if err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{tools: []string{"gemini"}}); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}
	for _, path := range []string{"AGENTS.md", "GEMINI.md", manifestFilename} {
		if _, err := os.Stat(filepath.Join(projectDir, path)); err != nil {
			t.Errorf("expected %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("CLAUDE.md exists despite tool selection; stat error = %v", err)
	}
	manifest, err := readManifest(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Tools, []string{"gemini"}) {
		t.Fatalf("manifest tools = %v", manifest.Tools)
	}
}

func TestRunAdoptCreatesStateWithoutChangingExistingFiles(t *testing.T) {
	projectDir := t.TempDir()
	writeProjectFile(t, projectDir, "AGENTS.md", "existing canonical\n")
	writeProjectFile(t, projectDir, "CLAUDE.md", "existing Claude adapter\n")

	if err := runAdopt(projectDir, []string{"claude"}, false, &bytes.Buffer{}, false); err != nil {
		t.Fatalf("runAdopt() error = %v", err)
	}
	for path, want := range map[string]string{
		"AGENTS.md": "existing canonical\n",
		"CLAUDE.md": "existing Claude adapter\n",
	} {
		data, err := os.ReadFile(filepath.Join(projectDir, path))
		if err != nil || string(data) != want {
			t.Fatalf("%s = %q, %v; want unchanged", path, data, err)
		}
	}
	manifest, err := readManifest(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Tools, []string{"claude"}) {
		t.Fatalf("manifest tools = %v", manifest.Tools)
	}
}

func TestHealthReportDistinguishesCanonicalAndAdapterEdits(t *testing.T) {
	templateDir, projectDir := initializeManagedProject(t)
	_ = templateDir
	writeProjectFile(t, projectDir, "AGENTS.md", "custom project instructions\n")

	report, err := inspectProjectHealth(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := healthCheckError(report, true); err != nil {
		t.Fatalf("customized AGENTS.md failed strict check: %v", err)
	}
	if !reportContains(report, "info", "AGENTS.md is customized") {
		t.Fatalf("report = %+v, want canonical customization info", report.findings)
	}

	writeProjectFile(t, projectDir, "CLAUDE.md", "custom adapter\n")
	report, err = inspectProjectHealth(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if !reportContains(report, "warning", "CLAUDE.md was modified") {
		t.Fatalf("report = %+v, want adapter warning", report.findings)
	}
	if err := healthCheckError(report, false); err != nil {
		t.Fatalf("non-strict check failed on warning: %v", err)
	}
	if err := healthCheckError(report, true); err == nil {
		t.Fatal("strict check passed despite adapter warning")
	}
}

func TestHealthReportFailsForMissingManagedFile(t *testing.T) {
	_, projectDir := initializeManagedProject(t)
	if err := os.Remove(filepath.Join(projectDir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	report, err := inspectProjectHealth(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if !reportContains(report, "error", "CLAUDE.md is missing") {
		t.Fatalf("report = %+v, want missing-file error", report.findings)
	}
	if err := healthCheckError(report, false); err == nil {
		t.Fatal("check passed despite missing adapter")
	}
}

func TestSyncRepairsMissingAndUpdatesUnmodifiedAdapters(t *testing.T) {
	templateDir, projectDir := initializeManagedProject(t)
	if err := os.Remove(filepath.Join(projectDir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	writeProjectFile(t, templateDir, "cursor.mdc", "---\nalwaysApply: true\n---\n@AGENTS.md\n")

	plan, err := buildLifecyclePlan(projectDir, templateDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.conflicts) != 0 || len(plan.changes) != 2 {
		t.Fatalf("plan = %+v, want two safe changes", plan)
	}
	if err := applyLifecyclePlan(projectDir, plan, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"CLAUDE.md":                   "@AGENTS.md\n",
		".cursor/rules/aicontext.mdc": "---\nalwaysApply: true\n---\n@AGENTS.md\n",
	} {
		data, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(path)))
		if err != nil || string(data) != want {
			t.Errorf("%s = %q, %v; want %q", path, data, err, want)
		}
	}
}

func TestSyncRefusesEditedAdapter(t *testing.T) {
	templateDir, projectDir := initializeManagedProject(t)
	writeProjectFile(t, projectDir, "CLAUDE.md", "my custom Claude rules\n")
	writeProjectFile(t, templateDir, "CLAUDE.md", "@AGENTS.md\nnew default\n")

	plan, err := buildLifecyclePlan(projectDir, templateDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.conflicts) != 1 || !strings.Contains(plan.conflicts[0], "edited") {
		t.Fatalf("conflicts = %v", plan.conflicts)
	}
	if err := applyLifecyclePlan(projectDir, plan, &bytes.Buffer{}); err == nil {
		t.Fatal("applyLifecyclePlan() overwrote an edited adapter")
	}
	data, _ := os.ReadFile(filepath.Join(projectDir, "CLAUDE.md"))
	if string(data) != "my custom Claude rules\n" {
		t.Fatalf("edited adapter changed to %q", data)
	}
}

func TestUpdateChangesToolSelectionSafely(t *testing.T) {
	templateDir, projectDir := initializeManagedProject(t)
	writeProjectFile(t, projectDir, "AGENTS.md", "user-owned instructions\n")

	plan, err := buildLifecyclePlan(projectDir, templateDir, []string{"codex", "gemini"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.conflicts) != 0 {
		t.Fatalf("conflicts = %v", plan.conflicts)
	}
	if err := applyLifecyclePlan(projectDir, plan, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "GEMINI.md")); err != nil {
		t.Fatalf("GEMINI.md not created: %v", err)
	}
	for _, path := range []string{"CLAUDE.md", ".cursor/rules/aicontext.mdc", ".github/copilot-instructions.md"} {
		if _, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(path))); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("deselected adapter %s remains; stat error = %v", path, err)
		}
	}
	agents, _ := os.ReadFile(filepath.Join(projectDir, "AGENTS.md"))
	if string(agents) != "user-owned instructions\n" {
		t.Fatalf("AGENTS.md was not preserved: %q", agents)
	}
	manifest, err := readManifest(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Tools, []string{"codex", "gemini"}) {
		t.Fatalf("manifest tools = %v", manifest.Tools)
	}
}

func TestDiffPrintsPlanWithoutWriting(t *testing.T) {
	templateDir, projectDir := initializeManagedProject(t)
	writeProjectFile(t, templateDir, "CLAUDE.md", "@AGENTS.md\nnew\n")
	plan, err := buildLifecyclePlan(projectDir, templateDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	printLifecyclePlan(&output, plan)
	if !strings.Contains(output.String(), "~ CLAUDE.md") || !strings.Contains(output.String(), "+++ desired") {
		t.Fatalf("diff output = %q", output.String())
	}
	data, _ := os.ReadFile(filepath.Join(projectDir, "CLAUDE.md"))
	if string(data) != "@AGENTS.md\n" {
		t.Fatalf("diff modified CLAUDE.md: %q", data)
	}
}

func TestLifecycleCommandsEndToEnd(t *testing.T) {
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)

	commands := []struct {
		args []string
		want string
	}{
		{
			args: []string{"init", "--target", projectDir, "--template-dir", templateDir, "--tools", "codex,gemini"},
			want: manifestFilename,
		},
		{args: []string{"doctor", "--target", projectDir}, want: "summary:"},
		{args: []string{"check", "--strict", "--target", projectDir}, want: "0 error(s)"},
		{
			args: []string{"update", "--dry-run", "--target", projectDir, "--template-dir", templateDir, "--tools", "codex,claude"},
			want: "+ CLAUDE.md",
		},
	}
	for _, command := range commands {
		var output bytes.Buffer
		if err := run(command.args, strings.NewReader(""), &output); err != nil {
			t.Fatalf("run(%v) error = %v\noutput: %s", command.args, err, output.String())
		}
		if !strings.Contains(output.String(), command.want) {
			t.Fatalf("run(%v) output = %q, want %q", command.args, output.String(), command.want)
		}
	}
}

func initializeManagedProject(t *testing.T) (string, string) {
	t.Helper()
	templateDir := t.TempDir()
	projectDir := t.TempDir()
	writeTestTemplates(t, templateDir)
	if err := runInit(projectDir, templateDir, &bytes.Buffer{}, initOptions{}); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}
	return templateDir, projectDir
}

func reportContains(report healthReport, severity, text string) bool {
	for _, finding := range report.findings {
		if finding.severity == severity && strings.Contains(finding.message, text) {
			return true
		}
	}
	return false
}
