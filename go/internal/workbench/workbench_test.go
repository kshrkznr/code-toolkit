package workbench

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

func TestGenerateFreezeDraftFromReusedLock(t *testing.T) {
	root := t.TempDir()
	cookbookRoot := filepath.Join(root, "cookbook")
	distPath := filepath.Join(root, "dist", "demo")
	for _, directory := range []string{
		filepath.Join(cookbookRoot, "recipe"), filepath.Join(cookbookRoot, "ingredient"),
		filepath.Join(distPath, ".meta"), filepath.Join(distPath, ".data"), filepath.Join(distPath, ".ext"),
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	recipeData := []byte("name: demo\nos: macos\nplatform: code\nruntime:\n  - golang\nprofile:\n  - work\nconfig:\n  dist-strategy:\n    lock-mode: reuse\n")
	recipePath := filepath.Join(distPath, ".meta", "recipe.yaml")
	if err := os.WriteFile(recipePath, recipeData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cookbookRoot, "recipe", "demo.macos.yaml"), recipeData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cookbookRoot, "recipe", "demo.windows.yaml"), []byte(strings.Replace(string(recipeData), "os: macos", "os: windows", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cookbookRoot, "ingredient", "runtime.golang.settings.json"), []byte("{\"editor.fontSize\": 14}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cookbookRoot, "ingredient", "runtime.golang.extensions"), []byte("set:golang\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cookbookRoot, "ingredient", "extension-set.golang.extensions"), []byte("golang.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cookbookRoot, "ingredient", "extension-set.golang.settings.json"), []byte("{\"set.enabled\":true}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := (cookbook.Repository{Root: filepath.Join(cookbookRoot, "ingredient")}).Resolve(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := runtimelock.Snapshot{
		FormatVersion: runtimelock.FormatVersion, RecipeName: "demo", Platform: "code",
		ObservedAt: time.Date(2026, 7, 23, 1, 2, 3, 0, time.UTC),
		Default:    runtimelock.ScopeSnapshot{Settings: settings.Document{"editor.fontSize": float64(16), "set.enabled": true}, Extensions: []runtimeio.Extension{{ID: "golang.go", Version: "1.2.3"}}},
		Profiles:   []runtimelock.ScopeSnapshot{{Name: "work", Settings: settings.Document{}, Extensions: []runtimeio.Extension{{ID: "golang.go", Version: "1.2.3"}}, Inheritance: cookbook.Inheritance{Settings: true, Keybindings: true, Tasks: true, MCP: true, Snippets: true}}},
	}
	changedSnapshot := snapshot
	changedSnapshot.Default.Extensions = append(append([]runtimeio.Extension(nil), snapshot.Default.Extensions...), runtimeio.Extension{ID: "manual.extension", Version: "1.0.0"})
	extensionsDraft, extensionResult := renderExtensions(&plan, changedSnapshot)
	if extensionResult.Status != "DIFFERENT" || !strings.Contains(extensionsDraft, "### runtime.draft.extensions") || strings.Contains(extensionsDraft, "extension-set.golang.extensions") {
		t.Fatalf("Freeze must keep concrete draft targets without reverse Set inference:\n%s", extensionsDraft)
	}
	if err := (runtimelock.Store{}).Seal(distPath, recipePath, snapshot, plan); err != nil {
		t.Fatal(err)
	}
	dist, err := distribution.Load(filepath.Join(root, "dist"), "demo")
	if err != nil {
		t.Fatal(err)
	}
	service := Service{CookbookRoot: cookbookRoot, Locks: runtimelock.Store{}, Now: func() time.Time { return time.Date(2026, 7, 23, 2, 0, 0, 0, time.UTC) }}
	result, err := service.Generate(context.Background(), FreezeDraft, dist, "abort")
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != filepath.Join(cookbookRoot, "draft") {
		t.Fatalf("Path = %s", result.Path)
	}

	settingsDraft := readTestFile(t, filepath.Join(result.Path, "settings.draft.md"))
	for _, expected := range []string{"## Inventory: Used by Recipe", "## Inventory: Available but Unused", "## Difference", `- ["editor.fontSize"]=14`, `+ ["editor.fontSize"]=16`, "### runtime.draft.settings.json", "### runtime.golang.settings.json", "### extension-set.golang.settings.json"} {
		if !strings.Contains(settingsDraft, expected) {
			t.Fatalf("settings Draft missing %q:\n%s", expected, settingsDraft)
		}
	}
	if strings.Contains(settingsDraft, "profile.work.draft.settings.json") {
		t.Fatalf("inherited physical Profile Settings must not differ:\n%s", settingsDraft)
	}
	summary := readTestFile(t, filepath.Join(result.Path, "summary.md"))
	for _, expected := range []string{"# Freeze Draft Summary", "| Recipe | SAME |", "| Settings | DIFFERENT | 0 | 0 | 1 |", "`reuse`", "`1.2.3`", "extension-set.golang.extensions"} {
		if !strings.Contains(summary, expected) {
			t.Fatalf("summary missing %q:\n%s", expected, summary)
		}
	}
	if strings.Contains(summary, "Available but Unused") {
		t.Fatalf("summary must not repeat unused Inventory:\n%s", summary)
	}
	if strings.Count(summary, "| `golang.go` |") != 1 {
		t.Fatalf("Extension summary must be unique by ID:\n%s", summary)
	}
	if _, err := os.Stat(filepath.Join(result.Path, "extensions.draft.md")); !os.IsNotExist(err) {
		t.Fatalf("SAME Freeze Draft Artifact must be omitted: %v", err)
	}
	if !strings.Contains(summary, "| Extensions | SAME | 0 | 0 | 0 | - |") {
		t.Fatalf("omitted Artifact must remain visible in summary:\n%s", summary)
	}
	if _, err := service.Generate(context.Background(), FreezeDraft, dist, "abort"); err == nil {
		t.Fatal("expected existing Workbench error")
	}
	if _, err := service.Generate(context.Background(), FreezeDraft, dist, "replace"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cookbookRoot, "draft.old", "summary.md")); err != nil {
		t.Fatalf("Freeze Draft .old missing: %v", err)
	}
	if _, err := service.Generate(context.Background(), View, dist, "abort"); err != nil {
		t.Fatal(err)
	}
	viewPath := filepath.Join(cookbookRoot, "inspect", "dist.demo")
	if _, err := os.Stat(filepath.Join(viewPath, "extensions.draft.md")); err != nil {
		t.Fatalf("View must retain Inventory Artifact: %v", err)
	}
	viewExtensions := readTestFile(t, filepath.Join(viewPath, "extensions.draft.md"))
	if !strings.Contains(viewExtensions, "# Extensions View") || !strings.Contains(viewExtensions, "+ golang.go") {
		t.Fatalf("View must render observations against an empty reference:\n%s", viewExtensions)
	}
	viewSummary := readTestFile(t, filepath.Join(viewPath, "summary.md"))
	if strings.Contains(viewSummary, "Extension Declaration Sources") || strings.Contains(viewSummary, "extension-set.golang.extensions") {
		t.Fatalf("Distribution View must not retain Cookbook-only Set provenance:\n%s", viewSummary)
	}
	for _, forbidden := range []string{"Recipe Difference", "Ingredient Context", "Available but Unused"} {
		if strings.Contains(viewSummary, forbidden) {
			t.Fatalf("View must not compare with current Cookbook (%s):\n%s", forbidden, viewSummary)
		}
	}
	if _, err := service.Generate(context.Background(), View, dist, "replace"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(viewPath + ".old"); !os.IsNotExist(err) {
		t.Fatalf("View must not retain .old: %v", err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestFindCurrentRecipeUsesDocumentIdentityNotFilename(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "recipe"), 0o755); err != nil {
		t.Fatal(err)
	}
	data := []byte("name: demo\nos: windows\nplatform: code\n")
	provenance := filepath.Join(root, "source.yaml")
	if err := os.WriteFile(provenance, data, 0o644); err != nil {
		t.Fatal(err)
	}
	oddName := filepath.Join(root, "recipe", "anything-at-all.yaml")
	if err := os.WriteFile(oddName, data, 0o644); err != nil {
		t.Fatal(err)
	}
	path, status, _ := findCurrentRecipe(root, recipe.Recipe{Name: "demo", OS: "windows", Platform: "code"}, provenance)
	if path != oddName || status != "SAME" {
		t.Fatalf("findCurrentRecipe() = %q, %q", path, status)
	}
	if err := os.WriteFile(filepath.Join(root, "recipe", "second.yaml"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, status, _ := findCurrentRecipe(root, recipe.Recipe{Name: "demo", OS: "windows", Platform: "code"}, provenance); status != "UNAVAILABLE" {
		t.Fatalf("duplicate identity status = %q", status)
	}
}
