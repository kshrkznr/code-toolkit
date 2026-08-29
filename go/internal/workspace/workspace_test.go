package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaultsToWorkspaceOwnedPaths(t *testing.T) {
	root := t.TempDir()
	makeCookbook(t, filepath.Join(root, "cookbook"))

	paths, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if paths.CookbookSource != filepath.Join(root, "cookbook") || paths.Workbench != filepath.Join(root, "cookbook") || paths.Dist != filepath.Join(root, "dist") {
		t.Fatalf("unexpected paths: %+v", paths)
	}
	if paths.Archive != filepath.Join(root, "archive") || paths.Pool != filepath.Join(root, ".vsix") {
		t.Fatalf("Workspace-owned paths moved: %+v", paths)
	}
}

func TestLoadRelocatesOnlyCookbookSourceAndDist(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "shared-cookbook")
	makeCookbook(t, source)
	if err := os.MkdirAll(filepath.Join(root, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "paths:\n  cookbook-source: " + source + "\n  dist: artifacts/dist\n"
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(configPath)), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if paths.CookbookSource != source || paths.Dist != filepath.Join(root, "artifacts", "dist") {
		t.Fatalf("overrides = %+v", paths)
	}
	if paths.Workbench != filepath.Join(root, "cookbook") || paths.Archive != filepath.Join(root, "archive") || paths.Pool != filepath.Join(root, ".vsix") {
		t.Fatalf("Workspace ownership changed = %+v", paths)
	}
	if !HasMarker(root) {
		t.Fatal("Workspace configuration should be a discovery marker")
	}
}

func TestLoadRejectsUnknownFieldsAndMissingCookbook(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, filepath.FromSlash(configPath))
	if err := os.WriteFile(path, []byte("paths:\n  archive: elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "field archive not found") {
		t.Fatalf("unknown field error = %v", err)
	}
	if err := os.WriteFile(path, []byte("paths:\n  cookbook-source: missing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "must contain recipe and ingredient") {
		t.Fatalf("missing Cookbook error = %v", err)
	}
}

func makeCookbook(t *testing.T, root string) {
	t.Helper()
	for _, name := range []string{"recipe", "ingredient"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}
