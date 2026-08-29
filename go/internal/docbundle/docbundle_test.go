package docbundle

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateRepositoryBundleIsDeterministic(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	metadata := Metadata{Version: "v0.4.0", Revision: "abc1234", Tag: "v0.4.0"}
	first, err := Generate(root, metadata)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate(root, metadata)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Archive, second.Archive) {
		t.Fatal("identical source did not produce an identical Bundle ZIP")
	}
	if first.Manifest.ContentSHA256 != second.Manifest.ContentSHA256 {
		t.Fatal("identical source did not produce an identical content digest")
	}

	bootstrap := string(first.Bootstrap)
	for _, included := range []string{"# Concept Domains", "# Documentation Resolver"} {
		if !strings.Contains(bootstrap, included) {
			t.Fatalf("Bootstrap does not contain %q", included)
		}
	}
	if !strings.Contains(bootstrap, "Bundle provenance: CTK `v0.4.0`, source `abc1234`") {
		t.Fatal("Bootstrap does not expose Bundle provenance")
	}
	for _, excluded := range []string{"## Why CTK?", "# Installation", "# Getting Started"} {
		if strings.Contains(bootstrap, excluded) {
			t.Fatalf("Bootstrap unexpectedly contains %q", excluded)
		}
	}
	if !strings.Contains(bootstrap, "https://github.com/kshrkznr/code-toolkit/tree/v0.4.0") {
		t.Fatal("Bootstrap does not route repository-only content to the exact Release tag")
	}

	documents := map[string]ManifestDocument{}
	for _, document := range first.Manifest.Documents {
		documents[document.Path] = document
	}
	core := documents["doc/core/README.md"]
	if core.Identity != "Knowledge.core.md" || len(core.Aliases) != 1 || core.Aliases[0] != "core" {
		t.Fatalf("Core document index = %+v", core)
	}
	if _, ok := documents["doc/project-knowledge/note/note.collaborative-review-surfaces.md"]; !ok {
		t.Fatal("Collaborative Review Surfaces exception is not bundled")
	}
	if _, ok := documents["doc/future/future.documentation-bundle.md"]; ok {
		t.Fatal("Future document must remain repository-only")
	}

	reader, err := zip.NewReader(bytes.NewReader(first.Archive), int64(len(first.Archive)))
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string][]byte{}
	for _, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			t.Fatal(err)
		}
		entries[file.Name] = content
	}
	for _, required := range []string{ManifestPath, BootstrapPath, "README.md", "doc/README.md", "go/README.md"} {
		if _, ok := entries[required]; !ok {
			t.Fatalf("Bundle ZIP does not contain %s", required)
		}
	}
}

func TestDecodeDefinitionIsStrict(t *testing.T) {
	valid := `format-version: 1
repository: https://github.com/kshrkznr/code-toolkit
documents:
  files: [README.md]
  trees: []
  exclude: []
nodes: {}
bootstrap-template: doc/template.md.tmpl
`
	if _, err := decodeDefinition([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	if _, err := decodeDefinition([]byte(valid + "unknown: true\n")); err == nil {
		t.Fatal("unknown Definition field was accepted")
	}
	if _, err := decodeDefinition([]byte(strings.Replace(valid, "format-version: 1", "format-version: 2", 1))); err == nil {
		t.Fatal("unsupported Definition version was accepted")
	}
}

func TestHeadingRangeRequiresUniqueOrderedHeadings(t *testing.T) {
	content := "# One\nfirst\n# Two\nsecond\n# Three\n"
	got, err := headingRange(content, "# One", "# Three")
	if err != nil {
		t.Fatal(err)
	}
	if got != "# One\nfirst\n# Two\nsecond\n" {
		t.Fatalf("range = %q", got)
	}
	if _, err := headingRange(content, "# Three", "# One"); err == nil {
		t.Fatal("reversed range was accepted")
	}
	if _, err := headingRange("# One\n# One\n# Two\n", "# One", "# Two"); err == nil {
		t.Fatal("duplicate range heading was accepted")
	}
}

func TestResolveRelativeLinkPreservesFragment(t *testing.T) {
	resolved, fragment, ok := resolveRelativeLink("doc/note/note.example.md", "../core/core.cookbook.md#responsibility")
	if !ok || resolved != "doc/core/core.cookbook.md" || fragment != "#responsibility" {
		t.Fatalf("resolved = %q %q %v", resolved, fragment, ok)
	}
	if _, _, ok := resolveRelativeLink("doc/README.md", "https://example.com"); ok {
		t.Fatal("external URL was treated as a relative link")
	}
}

func TestBundleLookupUsesNodeIdentityPathAndHeading(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(root, Metadata{Version: "dev", Revision: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Open(generated.Archive)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bundle.Bootstrap(), generated.Bootstrap) {
		t.Fatal("opened Bootstrap differs from generated Bootstrap")
	}

	byNode, err := bundle.ShowNode("core")
	if err != nil {
		t.Fatal(err)
	}
	byIdentity, err := bundle.Show("Knowledge.core.md")
	if err != nil {
		t.Fatal(err)
	}
	byPath, err := bundle.Show("doc/core/README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(byNode, byIdentity) || !bytes.Equal(byNode, byPath) {
		t.Fatal("Node, identity, and path did not resolve to the same document")
	}
	byFoldedIdentity, err := bundle.Show("knowledge.core.md")
	if err != nil || !bytes.Equal(byNode, byFoldedIdentity) {
		t.Fatalf("case-folded identity lookup failed: %v", err)
	}
	if !strings.Contains(string(byNode), "doc/core/core.cookbook.md") {
		t.Fatal("Show did not rewrite included links to repository-root-relative paths")
	}

	section, err := bundle.Show("Knowledge.core.md#responsibility")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(section), "## Responsibility\n") || strings.Contains(string(section), "## Navigate by question") {
		t.Fatalf("unexpected heading section:\n%s", section)
	}
	if _, err := bundle.Show("missing.md"); err == nil {
		t.Fatal("missing Show reference was accepted")
	}

	candidates := bundle.Resolve([]string{"core"})
	if len(candidates) == 0 || candidates[0].Path != "doc/core/README.md" {
		t.Fatalf("Core Resolve candidates = %+v", candidates)
	}
	exact := bundle.Resolve([]string{"Knowledge.core.cookbook.md"})
	if len(exact) == 0 || exact[0].Path != "doc/core/core.cookbook.md" {
		t.Fatalf("exact Resolve candidates = %+v", exact)
	}
	leaving := bundle.Resolve([]string{"leave CTK and restore editor safely"})
	if len(leaving) == 0 || leaving[0].Path != "doc/note/note.leaving-ctk.md" {
		t.Fatalf("natural-language Leaving CTK Resolve candidates = %+v", leaving)
	}
	settingsVariant := bundle.Resolve([]string{"Settings Variant precedence"})
	if len(settingsVariant) == 0 || settingsVariant[0].Path != "doc/note/note.variant.md" {
		t.Fatalf("Settings Variant Resolve candidates = %+v", settingsVariant)
	}
	workspace := bundle.Resolve([]string{"Workspace Dist Archive Workbench"})
	if len(workspace) == 0 || workspace[0].Path != "doc/integration/integration.workspace.md" {
		t.Fatalf("Workspace ownership Resolve candidates = %+v", workspace)
	}
}

func TestResolveDoesNotSearchDocumentBodies(t *testing.T) {
	bundle := &Bundle{
		manifest:  Manifest{Documents: []ManifestDocument{{Path: "doc.md", Identity: "Knowledge.index.md", Title: "Index Title", Headings: []string{"Index Heading"}}}},
		documents: map[string][]byte{"doc.md": []byte("bodyonlyneedle")},
	}
	if candidates := bundle.Resolve([]string{"bodyonlyneedle"}); len(candidates) != 0 {
		t.Fatalf("Resolve unexpectedly searched document bodies: %+v", candidates)
	}
}

func TestSelectHeadingUsesDuplicateMarkdownAnchorSuffix(t *testing.T) {
	content := "# Example\n## Responsibility\nfirst\n## Responsibility\nsecond\n## Next\nthird\n"
	section, err := selectHeading(content, "responsibility-1")
	if err != nil {
		t.Fatal(err)
	}
	if section != "## Responsibility\nsecond\n" {
		t.Fatalf("duplicate heading section = %q", section)
	}
}

func TestOpenAcceptsBundleAppendedToExecutablePrefix(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(root, Metadata{Version: "dev", Revision: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	executable := append([]byte("mock executable prefix\n"), generated.Archive...)
	bundle, err := Open(executable)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Manifest().ContentSHA256 != generated.Manifest.ContentSHA256 {
		t.Fatal("appended Bundle Manifest differs from generated Manifest")
	}
}

func TestExportPublishesVerifiedTreeToAbsentOrEmptyTarget(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(root, Metadata{Version: "dev", Revision: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Open(generated.Archive)
	if err != nil {
		t.Fatal(err)
	}

	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprintf("existing=%v", existing), func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "documentation")
			if existing {
				if err := os.Mkdir(target, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			result, err := bundle.Export(target)
			if err != nil {
				t.Fatal(err)
			}
			if result.Path != target || result.ContentSHA256 != generated.Manifest.ContentSHA256 || result.Documents != len(generated.Manifest.Documents) {
				t.Fatalf("Export result = %+v", result)
			}
			for _, relative := range []string{ManifestPath, BootstrapPath, "doc/note/note.variant.md"} {
				if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(relative))); err != nil {
					t.Fatalf("missing exported file %s: %v", relative, err)
				}
			}
			variant, err := os.ReadFile(filepath.Join(target, "doc", "note", "note.variant.md"))
			if err != nil || !bytes.Contains(variant, []byte("environmental difference")) {
				t.Fatalf("exported documents are not available for full-text search: %v", err)
			}
			if runtime.GOOS != "windows" {
				fileInfo, err := os.Stat(filepath.Join(target, filepath.FromSlash(ManifestPath)))
				if err != nil || fileInfo.Mode().Perm() != 0o644 {
					t.Fatalf("Manifest mode = %v, %v", fileInfo.Mode().Perm(), err)
				}
				directoryInfo, err := os.Stat(filepath.Join(target, "doc"))
				if err != nil || directoryInfo.Mode().Perm() != 0o755 {
					t.Fatalf("directory mode = %v, %v", directoryInfo.Mode().Perm(), err)
				}
			}
		})
	}
}

func TestExportRefusesNonEmptyFileAndSymlinkTargets(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(root, Metadata{Version: "dev", Revision: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Open(generated.Archive)
	if err != nil {
		t.Fatal(err)
	}

	nonEmpty := filepath.Join(t.TempDir(), "non-empty")
	if err := os.Mkdir(nonEmpty, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(nonEmpty, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := bundle.Export(nonEmpty); err == nil {
		t.Fatal("Export accepted a non-empty target")
	}
	if content, err := os.ReadFile(marker); err != nil || string(content) != "keep" {
		t.Fatalf("non-empty target changed: %q, %v", content, err)
	}

	fileTarget := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(fileTarget, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := bundle.Export(fileTarget); err == nil {
		t.Fatal("Export accepted a file target")
	}

	symlinkRoot := t.TempDir()
	symlinkTarget := filepath.Join(symlinkRoot, "link")
	if err := os.Symlink(t.TempDir(), symlinkTarget); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := bundle.Export(symlinkTarget); err == nil {
		t.Fatal("Export accepted a symlink target")
	}
}
