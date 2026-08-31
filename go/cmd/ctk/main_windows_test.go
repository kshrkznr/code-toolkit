//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/docbundle"
)

func TestOpenExecutableDocumentationBundleThroughDirectoryJunction(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := docbundle.Generate(repositoryRoot, docbundle.Metadata{
		Version:  "test",
		Revision: "junction-test",
	})
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	versionDirectory := filepath.Join(root, "0.0.0")
	if err := os.Mkdir(versionDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(versionDirectory, "ctk.exe")
	if err := os.WriteFile(executable, []byte("executable-prefix"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := docbundle.AppendExecutable(executable, generated.Archive); err != nil {
		t.Fatal(err)
	}

	current := filepath.Join(root, "current")
	if output, err := exec.Command("cmd", "/c", "mklink", "/J", current, versionDirectory).CombinedOutput(); err != nil {
		t.Fatalf("create directory junction: %v\n%s", err, output)
	}
	bundle, err := openExecutableDocumentationBundle(filepath.Join(current, "ctk.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if status := docbundle.PackagedSourceStatus(bundle); !strings.Contains(status.Revision, "junction-test") {
		t.Fatalf("packaged source revision = %q", status.Revision)
	}
}
