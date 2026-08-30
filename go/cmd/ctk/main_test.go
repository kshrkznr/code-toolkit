package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/docbundle"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
)

func TestRunDocsNavigatesPackagedBundle(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := docbundle.Generate(repositoryRoot, docbundle.Metadata{Version: "dev", Revision: "test-revision"})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := docbundle.Open(generated.Archive)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name     string
		args     []string
		contains []string
	}{
		{name: "help", args: []string{"--help"}, contains: []string{"--source <repository>", "ctk docs status", "ctk docs resolve", "ctk docs toc", "--depth <N|A..B>"}},
		{name: "status", args: []string{"status"}, contains: []string{"source: packaged", "source-revision: test-revision", "definition-sha256: ", "content-sha256: ", "repository: https://github.com/kshrkznr/code-toolkit"}},
		{name: "nodes", args: []string{"nodes"}, contains: []string{"core\tdoc/core/README.md"}},
		{name: "core", args: []string{"core"}, contains: []string{"# Concept Domain: Core", "doc/core/core.cookbook.md"}},
		{name: "resolve", args: []string{"resolve", "Settings Variant precedence"}, contains: []string{"IDENTITY\tPATH\tTITLE\tMATCHED", "Knowledge.note.variant.md\tdoc/note/note.variant.md"}},
		{name: "toc", args: []string{"toc", "knowledge.core.cookbook.md"}, contains: []string{"- [Concept API: Cookbook](doc/core/core.cookbook.md#concept-api-cookbook)", "  - [Responsibility](doc/core/core.cookbook.md#responsibility)"}},
		{name: "show folded identity and duplicate heading", args: []string{"show", "knowledge.core.cookbook.md#responsibility-1"}, contains: []string{"## Responsibility", "Ingredients provide reusable building blocks"}},
		{name: "show depth", args: []string{"show", "knowledge.core.cookbook.md#responsibility-1", "--depth", "0"}, contains: []string{"## Responsibility", "Ingredients provide reusable building blocks"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := runDocs(&output, bundle, test.args); err != nil {
				t.Fatal(err)
			}
			for _, expected := range test.contains {
				if !strings.Contains(output.String(), expected) {
					t.Fatalf("output does not contain %q:\n%s", expected, output.String())
				}
			}
		})
	}

	exportTarget := filepath.Join(t.TempDir(), "documentation")
	var exportOutput bytes.Buffer
	if err := runDocs(&exportOutput, bundle, []string{"export", exportTarget}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(exportOutput.String(), exportTarget+"\ncontent-sha256: ") {
		t.Fatalf("Export output = %q", exportOutput.String())
	}
	if _, err := os.Stat(filepath.Join(exportTarget, filepath.FromSlash(docbundle.ManifestPath))); err != nil {
		t.Fatalf("Export Manifest missing: %v", err)
	}
	for _, args := range [][]string{
		{"toc", "Knowledge.core.md#responsibility"},
		{"show", "Knowledge.core.md", "--depth", "0"},
		{"show", "Knowledge.core.md#responsibility", "--depth", "1..2"},
	} {
		if err := runDocs(io.Discard, bundle, args); err == nil {
			t.Fatalf("invalid docs arguments were accepted: %v", args)
		}
	}
}

func TestParseDocsDepth(t *testing.T) {
	for _, test := range []struct {
		value            string
		minimum, maximum int
	}{
		{value: "-1", minimum: -1, maximum: 0},
		{value: "0", minimum: 0, maximum: 0},
		{value: "2", minimum: 0, maximum: 2},
		{value: "-1..2", minimum: -1, maximum: 2},
	} {
		minimum, maximum, err := parseDocsDepth(test.value)
		if err != nil || minimum != test.minimum || maximum != test.maximum {
			t.Fatalf("parseDocsDepth(%q) = %d, %d, %v", test.value, minimum, maximum, err)
		}
	}
	for _, value := range []string{"", "one", "1..2", "-2..-1", "2..-1", "-1..0..2"} {
		if _, _, err := parseDocsDepth(value); err == nil {
			t.Fatalf("parseDocsDepth(%q) unexpectedly succeeded", value)
		}
	}
}

func TestParseDocsSourceRequiresLeadingExplicitSelection(t *testing.T) {
	for _, test := range []struct {
		args      []string
		source    string
		remaining []string
	}{
		{args: nil, remaining: nil},
		{args: []string{"status"}, remaining: []string{"status"}},
		{args: []string{"--source", "../clone", "show", "doc.md"}, source: "../clone", remaining: []string{"show", "doc.md"}},
		{args: []string{"--source=/clone", "toc", "doc.md"}, source: "/clone", remaining: []string{"toc", "doc.md"}},
	} {
		source, remaining, err := parseDocsSource(test.args)
		if err != nil || source != test.source || !slices.Equal(remaining, test.remaining) {
			t.Fatalf("parseDocsSource(%v) = %q, %v, %v", test.args, source, remaining, err)
		}
	}
	for _, args := range [][]string{{"--source"}, {"--source="}, {"show", "doc.md", "--source", "/clone"}, {"--source", "/one", "--source", "/two"}} {
		if _, _, err := parseDocsSource(args); err == nil {
			t.Fatalf("parseDocsSource(%v) unexpectedly succeeded", args)
		}
	}
}

func TestWriteDocsSourceStatusKeepsComparisonsIndependent(t *testing.T) {
	status := docbundle.SourceStatus{
		Kind:                     "local",
		Path:                     "/clone",
		Version:                  "v1.0.0",
		Revision:                 "local-revision",
		Repository:               "https://github.com/example/ctk",
		DefinitionSHA256:         "local-definition",
		ContentSHA256:            "local-content",
		ComparisonContentSHA256:  "comparison-content",
		PackagedRevision:         "packaged-revision",
		PackagedDefinitionSHA256: "packaged-definition",
		PackagedContentSHA256:    "packaged-content",
		RevisionMatch:            docbundle.Mismatch,
		DefinitionMatch:          docbundle.Match,
		ContentMatch:             docbundle.Mismatch,
		SelectedPathDirty:        docbundle.Dirty,
		SelectedDirtyPaths:       []string{"doc/one.md", "doc/two.md"},
		RepositoryDirty:          docbundle.Dirty,
	}
	var output bytes.Buffer
	if err := writeDocsSourceStatus(&output, status); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"source: local\n",
		"revision-match: mismatch\n",
		"definition-match: match\n",
		"comparison-content-sha256: comparison-content\n",
		"content-match: mismatch\n",
		"selected-path-dirty: dirty\n",
		"selected-dirty-path: doc/one.md\n",
		"repository-dirty: dirty\n",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("status does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestMaskHomePathHidesUserSpecificPrefix(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("home directory unavailable: %v", err)
	}
	inside := filepath.Join(home, "local", "ctk")
	if got := maskHomePath(inside); got != filepath.Join("<home>", "local", "ctk") {
		t.Fatalf("masked path = %q", got)
	}
	outside := filepath.Join(filepath.Dir(home), "shared", "ctk")
	if got := maskHomePath(outside); got != outside {
		t.Fatalf("outside path changed = %q", got)
	}
}

func TestWriteCurrentContextKeepsHelpDiagnosticsNonFatal(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("home directory unavailable: %v", err)
	}
	var output bytes.Buffer
	writeCurrentContext(&output, currentContext{
		workspacePath:       filepath.Join(home, "local", "my-ctk"),
		workspaceSource:     "CTK_HOME",
		workspaceDiagnostic: "parse " + filepath.Join(home, "local", "my-ctk", ".config", "workspace.yaml"),
		documentation:       "packaged v0.5.0 @ abc123",
	})
	for _, expected := range []string{
		"Current context:\n",
		"Workspace:      " + filepath.Join("~", "local", "my-ctk"),
		"source: CTK_HOME",
		"diagnostic: parse " + filepath.Join("~", "local", "my-ctk", ".config", "workspace.yaml"),
		"Documentation:  packaged v0.5.0 @ abc123",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("context does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestRunHelpSucceedsWithInvalidWorkspace(t *testing.T) {
	t.Setenv("CTK_HOME", filepath.Join(t.TempDir(), "missing"))
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = oldStdout })

	if err := run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Usage: ctk <command>", "Current context:", "Workspace:      unavailable", "source: CTK_HOME", "diagnostic: CTK_HOME is not a CTK workspace"} {
		if !strings.Contains(string(output), expected) {
			t.Fatalf("help does not contain %q:\n%s", expected, output)
		}
	}
}

func TestDisplayHomePathDoesNotRewriteSimilarPrefix(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("home directory unavailable: %v", err)
	}
	inside := filepath.Join(home, "local", "ctk")
	if got := displayHomePath(inside); got != filepath.Join("~", "local", "ctk") {
		t.Fatalf("display path = %q", got)
	}
	similar := home + "-shared" + string(filepath.Separator) + "ctk"
	if got := displayHomePath(similar); got != similar {
		t.Fatalf("similar prefix changed = %q", got)
	}
}

func TestDetectViewSourceFromContentAndKnownNames(t *testing.T) {
	root := t.TempDir()
	distDir := filepath.Join(root, "dist")
	cookbookDir := filepath.Join(root, "cookbook")
	distPath := filepath.Join(distDir, "demo")
	recipePath := filepath.Join(cookbookDir, "recipe", "sample.yaml")
	ingredientPath := filepath.Join(cookbookDir, "ingredient", "runtime", "golang")
	for _, path := range []string{filepath.Join(distPath, ".meta"), filepath.Dir(recipePath), ingredientPath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(distPath, ".meta", "recipe.yaml"), []byte("name: demo\n"), 0644)
	os.WriteFile(recipePath, []byte("name: sample\n"), 0644)
	for _, test := range []struct{ input, kind string }{{distPath, "dist"}, {recipePath, "recipe"}, {ingredientPath, "ingredient"}, {"demo", "dist"}, {"sample", "recipe"}, {"runtime.golang", "ingredient"}} {
		kind, _, err := detectViewSource(distDir, cookbookDir, "", test.input)
		if err != nil || kind != test.kind {
			t.Fatalf("detect %s = %s, %v", test.input, kind, err)
		}
	}
}

func TestDetectViewSourceRejectsAmbiguousKnownName(t *testing.T) {
	root := t.TempDir()
	distDir := filepath.Join(root, "dist")
	cookbookDir := filepath.Join(root, "cookbook")
	os.MkdirAll(filepath.Join(distDir, "same"), 0755)
	os.MkdirAll(filepath.Join(cookbookDir, "recipe"), 0755)
	os.WriteFile(filepath.Join(cookbookDir, "recipe", "same.yaml"), []byte("name: same\n"), 0644)
	if _, _, err := detectViewSource(distDir, cookbookDir, "", "same"); err == nil {
		t.Fatal("expected explicit View requirement")
	}
}

func TestParseSyncArgsAllowsInteractiveMissingSides(t *testing.T) {
	for _, args := range [][]string{nil, {"left"}, {"left", "right"}} {
		_, inputs, err := parseSyncArgs(args)
		if err != nil || len(inputs) != len(args) {
			t.Fatalf("parseSyncArgs(%v) = %v, %v", args, inputs, err)
		}
	}
	if _, _, err := parseSyncArgs([]string{"a", "b", "c"}); err == nil {
		t.Fatal("expected too many Sync sources error")
	}
}

func TestParseLaunchArgsSeparatesDistributionAndPlatformTargets(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		dist    string
		forward []string
	}{
		{name: "interactive", args: nil},
		{name: "explicit distribution", args: []string{"vscode-default"}, dist: "vscode-default"},
		{name: "explicit target", args: []string{"vscode-default", "."}, dist: "vscode-default", forward: []string{"."}},
		{name: "empty distribution selects", args: []string{"", "."}, forward: []string{"."}},
		{name: "separator selects", args: []string{"--", "."}, forward: []string{"."}},
		{name: "separator after distribution", args: []string{"vscode-default", "--", "file.go", "directory"}, dist: "vscode-default", forward: []string{"file.go", "directory"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dist, forward, err := parseLaunchArgs(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if dist != test.dist || !slices.Equal(forward, test.forward) {
				t.Fatalf("parseLaunchArgs(%v) = %q, %v; want %q, %v", test.args, dist, forward, test.dist, test.forward)
			}
		})
	}
}

func TestBuildAndApplyForceParsing(t *testing.T) {
	build, err := parseBuildArgs([]string{"sample", "--force", "--keep-staging"})
	if err != nil || build.recipe != "sample" || !build.force || !build.keepStaging {
		t.Fatalf("build=%+v err=%v", build, err)
	}
	apply, err := parseApplyArgs([]string{"sample", "dist-name", "--force"})
	if err != nil || apply.source != "sample" || apply.dist != "dist-name" || !apply.force {
		t.Fatalf("apply=%+v err=%v", apply, err)
	}
}

func TestDeactivateForceEmptyParsing(t *testing.T) {
	platformName, force, forceEmpty, err := parseDeactivateArgs([]string{"code", "--force-empty"})
	if err != nil || platformName != "code" || force || !forceEmpty {
		t.Fatalf("platform=%q force=%v empty=%v err=%v", platformName, force, forceEmpty, err)
	}
	if _, _, _, err := parseDeactivateArgs([]string{"code", "--force", "--force-empty"}); err == nil {
		t.Fatal("combined force modes must be rejected")
	}
}

func TestPrintActivePlatforms(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"vscode-default", "kiro-default"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(root, "kiro-default"), filepath.Join(root, "current.kiro")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "vscode-default"), filepath.Join(root, "current.code")); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := printActivePlatforms(&output, root); err != nil {
		t.Fatal(err)
	}
	want := "[active] code: vscode-default\n" +
		"[active] kiro: kiro-default\n" +
		"[hint] deactivate remaining Platforms individually:\n" +
		"  ctk deactivate code\n" +
		"  ctk deactivate kiro\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestPrintActivePlatformsWhenNone(t *testing.T) {
	var output bytes.Buffer
	if err := printActivePlatforms(&output, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if want := "[active] none\n"; output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestParseOpenWorkbenchArgs(t *testing.T) {
	options, err := parseOpenWorkbenchArgs([]string{"inspect", "dist.demo", "--editor", "code"})
	if err != nil {
		t.Fatal(err)
	}
	if options.kind != "inspect" || options.viewpoint != "dist.demo" || options.editor != "code" {
		t.Fatalf("options = %+v", options)
	}
	if _, err := parseOpenWorkbenchArgs([]string{"draft", "unexpected"}); err == nil {
		t.Fatal("draft viewpoint error = nil")
	}
}

func TestInspectWorkbenchesAreDirectoriesOnlyAndSorted(t *testing.T) {
	root := t.TempDir()
	inspect := filepath.Join(root, "inspect")
	if err := os.MkdirAll(filepath.Join(inspect, "sync.z"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(inspect, "dist.a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(inspect, ".staging"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inspect, "summary.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := inspectWorkbenches(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dist.a", "sync.z"}
	if !slices.Equal(got, want) {
		t.Fatalf("inspectWorkbenches() = %v, want %v", got, want)
	}
}

func TestIngredientLayersAreUniqueAndSorted(t *testing.T) {
	layers := ingredientLayers([]string{"runtime.go", "profile.work", "runtime.java"})
	if len(layers) != 2 || layers[0] != "profile" || layers[1] != "runtime" {
		t.Fatalf("layers = %v", layers)
	}
}

func TestIngredientCandidatesIncludeUnreferencedPhysicalResources(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "recipe"), 0755)
	os.MkdirAll(filepath.Join(root, "ingredient", "custom", "thing"), 0755)
	os.WriteFile(filepath.Join(root, "recipe", "base.yaml"), []byte("name: base\nos: macos\nplatform: code\n"), 0644)
	os.WriteFile(filepath.Join(root, "ingredient", "custom", "thing", "extensions"), []byte("x\n"), 0644)
	values, err := ingredientCandidates(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, value := range values {
		if value == "custom.thing" {
			found = true
		}
	}
	if !found {
		t.Fatalf("candidates = %v", values)
	}
}

func TestSelectRuntimeSourceDetectsRecipeAndArchive(t *testing.T) {
	root := t.TempDir()
	cookbook := filepath.Join(root, "cookbook")
	os.MkdirAll(filepath.Join(cookbook, "recipe"), 0755)
	os.MkdirAll(filepath.Join(root, "archive", "saved"), 0755)
	os.WriteFile(filepath.Join(cookbook, "recipe", "sample.yaml"), []byte("name: sample\n"), 0644)
	os.WriteFile(filepath.Join(root, "archive", "saved", "manifest.json"), []byte("{}"), 0644)
	kind, path, err := selectRuntimeSource(filepath.Join(root, "archive"), cookbook, "sample", nil)
	if err != nil || kind != "recipe" || filepath.Base(path) != "sample.yaml" {
		t.Fatalf("Recipe source = %s %s %v", kind, path, err)
	}
	kind, path, err = selectRuntimeSource(filepath.Join(root, "archive"), cookbook, "saved", nil)
	if err != nil || kind != "archive" || filepath.Base(path) != "saved" {
		t.Fatalf("Archive source = %s %s %v", kind, path, err)
	}
}

func TestSelectIdentityDistributionUsesRecipeMetadata(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "custom-name")
	for _, directory := range []string{filepath.Join(path, ".meta"), filepath.Join(path, ".data"), filepath.Join(path, ".ext")} {
		os.MkdirAll(directory, 0755)
	}
	os.WriteFile(filepath.Join(path, ".meta", "recipe.yaml"), []byte("name: sample\nos: macos\nplatform: code\n"), 0644)
	name, err := selectIdentityDistribution(root, recipe.Recipe{Name: "sample", OS: "macos", Platform: "code"}, nil)
	if err != nil || name != "custom-name" {
		t.Fatalf("identity = %s, %v", name, err)
	}
}

func TestFindProjectRootResolution(t *testing.T) {
	workspace := t.TempDir()
	for _, directory := range []string{filepath.Join(workspace, "cookbook", "recipe"), filepath.Join(workspace, "cookbook", "ingredient"), filepath.Join(workspace, "dist", "sample")} {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}

	fromWorkingTree, err := findProjectRoot("", filepath.Join(workspace, "dist", "sample"), filepath.Join(t.TempDir(), "bin", "ctk"))
	if err != nil || fromWorkingTree != workspace {
		t.Fatalf("working tree root = %q, %v", fromWorkingTree, err)
	}

	fromConfigured, err := findProjectRoot(workspace, t.TempDir(), filepath.Join(t.TempDir(), "bin", "ctk"))
	if err != nil || fromConfigured != workspace {
		t.Fatalf("configured root = %q, %v", fromConfigured, err)
	}

	fromExecutable, err := findProjectRoot("", t.TempDir(), filepath.Join(workspace, "bin", "ctk"))
	if err != nil || fromExecutable != workspace {
		t.Fatalf("executable root = %q, %v", fromExecutable, err)
	}

	_, source, err := findProjectRootWithSource(workspace, t.TempDir(), filepath.Join(t.TempDir(), "bin", "ctk"))
	if err != nil || source != "CTK_HOME" {
		t.Fatalf("configured source = %q, %v", source, err)
	}
	_, source, err = findProjectRootWithSource("", filepath.Join(workspace, "dist", "sample"), filepath.Join(t.TempDir(), "bin", "ctk"))
	if err != nil || source != "current directory" {
		t.Fatalf("working-directory source = %q, %v", source, err)
	}
	_, source, err = findProjectRootWithSource("", t.TempDir(), filepath.Join(workspace, "bin", "ctk"))
	if err != nil || source != "executable-relative" {
		t.Fatalf("executable source = %q, %v", source, err)
	}
}

func TestFindProjectRootAcceptsWorkspaceConfigurationMarker(t *testing.T) {
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, ".config", "workspace.yaml"), []byte("paths: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := findProjectRoot("", filepath.Join(workspace, "nested"), filepath.Join(t.TempDir(), "bin", "ctk"))
	if err != nil || got != workspace {
		t.Fatalf("configured Workspace root = %q, %v", got, err)
	}
}

func TestRunListUsesConfiguredCookbookAndDist(t *testing.T) {
	workspace := t.TempDir()
	source := t.TempDir()
	dist := t.TempDir()
	for _, path := range []string{filepath.Join(workspace, ".config"), filepath.Join(source, "recipe"), filepath.Join(source, "ingredient"), filepath.Join(dist, "external-dist")} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	config := "paths:\n  cookbook-source: " + source + "\n  dist: " + dist + "\n"
	if err := os.WriteFile(filepath.Join(workspace, ".config", "workspace.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CTK_HOME", workspace)

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = oldStdout })
	if err := run([]string{"list"}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(output) != "external-dist\n" {
		t.Fatalf("list output = %q", output)
	}
}
