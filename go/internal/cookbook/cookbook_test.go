package cookbook

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestResolveExistingGolangRecipe(t *testing.T) {
	repositoryRoot := filepath.Join("testdata", "cookbook")
	plan, err := (Repository{Root: filepath.Join(repositoryRoot, "ingredient")}).Resolve(filepath.Join(repositoryRoot, "recipe", "vscode-golang.macos.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Name != "vscode-golang" || plan.Platform != "code" {
		t.Fatalf("identity = %s/%s", plan.Name, plan.Platform)
	}
	if !plan.ExtensionMarketplace {
		t.Fatal("extension marketplace should use the Recipe default")
	}
	if plan.ExtensionPool != "reuse" {
		t.Fatalf("extension Pool mode = %q", plan.ExtensionPool)
	}
	if len(plan.Default.Extensions) != 0 {
		t.Fatalf("default extensions = %v", plan.Default.Extensions)
	}
	if plan.Default.Settings["workbench.colorTheme"] != "Solarized Light" {
		t.Fatalf("theme = %#v", plan.Default.Settings["workbench.colorTheme"])
	}
	if len(plan.Profiles) != 5 {
		t.Fatalf("profiles = %d", len(plan.Profiles))
	}
	if !slices.Contains(plan.Profiles[0].Extensions, "golang.go") || !slices.Contains(plan.Profiles[0].Extensions, "openai.chatgpt") {
		t.Fatalf("core extensions = %v", plan.Profiles[0].Extensions)
	}
}

func TestResolveRepresentativeExtensionSetRecipes(t *testing.T) {
	repositoryRoot := filepath.Join("testdata", "cookbook")
	for _, osName := range []string{"macos", "windows"} {
		t.Run(osName, func(t *testing.T) {
			plan, err := (Repository{Root: filepath.Join(repositoryRoot, "ingredient")}).Resolve(filepath.Join(repositoryRoot, "recipe", "vscode-golang."+osName+".yaml"))
			if err != nil {
				t.Fatal(err)
			}
			if plan.OS != osName || len(plan.Default.Extensions) != 0 {
				t.Fatalf("identity/default Extensions = %s/%v", plan.OS, plan.Default.Extensions)
			}
			if plan.Default.Settings["ctk.fixture.extension-set"] != true {
				t.Fatalf("Set member Extension Settings missing: %#v", plan.Default.Settings)
			}
			want := []string{"golang.go", "openai.chatgpt"}
			for _, profile := range plan.Profiles {
				if !slices.Equal(profile.Extensions, want) {
					t.Fatalf("profile %s Extensions = %v, want %v", profile.Name, profile.Extensions, want)
				}
			}
			core := plan.Profiles[0]
			if len(core.ExtensionOrigins["openai.chatgpt"]) != 2 {
				t.Fatalf("direct and Set origins = %#v", core.ExtensionOrigins["openai.chatgpt"])
			}
		})
	}
}

func TestResolveRejectsUnknownExtensionPoolMode(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "recipe.yaml")
	mustWrite(t, path, "name: test\nos: macos\nplatform: code\nconfig:\n  dist-strategy:\n    extension-pool: unknown\n")
	_, err := (Repository{Root: filepath.Join(root, "ingredient")}).Resolve(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported extension-pool") {
		t.Fatalf("error = %v", err)
	}
}

func TestResolveAllFixtureRecipes(t *testing.T) {
	repositoryRoot := filepath.Join("testdata", "cookbook")
	paths, err := filepath.Glob(filepath.Join(repositoryRoot, "recipe", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no fixture Recipes found")
	}
	repository := Repository{Root: filepath.Join(repositoryRoot, "ingredient")}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			plan, err := repository.Resolve(path)
			if err != nil {
				t.Fatal(err)
			}
			if plan.Name == "" || plan.Platform == "" {
				t.Fatalf("incomplete plan: %#v", plan)
			}
		})
	}
}

func TestResolveVariantsInOrderAndJSONC(t *testing.T) {
	root := t.TempDir()
	ingredients := filepath.Join(root, "ingredient")
	recipes := filepath.Join(root, "recipe")
	mustWrite(t, filepath.Join(recipes, "test.yaml"), "name: test\nos: macos\nplatform: code\nruntime: [sample]\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.sample.settings.jsonc"), `{"value":"base","object":{"base":true}}`)
	mustWrite(t, filepath.Join(ingredients, "runtime.sample.macos.settings.json"), `{"value":"os","array":[1]}`)
	mustWrite(t, filepath.Join(ingredients, "runtime.sample.code.settings.json"), `{/*c*/"value":"platform","array":[2,],}`)
	plan, err := (Repository{Root: ingredients}).Resolve(filepath.Join(recipes, "test.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Default.Settings["value"] != "platform" {
		t.Fatalf("value = %#v", plan.Default.Settings["value"])
	}
	if len(plan.Default.Sources) < 3 || plan.Default.Sources[len(plan.Default.Sources)-1].Variant != "code" {
		t.Fatalf("sources = %#v", plan.Default.Sources)
	}
}

func TestResolveAppliesCookbookMergeRules(t *testing.T) {
	root := t.TempDir()
	ingredients := filepath.Join(root, "ingredient")
	recipePath := filepath.Join(root, "recipe", "test.yaml")
	mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\nruntime: [sample]\n")
	mustWrite(t, filepath.Join(ingredients, "os.macos.settings.json"), `{"values":[1]}`)
	mustWrite(t, filepath.Join(ingredients, "runtime.sample.settings.json"), `{"values":[1,2]}`)
	mustWrite(t, filepath.Join(root, "kitchen-notes", "go.merge-rules.yaml"), "format-version: 1\nsettings:\n  union:\n    - [values]\n")
	plan, err := (Repository{Root: ingredients}).Resolve(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	values := plan.Default.Settings["values"].([]any)
	if len(values) != 2 || values[0] != float64(1) || values[1] != float64(2) {
		t.Fatalf("values = %#v", values)
	}
}

func TestResolveAllowsMissingResources(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "recipe.yaml")
	mustWrite(t, path, "name: empty\nos: macos\nplatform: code\nruntime: [absent]\nprofile: [also-absent]\n")
	plan, err := (Repository{Root: filepath.Join(root, "ingredient")}).Resolve(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Default.Settings) != 0 || len(plan.Profiles) != 1 {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestResolveExtensionSetsFromRuntimeAndProfile(t *testing.T) {
	for _, test := range []struct {
		name       string
		selection  string
		layer      string
		ingredient string
	}{
		{name: "runtime", selection: "runtime: [sample]", layer: "runtime", ingredient: "sample"},
		{name: "profile", selection: "profile: [sample]", layer: "profile", ingredient: "sample"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			recipePath := filepath.Join(root, "recipe", "test.yaml")
			mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\n"+test.selection+"\n")
			ingredients := filepath.Join(root, "ingredient")
			mustWrite(t, filepath.Join(ingredients, test.layer+"."+test.ingredient+".extensions"), "z.direct\nset:shared\nset:secondary\na.direct\n")
			setPath := filepath.Join(ingredients, "extension-set.shared.extensions")
			mustWrite(t, setPath, "z.member\na.member\nz.direct\n")
			mustWrite(t, filepath.Join(ingredients, "extension-set.secondary.extensions"), "a.member\n")
			mustWrite(t, filepath.Join(ingredients, "extension.a.member.settings.json"), `{"from-extension-set":true}`)

			plan, err := (Repository{Root: ingredients}).Resolve(recipePath)
			if err != nil {
				t.Fatal(err)
			}
			want := []string{"a.direct", "a.member", "z.direct", "z.member"}
			var got []string
			if test.layer == "runtime" {
				got = plan.Default.Extensions
			} else {
				got = plan.Profiles[0].Extensions
			}
			if !slices.Equal(got, want) {
				t.Fatalf("extensions = %v, want %v", got, want)
			}
			if plan.Default.Settings["from-extension-set"] != true {
				t.Fatalf("expanded Extension Settings were not resolved: %#v", plan.Default.Settings)
			}
			var origins map[string][]Source
			if test.layer == "runtime" {
				origins = plan.Default.ExtensionOrigins
			} else {
				origins = plan.Profiles[0].ExtensionOrigins
			}
			if len(origins["a.member"]) != 2 || len(origins["z.direct"]) != 2 {
				t.Fatalf("Extension origins = %#v", origins)
			}
			if !slices.Contains(plan.Default.Sources, Source{Layer: "extension-set", Ingredient: "shared", Path: setPath}) &&
				(test.layer != "profile" || !slices.Contains(plan.Profiles[0].Sources, Source{Layer: "extension-set", Ingredient: "shared", Path: setPath})) {
				t.Fatalf("Extension Set source not preserved: default=%#v profiles=%#v", plan.Default.Sources, plan.Profiles)
			}
		})
	}
}

func TestResolveExtensionSetLayouts(t *testing.T) {
	for _, relative := range []string{
		"extension-set.shared.extensions",
		filepath.Join("extension-set", "shared.extensions"),
		filepath.Join("extension-set", "shared", "extensions"),
	} {
		t.Run(relative, func(t *testing.T) {
			root := t.TempDir()
			ingredients := filepath.Join(root, "ingredient")
			recipePath := filepath.Join(root, "recipe.yaml")
			mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\nruntime: [sample]\n")
			mustWrite(t, filepath.Join(ingredients, "runtime.sample.extensions"), "set:shared\n")
			mustWrite(t, filepath.Join(ingredients, relative), "member.extension\n")
			plan, err := (Repository{Root: ingredients}).Resolve(recipePath)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(plan.Default.Extensions, []string{"member.extension"}) {
				t.Fatalf("extensions = %v", plan.Default.Extensions)
			}
		})
	}
}

func TestResolveExtensionSetAllowsAbsentAndEmptyResources(t *testing.T) {
	root := t.TempDir()
	ingredients := filepath.Join(root, "ingredient")
	recipePath := filepath.Join(root, "recipe.yaml")
	mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\nruntime: [sample]\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.sample.extensions"), "set:absent\nset:empty\n")
	mustWrite(t, filepath.Join(ingredients, "extension-set.empty.extensions"), "\n")
	plan, err := (Repository{Root: ingredients}).Resolve(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Default.Extensions) != 0 {
		t.Fatalf("extensions = %v", plan.Default.Extensions)
	}
}

func TestResolveExtensionSetRejectsInvalidDeclarations(t *testing.T) {
	for _, declaration := range []string{"set:", "set:-leading", "set:.leading", "set:with space", "set:path/name", `set:path\name`, "set:nested:name"} {
		t.Run(declaration, func(t *testing.T) {
			root := t.TempDir()
			recipePath := filepath.Join(root, "recipe.yaml")
			mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\nruntime: [sample]\n")
			mustWrite(t, filepath.Join(root, "ingredient", "runtime.sample.extensions"), declaration+"\n")
			_, err := (Repository{Root: filepath.Join(root, "ingredient")}).Resolve(recipePath)
			if err == nil || !strings.Contains(err.Error(), "invalid Extension Set declaration") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestResolveExtensionSetRejectsNestedAndAmbiguousResources(t *testing.T) {
	for _, test := range []struct {
		name    string
		prepare func(string)
		want    string
	}{
		{name: "nested", prepare: func(root string) {
			mustWrite(t, filepath.Join(root, "extension-set.shared.extensions"), "set:other\n")
		}, want: "nested Extension Set declaration"},
		{name: "ambiguous", prepare: func(root string) {
			mustWrite(t, filepath.Join(root, "extension-set.shared.extensions"), "one.extension\n")
			mustWrite(t, filepath.Join(root, "extension-set", "shared.extensions"), "two.extension\n")
		}, want: "ambiguous extensions resource for extension-set.shared"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			ingredients := filepath.Join(root, "ingredient")
			recipePath := filepath.Join(root, "recipe.yaml")
			mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\nruntime: [sample]\n")
			mustWrite(t, filepath.Join(ingredients, "runtime.sample.extensions"), "set:shared\n")
			test.prepare(ingredients)
			_, err := (Repository{Root: ingredients}).Resolve(recipePath)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestResolveRejectsAmbiguousLayout(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "recipe.yaml")
	mustWrite(t, path, "name: ambiguous\nos: macos\nplatform: code\nruntime: [sample]\n")
	mustWrite(t, filepath.Join(root, "ingredient", "runtime.sample.extensions"), "one.extension\n")
	mustWrite(t, filepath.Join(root, "ingredient", "runtime", "sample.extensions"), "two.extension\n")
	_, err := (Repository{Root: filepath.Join(root, "ingredient")}).Resolve(path)
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("error = %v", err)
	}
}

func TestResolveRuntimeArtifactsWithoutSemanticDeduplication(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "recipe", "test.yaml")
	ingredients := filepath.Join(root, "ingredient")
	mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\nruntime: [one, two]\nconfig:\n  dist-strategy:\n    default-profile:\n      mcp: runtime\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.one.keybindings.jsonc"), `[{"key":"same","command":"one"},]`)
	mustWrite(t, filepath.Join(ingredients, "runtime.two.keybindings.json"), `[{"key":"same","command":"two"}]`)
	mustWrite(t, filepath.Join(ingredients, "runtime.one.tasks.json"), `{"version":"2.0.0","tasks":[{"label":"same","command":"one"}],"inputs":[{"id":"same"}]}`)
	mustWrite(t, filepath.Join(ingredients, "runtime.two.tasks.json"), `{"version":"2.0.0","tasks":[{"label":"same","command":"two"}],"inputs":[{"id":"same"}]}`)
	mustWrite(t, filepath.Join(ingredients, "runtime.one.mcp.json"), `{"servers":{"sample":{"url":"old"}},"inputs":[{"id":"same"}]}`)
	mustWrite(t, filepath.Join(ingredients, "runtime.two.mcp.json"), `{"servers":{"sample":{"url":"new"}},"inputs":[{"id":"same"}]}`)
	mustWrite(t, filepath.Join(ingredients, "runtime.one.snippets.go.json"), `{"Same":{"prefix":"old","body":["old"]},"OnlyOne":{"body":["one"]}}`)
	mustWrite(t, filepath.Join(ingredients, "runtime", "two", "snippets", "go.json"), `{"Same":{"prefix":"new","body":["new"]}}`)
	plan, err := (Repository{Root: ingredients}).Resolve(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Default.Keybindings) != 2 {
		t.Fatalf("keybindings = %#v", plan.Default.Keybindings)
	}
	if len(plan.Default.Tasks["tasks"].([]any)) != 2 || len(plan.Default.Tasks["inputs"].([]any)) != 2 {
		t.Fatalf("tasks = %#v", plan.Default.Tasks)
	}
	server := plan.Default.MCP["servers"].(map[string]any)["sample"].(map[string]any)
	if server["url"] != "new" || len(plan.Default.MCP["inputs"].([]any)) != 2 {
		t.Fatalf("mcp = %#v", plan.Default.MCP)
	}
	if len(plan.Default.Snippets["go.json"]) != 2 || plan.Default.Snippets["go.json"]["Same"].(map[string]any)["prefix"] != "new" {
		t.Fatalf("snippets = %#v", plan.Default.Snippets)
	}
}

func TestMCPIsUnmanagedByDefault(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "recipe.yaml")
	mustWrite(t, recipePath, "name: test\nos: macos\nplatform: code\nruntime: [sample]\n")
	mustWrite(t, filepath.Join(root, "ingredient", "runtime.sample.mcp.json"), `{"servers":{"secret":{"url":"https://example.invalid"}}}`)
	plan, err := (Repository{Root: filepath.Join(root, "ingredient")}).Resolve(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Default.MCP != nil {
		t.Fatalf("default MCP must be unmanaged: %#v", plan.Default.MCP)
	}
	if !plan.Default.Inheritance.Unmanaged["mcp"] {
		t.Fatalf("default MCP unmanaged strategy was not carried into Plan: %#v", plan.Default.Inheritance)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
