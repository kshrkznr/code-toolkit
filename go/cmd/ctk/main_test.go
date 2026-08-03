package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"code-toolkit/internal/recipe"
)

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
	kind, path, err := selectRuntimeSource(root, cookbook, "sample", nil)
	if err != nil || kind != "recipe" || filepath.Base(path) != "sample.yaml" {
		t.Fatalf("Recipe source = %s %s %v", kind, path, err)
	}
	kind, path, err = selectRuntimeSource(root, cookbook, "saved", nil)
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
}
