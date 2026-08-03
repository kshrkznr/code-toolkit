package recipe

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Recipe struct {
	Name     string   `yaml:"name"`
	OS       string   `yaml:"os"`
	Platform string   `yaml:"platform"`
	Runtime  []string `yaml:"runtime"`
	Profile  []string `yaml:"profile"`
	Config   Config   `yaml:"config"`
}

type Config struct {
	DistStrategy    DistStrategy               `yaml:"dist-strategy"`
	ProfileStrategy map[string]ProfileStrategy `yaml:"profile-strategy"`
}

type DistStrategy struct {
	ExtensionMarketplace *bool          `yaml:"extension-marketplace"`
	LockMode             string         `yaml:"lock-mode"`
	DefaultProfile       DefaultProfile `yaml:"default-profile"`
}

type DefaultProfile struct {
	Extensions  string `yaml:"extensions"`
	Settings    string `yaml:"settings"`
	Keybindings string `yaml:"keybindings"`
	Tasks       string `yaml:"tasks"`
	MCP         string `yaml:"mcp"`
	Snippets    string `yaml:"snippets"`
}

type ProfileStrategy struct {
	Settings    string `yaml:"settings"`
	Keybindings string `yaml:"keybindings"`
	Tasks       string `yaml:"tasks"`
	MCP         string `yaml:"mcp"`
	Snippets    string `yaml:"snippets"`
}

func (s ProfileStrategy) Content(content string) string {
	switch content {
	case "settings":
		return s.Settings
	case "keybindings":
		return s.Keybindings
	case "tasks":
		return s.Tasks
	case "mcp":
		return s.MCP
	case "snippets":
		return s.Snippets
	default:
		return ""
	}
}

func (r Recipe) DefaultExtensionMode() string {
	if r.Config.DistStrategy.DefaultProfile.Extensions == "" {
		return "runtime"
	}
	return r.Config.DistStrategy.DefaultProfile.Extensions
}

func (r Recipe) DefaultContent(content string) string {
	var value string
	switch content {
	case "extensions":
		value = r.Config.DistStrategy.DefaultProfile.Extensions
	case "settings":
		value = r.Config.DistStrategy.DefaultProfile.Settings
	case "keybindings":
		value = r.Config.DistStrategy.DefaultProfile.Keybindings
	case "tasks":
		value = r.Config.DistStrategy.DefaultProfile.Tasks
	case "mcp":
		value = r.Config.DistStrategy.DefaultProfile.MCP
	case "snippets":
		value = r.Config.DistStrategy.DefaultProfile.Snippets
	default:
		return ""
	}
	if value != "" {
		return value
	}
	if content == "mcp" {
		return "unmanaged"
	}
	return "runtime"
}

func (r Recipe) ExtensionMarketplace() bool {
	if r.Config.DistStrategy.ExtensionMarketplace == nil {
		return true
	}
	return *r.Config.DistStrategy.ExtensionMarketplace
}

func (r Recipe) LockMode() string {
	if r.Config.DistStrategy.LockMode == "" {
		return "refresh"
	}
	return r.Config.DistStrategy.LockMode
}

func (r Recipe) ProfileContent(name string) ProfileStrategy {
	strategy := r.Config.ProfileStrategy[name]
	if strategy.Settings == "" {
		strategy.Settings = "default"
	}
	if strategy.Keybindings == "" {
		strategy.Keybindings = "default"
	}
	if strategy.Tasks == "" {
		strategy.Tasks = "default"
	}
	if strategy.MCP == "" {
		strategy.MCP = "default"
	}
	if strategy.Snippets == "" {
		strategy.Snippets = "default"
	}
	return strategy
}

func Load(path string) (Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Recipe{}, fmt.Errorf("read recipe %s: %w", path, err)
	}

	var value Recipe
	if err := yaml.Unmarshal(data, &value); err != nil {
		return Recipe{}, fmt.Errorf("parse recipe %s: %w", path, err)
	}
	if value.Name == "" {
		return Recipe{}, fmt.Errorf("recipe.name missing: %s", path)
	}
	return value, nil
}
