package main

import (
	"fmt"
	"strings"
)

const manifestFilename = ".aicontext.json"

type toolDefinition struct {
	name        string
	description string
	template    *templateSpec
}

var canonicalTemplate = templateSpec{source: "AGENTS.md", destination: "AGENTS.md"}

var toolDefinitions = []toolDefinition{
	{name: "codex", description: "AGENTS.md (no additional adapter)"},
	{name: "claude", description: "Claude Code", template: &templateSpec{source: "CLAUDE.md", destination: "CLAUDE.md"}},
	{name: "cursor", description: "Cursor", template: &templateSpec{source: "cursor.mdc", destination: ".cursor/rules/aicontext.mdc"}},
	{name: "copilot", description: "GitHub Copilot", template: &templateSpec{source: "copilot-instructions.md", destination: ".github/copilot-instructions.md"}},
	{name: "gemini", description: "Gemini CLI", template: &templateSpec{source: "GEMINI.md", destination: "GEMINI.md"}},
}

var defaultTools = []string{"codex", "claude", "cursor", "copilot"}

func parseToolSelection(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("tool selection cannot be empty")
	}
	requested := make(map[string]bool)
	for _, item := range strings.Split(value, ",") {
		name := strings.ToLower(strings.TrimSpace(item))
		if name == "all" {
			for _, tool := range toolDefinitions {
				requested[tool.name] = true
			}
			continue
		}
		if !isSupportedTool(name) {
			return nil, fmt.Errorf("unknown tool %q (supported: %s)", name, supportedToolList())
		}
		requested[name] = true
	}

	tools := make([]string, 0, len(requested))
	for _, tool := range toolDefinitions {
		if requested[tool.name] {
			tools = append(tools, tool.name)
		}
	}
	if len(tools) == 0 {
		return nil, fmt.Errorf("tool selection cannot be empty")
	}
	return tools, nil
}

func isSupportedTool(name string) bool {
	for _, tool := range toolDefinitions {
		if tool.name == name {
			return true
		}
	}
	return false
}

func supportedToolList() string {
	names := make([]string, 0, len(toolDefinitions))
	for _, tool := range toolDefinitions {
		names = append(names, tool.name)
	}
	return strings.Join(names, ",")
}

func templateSpecsForTools(tools []string) []templateSpec {
	specs := []templateSpec{canonicalTemplate}
	selected := make(map[string]bool, len(tools))
	for _, name := range tools {
		selected[name] = true
	}
	for _, tool := range toolDefinitions {
		if selected[tool.name] && tool.template != nil {
			specs = append(specs, *tool.template)
		}
	}
	return specs
}

func adapterSpecsForTools(tools []string) []templateSpec {
	all := templateSpecsForTools(tools)
	return all[1:]
}

func setupTemplateSpecs() []templateSpec {
	specs := []templateSpec{canonicalTemplate}
	for _, tool := range toolDefinitions {
		if tool.template != nil {
			specs = append(specs, *tool.template)
		}
	}
	return specs
}

func toolForDestination(destination string) string {
	if destination == canonicalTemplate.destination {
		return "canonical"
	}
	for _, tool := range toolDefinitions {
		if tool.template != nil && tool.template.destination == destination {
			return tool.name
		}
	}
	return ""
}
