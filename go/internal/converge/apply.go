package converge

import (
	"context"
	"reflect"

	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/runtimeio"
)

func Plan(ctx context.Context, runtime runtimeio.Runtime, plan cookbook.Plan, pool ArtifactResolver, forceExtensions bool) Report {
	report := Report{}
	applyScope(ctx, runtime, runtimeio.DefaultScope(), plan.Default, plan.Platform, plan.ExtensionMarketplace, pool, forceExtensions, &report)
	for _, profile := range plan.Profiles {
		if err := runtime.EnsureProfile(ctx, profile.Name); err != nil {
			report.Add(Operation{Scope: profile.Name, Action: "ensure profile", Subject: profile.Name, Status: Failed, Err: err})
			continue
		}
		report.Add(Operation{Scope: profile.Name, Action: "ensure profile", Subject: profile.Name, Status: Completed})
		if err := runtime.SetInheritance(ctx, runtimeio.ProfileScope(profile.Name), profile.Inheritance); err != nil {
			report.Add(Operation{Scope: profile.Name, Action: "set inheritance", Subject: profile.Name, Status: Failed, Err: err})
		} else {
			report.Add(Operation{Scope: profile.Name, Action: "set inheritance", Subject: profile.Name, Status: Completed})
		}
		applyScope(ctx, runtime, runtimeio.ProfileScope(profile.Name), profile, plan.Platform, plan.ExtensionMarketplace, pool, forceExtensions, &report)
	}
	return report
}

func applyScope(ctx context.Context, runtime runtimeio.Runtime, scope runtimeio.Scope, plan cookbook.ScopePlan, platform string, marketplace bool, pool ArtifactResolver, forceExtensions bool, report *Report) {
	if !plan.Inheritance.Unmanaged["settings"] && plan.Settings != nil && (scope.IsDefault() || !plan.Inheritance.Settings) {
		if err := runtime.WriteSettings(ctx, scope, plan.Settings); err != nil {
			report.Add(Operation{Scope: scope.Name, Action: "write settings", Status: Failed, Err: err})
		} else {
			report.Add(Operation{Scope: scope.Name, Action: "write settings", Status: Completed})
		}
	}
	if artifacts, ok := runtime.(runtimeio.ArtifactRuntime); ok {
		write := func(kind string, inherited bool, unmanaged bool, value any, operation func() error) {
			if unmanaged || !scope.IsDefault() && inherited || nilArtifact(value) {
				return
			}
			if err := operation(); err != nil {
				report.Add(Operation{Scope: scope.Name, Action: "write " + kind, Status: Failed, Err: err})
			} else {
				report.Add(Operation{Scope: scope.Name, Action: "write " + kind, Status: Completed})
			}
		}
		write("keybindings", plan.Inheritance.Keybindings, plan.Inheritance.Unmanaged["keybindings"], plan.Keybindings, func() error { return artifacts.WriteKeybindings(ctx, scope, plan.Keybindings) })
		write("tasks", plan.Inheritance.Tasks, plan.Inheritance.Unmanaged["tasks"], plan.Tasks, func() error { return artifacts.WriteTasks(ctx, scope, plan.Tasks) })
		write("mcp", plan.Inheritance.MCP, plan.Inheritance.Unmanaged["mcp"], plan.MCP, func() error { return artifacts.WriteMCP(ctx, scope, plan.MCP) })
		write("snippets", plan.Inheritance.Snippets, plan.Inheritance.Unmanaged["snippets"], plan.Snippets, func() error { return artifacts.WriteSnippets(ctx, scope, plan.Snippets) })
	}
	if !plan.Inheritance.Unmanaged["extensions"] {
		Extensions(ctx, runtime, scope, plan.Extensions, MarketplacePolicy{Platform: platform, Allowed: marketplace, Pool: pool, Force: forceExtensions}, report)
	}
}

func nilArtifact(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Map, reflect.Slice, reflect.Pointer, reflect.Interface:
		return reflected.IsNil()
	default:
		return false
	}
}
