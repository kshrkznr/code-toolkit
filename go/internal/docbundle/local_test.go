package docbundle

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const localFixtureDocument = `# Knowledge.local.md
============================================================

# Local Documentation

Original local body.
`

const localFixtureDefinition = `format-version: 1
repository: https://github.com/example/ctk
documents:
  files: [README.md, doc.md]
  trees: []
  exclude: []
nodes:
  local: doc.md
bootstrap-template: doc/bootstrap.md.tmpl
`

func TestOpenLocalReportsIndependentSourceStatus(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	writeLocalFixture(t, root)
	gitRun(t, root, "init")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "-c", "user.name=CTK Test", "-c", "user.email=ctk@example.invalid", "commit", "-m", "initial")
	packagedRevision := strings.TrimSpace(gitRun(t, root, "rev-parse", "HEAD"))
	generated, err := Generate(root, Metadata{Version: "v1.0.0", Revision: packagedRevision})
	if err != nil {
		t.Fatal(err)
	}
	packaged, err := Open(generated.Archive)
	if err != nil {
		t.Fatal(err)
	}

	matching, err := OpenLocal(root, packaged)
	if err != nil {
		t.Fatal(err)
	}
	if matching.Status.RevisionMatch != Match || matching.Status.DefinitionMatch != Match || matching.Status.ContentMatch != Match {
		t.Fatalf("matching comparisons = %+v", matching.Status)
	}
	if matching.Status.SelectedPathDirty != Clean || matching.Status.RepositoryDirty != Clean {
		t.Fatalf("matching dirty status = %+v", matching.Status)
	}

	if err := os.WriteFile(filepath.Join(root, "unrelated.txt"), []byte("unrelated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unrelated, err := OpenLocal(root, packaged)
	if err != nil {
		t.Fatal(err)
	}
	if unrelated.Status.RepositoryDirty != Dirty || unrelated.Status.SelectedPathDirty != Clean || len(unrelated.Status.SelectedDirtyPaths) != 0 || unrelated.Status.ContentMatch != Match {
		t.Fatalf("unrelated dirty status = %+v", unrelated.Status)
	}

	gitRun(t, root, "add", "unrelated.txt")
	gitRun(t, root, "-c", "user.name=CTK Test", "-c", "user.email=ctk@example.invalid", "commit", "-m", "unrelated")
	revisionMismatch, err := OpenLocal(root, packaged)
	if err != nil {
		t.Fatal(err)
	}
	if revisionMismatch.Status.RevisionMatch != Mismatch || revisionMismatch.Status.ContentMatch != Match || revisionMismatch.Status.SelectedPathDirty != Clean || revisionMismatch.Status.RepositoryDirty != Clean {
		t.Fatalf("revision-only mismatch = %+v", revisionMismatch.Status)
	}

	edited := strings.Replace(localFixtureDocument, "Original local body.", "Edited local body.", 1)
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	selectedDirty, err := OpenLocal(root, packaged)
	if err != nil {
		t.Fatal(err)
	}
	if selectedDirty.Status.ContentMatch != Mismatch || selectedDirty.Status.SelectedPathDirty != Dirty || !slices.Equal(selectedDirty.Status.SelectedDirtyPaths, []string{"doc.md"}) {
		t.Fatalf("selected dirty status = %+v", selectedDirty.Status)
	}
	shown, err := selectedDirty.Bundle.Show("Knowledge.local.md")
	if err != nil || !bytes.Contains(shown, []byte("Edited local body.")) {
		t.Fatalf("local Show did not use edited content: %v\n%s", err, shown)
	}

	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte(localFixtureDocument), 0o644); err != nil {
		t.Fatal(err)
	}
	definitionPath := filepath.Join(root, filepath.FromSlash(DefinitionPath))
	definition, err := os.ReadFile(definitionPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(definitionPath, append(definition, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	definitionDirty, err := OpenLocal(root, packaged)
	if err != nil {
		t.Fatal(err)
	}
	if definitionDirty.Status.DefinitionMatch != Mismatch || definitionDirty.Status.ContentMatch != Match || definitionDirty.Status.SelectedPathDirty != Clean {
		t.Fatalf("Definition-only mismatch = %+v", definitionDirty.Status)
	}
}

func TestOpenLocalSupportsNonGitSourceAndRejectsSymlinks(t *testing.T) {
	root := t.TempDir()
	writeLocalFixture(t, root)
	generated, err := Generate(root, Metadata{Version: "dev", Revision: "packaged"})
	if err != nil {
		t.Fatal(err)
	}
	packaged, err := Open(generated.Archive)
	if err != nil {
		t.Fatal(err)
	}
	local, err := OpenLocal(root, packaged)
	if err != nil {
		t.Fatal(err)
	}
	if local.Status.RevisionMatch != Unknown || local.Status.SelectedPathDirty != DirtyUnknown || local.Status.RepositoryDirty != DirtyUnknown {
		t.Fatalf("non-Git status = %+v", local.Status)
	}
	if local.Status.DefinitionMatch != Match || local.Status.ContentMatch != Match {
		t.Fatalf("non-Git comparisons = %+v", local.Status)
	}

	link := filepath.Join(t.TempDir(), "repository-link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := OpenLocal(link, packaged); err == nil {
		t.Fatal("local source accepted a symlink root")
	}

	target := filepath.Join(t.TempDir(), "replacement.md")
	if err := os.WriteFile(target, []byte(localFixtureDocument), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "doc.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "doc.md")); err != nil {
		t.Skipf("selected-file symlink unavailable: %v", err)
	}
	if _, err := OpenLocal(root, packaged); err == nil {
		t.Fatal("local source accepted a selected-file symlink")
	}
}

func TestOpenLocalReusesStrictDefinitionValidation(t *testing.T) {
	baseline := t.TempDir()
	writeLocalFixture(t, baseline)
	generated, err := Generate(baseline, Metadata{Version: "dev", Revision: "packaged"})
	if err != nil {
		t.Fatal(err)
	}
	packaged, err := Open(generated.Archive)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		files string
		want  string
	}{
		{name: "traversal", files: "[../outside.md]", want: "escapes the repository"},
		{name: "duplicate", files: "[doc.md, doc.md]", want: "duplicate selected document"},
		{name: "case collision", files: "[doc.md, DOC.md]", want: "case-insensitive document collision"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeLocalFixture(t, root)
			definition := strings.Replace(localFixtureDefinition, "[README.md, doc.md]", test.files, 1)
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(DefinitionPath)), []byte(definition), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := OpenLocal(root, packaged); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validation error = %v, want %q", err, test.want)
			}
		})
	}
	t.Run("non-regular selected path", func(t *testing.T) {
		root := t.TempDir()
		writeLocalFixture(t, root)
		definition := strings.Replace(localFixtureDefinition, "[README.md, doc.md]", "[special.md]", 1)
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(DefinitionPath)), []byte(definition), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(root, "special.md"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := OpenLocal(root, packaged); err == nil || !strings.Contains(err.Error(), "regular file") {
			t.Fatalf("non-regular validation error = %v", err)
		}
	})
}

func writeLocalFixture(t *testing.T, root string) {
	t.Helper()
	for name, content := range map[string]string{
		"README.md":             "# Knowledge.local-root.md\n============================================================\n\n# Local Root\n",
		"doc.md":                localFixtureDocument,
		DefinitionPath:          localFixtureDefinition,
		"doc/bootstrap.md.tmpl": "{{ bundle-provenance }}\n{{ include-document \"doc.md\" }}\n",
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("Git unavailable: %v", err)
	}
}

func gitRun(t *testing.T, root string, args ...string) string {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	output, err := exec.Command("git", commandArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return string(output)
}
