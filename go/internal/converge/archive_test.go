package converge

import (
	"context"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type exactResolver struct{}

func (exactResolver) ResolveExact(extension runtimeio.Extension) (string, error) {
	return "/archive/" + extension.ID + "-" + extension.Version + ".vsix", nil
}

func TestArchiveSnapshotReinstallsWrongVersionFromExactArtifact(t *testing.T) {
	runtime := &fakeRuntime{extensions: map[string][]runtimeio.Extension{"": {{ID: "sample.ext", Version: "1.0"}, {ID: "remove.ext", Version: "1.0"}}}}
	plan := cookbook.Plan{Platform: "code", Default: cookbook.ScopePlan{Settings: settings.Document{}, Extensions: []string{"sample.ext"}}}
	snapshot := runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Settings: settings.Document{}, Extensions: []runtimeio.Extension{{ID: "sample.ext", Version: "2.0"}}}}
	report := ArchiveSnapshot(context.Background(), runtime, plan, snapshot, exactResolver{})
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
	if len(runtime.installed) != 1 || runtime.installed[0] != "/archive/sample.ext-2.0.vsix" {
		t.Fatalf("installed = %v", runtime.installed)
	}
	if len(runtime.uninstalled) != 1 || runtime.uninstalled[0] != "remove.ext" {
		t.Fatalf("uninstalled = %v", runtime.uninstalled)
	}
}
