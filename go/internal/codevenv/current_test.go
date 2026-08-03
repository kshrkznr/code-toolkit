package codevenv

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCurrent(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"vscode-golang", "kiro-default"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(dir, "vscode-golang"), filepath.Join(dir, "current.code")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "kiro-default"), filepath.Join(dir, "current.kiro")); err != nil {
		t.Fatal(err)
	}

	got, err := Current(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"code": "vscode-golang", "kiro": "kiro-default"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Current() = %v, want %v", got, want)
	}
}

func TestCurrentInactivePlatform(t *testing.T) {
	got, err := Current(t.TempDir(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if got["code"] != "none" {
		t.Fatalf("Current()[code] = %q, want none", got["code"])
	}
}
