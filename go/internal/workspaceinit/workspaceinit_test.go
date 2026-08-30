package workspaceinit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
)

func TestInitializeCreatesDiscoverableWorkspaceWithSample(t *testing.T) {
	target := filepath.Join(t.TempDir(), "new-workspace")
	result, err := Initialize(target, Options{IncludeSample: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != target || len(result.Created) != 6 || len(result.Unchanged) != 0 {
		t.Fatalf("result = %+v", result)
	}
	for _, relative := range []string{
		"cookbook/recipe/vscode-sample.macos.yaml",
		"cookbook/recipe/vscode-sample.windows.yaml",
		"cookbook/ingredient/runtime.common.settings.json",
		"cookbook/ingredient/runtime.text.settings.json",
		"cookbook/ingredient/runtime.markdown.extensions",
		"cookbook/ingredient/extension/yzhang.markdown-all-in-one.settings.json",
	} {
		if info, err := os.Stat(filepath.Join(target, filepath.FromSlash(relative))); err != nil || info.IsDir() {
			t.Fatalf("sample file %s: %v", relative, err)
		}
	}
	repository := cookbook.Repository{Root: filepath.Join(target, "cookbook", "ingredient")}
	for _, recipeName := range []string{"vscode-sample.macos.yaml", "vscode-sample.windows.yaml"} {
		plan, err := repository.Resolve(filepath.Join(target, "cookbook", "recipe", recipeName))
		if err != nil {
			t.Fatalf("resolve initialized %s: %v", recipeName, err)
		}
		if plan.Name != "vscode-sample" || len(plan.Default.Extensions) != 1 {
			t.Fatalf("initialized plan = %#v", plan)
		}
	}

	repeated, err := Initialize(target, Options{IncludeSample: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(repeated.Created) != 0 || len(repeated.Unchanged) != 6 {
		t.Fatalf("repeated result = %+v", repeated)
	}
}

func TestInitializeCanExcludeSample(t *testing.T) {
	target := filepath.Join(t.TempDir(), "empty-workspace")
	result, err := Initialize(target, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 0 || len(result.Unchanged) != 0 {
		t.Fatalf("result = %+v", result)
	}
	for _, relative := range []string{"cookbook/recipe", "cookbook/ingredient"} {
		if info, err := os.Stat(filepath.Join(target, filepath.FromSlash(relative))); err != nil || !info.IsDir() {
			t.Fatalf("Workspace directory %s: %v", relative, err)
		}
	}
}

func TestInitializeRejectsAllConflictsBeforeWriting(t *testing.T) {
	target := t.TempDir()
	conflict := filepath.Join(target, "cookbook", "ingredient", "runtime.common.settings.json")
	if err := os.MkdirAll(filepath.Dir(conflict), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(conflict, []byte("different\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Initialize(target, Options{IncludeSample: true})
	if err == nil || !strings.Contains(err.Error(), "runtime.common.settings.json") {
		t.Fatalf("conflict error = %v", err)
	}
	missing := filepath.Join(target, "cookbook", "recipe", "vscode-sample.macos.yaml")
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("initialization wrote before rejecting conflict: %v", err)
	}
}
