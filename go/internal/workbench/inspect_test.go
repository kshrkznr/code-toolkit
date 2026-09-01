package workbench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecipeAndIngredientViewUseExistingResolvers(t *testing.T) {
	workspace := t.TempDir()
	root := filepath.Join(workspace, "cookbook")
	recipePath := filepath.Join(root, "recipe", "sample.yaml")
	writeInspectTest(t, recipePath, "name: sample\nos: macos\nplatform: code\nruntime: [sample]\n")
	writeInspectTest(t, filepath.Join(root, "ingredient", "runtime.sample.settings.json"), `{"editor.fontSize":14}`)
	writeInspectTest(t, filepath.Join(root, "ingredient", "runtime.sample.macos.settings.jsonc"), `{/* variant */"editor.fontSize":16}`)
	writeInspectTest(t, filepath.Join(root, "ingredient", "runtime.sample.extensions"), "one.extension\nset:shared\n")
	writeInspectTest(t, filepath.Join(root, "ingredient", "extension-set.shared.extensions"), "set.extension\n")
	writeInspectTest(t, filepath.Join(root, "ingredient", "profile.work.extensions"), "profile.extension\n")
	service := Service{CookbookRoot: root}
	source, err := service.RecipeSource(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	recipeResult, err := service.GenerateRecipeView(source, "replace")
	if err != nil {
		t.Fatal(err)
	}
	settingsData, err := os.ReadFile(filepath.Join(recipeResult.Path, "settings.draft.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(settingsData), `+ ["editor.fontSize"]=16`) {
		t.Fatalf("Recipe View = %s", settingsData)
	}
	summaryData, err := os.ReadFile(filepath.Join(recipeResult.Path, "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Result", "## Profiles", "## Extensions Used by Recipe", "one.extension", "set.extension", "## Extension Declaration Sources", "| `one.extension` | default | `runtime.sample.extensions` |", "| `set.extension` | default | `extension-set.shared.extensions` |", "## Resolved Ingredient Resources", "runtime.sample.macos.settings.jsonc", "extension-set.shared.extensions", "## Resolution", "Recipe source: `$CTK_HOME/cookbook/recipe/sample.yaml`", "Generated:"} {
		if !strings.Contains(string(summaryData), want) {
			t.Fatalf("Recipe summary missing %q: %s", want, summaryData)
		}
	}
	if strings.Count(string(summaryData), "one.extension") != 2 {
		t.Fatalf("Recipe summary must list the Extension once and its provenance once: %s", summaryData)
	}
	ingredientResult, err := service.GenerateIngredientView("runtime.sample", "replace")
	if err != nil {
		t.Fatal(err)
	}
	ingredientData, err := os.ReadFile(filepath.Join(ingredientResult.Path, "settings.draft.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"runtime.sample.settings.json", "runtime.sample.macos.settings.jsonc"} {
		if !strings.Contains(string(ingredientData), want) {
			t.Fatalf("Ingredient View missing %s: %s", want, ingredientData)
		}
	}
	layerResult, err := service.GenerateIngredientView("runtime", "replace")
	if err != nil {
		t.Fatal(err)
	}
	layerExtensions, err := os.ReadFile(filepath.Join(layerResult.Path, "extensions.draft.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(layerExtensions), "profile.extension") {
		t.Fatalf("layer View crossed Resource layer: %s", layerExtensions)
	}
	allResult, err := service.GenerateIngredientView("all", "replace")
	if err != nil {
		t.Fatal(err)
	}
	allExtensions, err := os.ReadFile(filepath.Join(allResult.Path, "extensions.draft.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"one.extension", "set.extension", "profile.extension"} {
		if !strings.Contains(string(allExtensions), want) {
			t.Fatalf("all View missing %s: %s", want, allExtensions)
		}
	}
	allSummary, err := os.ReadFile(filepath.Join(allResult.Path, "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Resolved Ingredient Resources", "runtime.sample.settings.json", "profile.work.extensions", "## Resolved Ingredient", "runtime.sample", "profile.work", "Recipe authoring hints", "## Resolution", "Query scope: `all`", "Ingredient root: `$CTK_HOME/cookbook/ingredient`", "raw Resource Inventory", "Generated:"} {
		if !strings.Contains(string(allSummary), want) {
			t.Fatalf("Ingredient summary missing %q: %s", want, allSummary)
		}
	}
}

func TestDisplayPathPreservesSourcesOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	service := Service{CookbookRoot: filepath.Join(workspace, "cookbook")}
	outside := filepath.Join(filepath.Dir(workspace), "outside", "recipe.yaml")
	if got := service.displayPath(outside); got != outside {
		t.Fatalf("displayPath outside Workspace = %q, want %q", got, outside)
	}
}

func TestRepresentativeExtensionSetRecipeViews(t *testing.T) {
	cookbookRoot := filepath.Join("..", "cookbook", "testdata", "cookbook")
	for _, osName := range []string{"macos", "windows"} {
		t.Run(osName, func(t *testing.T) {
			workbench := t.TempDir()
			recipePath := filepath.Join(cookbookRoot, "recipe", "vscode-golang."+osName+".yaml")
			service := Service{CookbookRoot: cookbookRoot, WorkbenchRoot: workbench}
			source, err := service.RecipeSource(recipePath)
			if err != nil {
				t.Fatal(err)
			}
			result, err := service.GenerateRecipeView(source, "replace")
			if err != nil {
				t.Fatal(err)
			}
			summary, err := os.ReadFile(filepath.Join(result.Path, "summary.md"))
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"OS: `" + osName + "`", "openai.chatgpt", "extension-set.editor-core.extensions", "profile.core.extensions", "runtime.golang.extensions"} {
				if !strings.Contains(string(summary), want) {
					t.Fatalf("Recipe View missing %q: %s", want, summary)
				}
			}
		})
	}
}

func TestRecipeViewReadsExternalCookbookAndWritesWorkspaceInspect(t *testing.T) {
	workspace := t.TempDir()
	source := t.TempDir()
	workbench := filepath.Join(workspace, "cookbook")
	recipePath := filepath.Join(source, "recipe", "sample.yaml")
	writeInspectTest(t, recipePath, "name: sample\nos: macos\nplatform: code\nruntime: [sample]\n")
	writeInspectTest(t, filepath.Join(source, "ingredient", "runtime.sample.extensions"), "one.extension\n")
	service := Service{WorkspaceRoot: workspace, CookbookRoot: source, WorkbenchRoot: workbench}

	completed, err := service.RecipeSource(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.GenerateRecipeView(completed, "replace")
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != filepath.Join(workbench, "inspect", "recipe.sample") {
		t.Fatalf("Recipe View path = %s", result.Path)
	}
	if _, err := os.Stat(filepath.Join(source, "inspect")); !os.IsNotExist(err) {
		t.Fatalf("Inspect leaked into Cookbook Source: %v", err)
	}
	summary, err := os.ReadFile(filepath.Join(result.Path, "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(summary), "Recipe source: `"+recipePath+"`") {
		t.Fatalf("external Recipe provenance = %s", summary)
	}
}

func TestSyncComparesRecipeCompletedStates(t *testing.T) {
	root := t.TempDir()
	leftPath := filepath.Join(root, "left.yaml")
	rightPath := filepath.Join(root, "right.yaml")
	writeInspectTest(t, leftPath, "name: left\nos: macos\nplatform: code\nruntime: [left]\n")
	writeInspectTest(t, rightPath, "name: right\nos: macos\nplatform: code\nruntime: [right]\n")
	writeInspectTest(t, filepath.Join(root, "ingredient", "runtime.left.settings.json"), `{"value":"left"}`)
	writeInspectTest(t, filepath.Join(root, "ingredient", "runtime.right.settings.json"), `{"value":"right"}`)
	service := Service{CookbookRoot: root}
	left, err := service.RecipeSource(leftPath)
	if err != nil {
		t.Fatal(err)
	}
	right, err := service.RecipeSource(rightPath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.GenerateSync(left, right, "replace")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(result.Path, "settings.draft.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, `- ["value"]="left"`) || !strings.Contains(text, `+ ["value"]="right"`) {
		t.Fatalf("Sync = %s", text)
	}
}

func writeInspectTest(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
