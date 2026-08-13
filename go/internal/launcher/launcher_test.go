package launcher

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
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
	l := &Launcher{GOOS: "darwin", lookPath: func(command string) (string, error) { return "/bin/" + command, nil }}
	dist := distribution.Distribution{Path: dir, Recipe: recipe.Recipe{Platform: "code"}}
	command, args, isOverride, err := l.command(dist, []string{"extra"})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(command) != "code" || isOverride {
		t.Fatalf("command = %q override=%v", command, isOverride)
	}
	want := []string{"--user-data-dir", filepath.Join(dir, ".data"), "--extensions-dir", filepath.Join(dir, ".ext"), "extra"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
}

func TestCommandRejectsUnknownNativePlatform(t *testing.T) {
	l := &Launcher{GOOS: "darwin", lookPath: func(string) (string, error) { return "", errors.New("must not be called") }}
	_, _, _, err := l.command(distribution.Distribution{Path: t.TempDir(), Recipe: recipe.Recipe{Platform: "unknown"}}, nil)
	if err == nil {
		t.Fatal("expected unknown Platform error")
	}
}
