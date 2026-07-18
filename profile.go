package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type languageDefinition struct {
	name        string
	description string
}

var builtInProfiles = []string{"minimal", "standard", "strict", "security"}

var languageDefinitions = []languageDefinition{
	{name: "go", description: "Go"},
	{name: "rust", description: "Rust"},
	{name: "typescript", description: "TypeScript"},
	{name: "javascript", description: "JavaScript"},
	{name: "python", description: "Python"},
	{name: "ruby", description: "Ruby"},
	{name: "java", description: "Java"},
	{name: "kotlin", description: "Kotlin"},
	{name: "csharp", description: "C# / .NET"},
	{name: "swift", description: "Swift"},
	{name: "php", description: "PHP"},
	{name: "terraform", description: "Terraform"},
}

type setupAsset struct {
	embedded    string
	destination string
	label       string
}

func setupAssets(templateDir string) []setupAsset {
	assets := make([]setupAsset, 0, len(setupTemplateSpecs())+len(builtInProfiles)+len(languageDefinitions))
	for _, spec := range setupTemplateSpecs() {
		assets = append(assets, setupAsset{
			embedded:    "templates/" + spec.source,
			destination: filepath.Join(templateDir, spec.source),
			label:       "templates/" + spec.source,
		})
	}
	root := filepath.Dir(filepath.Clean(templateDir))
	for _, profile := range builtInProfiles {
		name := profile + ".md"
		assets = append(assets, setupAsset{
			embedded:    "profiles/" + name,
			destination: filepath.Join(root, "profiles", name),
			label:       "profiles/" + name,
		})
	}
	for _, language := range languageDefinitions {
		name := language.name + ".md"
		assets = append(assets, setupAsset{
			embedded:    "guidelines/" + name,
			destination: filepath.Join(root, "guidelines", name),
			label:       "guidelines/" + name,
		})
	}
	return assets
}

func loadProfileGuidance(templateDir, profile string, languages []string) (string, string, error) {
	if err := validateProfileName(profile); err != nil {
		return "", "", err
	}
	root := filepath.Dir(filepath.Clean(templateDir))
	profilePath := filepath.Join(root, "profiles", profile+".md")
	profileContent, err := os.ReadFile(profilePath)
	if err != nil {
		return "", "", fmt.Errorf("cannot read profile %q (run: aiContext setup): %w", profile, err)
	}

	sections := make([]string, 0, len(languages))
	for _, language := range languages {
		if !isSupportedLanguage(language) {
			return "", "", fmt.Errorf("unknown language guideline %q", language)
		}
		path := filepath.Join(root, "guidelines", language+".md")
		content, err := os.ReadFile(path)
		if err != nil {
			return "", "", fmt.Errorf("cannot read %s guidelines (run: aiContext setup): %w", language, err)
		}
		sections = append(sections, strings.TrimSpace(string(content)))
	}
	languageContent := "<!-- No language-specific guidelines selected. -->"
	if len(sections) > 0 {
		languageContent = strings.Join(sections, "\n\n")
	}
	return strings.TrimSpace(string(profileContent)), languageContent, nil
}

func resolveLanguageSelection(root, value string) ([]string, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "", "none":
		return nil, nil
	case "auto":
		return detectProjectLanguages(root)
	}
	requested := make(map[string]bool)
	for _, item := range strings.Split(value, ",") {
		name := strings.TrimSpace(item)
		if name == "all" {
			for _, language := range languageDefinitions {
				requested[language.name] = true
			}
			continue
		}
		if !isSupportedLanguage(name) {
			return nil, fmt.Errorf("unknown language guideline %q (supported: %s)", name, supportedLanguageList())
		}
		requested[name] = true
	}
	languages := make([]string, 0, len(requested))
	for _, language := range languageDefinitions {
		if requested[language.name] {
			languages = append(languages, language.name)
		}
	}
	return languages, nil
}

func detectProjectLanguages(root string) ([]string, error) {
	detected := make(map[string]bool)
	if fileExists(root, "go.mod") {
		detected["go"] = true
	}
	if fileExists(root, "Cargo.toml") {
		detected["rust"] = true
	}
	if data, exists, err := readProjectFile(root, "package.json"); err != nil {
		return nil, err
	} else if exists {
		var manifest nodePackage
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("cannot parse package.json for language detection: %w", err)
		}
		if fileExists(root, "tsconfig.json") || manifest.hasDependency("typescript") {
			detected["typescript"] = true
		} else {
			detected["javascript"] = true
		}
	}
	for _, marker := range []string{"pyproject.toml", "requirements.txt", "setup.py"} {
		if fileExists(root, marker) {
			detected["python"] = true
		}
	}
	if fileExists(root, "Gemfile") {
		detected["ruby"] = true
	}
	if fileExists(root, "pom.xml") || fileExists(root, "build.gradle") {
		detected["java"] = true
	}
	if fileExists(root, "build.gradle.kts") {
		detected["kotlin"] = true
	}
	solutions, _ := filepath.Glob(filepath.Join(root, "*.sln"))
	projects, _ := filepath.Glob(filepath.Join(root, "*.csproj"))
	if len(solutions) > 0 || len(projects) > 0 {
		detected["csharp"] = true
	}
	if fileExists(root, "Package.swift") {
		detected["swift"] = true
	}
	if fileExists(root, "composer.json") {
		detected["php"] = true
	}
	terraforms, _ := filepath.Glob(filepath.Join(root, "*.tf"))
	if len(terraforms) > 0 {
		detected["terraform"] = true
	}

	languages := make([]string, 0, len(detected))
	for _, language := range languageDefinitions {
		if detected[language.name] {
			languages = append(languages, language.name)
		}
	}
	return languages, nil
}

func validateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	for _, char := range name {
		if unicode.IsLower(char) || unicode.IsDigit(char) || char == '-' || char == '_' {
			continue
		}
		return fmt.Errorf("invalid profile name %q; use lowercase letters, numbers, hyphens, or underscores", name)
	}
	return nil
}

func isSupportedLanguage(name string) bool {
	for _, language := range languageDefinitions {
		if language.name == name {
			return true
		}
	}
	return false
}

func supportedLanguageList() string {
	names := make([]string, 0, len(languageDefinitions))
	for _, language := range languageDefinitions {
		names = append(names, language.name)
	}
	return strings.Join(names, ",")
}

func printProfiles(w io.Writer, templateDir string) error {
	root := filepath.Dir(filepath.Clean(templateDir))
	profileDir := filepath.Join(root, "profiles")
	entries, err := os.ReadDir(profileDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("profile directory is missing (run: aiContext setup)")
		}
		return fmt.Errorf("cannot read profile directory: %w", err)
	}
	profiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		profiles = append(profiles, strings.TrimSuffix(entry.Name(), ".md"))
	}
	sort.Strings(profiles)
	fmt.Fprintln(w, "Profiles:")
	for _, profile := range profiles {
		fmt.Fprintln(w, " ", profile)
	}
	fmt.Fprintln(w, "Language guideline packs:")
	for _, language := range languageDefinitions {
		fmt.Fprintf(w, "  %-12s %s\n", language.name, language.description)
	}
	fmt.Fprintln(w, "Config root:", root)
	return nil
}
