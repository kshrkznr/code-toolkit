package recipe

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recipe.yaml")
	content := []byte(`name: vscode-golang
os: macos
platform: code
runtime:
  - common
  - golang
profile:
  - review
config:
  future-field: preserved-by-source
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := Recipe{
		Name:     "vscode-golang",
		OS:       "macos",
		Platform: "code",
		Runtime:  []string{"common", "golang"},
		Profile:  []string{"review"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestStrategyDefaults(t *testing.T) {
	recipe := Recipe{}
	if recipe.LockMode() != "refresh" {
		t.Fatalf("lock mode = %q", recipe.LockMode())
	}
	if recipe.DefaultExtensionMode() != "runtime" {
		t.Fatalf("extension mode = %q", recipe.DefaultExtensionMode())
	}
	for _, content := range []string{"extensions", "settings", "keybindings", "tasks", "snippets"} {
		if mode := recipe.DefaultContent(content); mode != "runtime" {
			t.Fatalf("default %s mode = %q", content, mode)
		}
	}
	if mode := recipe.DefaultContent("mcp"); mode != "unmanaged" {
		t.Fatalf("default mcp mode = %q", mode)
	}
	if !recipe.ExtensionMarketplace() {
		t.Fatal("extension marketplace should default to true")
	}
}

func TestLoadRequiresName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recipe.yaml")
	if err := os.WriteFile(path, []byte("platform: code\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want missing name error")
	}
}
