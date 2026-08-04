package launcher

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
)

func TestCommandUsesOverride(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(override, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	l := &Launcher{GOOS: "darwin"}
	command, args, isOverride, err := l.command(distribution.Distribution{Path: dir}, []string{"."})
	if err != nil {
		t.Fatal(err)
	}
	if command != override || !isOverride || !reflect.DeepEqual(args, []string{"."}) {
		t.Fatalf("command = %q %v override=%v", command, args, isOverride)
	}
}

func TestCommandUsesNativePlatform(t *testing.T) {
	dir := t.TempDir()
	platform := "sh"
	if runtime.GOOS == "windows" {
		platform = "cmd.exe"
	}
	l := &Launcher{GOOS: runtime.GOOS}
	dist := distribution.Distribution{Path: dir, Recipe: recipe.Recipe{Platform: platform}}
	command, args, isOverride, err := l.command(dist, []string{"extra"})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(command) != platform || isOverride {
		t.Fatalf("command = %q override=%v", command, isOverride)
	}
	want := []string{"--user-data-dir", filepath.Join(dir, ".data"), "--extensions-dir", filepath.Join(dir, ".ext"), "extra"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
}
