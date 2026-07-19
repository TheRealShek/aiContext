package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type cleanPlan struct {
	remove   []managedFile
	preserve []managedFile
	missing  []managedFile
}

func runCleanCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("clean", flag.ContinueOnError)
	flags.SetOutput(stdout)
	target := flags.String("target", "", "project directory (default: current directory)")
	dryRun := flags.Bool("dry-run", false, "show what would be removed without changing files")
	flags.Usage = func() { printCleanUsage(stdout) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("clean does not accept positional arguments")
	}
	projectDir, err := resolveTargetDir(*target)
	if err != nil {
		return err
	}
	plan, err := buildCleanPlan(projectDir)
	if err != nil {
		return err
	}
	printCleanPlan(stdout, plan)
	if *dryRun {
		return nil
	}
	return applyCleanPlan(projectDir, plan, stdout)
}

func buildCleanPlan(root string) (cleanPlan, error) {
	manifest, err := readManifest(root)
	if err != nil {
		return cleanPlan{}, err
	}
	plan := cleanPlan{}
	for _, file := range manifest.Files {
		path := filepath.Join(root, filepath.FromSlash(file.Path))
		hasSymlink, err := pathContainsSymlink(root, file.Path)
		if err != nil {
			return cleanPlan{}, fmt.Errorf("cannot safely inspect %s: %w", file.Path, err)
		}
		if hasSymlink {
			plan.preserve = append(plan.preserve, file)
			continue
		}
		content, exists, err := readOptionalFile(path)
		if err != nil {
			return cleanPlan{}, fmt.Errorf("cannot inspect %s: %w", file.Path, err)
		}
		switch {
		case !exists:
			plan.missing = append(plan.missing, file)
		case contentSHA256(content) == file.SHA256:
			plan.remove = append(plan.remove, file)
		default:
			plan.preserve = append(plan.preserve, file)
		}
	}
	return plan, nil
}

func printCleanPlan(w io.Writer, plan cleanPlan) {
	for _, file := range plan.remove {
		fmt.Fprintln(w, "-", file.Path, "— managed and unchanged")
	}
	for _, file := range plan.missing {
		fmt.Fprintln(w, "i", file.Path, "— already absent")
	}
	for _, file := range plan.preserve {
		fmt.Fprintln(w, "!", file.Path, "— modified or linked; preserving")
	}
	fmt.Fprintln(w, "-", manifestFilename, "— stop managing this project")
}

func applyCleanPlan(root string, plan cleanPlan, stdout io.Writer) (err error) {
	type backup struct {
		path    string
		content []byte
	}
	backups := make([]backup, 0, len(plan.remove)+1)
	manifestPath := filepath.Join(root, manifestFilename)
	manifestContent, exists, err := readOptionalFile(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot back up %s: %w", manifestFilename, err)
	}
	if !exists {
		return fmt.Errorf("cannot clean project: %s disappeared", manifestFilename)
	}
	backups = append(backups, backup{path: manifestPath, content: manifestContent})

	defer func() {
		if err == nil {
			return
		}
		for i := len(backups) - 1; i >= 0; i-- {
			item := backups[i]
			if restoreErr := os.MkdirAll(filepath.Dir(item.path), 0o755); restoreErr == nil {
				restoreErr = os.WriteFile(item.path, item.content, 0o644)
				if restoreErr != nil {
					err = errors.Join(err, fmt.Errorf("cannot restore %s: %w", item.path, restoreErr))
				}
			} else {
				err = errors.Join(err, fmt.Errorf("cannot restore directory for %s: %w", item.path, restoreErr))
			}
		}
	}()

	for _, file := range plan.remove {
		path := filepath.Join(root, filepath.FromSlash(file.Path))
		hasSymlink, inspectErr := pathContainsSymlink(root, file.Path)
		if inspectErr != nil {
			return fmt.Errorf("cannot safely recheck %s: %w", file.Path, inspectErr)
		}
		if hasSymlink {
			return fmt.Errorf("cannot remove %s: path became linked after the clean plan was built", file.Path)
		}
		content, exists, readErr := readOptionalFile(path)
		if readErr != nil {
			return fmt.Errorf("cannot recheck %s: %w", file.Path, readErr)
		}
		if !exists {
			continue
		}
		if contentSHA256(content) != file.SHA256 {
			return fmt.Errorf("cannot remove %s: file changed after the clean plan was built", file.Path)
		}
		backups = append(backups, backup{path: path, content: content})
		if removeErr := os.Remove(path); removeErr != nil {
			return fmt.Errorf("cannot remove %s: %w", file.Path, removeErr)
		}
	}
	if removeErr := os.Remove(manifestPath); removeErr != nil {
		return fmt.Errorf("cannot remove %s: %w", manifestFilename, removeErr)
	}
	if cleanupErr := removeEmptyAdapterDirectories(root); cleanupErr != nil {
		return cleanupErr
	}

	for _, file := range plan.remove {
		fmt.Fprintln(stdout, "✓ removed", file.Path)
	}
	for _, file := range plan.preserve {
		fmt.Fprintln(stdout, "✓ preserved", file.Path)
	}
	fmt.Fprintln(stdout, "✓ removed", manifestFilename)
	return nil
}

func pathContainsSymlink(root, relative string) (bool, error) {
	current := root
	for _, part := range splitManagedPath(relative) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true, nil
		}
	}
	return false, nil
}

func splitManagedPath(path string) []string {
	clean := filepath.Clean(filepath.FromSlash(path))
	parts := make([]string, 0, 4)
	for clean != "." {
		dir, file := filepath.Split(clean)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		clean = filepath.Clean(dir)
	}
	return parts
}

func removeEmptyAdapterDirectories(root string) error {
	for _, relative := range []string{".cursor/rules", ".cursor", ".github"} {
		path := filepath.Join(root, filepath.FromSlash(relative))
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("cannot inspect generated directory %s: %w", relative, err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("cannot inspect generated directory %s: %w", relative, err)
		}
		if len(entries) == 0 {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("cannot remove empty generated directory %s: %w", relative, err)
			}
		}
	}
	return nil
}
