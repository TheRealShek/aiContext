package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type lifecycleChange struct {
	kind    string
	path    string
	source  string
	current []byte
	desired []byte
	reason  string
}

type lifecyclePlan struct {
	currentManifest projectManifest
	nextManifest    projectManifest
	changes         []lifecycleChange
	conflicts       []string
	notes           []string
}

func buildLifecyclePlan(root, templateDir string, selectedTools []string) (lifecyclePlan, error) {
	manifest, err := readManifest(root)
	if err != nil {
		return lifecyclePlan{}, err
	}
	if len(selectedTools) == 0 {
		selectedTools = append([]string(nil), manifest.Tools...)
	}
	for _, tool := range selectedTools {
		if !isSupportedTool(tool) {
			return lifecyclePlan{}, fmt.Errorf("unknown tool %q", tool)
		}
	}

	desiredFiles, err := renderAdapterFiles(root, templateDir, selectedTools)
	if err != nil {
		return lifecyclePlan{}, err
	}
	recorded := manifestFileMap(manifest)
	plan := lifecyclePlan{currentManifest: manifest}

	canonical, ok := recorded[canonicalTemplate.destination]
	if !ok || canonical.Role != "canonical" {
		plan.conflicts = append(plan.conflicts, "canonical AGENTS.md is not recorded correctly in the manifest")
	} else {
		content, exists, readErr := readOptionalFile(filepath.Join(root, canonical.Path))
		if readErr != nil {
			return lifecyclePlan{}, fmt.Errorf("cannot read %s: %w", canonical.Path, readErr)
		}
		if !exists {
			plan.conflicts = append(plan.conflicts, "AGENTS.md is missing; lifecycle commands never recreate the user-owned canonical file")
		} else if contentSHA256(content) != canonical.SHA256 {
			plan.notes = append(plan.notes, "AGENTS.md is customized and will be preserved")
		}
	}

	desiredByPath := make(map[string]pendingFile, len(desiredFiles))
	for _, desired := range desiredFiles {
		desiredByPath[desired.display] = desired
		record, wasManaged := recorded[desired.display]
		current, exists, readErr := readOptionalFile(desired.path)
		if readErr != nil {
			return lifecyclePlan{}, fmt.Errorf("cannot read %s: %w", desired.display, readErr)
		}
		switch {
		case !exists:
			plan.changes = append(plan.changes, lifecycleChange{kind: "create", path: desired.display, source: desired.source, desired: desired.content, reason: "missing adapter"})
		case bytes.Equal(current, desired.content):
			if !wasManaged {
				plan.notes = append(plan.notes, desired.display+" already matches and will be adopted")
			}
		case wasManaged && contentSHA256(current) == record.SHA256:
			plan.changes = append(plan.changes, lifecycleChange{kind: "update", path: desired.display, source: desired.source, current: current, desired: desired.content, reason: "managed template changed"})
		case !wasManaged:
			plan.conflicts = append(plan.conflicts, desired.display+" exists but is not managed by aiContext")
		default:
			plan.conflicts = append(plan.conflicts, desired.display+" was edited after generation")
		}
	}

	for _, record := range manifest.Files {
		if record.Role != "adapter" {
			continue
		}
		if _, wanted := desiredByPath[record.Path]; wanted {
			continue
		}
		current, exists, readErr := readOptionalFile(filepath.Join(root, filepath.FromSlash(record.Path)))
		if readErr != nil {
			return lifecyclePlan{}, fmt.Errorf("cannot read %s: %w", record.Path, readErr)
		}
		if !exists {
			plan.notes = append(plan.notes, record.Path+" is already absent and will be forgotten")
		} else if contentSHA256(current) == record.SHA256 {
			plan.changes = append(plan.changes, lifecycleChange{kind: "remove", path: record.Path, current: current, reason: "tool deselected"})
		} else {
			plan.conflicts = append(plan.conflicts, record.Path+" was edited and cannot be removed safely")
		}
	}

	next := manifest
	next.GeneratorVersion = version
	next.Tools = append([]string(nil), selectedTools...)
	next.Files = nil
	if ok {
		next.Files = append(next.Files, canonical)
	}
	for _, desired := range desiredFiles {
		next.Files = append(next.Files, managedFile{
			Path:     desired.display,
			Template: desired.source,
			Role:     "adapter",
			SHA256:   contentSHA256(desired.content),
		})
	}
	plan.nextManifest = next
	return plan, nil
}

func renderAdapterFiles(root, templateDir string, tools []string) ([]pendingFile, error) {
	projectName := filepath.Base(filepath.Clean(root))
	replacer := strings.NewReplacer("{{PROJECT_NAME}}", projectName)
	specs := adapterSpecsForTools(tools)
	files := make([]pendingFile, 0, len(specs))
	for _, spec := range specs {
		raw, err := os.ReadFile(filepath.Join(templateDir, spec.source))
		if err != nil {
			return nil, fmt.Errorf("cannot read template %s (run: aiContext setup): %w", spec.source, err)
		}
		files = append(files, pendingFile{
			path:    filepath.Join(root, filepath.FromSlash(spec.destination)),
			display: spec.destination,
			source:  spec.source,
			content: []byte(replacer.Replace(string(raw))),
		})
	}
	return files, nil
}

func printLifecyclePlan(w io.Writer, plan lifecyclePlan) {
	for _, note := range plan.notes {
		fmt.Fprintln(w, "i", note)
	}
	for _, change := range plan.changes {
		marker := map[string]string{"create": "+", "update": "~", "remove": "-"}[change.kind]
		fmt.Fprintf(w, "%s %s — %s\n", marker, change.path, change.reason)
		if change.kind == "update" {
			printContentDiff(w, change.current, change.desired)
		}
	}
	for _, conflict := range plan.conflicts {
		fmt.Fprintln(w, "!", conflict)
	}
	if len(plan.changes) == 0 && len(plan.conflicts) == 0 {
		fmt.Fprintln(w, "✓ generated adapters are in sync")
	}
}

func printContentDiff(w io.Writer, current, desired []byte) {
	fmt.Fprintln(w, "  --- current")
	for _, line := range strings.Split(strings.TrimSuffix(string(current), "\n"), "\n") {
		fmt.Fprintln(w, "  -", line)
	}
	fmt.Fprintln(w, "  +++ desired")
	for _, line := range strings.Split(strings.TrimSuffix(string(desired), "\n"), "\n") {
		fmt.Fprintln(w, "  +", line)
	}
}

func applyLifecyclePlan(root string, plan lifecyclePlan, stdout io.Writer) (err error) {
	if len(plan.conflicts) > 0 {
		return fmt.Errorf("cannot apply lifecycle plan: resolve %d conflict(s) first", len(plan.conflicts))
	}
	type backup struct {
		path    string
		content []byte
		existed bool
	}
	backups := make([]backup, 0, len(plan.changes)+1)
	manifestPath := filepath.Join(root, manifestFilename)
	manifestContent, _, readErr := readOptionalFile(manifestPath)
	if readErr != nil {
		return fmt.Errorf("cannot back up %s: %w", manifestFilename, readErr)
	}
	backups = append(backups, backup{path: manifestPath, content: manifestContent, existed: true})

	defer func() {
		if err == nil {
			return
		}
		for i := len(backups) - 1; i >= 0; i-- {
			item := backups[i]
			if item.existed {
				restoreErr := os.MkdirAll(filepath.Dir(item.path), 0o755)
				if restoreErr == nil {
					restoreErr = os.WriteFile(item.path, item.content, 0o644)
				}
				if restoreErr != nil {
					err = errors.Join(err, fmt.Errorf("cannot restore %s: %w", item.path, restoreErr))
				}
			} else if removeErr := os.Remove(item.path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("cannot roll back %s: %w", item.path, removeErr))
			}
		}
	}()

	for _, change := range plan.changes {
		path := filepath.Join(root, filepath.FromSlash(change.path))
		current, exists, backupErr := readOptionalFile(path)
		if backupErr != nil {
			return fmt.Errorf("cannot back up %s: %w", change.path, backupErr)
		}
		backups = append(backups, backup{path: path, content: current, existed: exists})
		switch change.kind {
		case "create", "update":
			if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o755); mkdirErr != nil {
				return fmt.Errorf("cannot create directory for %s: %w", change.path, mkdirErr)
			}
			if writeErr := os.WriteFile(path, change.desired, 0o644); writeErr != nil {
				return fmt.Errorf("cannot write %s: %w", change.path, writeErr)
			}
		case "remove":
			if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				return fmt.Errorf("cannot remove %s: %w", change.path, removeErr)
			}
		default:
			return fmt.Errorf("unknown lifecycle action %q", change.kind)
		}
	}

	manifestData, marshalErr := marshalManifest(plan.nextManifest)
	if marshalErr != nil {
		return marshalErr
	}
	if writeErr := os.WriteFile(manifestPath, manifestData, 0o644); writeErr != nil {
		return fmt.Errorf("cannot update %s: %w", manifestFilename, writeErr)
	}
	for _, change := range plan.changes {
		fmt.Fprintln(stdout, "✓", change.kind, change.path)
	}
	if len(plan.changes) == 0 {
		fmt.Fprintln(stdout, "✓ manifest metadata is current")
	}
	return nil
}

type healthFinding struct {
	severity string
	message  string
}

type healthReport struct {
	findings []healthFinding
}

func inspectProjectHealth(root string) (healthReport, error) {
	report := healthReport{}
	manifest, err := readManifest(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			report.add("error", manifestFilename+" is missing; run aiContext init --adopt for existing instruction files")
			return report, nil
		}
		report.add("error", err.Error())
		return report, nil
	}
	if version != "dev" && manifest.GeneratorVersion != version {
		report.add("warning", fmt.Sprintf("manifest was generated by aiContext %s; current version is %s", manifest.GeneratorVersion, version))
	}

	recorded := manifestFileMap(manifest)
	for _, spec := range templateSpecsForTools(manifest.Tools) {
		if _, ok := recorded[spec.destination]; !ok {
			report.add("error", spec.destination+" is selected but not recorded in the manifest")
		}
	}
	for _, file := range manifest.Files {
		content, exists, readErr := readOptionalFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if readErr != nil {
			return healthReport{}, fmt.Errorf("cannot inspect %s: %w", file.Path, readErr)
		}
		if !exists {
			report.add("error", file.Path+" is missing")
			continue
		}
		modified := contentSHA256(content) != file.SHA256
		if file.Role == "canonical" && modified {
			report.add("info", file.Path+" is customized (expected for the canonical instructions)")
		} else if file.Role == "adapter" && modified {
			report.add("warning", file.Path+" was modified after generation")
		}
		if containsUnresolvedPlaceholder(content) {
			report.add("warning", file.Path+" contains an unresolved template placeholder")
		}
		if file.Role == "adapter" && !bytes.Contains(content, []byte("AGENTS.md")) {
			report.add("warning", file.Path+" does not reference AGENTS.md")
		}
	}

	if _, err := os.Lstat(filepath.Join(root, ".cursorrules")); err == nil {
		report.add("warning", ".cursorrules is legacy; use .cursor/rules/*.mdc")
	} else if !errors.Is(err, os.ErrNotExist) {
		return healthReport{}, fmt.Errorf("cannot inspect .cursorrules: %w", err)
	}
	selected := make(map[string]bool, len(manifest.Tools))
	for _, tool := range manifest.Tools {
		selected[tool] = true
	}
	for _, tool := range toolDefinitions {
		if tool.template == nil || selected[tool.name] {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(tool.template.destination))
		if _, err := os.Lstat(path); err == nil {
			report.add("warning", tool.template.destination+" exists although "+tool.name+" is not selected")
		}
	}
	if len(report.findings) == 0 {
		report.add("info", "project instruction files are healthy")
	}
	return report, nil
}

func containsUnresolvedPlaceholder(content []byte) bool {
	for _, placeholder := range [][]byte{[]byte("{{PROJECT_NAME}}"), []byte("{{STACK}}"), []byte("{{COMMANDS}}"), []byte("{{GUIDELINES}}")} {
		if bytes.Contains(content, placeholder) {
			return true
		}
	}
	return false
}

func (r *healthReport) add(severity, message string) {
	r.findings = append(r.findings, healthFinding{severity: severity, message: message})
}

func printHealthReport(w io.Writer, report healthReport) {
	counts := map[string]int{}
	for _, finding := range report.findings {
		marker := map[string]string{"info": "i", "warning": "!", "error": "✗"}[finding.severity]
		fmt.Fprintf(w, "%s %-7s %s\n", marker, finding.severity, finding.message)
		counts[finding.severity]++
	}
	fmt.Fprintf(w, "summary: %d error(s), %d warning(s), %d info\n", counts["error"], counts["warning"], counts["info"])
}

func healthCheckError(report healthReport, strict bool) error {
	for _, finding := range report.findings {
		if finding.severity == "error" || (strict && finding.severity == "warning") {
			return fmt.Errorf("project instruction check failed")
		}
	}
	return nil
}
