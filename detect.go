package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type detectedProjectContext struct {
	stack    string
	commands string
}

type detectedCommand struct {
	action  string
	command string
	note    string
}

type projectDetector struct {
	root     string
	stack    []string
	commands []detectedCommand
}

func detectProjectContext(root string) (detectedProjectContext, error) {
	detector := &projectDetector{root: root}
	if err := detector.detectGo(); err != nil {
		return detectedProjectContext{}, err
	}
	if err := detector.detectNode(); err != nil {
		return detectedProjectContext{}, err
	}
	detector.detectRust()
	if err := detector.detectPython(); err != nil {
		return detectedProjectContext{}, err
	}
	if err := detector.detectRuby(); err != nil {
		return detectedProjectContext{}, err
	}
	detector.detectJava()
	detector.detectDotNet()
	detector.detectOtherLanguages()
	detector.detectInfrastructure()

	stack := "Unknown — no supported project manifests detected"
	if len(detector.stack) > 0 {
		stack = strings.Join(detector.stack, " · ")
	}
	commands := "<!-- No common commands were detected; add project-specific commands here. -->"
	if len(detector.commands) > 0 {
		var rows strings.Builder
		for _, command := range detector.commands {
			note := command.note
			if note == "" {
				note = "—"
			}
			fmt.Fprintf(&rows, "| %s | `%s` | %s |\n", command.action, command.command, note)
		}
		commands = strings.TrimSuffix(rows.String(), "\n")
	}
	return detectedProjectContext{stack: stack, commands: commands}, nil
}

func (d *projectDetector) detectGo() error {
	data, exists, err := readProjectFile(d.root, "go.mod")
	if err != nil || !exists {
		return err
	}
	d.addStack("Go")
	modules := string(data)
	for _, framework := range []struct{ dependency, name string }{
		{dependency: "github.com/gin-gonic/gin", name: "Gin"},
		{dependency: "github.com/gofiber/fiber", name: "Fiber"},
		{dependency: "github.com/labstack/echo", name: "Echo"},
		{dependency: "gorm.io/gorm", name: "GORM"},
	} {
		if strings.Contains(modules, framework.dependency) {
			d.addStack(framework.name)
		}
	}
	d.addCommand("Build", "go build ./...", "")
	d.addCommand("Test", "go test ./...", "")
	d.addCommand("Vet", "go vet ./...", "")
	return nil
}

type nodePackage struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func (d *projectDetector) detectNode() error {
	data, exists, err := readProjectFile(d.root, "package.json")
	if err != nil || !exists {
		return err
	}
	var manifest nodePackage
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("cannot parse package.json for project detection: %w", err)
	}

	if fileExists(d.root, "tsconfig.json") || manifest.hasDependency("typescript") {
		d.addStack("TypeScript")
	} else {
		d.addStack("JavaScript")
	}
	switch {
	case manifest.hasDependency("next"):
		d.addStack("Next.js")
	case manifest.hasDependency("@angular/core"):
		d.addStack("Angular")
	case manifest.hasDependency("vue"):
		d.addStack("Vue")
	case manifest.hasDependency("svelte"):
		d.addStack("Svelte")
	case manifest.hasDependency("astro"):
		d.addStack("Astro")
	case manifest.hasDependency("react"):
		d.addStack("React")
	}
	for _, database := range []struct{ dependency, name string }{
		{dependency: "@prisma/client", name: "Prisma"},
		{dependency: "drizzle-orm", name: "Drizzle ORM"},
		{dependency: "mongoose", name: "Mongoose"},
		{dependency: "typeorm", name: "TypeORM"},
	} {
		if manifest.hasDependency(database.dependency) {
			d.addStack(database.name)
		}
	}

	packageManager := "npm"
	switch {
	case fileExists(d.root, "pnpm-lock.yaml"):
		packageManager = "pnpm"
	case fileExists(d.root, "yarn.lock"):
		packageManager = "yarn"
	case fileExists(d.root, "bun.lock"), fileExists(d.root, "bun.lockb"):
		packageManager = "bun"
	}
	for _, script := range []struct{ name, action string }{
		{name: "dev", action: "Dev"},
		{name: "build", action: "Build"},
		{name: "test", action: "Test"},
		{name: "lint", action: "Lint"},
		{name: "typecheck", action: "Typecheck"},
	} {
		if _, ok := manifest.Scripts[script.name]; ok {
			d.addCommand(script.action, packageManager+" run "+script.name, "")
		}
	}
	return nil
}

func (p nodePackage) hasDependency(name string) bool {
	_, production := p.Dependencies[name]
	_, development := p.DevDependencies[name]
	return production || development
}

func (d *projectDetector) detectRust() {
	if !fileExists(d.root, "Cargo.toml") {
		return
	}
	d.addStack("Rust")
	d.addCommand("Build", "cargo build", "")
	d.addCommand("Test", "cargo test", "")
	d.addCommand("Lint", "cargo clippy", "")
}

func (d *projectDetector) detectPython() error {
	var contents strings.Builder
	found := false
	for _, name := range []string{"pyproject.toml", "requirements.txt", "setup.py"} {
		data, exists, err := readProjectFile(d.root, name)
		if err != nil {
			return err
		}
		if exists {
			found = true
			contents.Write(data)
		}
	}
	if !found {
		return nil
	}
	d.addStack("Python")
	lower := strings.ToLower(contents.String())
	for _, framework := range []struct{ marker, name string }{
		{marker: "django", name: "Django"},
		{marker: "fastapi", name: "FastAPI"},
		{marker: "flask", name: "Flask"},
	} {
		if strings.Contains(lower, framework.marker) {
			d.addStack(framework.name)
		}
	}
	if strings.Contains(lower, "pytest") || fileExists(d.root, "pytest.ini") {
		d.addCommand("Test", "python -m pytest", "")
	}
	return nil
}

func (d *projectDetector) detectRuby() error {
	data, exists, err := readProjectFile(d.root, "Gemfile")
	if err != nil || !exists {
		return err
	}
	d.addStack("Ruby")
	lower := strings.ToLower(string(data))
	if strings.Contains(lower, "rails") {
		d.addStack("Rails")
		d.addCommand("Test", "bin/rails test", "")
	} else if strings.Contains(lower, "rspec") {
		d.addCommand("Test", "bundle exec rspec", "")
	}
	return nil
}

func (d *projectDetector) detectJava() {
	if fileExists(d.root, "pom.xml") {
		d.addStack("Java")
		d.addStack("Maven")
		maven := "mvn"
		if fileExists(d.root, "mvnw") {
			maven = "./mvnw"
		}
		d.addCommand("Build", maven+" package", "")
		d.addCommand("Test", maven+" test", "")
	}
	if fileExists(d.root, "build.gradle") || fileExists(d.root, "build.gradle.kts") {
		d.addStack("Java/Kotlin")
		d.addStack("Gradle")
		gradle := "gradle"
		if fileExists(d.root, "gradlew") {
			gradle = "./gradlew"
		}
		d.addCommand("Build", gradle+" build", "")
		d.addCommand("Test", gradle+" test", "")
	}
}

func (d *projectDetector) detectDotNet() {
	solutions, _ := filepath.Glob(filepath.Join(d.root, "*.sln"))
	projects, _ := filepath.Glob(filepath.Join(d.root, "*.csproj"))
	if len(solutions) == 0 && len(projects) == 0 {
		return
	}
	d.addStack(".NET")
	d.addCommand("Build", "dotnet build", "")
	d.addCommand("Test", "dotnet test", "")
}

func (d *projectDetector) detectOtherLanguages() {
	if fileExists(d.root, "Package.swift") {
		d.addStack("Swift")
		d.addStack("Swift Package Manager")
		d.addCommand("Build", "swift build", "")
		d.addCommand("Test", "swift test", "")
	}
	if fileExists(d.root, "composer.json") {
		d.addStack("PHP")
		d.addStack("Composer")
		d.addCommand("Setup", "composer install", "")
	}
}

func (d *projectDetector) detectInfrastructure() {
	if fileExists(d.root, "Dockerfile") {
		d.addStack("Docker")
		d.addCommand("Container", "docker build .", "")
	}
	if fileExists(d.root, "compose.yaml") || fileExists(d.root, "compose.yml") || fileExists(d.root, "docker-compose.yml") {
		d.addStack("Docker Compose")
		d.addCommand("Services", "docker compose up", "")
	}
	terraforms, _ := filepath.Glob(filepath.Join(d.root, "*.tf"))
	if len(terraforms) > 0 {
		d.addStack("Terraform")
		d.addCommand("Validate", "terraform validate", "")
	}
}

func (d *projectDetector) addStack(name string) {
	for _, existing := range d.stack {
		if existing == name {
			return
		}
	}
	d.stack = append(d.stack, name)
}

func (d *projectDetector) addCommand(action, command, note string) {
	for _, existing := range d.commands {
		if existing.command == command {
			return
		}
	}
	d.commands = append(d.commands, detectedCommand{action: action, command: command, note: note})
}

func readProjectFile(root, name string) ([]byte, bool, error) {
	path := filepath.Join(root, filepath.FromSlash(name))
	data, err := os.ReadFile(path)
	if err == nil {
		return data, true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("cannot read %s for project detection: %w", name, err)
}

func fileExists(root, name string) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(name)))
	return err == nil && !info.IsDir()
}
