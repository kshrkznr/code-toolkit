package codevenv

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
)

type fakeStopper struct {
	called bool
}

func (s *fakeStopper) StopForSelection(context.Context, string, ...string) error {
	s.called = true
	return nil
}

func (s *fakeStopper) StopRuntime(context.Context, string, ...string) error { return nil }

func TestUseReplacesSelection(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "old")
	newPath := filepath.Join(root, "new")
	for _, path := range []string{oldPath, newPath} {
		if err := os.MkdirAll(filepath.Join(path, ".data"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	current := filepath.Join(root, "current.code")
	if err := os.Symlink(oldPath, current); err != nil {
		t.Fatal(err)
	}
	stopper := &fakeStopper{}
	target := distribution.Distribution{Name: "new", Path: newPath, Recipe: recipe.Recipe{Platform: "code"}}
	result, err := Use(context.Background(), root, target, stopper)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || !stopper.called {
		t.Fatalf("result=%+v stopper.called=%v", result, stopper.called)
	}
	got, err := os.Readlink(current)
	if err != nil {
		t.Fatal(err)
	}
	if !samePath(got, newPath) {
		t.Fatalf("current target = %q, want %q", got, newPath)
	}
}

func TestUseSameSelectionIsNoOp(t *testing.T) {
	root := t.TempDir()
	targetPath := filepath.Join(root, "same")
	if err := os.Mkdir(targetPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetPath, filepath.Join(root, "current.code")); err != nil {
		t.Fatal(err)
	}
	stopper := &fakeStopper{}
	target := distribution.Distribution{Name: "same", Path: targetPath, Recipe: recipe.Recipe{Platform: "code"}}
	result, err := Use(context.Background(), root, target, stopper)
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed || stopper.called {
		t.Fatalf("result=%+v stopper.called=%v", result, stopper.called)
	}
}

func TestReplaceSelectionRollsBack(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "old")
	newPath := filepath.Join(root, "new")
	for _, path := range []string{oldPath, newPath} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	current := filepath.Join(root, "current.code")
	if err := os.Symlink(oldPath, current); err != nil {
		t.Fatal(err)
	}

	failed := false
	rename := func(old, new string) error {
		if !failed && filepath.Base(old) != "current.code" && filepath.Base(new) == "current.code" {
			failed = true
			return errors.New("injected replacement failure")
		}
		return os.Rename(old, new)
	}
	if err := replaceSelectionWith(current, newPath, rename); err == nil {
		t.Fatal("replaceSelectionWith() error = nil")
	}
	got, err := os.Readlink(current)
	if err != nil {
		t.Fatal(err)
	}
	if !samePath(got, oldPath) {
		t.Fatalf("current target = %q, want restored %q", got, oldPath)
	}
}
