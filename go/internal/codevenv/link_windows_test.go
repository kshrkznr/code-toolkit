//go:build windows

package codevenv

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWindowsJunctionIsManagedLink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	link := filepath.Join(root, "current.code")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := createManagedLink(target, link); err != nil {
		t.Fatal(err)
	}

	resolved, exists, err := linkTarget(link)
	if err != nil {
		t.Fatal(err)
	}
	if !exists || !samePath(resolved, target) {
		t.Fatalf("junction target = %q, exists = %t; want %q, true", resolved, exists, target)
	}
	if state := pathState(link, target); state != "linked" {
		t.Fatalf("junction state = %q; want linked", state)
	}
}

func TestCurrentListsWindowsJunction(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "origin.kiro")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := createManagedLink(target, filepath.Join(root, "current.kiro")); err != nil {
		t.Fatal(err)
	}

	got, err := Current(root, "")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"kiro": "origin.kiro"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Current() = %v, want %v", got, want)
	}
}

func TestReplaceSelectionWithWindowsJunction(t *testing.T) {
	root := t.TempDir()
	oldTarget := filepath.Join(root, "old")
	newTarget := filepath.Join(root, "new")
	for _, target := range []string{oldTarget, newTarget} {
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	current := filepath.Join(root, "current.code")
	if err := createManagedLink(oldTarget, current); err != nil {
		t.Fatal(err)
	}

	if err := replaceSelection(current, newTarget); err != nil {
		t.Fatal(err)
	}
	resolved, exists, err := linkTarget(current)
	if err != nil {
		t.Fatal(err)
	}
	if !exists || !samePath(resolved, newTarget) {
		t.Fatalf("junction target = %q, exists = %t; want %q, true", resolved, exists, newTarget)
	}
}
