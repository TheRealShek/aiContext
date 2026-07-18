package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const manifestSchemaVersion = 1

type projectManifest struct {
	SchemaVersion    int           `json:"schemaVersion"`
	GeneratorVersion string        `json:"generatorVersion"`
	Tools            []string      `json:"tools"`
	Profile          string        `json:"profile,omitempty"`
	Languages        []string      `json:"languages,omitempty"`
	Detect           bool          `json:"detect"`
	Files            []managedFile `json:"files"`
}

type managedFile struct {
	Path     string `json:"path"`
	Template string `json:"template"`
	Role     string `json:"role"`
	SHA256   string `json:"sha256"`
}

func newProjectManifest(tools []string, detect bool, profile string, languages []string, files []pendingFile) projectManifest {
	manifest := projectManifest{
		SchemaVersion:    manifestSchemaVersion,
		GeneratorVersion: version,
		Tools:            append([]string(nil), tools...),
		Profile:          profile,
		Languages:        append([]string(nil), languages...),
		Detect:           detect,
		Files:            make([]managedFile, 0, len(files)),
	}
	for _, file := range files {
		role := "adapter"
		if file.display == canonicalTemplate.destination {
			role = "canonical"
		}
		manifest.Files = append(manifest.Files, managedFile{
			Path:     file.display,
			Template: file.source,
			Role:     role,
			SHA256:   contentSHA256(file.content),
		})
	}
	return manifest
}

func marshalManifest(manifest projectManifest) ([]byte, error) {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("cannot encode %s: %w", manifestFilename, err)
	}
	return append(data, '\n'), nil
}

func readManifest(root string) (projectManifest, error) {
	path := filepath.Join(root, manifestFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		return projectManifest{}, fmt.Errorf("cannot read %s: %w", manifestFilename, err)
	}
	var manifest projectManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return projectManifest{}, fmt.Errorf("cannot parse %s: %w", manifestFilename, err)
	}
	if err := validateManifest(manifest); err != nil {
		return projectManifest{}, err
	}
	return manifest, nil
}

func validateManifest(manifest projectManifest) error {
	if manifest.SchemaVersion != manifestSchemaVersion {
		return fmt.Errorf("unsupported %s schema version %d", manifestFilename, manifest.SchemaVersion)
	}
	if len(manifest.Tools) == 0 {
		return fmt.Errorf("%s has no selected tools", manifestFilename)
	}
	for _, tool := range manifest.Tools {
		if !isSupportedTool(tool) {
			return fmt.Errorf("%s contains unknown tool %q", manifestFilename, tool)
		}
	}
	if manifest.Profile != "" {
		if err := validateProfileName(manifest.Profile); err != nil {
			return fmt.Errorf("%s contains an invalid profile: %w", manifestFilename, err)
		}
	}
	seenLanguages := make(map[string]bool, len(manifest.Languages))
	for _, language := range manifest.Languages {
		if !isSupportedLanguage(language) {
			return fmt.Errorf("%s contains unknown language guideline %q", manifestFilename, language)
		}
		if seenLanguages[language] {
			return fmt.Errorf("%s contains duplicate language guideline %q", manifestFilename, language)
		}
		seenLanguages[language] = true
	}
	seen := make(map[string]bool)
	for _, file := range manifest.Files {
		if err := validateManagedPath(file.Path); err != nil {
			return err
		}
		if seen[file.Path] {
			return fmt.Errorf("%s contains duplicate path %s", manifestFilename, file.Path)
		}
		seen[file.Path] = true
		if file.Role != "canonical" && file.Role != "adapter" {
			return fmt.Errorf("%s contains invalid role %q for %s", manifestFilename, file.Role, file.Path)
		}
		if len(file.SHA256) != sha256.Size*2 {
			return fmt.Errorf("%s contains invalid SHA-256 for %s", manifestFilename, file.Path)
		}
		if _, err := hex.DecodeString(file.SHA256); err != nil {
			return fmt.Errorf("%s contains invalid SHA-256 for %s", manifestFilename, file.Path)
		}
	}
	return nil
}

func validateManagedPath(path string) error {
	if path == "" || filepath.IsAbs(path) {
		return fmt.Errorf("%s contains unsafe managed path %q", manifestFilename, path)
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || clean != path {
		return fmt.Errorf("%s contains unsafe managed path %q", manifestFilename, path)
	}
	return nil
}

func manifestFileMap(manifest projectManifest) map[string]managedFile {
	files := make(map[string]managedFile, len(manifest.Files))
	for _, file := range manifest.Files {
		files[file.Path] = file
	}
	return files
}

func contentSHA256(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func readOptionalFile(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return nil, false, err
}
