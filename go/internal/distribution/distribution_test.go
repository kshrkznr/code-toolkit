package distribution

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestList(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"vscode-golang", "origin.code", "current.code", ".hidden"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "not-a-dist"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"origin.code", "vscode-golang"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
}

func TestRecipe(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "vscode-golang", ".meta")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "recipe.yaml"), []byte("name: vscode-golang\nplatform: code\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := Recipe(filepath.Dir(dir))
	if err != nil {
		t.Fatal(err)
	}
	if got.Platform != "code" {
		t.Fatalf("Recipe().Platform = %q, want code", got.Platform)
	}
}

func TestLoadAcceptsNameAndWorkspaceRelativePath(t *testing.T) {
	workspace := t.TempDir()
	distRoot := filepath.Join(workspace, "dist")
	distPath := filepath.Join(distRoot, "vscode-golang.1")
	if err := os.MkdirAll(filepath.Join(distPath, ".meta"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".data", ".ext"} {
		if err := os.Mkdir(filepath.Join(distPath, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(distPath, ".meta", "recipe.yaml"), []byte("name: vscode-golang\nplatform: code\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	byName, err := Load(distRoot, "vscode-golang.1")
	if err != nil {
		t.Fatal(err)
	}
	if byName.Path != distPath {
		t.Fatalf("Load(name).Path = %q, want %q", byName.Path, distPath)
	}

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	byPath, err := Load(distRoot, filepath.Join("dist", "vscode-golang.1"))
	if err != nil {
		t.Fatal(err)
	}
	gotInfo, err := os.Stat(byPath.Path)
	if err != nil {
		t.Fatal(err)
	}
	wantInfo, err := os.Stat(distPath)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(gotInfo, wantInfo) {
		t.Fatalf("Load(path).Path = %q, want same directory as %q", byPath.Path, distPath)
	}

	if err := os.Chdir(distRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(distRoot, filepath.Join("dist", "vscode-golang.1")); err == nil {
		t.Fatal("workspace-relative path unexpectedly resolved from inside dist")
	}
}

func TestLoadForLaunchAcceptsOverrideOnly(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "java-home")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "run.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	dist, err := LoadForLaunch(root, "java-home", "run.sh")
	if err != nil {
		t.Fatal(err)
	}
	if dist.Name != "java-home" || dist.Path != path {
		t.Fatalf("Distribution = %+v", dist)
	}
	if _, err := Load(root, "java-home"); err == nil {
		t.Fatal("ordinary Load accepted launch-only input")
	}
	if _, err := LoadForLaunch(root, "java-home", "run.cmd"); err == nil {
		t.Fatal("non-matching platform Override accepted launch-only input")
	}
}
