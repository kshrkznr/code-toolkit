package converge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type fakeRuntime struct {
	extensions  map[string][]runtimeio.Extension
	installed   []string
	uninstalled []string
	installErrs map[string]error
}

func (f *fakeRuntime) Scopes(context.Context) ([]runtimeio.Scope, error) {
	return []runtimeio.Scope{runtimeio.DefaultScope()}, nil
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
func (f *fakeRuntime) Extensions(_ context.Context, scope runtimeio.Scope) ([]runtimeio.Extension, error) {
	return append([]runtimeio.Extension(nil), f.extensions[scope.Name]...), nil
}
func (f *fakeRuntime) InstallExtension(_ context.Context, _ runtimeio.Scope, id string) error {
	f.installed = append(f.installed, id)
	return f.installErrs[id]
}
func (f *fakeRuntime) UninstallExtension(_ context.Context, _ runtimeio.Scope, id string) error {
	f.uninstalled = append(f.uninstalled, id)
	return nil
}

func TestExtensionsConvergesExactIDs(t *testing.T) {
	runtime := &fakeRuntime{extensions: map[string][]runtimeio.Extension{"": {{ID: "keep.id"}, {ID: "remove.id"}}}}
	report := Report{}
	Extensions(context.Background(), runtime, runtimeio.DefaultScope(), []string{"keep.id", "new.id"}, MarketplacePolicy{Allowed: true}, &report)
	if !slices.Equal(runtime.installed, []string{"new.id"}) || !slices.Equal(runtime.uninstalled, []string{"remove.id"}) {
		t.Fatalf("installed=%v uninstalled=%v", runtime.installed, runtime.uninstalled)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func TestExtensionsRejectsCaseOnlyConflictWithoutMutation(t *testing.T) {
	runtime := &fakeRuntime{extensions: map[string][]runtimeio.Extension{"": {{ID: "golang.go"}}}}
	report := Report{}
	Extensions(context.Background(), runtime, runtimeio.DefaultScope(), []string{"Golang.Go"}, MarketplacePolicy{Allowed: true}, &report)
	if len(runtime.installed) != 0 || len(runtime.uninstalled) != 0 {
		t.Fatalf("mutated: %#v", runtime)
	}
	if !report.HasFailures() {
		t.Fatal("expected conflict failure")
	}
}

func TestExtensionsRejectsCaseOnlyDesiredDuplicates(t *testing.T) {
	runtime := &fakeRuntime{extensions: map[string][]runtimeio.Extension{}}
	report := Report{}
	Extensions(context.Background(), runtime, runtimeio.DefaultScope(), []string{"Golang.Go", "golang.go"}, MarketplacePolicy{Allowed: true}, &report)
	if len(runtime.installed) != 0 || !report.HasFailures() {
		t.Fatalf("runtime=%#v report=%#v", runtime, report)
	}
}

func TestExtensionsPoolMissIsUnresolvedNotFailure(t *testing.T) {
	runtime := &fakeRuntime{extensions: map[string][]runtimeio.Extension{}}
	report := Report{}
	Extensions(context.Background(), runtime, runtimeio.DefaultScope(), []string{"missing.id"}, MarketplacePolicy{Allowed: false}, &report)
	if report.HasFailures() || report.Error() != nil {
		t.Fatalf("report = %#v", report)
	}
	if len(report.Operations) != 1 || report.Operations[0].Status != Unresolved {
		t.Fatalf("operations = %#v", report.Operations)
	}
}

func TestReportErrorWrapsFailures(t *testing.T) {
	want := errors.New("failed")
	report := Report{Operations: []Operation{{Status: Failed, Err: want}}}
	if !errors.Is(report.Error(), want) {
		t.Fatalf("error = %v", report.Error())
	}
}

func TestPoolUsesKiroRepositoryThenVisualStudioFallback(t *testing.T) {
	root := t.TempDir()
	primary := filepath.Join(root, "open-vsx", "sample.id-2.0.vsix")
	marketplace := filepath.Join(root, "visual-studio-marketplace", "sample.id-1.0.vsix")
	for _, path := range []string{primary, marketplace} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := (Pool{Root: root}).ResolveCandidates("kiro", "sample.id")
	if err != nil || len(got) != 2 || got[0].Path != primary || !got[0].Primary || got[1].Path != marketplace || got[1].Primary {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}

func TestPoolUsesCursorMarketplaceThenVisualStudioFallback(t *testing.T) {
	root := t.TempDir()
	cursor := filepath.Join(root, "cursor-marketplace", "sample.id-2.0.vsix")
	openVSX := filepath.Join(root, "open-vsx", "sample.id-3.0.vsix")
	visualStudio := filepath.Join(root, "visual-studio-marketplace", "sample.id-1.0.vsix")
	for _, path := range []string{cursor, openVSX, visualStudio} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := (Pool{Root: root}).ResolveCandidates("cursor", "sample.id")
	if err != nil || len(got) != 2 || got[0].Path != cursor || !got[0].Primary || got[1].Path != visualStudio || got[1].Primary {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}

func TestPoolUsesOnlyWindsurfMarketplaceForDevinDesktop(t *testing.T) {
	root := t.TempDir()
	windsurf := filepath.Join(root, "windsurf-marketplace", "sample.id-2.0.vsix")
	visualStudio := filepath.Join(root, "visual-studio-marketplace", "sample.id-1.0.vsix")
	for _, path := range []string{windsurf, visualStudio} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := (Pool{Root: root}).ResolveCandidates("devin-desktop", "sample.id")
	if err != nil || len(got) != 1 || got[0].Path != windsurf || !got[0].Primary {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}

func TestPlatformRepositories(t *testing.T) {
	tests := []struct {
		platform string
		want     []string
	}{
		{"code", []string{"visual-studio-marketplace"}},
		{"kiro", []string{"open-vsx", "visual-studio-marketplace"}},
		{"cursor", []string{"cursor-marketplace", "visual-studio-marketplace"}},
		{"devin-desktop", []string{"windsurf-marketplace"}},
	}
	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			if got := platformRepositories(tt.platform); !slices.Equal(got, tt.want) {
				t.Fatalf("platformRepositories() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoolReadsLegacyMixedCaseArtifact(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, "open-vsx", "emilast.LogFileHighlighter-2.8.0.vsix")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := (Pool{Root: root}).ResolveCandidates("kiro", "emilast.logfilehighlighter")
	if err != nil || len(got) != 1 || got[0].Path != legacy || !got[0].Primary {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}

type candidateResolver struct{ candidates []ArtifactCandidate }

func (r candidateResolver) ResolveCandidates(string, string) ([]ArtifactCandidate, error) {
	return r.candidates, nil
}

func TestExtensionsTriesPlatformRepositoryBeforeSecondaryPool(t *testing.T) {
	secondary := "/pool/visual-studio-marketplace/sample.id-1.0.vsix"
	runtime := &fakeRuntime{
		extensions:  map[string][]runtimeio.Extension{},
		installErrs: map[string]error{"sample.id": errors.New("not in Open VSX")},
	}
	report := Report{}
	Extensions(context.Background(), runtime, runtimeio.DefaultScope(), []string{"sample.id"}, MarketplacePolicy{
		Platform: "kiro", Allowed: true,
		Pool: candidateResolver{candidates: []ArtifactCandidate{{Path: secondary, Repository: "visual-studio-marketplace"}}},
	}, &report)
	if !slices.Equal(runtime.installed, []string{"sample.id", secondary}) || report.HasFailures() {
		t.Fatalf("installed=%v report=%#v", runtime.installed, report)
	}
}

func TestExtensionsForceLeavesFailedInstallUnresolved(t *testing.T) {
	want := errors.New("unavailable")
	runtime := &fakeRuntime{extensions: map[string][]runtimeio.Extension{}, installErrs: map[string]error{"sample.id": want}}
	report := Report{}
	Extensions(context.Background(), runtime, runtimeio.DefaultScope(), []string{"sample.id"}, MarketplacePolicy{Allowed: true, Force: true}, &report)
	if report.HasFailures() || len(report.Operations) != 1 || report.Operations[0].Status != Unresolved || !errors.Is(report.Operations[0].Err, want) {
		t.Fatalf("report=%#v", report)
	}
}
