package recovery

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type fakeRuntime struct {
	settings    map[string]settings.Document
	extensions  map[string][]runtimeio.Extension
	inheritance map[string]cookbook.Inheritance
}

func newFakeRuntime() *fakeRuntime {
	return &fakeRuntime{settings: map[string]settings.Document{"": {}}, extensions: map[string][]runtimeio.Extension{"": {}}, inheritance: map[string]cookbook.Inheritance{}}
}
func (f *fakeRuntime) Scopes(context.Context) ([]runtimeio.Scope, error) {
	names := []string{""}
	for name := range f.settings {
		if name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	result := make([]runtimeio.Scope, len(names))
	for i, name := range names {
		result[i] = runtimeio.Scope{Name: name}
	}
	return result, nil
}
func (f *fakeRuntime) EnsureProfile(_ context.Context, name string) error {
	if _, ok := f.settings[name]; !ok {
		f.settings[name] = settings.Document{}
		f.extensions[name] = []runtimeio.Extension{}
	}
	return nil
}
func (f *fakeRuntime) SetInheritance(_ context.Context, scope runtimeio.Scope, value cookbook.Inheritance) error {
	f.inheritance[scope.Name] = value
	return nil
}
func (f *fakeRuntime) ReadInheritance(_ context.Context, scope runtimeio.Scope) (cookbook.Inheritance, error) {
	return f.inheritance[scope.Name], nil
}
func (f *fakeRuntime) ReadSettings(_ context.Context, scope runtimeio.Scope) (settings.Document, error) {
	return f.settings[scope.Name], nil
}
func (f *fakeRuntime) WriteSettings(_ context.Context, scope runtimeio.Scope, value settings.Document) error {
	f.settings[scope.Name] = value
	return nil
}
func (f *fakeRuntime) Extensions(_ context.Context, scope runtimeio.Scope) ([]runtimeio.Extension, error) {
	return f.extensions[scope.Name], nil
}
func (f *fakeRuntime) InstallExtension(_ context.Context, scope runtimeio.Scope, id string) error {
	f.extensions[scope.Name] = append(f.extensions[scope.Name], runtimeio.Extension{ID: id, Version: "new"})
	return nil
}
func (f *fakeRuntime) UninstallExtension(_ context.Context, scope runtimeio.Scope, id string) error {
	var result []runtimeio.Extension
	for _, extension := range f.extensions[scope.Name] {
		if extension.ID != id {
			result = append(result, extension)
		}
	}
	f.extensions[scope.Name] = result
	return nil
}

func TestPrepareReportsRecipeAndSourceProfileDifferences(t *testing.T) {
	root := t.TempDir()
	recipeText := "name: sample\nos: macos\nplatform: code\nprofile: [recipe-only]\n"
	snapshot := runtimelock.Snapshot{FormatVersion: 1, RecipeName: "sample", Platform: "code", Default: scope("", nil), Profiles: []runtimelock.ScopeSnapshot{scope("lock-only", nil)}}
	writeLock(t, root, recipeText, snapshot)
	prepared, err := Prepare(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.Before.Differences) != 2 {
		t.Fatalf("differences = %#v", prepared.Before.Differences)
	}
	if len(prepared.Plan.Profiles) != 1 || prepared.Plan.Profiles[0].Name != "recipe-only" {
		t.Fatalf("plan = %#v", prepared.Plan)
	}
}

func TestCompareTreatsTasksEnvelopeAsSemanticEmpty(t *testing.T) {
	source := runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Tasks: runtimeartifact.Object{
		"version": "2.0.0", "tasks": []any{}, "inputs": []any{},
	}}}
	recovered := runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Tasks: runtimeartifact.Object{}}}
	if verification := Compare(cookbook.Plan{}, source, recovered); len(verification.Differences) != 0 {
		t.Fatalf("differences = %#v", verification.Differences)
	}
}

func TestRecoverUsesIDsAndIgnoresVersionDifference(t *testing.T) {
	lockRoot := t.TempDir()
	inheritance := cookbook.Inheritance{Settings: true, Keybindings: true, Tasks: true, MCP: true, Snippets: true}
	snapshot := runtimelock.Snapshot{
		FormatVersion: 1, RecipeName: "sample", Platform: "code", ObservedAt: time.Unix(1, 0),
		Default:  scope("", []runtimeio.Extension{{ID: "sample.default", Version: "old"}}),
		Profiles: []runtimelock.ScopeSnapshot{{Name: "work", Settings: settings.Document{}, Extensions: []runtimeio.Extension{{ID: "sample.profile", Version: "old"}}, Inheritance: inheritance}},
	}
	writeLock(t, lockRoot, "name: sample\nos: macos\nplatform: code\nprofile: [work]\n", snapshot)
	prepared, err := Prepare(lockRoot)
	if err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(target, ".ext"), 0o755); err != nil {
		t.Fatal(err)
	}
	runtime := newFakeRuntime()
	result, err := (Service{Locks: runtimelock.Store{Attempts: 1}}).Recover(context.Background(), prepared, target, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Before.Matches() || !result.After.Matches() {
		t.Fatalf("before=%#v after=%#v", result.Before, result.After)
	}
	if _, err := os.Stat(filepath.Join(target, ".lock", "manifest.json")); err != nil {
		t.Fatal(err)
	}
}

func TestRecoverRejectsNonEmptyExtensionArea(t *testing.T) {
	target := t.TempDir()
	extensions := filepath.Join(target, ".ext")
	if err := os.Mkdir(extensions, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extensions, "extensions.json"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := (Service{}).Recover(context.Background(), Prepared{}, target, newFakeRuntime())
	if err == nil {
		t.Fatal("expected non-empty Extension area failure")
	}
}

func TestCompareReportsSemanticDifferencesButIgnoresVersions(t *testing.T) {
	plan := cookbook.Plan{Profiles: []cookbook.ScopePlan{{Name: "work"}}}
	source := runtimelock.Snapshot{
		Default:  runtimelock.ScopeSnapshot{Settings: settings.Document{"value": "source"}, Extensions: []runtimeio.Extension{{ID: "same.id", Version: "1"}}},
		Profiles: []runtimelock.ScopeSnapshot{scope("work", []runtimeio.Extension{})},
	}
	recovered := runtimelock.Snapshot{
		Default:  runtimelock.ScopeSnapshot{Settings: settings.Document{"value": "recovered"}, Extensions: []runtimeio.Extension{{ID: "same.id", Version: "2"}}},
		Profiles: []runtimelock.ScopeSnapshot{scope("work", []runtimeio.Extension{{ID: "extra.id", Version: "1"}})},
	}
	verification := Compare(plan, source, recovered)
	if len(verification.Differences) != 2 {
		t.Fatalf("differences = %#v", verification.Differences)
	}
	if verification.Differences[0].Kind != "settings" || verification.Differences[1].Kind != "extensions" {
		t.Fatalf("differences = %#v", verification.Differences)
	}
}

func scope(name string, extensions []runtimeio.Extension) runtimelock.ScopeSnapshot {
	if extensions == nil {
		extensions = []runtimeio.Extension{}
	}
	return runtimelock.ScopeSnapshot{Name: name, Settings: settings.Document{}, Extensions: extensions}
}

func writeLock(t *testing.T, root, recipeText string, snapshot runtimelock.Snapshot) {
	t.Helper()
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "recipe.yaml"), []byte(recipeText), 0o644); err != nil {
		t.Fatal(err)
	}
}
