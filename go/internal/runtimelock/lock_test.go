package runtimelock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/runtimeio"
	"code-toolkit/internal/settings"
)

type fakeRuntime struct {
	scopes   []runtimeio.Scope
	failures int
}

func (f *fakeRuntime) Scopes(context.Context) ([]runtimeio.Scope, error) {
	if f.failures > 0 {
		f.failures--
		return nil, errors.New("transient observation")
	}
	return f.scopes, nil
}
func (f *fakeRuntime) EnsureProfile(context.Context, string) error { return nil }
func (f *fakeRuntime) SetInheritance(context.Context, runtimeio.Scope, cookbook.Inheritance) error {
	return nil
}
func (f *fakeRuntime) ReadInheritance(context.Context, runtimeio.Scope) (cookbook.Inheritance, error) {
	return cookbook.Inheritance{}, nil
}
func (f *fakeRuntime) ReadSettings(context.Context, runtimeio.Scope) (settings.Document, error) {
	return settings.Document{}, nil
}
func (f *fakeRuntime) WriteSettings(context.Context, runtimeio.Scope, settings.Document) error {
	return nil
}
func (f *fakeRuntime) Extensions(context.Context, runtimeio.Scope) ([]runtimeio.Extension, error) {
	return []runtimeio.Extension{}, nil
}
func (f *fakeRuntime) InstallExtension(context.Context, runtimeio.Scope, string) error   { return nil }
func (f *fakeRuntime) UninstallExtension(context.Context, runtimeio.Scope, string) error { return nil }

func TestCollectRequiresRecipeProfilesAndObservesExtraProfiles(t *testing.T) {
	plan := cookbook.Plan{Name: "sample", Platform: "code", Profiles: []cookbook.ScopePlan{{Name: "required"}}}
	runtime := &fakeRuntime{scopes: []runtimeio.Scope{runtimeio.DefaultScope(), runtimeio.ProfileScope("extra"), runtimeio.ProfileScope("required")}}
	snapshot, err := (Collector{Now: func() time.Time { return time.Unix(1, 0) }}).Collect(context.Background(), runtime, plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Profiles) != 2 || snapshot.Profiles[0].Name != "extra" || snapshot.Profiles[1].Name != "required" {
		t.Fatalf("profiles = %#v", snapshot.Profiles)
	}
}

func TestRefreshRetriesAndPreservesPreviousLockUntilPublish(t *testing.T) {
	dist := t.TempDir()
	recipePath := filepath.Join(dist, "recipe-source.yaml")
	if err := os.WriteFile(recipePath, []byte("name: sample\nplatform: code\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(dist, ".lock")
	if err := os.Mkdir(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "old"), []byte("trusted"), 0o644); err != nil {
		t.Fatal(err)
	}
	runtime := &fakeRuntime{scopes: []runtimeio.Scope{runtimeio.DefaultScope()}, failures: 2}
	delays := 0
	store := Store{Attempts: 3, Delay: func(int) { delays++ }}
	if _, err := store.Refresh(context.Background(), dist, recipePath, runtime, cookbook.Plan{Name: "sample", Platform: "code"}); err != nil {
		t.Fatal(err)
	}
	if delays != 2 {
		t.Fatalf("delays = %d", delays)
	}
	if _, err := os.Stat(filepath.Join(old, "manifest.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(old, "old")); !os.IsNotExist(err) {
		t.Fatalf("old Lock retained inside new Lock: %v", err)
	}
}

func TestReadRejectsBashOnlyLock(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"recipe.yaml", "settings.jsonc", "runtime.extensions.lock"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	_, _, err := Read(root, cookbook.Plan{Name: "sample", Platform: "code"})
	if err == nil {
		t.Fatal("Bash-only Lock must not be trusted by Go")
	}
}
