package converge

import (
	"context"
	"fmt"
	"sort"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
)

type ExactArtifactResolver interface {
	ResolveExact(runtimeio.Extension) (string, error)
}

func ArchiveSnapshot(ctx context.Context, runtime runtimeio.Runtime, plan cookbook.Plan, snapshot runtimelock.Snapshot, resolver ExactArtifactResolver) Report {
	report := Report{}
	applyArchiveScope(ctx, runtime, runtimeio.DefaultScope(), plan.Default, snapshot.Default, resolver, &report)
	for index, scope := range plan.Profiles {
		if err := runtime.EnsureProfile(ctx, scope.Name); err != nil {
			report.Add(Operation{Scope: scope.Name, Action: "ensure profile", Status: Failed, Err: err})
			continue
		}
		report.Add(Operation{Scope: scope.Name, Action: "ensure profile", Status: Completed})
		if err := runtime.SetInheritance(ctx, runtimeio.ProfileScope(scope.Name), scope.Inheritance); err != nil {
			report.Add(Operation{Scope: scope.Name, Action: "set inheritance", Status: Failed, Err: err})
		} else {
			report.Add(Operation{Scope: scope.Name, Action: "set inheritance", Status: Completed})
		}
		applyArchiveScope(ctx, runtime, runtimeio.ProfileScope(scope.Name), scope, snapshot.Profiles[index], resolver, &report)
	}
	return report
}

func applyArchiveScope(ctx context.Context, runtime runtimeio.Runtime, scope runtimeio.Scope, plan cookbook.ScopePlan, snapshot runtimelock.ScopeSnapshot, resolver ExactArtifactResolver, report *Report) {
	if !plan.Inheritance.Unmanaged["settings"] && plan.Settings != nil && (scope.IsDefault() || !plan.Inheritance.Settings) {
		if err := runtime.WriteSettings(ctx, scope, plan.Settings); err != nil {
			report.Add(Operation{Scope: scope.Name, Action: "write settings", Status: Failed, Err: err})
		} else {
			report.Add(Operation{Scope: scope.Name, Action: "write settings", Status: Completed})
		}
	}
	if artifacts, ok := runtime.(runtimeio.ArtifactRuntime); ok {
		write := func(kind string, inherited bool, unmanaged bool, value any, operation func() error) {
			if unmanaged || !scope.IsDefault() && inherited || value == nil {
				return
			}
			if err := operation(); err != nil {
				report.Add(Operation{Scope: scope.Name, Action: "write archived " + kind, Status: Failed, Err: err})
			} else {
				report.Add(Operation{Scope: scope.Name, Action: "write archived " + kind, Status: Completed})
			}
		}
		write("keybindings", plan.Inheritance.Keybindings, plan.Inheritance.Unmanaged["keybindings"], snapshot.Keybindings, func() error { return artifacts.WriteKeybindings(ctx, scope, snapshot.Keybindings) })
		write("tasks", plan.Inheritance.Tasks, plan.Inheritance.Unmanaged["tasks"], snapshot.Tasks, func() error { return artifacts.WriteTasks(ctx, scope, snapshot.Tasks) })
		write("mcp", plan.Inheritance.MCP, plan.Inheritance.Unmanaged["mcp"], snapshot.MCP, func() error { return artifacts.WriteMCP(ctx, scope, snapshot.MCP) })
		write("snippets", plan.Inheritance.Snippets, plan.Inheritance.Unmanaged["snippets"], snapshot.Snippets, func() error { return artifacts.WriteSnippets(ctx, scope, snapshot.Snippets) })
	}
	if plan.Inheritance.Unmanaged["extensions"] {
		return
	}
	installed, err := runtime.Extensions(ctx, scope)
	if err != nil {
		report.Add(Operation{Scope: scope.Name, Action: "observe extension", Status: Failed, Err: err})
		return
	}
	installedByID := map[string]runtimeio.Extension{}
	for _, extension := range installed {
		installedByID[extension.ID] = extension
	}
	desired := map[string]runtimeio.Extension{}
	for _, extension := range snapshot.Extensions {
		desired[extension.ID] = extension
	}
	ids := make([]string, 0, len(desired))
	for id := range desired {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		want := desired[id]
		if got, ok := installedByID[id]; ok && got.Version == want.Version {
			report.Add(Operation{Scope: scope.Name, Action: "retain exact extension", Subject: id + "@" + want.Version, Status: Completed})
			continue
		}
		artifact, resolveErr := resolver.ResolveExact(want)
		if resolveErr != nil {
			report.Add(Operation{Scope: scope.Name, Action: "resolve archived extension", Subject: id + "@" + want.Version, Status: Failed, Err: resolveErr})
			continue
		}
		if err := runtime.InstallExtension(ctx, scope, artifact); err != nil {
			report.Add(Operation{Scope: scope.Name, Action: "install archived extension", Subject: id + "@" + want.Version, Status: Failed, Err: err})
		} else {
			report.Add(Operation{Scope: scope.Name, Action: "install archived extension", Subject: id + "@" + want.Version, Status: Completed})
		}
	}
	installedIDs := make([]string, 0, len(installedByID))
	for id := range installedByID {
		installedIDs = append(installedIDs, id)
	}
	sort.Strings(installedIDs)
	for _, id := range installedIDs {
		if _, ok := desired[id]; ok {
			continue
		}
		if err := runtime.UninstallExtension(ctx, scope, id); err != nil {
			report.Add(Operation{Scope: scope.Name, Action: "uninstall extension", Subject: id, Status: Failed, Err: err})
		} else {
			report.Add(Operation{Scope: scope.Name, Action: "uninstall extension", Subject: id, Status: Completed})
		}
	}
}

func ExactSubject(extension runtimeio.Extension) string {
	return fmt.Sprintf("%s@%s", extension.ID, extension.Version)
}
