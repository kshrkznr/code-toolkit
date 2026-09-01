package lifecycle

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ctkarchive "github.com/kshrkznr/code-toolkit/go/internal/archive"
	"github.com/kshrkznr/code-toolkit/go/internal/converge"
	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type archiveRuntime struct {
	settings   settings.Document
	extensions []runtimeio.Extension
}

func (r *archiveRuntime) Scopes(context.Context) ([]runtimeio.Scope, error) {
	return []runtimeio.Scope{runtimeio.DefaultScope()}, nil
}
func (r *archiveRuntime) EnsureProfile(context.Context, string) error { return nil }
func (r *archiveRuntime) SetInheritance(context.Context, runtimeio.Scope, cookbook.Inheritance) error {
	return nil
}
func (r *archiveRuntime) ReadInheritance(context.Context, runtimeio.Scope) (cookbook.Inheritance, error) {
	return cookbook.Inheritance{}, nil
}
func (r *archiveRuntime) ReadSettings(context.Context, runtimeio.Scope) (settings.Document, error) {
	return r.settings, nil
}
func (r *archiveRuntime) WriteSettings(_ context.Context, _ runtimeio.Scope, value settings.Document) error {
	r.settings = value
	return nil
}
func (r *archiveRuntime) Extensions(context.Context, runtimeio.Scope) ([]runtimeio.Extension, error) {
	return r.extensions, nil
}
func (r *archiveRuntime) InstallExtension(_ context.Context, _ runtimeio.Scope, _ string) error {
	r.extensions = []runtimeio.Extension{{ID: "sample.ext", Version: "2.0"}}
	return nil
}
func (r *archiveRuntime) UninstallExtension(context.Context, runtimeio.Scope, string) error {
	return nil
}

type fakeRuntime struct {
	installErr error
	installed  []string
}

type unresolvedUpdater struct{}

func (unresolvedUpdater) Update(_ context.Context, _ string, _ runtimelock.Snapshot, report *converge.Report) {
	report.Add(converge.Operation{Action: "update Extension Pool", Status: converge.Unresolved})
}

func (*fakeRuntime) Scopes(context.Context) ([]runtimeio.Scope, error) {
	return []runtimeio.Scope{runtimeio.DefaultScope()}, nil
}
func (*fakeRuntime) EnsureProfile(context.Context, string) error { return nil }
func (*fakeRuntime) SetInheritance(context.Context, runtimeio.Scope, cookbook.Inheritance) error {
	return nil
}
func (*fakeRuntime) ReadInheritance(context.Context, runtimeio.Scope) (cookbook.Inheritance, error) {
	return cookbook.Inheritance{}, nil
}
func (*fakeRuntime) ReadSettings(context.Context, runtimeio.Scope) (settings.Document, error) {
	return settings.Document{}, nil
}
func (*fakeRuntime) WriteSettings(context.Context, runtimeio.Scope, settings.Document) error {
	return nil
}
func (f *fakeRuntime) Extensions(context.Context, runtimeio.Scope) ([]runtimeio.Extension, error) {
	extensions := make([]runtimeio.Extension, 0, len(f.installed))
	for _, id := range f.installed {
		extensions = append(extensions, runtimeio.Extension{ID: id, Version: "test"})
	}
	return extensions, nil
}
func (f *fakeRuntime) InstallExtension(_ context.Context, _ runtimeio.Scope, id string) error {
	if f.installErr == nil {
		f.installed = append(f.installed, id)
	}
	return f.installErr
}
func (*fakeRuntime) UninstallExtension(context.Context, runtimeio.Scope, string) error { return nil }

func TestBuildPublishesCompletedStaging(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [empty]\n")
	service := Service{Cookbook: cookbook.Repository{Root: ingredients}, Runtime: func(distribution.Distribution) (runtimeio.Runtime, error) { return &fakeRuntime{}, nil }}
	result, err := service.Build(context.Background(), recipePath, filepath.Join(root, "dist"), "sample", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Distribution.Name != "sample" || result.StagingPath != "" {
		t.Fatalf("result = %#v", result)
	}
	for _, required := range []string{".meta/recipe.yaml", ".lock/manifest.json", ".lock/runtime.extensions.lock", "sample"} {
		if _, err := os.Stat(filepath.Join(result.Distribution.Path, required)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBuildResolvesExtensionSetBeforeRuntimeMutation(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [reserved]\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.reserved.extensions"), "set:shared\n")
	mustWrite(t, filepath.Join(ingredients, "extension-set.shared.extensions"), "shared.extension\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.reserved.settings.json"), `{"wouldMutate":true}`)
	runtimeCalled := false
	runtime := &fakeRuntime{}
	service := Service{Cookbook: cookbook.Repository{Root: ingredients}, Runtime: func(distribution.Distribution) (runtimeio.Runtime, error) {
		runtimeCalled = true
		return runtime, nil
	}}
	distRoot := filepath.Join(root, "dist")
	result, err := service.Build(context.Background(), recipePath, distRoot, "sample", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !runtimeCalled || len(runtime.installed) != 1 || runtime.installed[0] != "shared.extension" {
		t.Fatalf("Runtime called=%t installed=%v", runtimeCalled, runtime.installed)
	}
	lockData, err := os.ReadFile(filepath.Join(result.Distribution.Path, ".lock", "runtime.extensions.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(lockData), "shared.extension") || strings.Contains(string(lockData), "set:") {
		t.Fatalf("Lock must remain concrete-only: %s", lockData)
	}
}

func TestBuildRejectsNestedExtensionSetBeforeRuntimeMutation(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [reserved]\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.reserved.extensions"), "set:shared\n")
	mustWrite(t, filepath.Join(ingredients, "extension-set.shared.extensions"), "set:nested\n")
	runtimeCalled := false
	service := Service{Cookbook: cookbook.Repository{Root: ingredients}, Runtime: func(distribution.Distribution) (runtimeio.Runtime, error) {
		runtimeCalled = true
		return &fakeRuntime{}, nil
	}}
	distRoot := filepath.Join(root, "dist")
	_, err := service.Build(context.Background(), recipePath, distRoot, "sample", false, false)
	if err == nil || !strings.Contains(err.Error(), `nested Extension Set declaration "set:nested"`) {
		t.Fatalf("error = %v", err)
	}
	if runtimeCalled {
		t.Fatal("Runtime factory called after nested Extension Set declaration")
	}
	if _, statErr := os.Stat(distRoot); !os.IsNotExist(statErr) {
		t.Fatalf("Distribution root changed: %v", statErr)
	}
}

func TestBuildStagingPathDoesNotIncludeDistributionName(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [broken]\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.broken.extensions"), "broken.id\n")
	const name = "distribution-with-a-deliberately-long-name"
	service := Service{Cookbook: cookbook.Repository{Root: ingredients}, Runtime: func(distribution.Distribution) (runtimeio.Runtime, error) {
		return &fakeRuntime{installErr: errors.New("rejected")}, nil
	}}
	result, err := service.Build(context.Background(), recipePath, filepath.Join(root, "dist"), name, true, false)
	if err == nil {
		t.Fatal("expected failure")
	}
	base := filepath.Base(result.StagingPath)
	if !strings.HasPrefix(base, ".build-") || strings.Contains(base, name) {
		t.Fatalf("staging directory = %q", base)
	}
}

func TestBuildArchiveUsesExactAssetsAndPublishesFreshLock(t *testing.T) {
	root := t.TempDir()
	bundleRoot := filepath.Join(root, "archive", "sample")
	mustWrite(t, filepath.Join(bundleRoot, "lock", "recipe.yaml"), "name: sample\nos: macos\nplatform: code\n")
	writeLifecycleVSIX(t, filepath.Join(bundleRoot, "vsix", "sample.ext-2.0.vsix"))
	snapshot := runtimelock.Snapshot{FormatVersion: runtimelock.FormatVersion, RecipeName: "sample", Platform: "code", Default: runtimelock.ScopeSnapshot{Settings: settings.Document{"value": "archived"}, Extensions: []runtimeio.Extension{{ID: "sample.ext", Version: "2.0"}}}}
	bundle := ctkarchive.Bundle{Path: bundleRoot, Manifest: ctkarchive.Manifest{RecipeName: "sample", OS: "macos", Platform: "code", LaunchOverrides: []string{"run.sh"}}, Recipe: recipe.Recipe{Name: "sample", OS: "macos", Platform: "code"}, Snapshot: snapshot}
	runtime := &archiveRuntime{}
	service := Service{Runtime: func(distribution.Distribution) (runtimeio.Runtime, error) { return runtime, nil }, Locks: runtimelock.Store{}}
	result, err := service.BuildArchive(context.Background(), bundle, filepath.Join(root, "dist"), "sample", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Distribution.Name != "sample" || len(result.Warnings) != 1 {
		t.Fatalf("result = %#v", result)
	}
	if runtime.settings["value"] != "archived" || len(runtime.extensions) != 1 || runtime.extensions[0].Version != "2.0" {
		t.Fatalf("runtime = %#v", runtime)
	}
	if _, err := os.Stat(filepath.Join(result.Distribution.Path, ".lock", "manifest.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(result.Distribution.Path, "sample")); err != nil {
		t.Fatal(err)
	}
	runtime.settings = settings.Document{"value": "changed"}
	runtime.extensions = []runtimeio.Extension{{ID: "sample.ext", Version: "1.0"}}
	applied, err := service.ApplyArchive(context.Background(), bundle, result.Distribution)
	if err != nil {
		t.Fatal(err)
	}
	if applied.Distribution.Name != "sample" || runtime.settings["value"] != "archived" || runtime.extensions[0].Version != "2.0" {
		t.Fatalf("applied=%#v runtime=%#v", applied, runtime)
	}
}

func writeLifecycleVSIX(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("extension/package.json")
	if err == nil {
		_, err = entry.Write([]byte(`{"publisher":"sample","name":"ext","version":"2.0"}`))
	}
	if closeErr := writer.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildFailureRemovesOrKeepsStaging(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [broken]\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.broken.extensions"), "broken.id\n")
	service := Service{Cookbook: cookbook.Repository{Root: ingredients}, Runtime: func(distribution.Distribution) (runtimeio.Runtime, error) {
		return &fakeRuntime{installErr: errors.New("rejected")}, nil
	}}
	removed, err := service.Build(context.Background(), recipePath, filepath.Join(root, "dist"), "removed", false, false)
	if err == nil {
		t.Fatal("expected failure")
	}
	if _, statErr := os.Stat(removed.StagingPath); !os.IsNotExist(statErr) {
		t.Fatalf("staging remains: %v", statErr)
	}
	kept, err := service.Build(context.Background(), recipePath, filepath.Join(root, "dist"), "kept", true, false)
	if err == nil {
		t.Fatal("expected failure")
	}
	if _, statErr := os.Stat(kept.StagingPath); statErr != nil {
		t.Fatalf("staging missing: %v", statErr)
	}
}

func TestBuildForcePublishesWithUnresolvedExtension(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [broken]\n")
	mustWrite(t, filepath.Join(ingredients, "runtime.broken.extensions"), "broken.id\n")
	service := Service{Cookbook: cookbook.Repository{Root: ingredients}, Runtime: func(distribution.Distribution) (runtimeio.Runtime, error) {
		return &fakeRuntime{installErr: errors.New("rejected")}, nil
	}}
	result, err := service.Build(context.Background(), recipePath, filepath.Join(root, "dist"), "forced", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Distribution.Name != "forced" || result.Report.HasFailures() {
		t.Fatalf("result=%#v", result)
	}
	found := false
	for _, operation := range result.Report.Operations {
		if operation.Action == "install extension" && operation.Status == converge.Unresolved {
			found = true
		}
	}
	if !found {
		t.Fatalf("report=%#v", result.Report)
	}
}

func TestNextAvailableName(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "sample"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "sample.1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got, err := NextAvailableName(root, "sample"); err != nil || got != "sample.2" {
		t.Fatalf("name = %q", got)
	}
}

func TestPoolUpdateUnresolvedDoesNotFailBuild(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [empty]\nconfig:\n  dist-strategy:\n    extension-pool: refresh\n")
	service := Service{
		Cookbook:   cookbook.Repository{Root: ingredients},
		Runtime:    func(distribution.Distribution) (runtimeio.Runtime, error) { return &fakeRuntime{}, nil },
		PoolUpdate: unresolvedUpdater{},
	}
	result, err := service.Build(context.Background(), recipePath, filepath.Join(root, "dist"), "sample", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Report.Operations) == 0 || result.Report.Operations[len(result.Report.Operations)-1].Status != converge.Unresolved {
		t.Fatalf("report = %#v", result.Report)
	}
}

func TestPoolUpdateIsDisabledByDefault(t *testing.T) {
	root := t.TempDir()
	recipePath, ingredients := fixture(t, root, "runtime: [empty]\n")
	service := Service{
		Cookbook:   cookbook.Repository{Root: ingredients},
		Runtime:    func(distribution.Distribution) (runtimeio.Runtime, error) { return &fakeRuntime{}, nil },
		PoolUpdate: unresolvedUpdater{},
	}
	result, err := service.Build(context.Background(), recipePath, filepath.Join(root, "dist"), "sample", false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range result.Report.Operations {
		if operation.Action == "update Extension Pool" {
			t.Fatalf("default Build contacted Extension Pool updater: %#v", result.Report)
		}
	}
}

func fixture(t *testing.T, root, extra string) (string, string) {
	t.Helper()
	recipePath := filepath.Join(root, "recipe", "sample.yaml")
	ingredients := filepath.Join(root, "ingredient")
	mustWrite(t, recipePath, "name: sample\nos: macos\nplatform: code\n"+extra)
	return recipePath, ingredients
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
