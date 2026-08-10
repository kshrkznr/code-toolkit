package recovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/kshrkznr/code-toolkit/go/internal/converge"
	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type Difference struct {
	Phase    string `json:"phase"`
	Kind     string `json:"kind"`
	Scope    string `json:"scope,omitempty"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
	Risk     string `json:"risk"`
}

type Verification struct {
	Differences []Difference `json:"differences"`
}

func (v Verification) Matches() bool { return len(v.Differences) == 0 }

type Prepared struct {
	RecipePath string
	Recipe     recipe.Recipe
	Source     runtimelock.Snapshot
	Plan       cookbook.Plan
	Before     Verification
}

func Prepare(lockRoot string) (Prepared, error) {
	recipePath := filepath.Join(lockRoot, "recipe.yaml")
	definition, err := recipe.Load(recipePath)
	if err != nil {
		return Prepared{}, fmt.Errorf("load trusted Lock Recipe: %w", err)
	}
	skeleton := planSkeleton(definition)
	source, _, err := runtimelock.Read(lockRoot, skeleton)
	if err != nil {
		return Prepared{}, err
	}
	prepared := Prepared{RecipePath: recipePath, Recipe: definition, Source: source}
	prepared.Plan = recoveryPlan(definition, source)
	prepared.Before = compareRecipeAndSource(definition, source)
	return prepared, nil
}

func planSkeleton(definition recipe.Recipe) cookbook.Plan {
	plan := cookbook.Plan{Name: definition.Name, OS: definition.OS, Platform: definition.Platform, ExtensionMarketplace: definition.ExtensionMarketplace(), ExtensionPool: definition.ExtensionPoolMode(), LockMode: "refresh"}
	for _, name := range definition.Profile {
		strategy := definition.ProfileContent(name)
		plan.Profiles = append(plan.Profiles, cookbook.ScopePlan{Name: name, Inheritance: inheritance(strategy)})
	}
	return plan
}

func recoveryPlan(definition recipe.Recipe, source runtimelock.Snapshot) cookbook.Plan {
	plan := planSkeleton(definition)
	plan.Default = scopePlan(source.Default)
	profiles := snapshotProfiles(source)
	for index := range plan.Profiles {
		observed, ok := profiles[plan.Profiles[index].Name]
		if ok {
			plan.Profiles[index] = scopePlan(observed)
			plan.Profiles[index].Name = definition.Profile[index]
		} else {
			plan.Profiles[index].Settings = settings.Document{}
			plan.Profiles[index].Extensions = []string{}
		}
	}
	return plan
}

func scopePlan(snapshot runtimelock.ScopeSnapshot) cookbook.ScopePlan {
	ids := make([]string, 0, len(snapshot.Extensions))
	for _, extension := range snapshot.Extensions {
		ids = append(ids, extension.ID)
	}
	sort.Strings(ids)
	return cookbook.ScopePlan{Name: snapshot.Name, Settings: snapshot.Settings, Keybindings: snapshot.Keybindings, Tasks: snapshot.Tasks, MCP: snapshot.MCP, Snippets: snapshot.Snippets, Extensions: ids, Inheritance: snapshot.Inheritance}
}

func inheritance(strategy recipe.ProfileStrategy) cookbook.Inheritance {
	unmanaged := map[string]bool{}
	for _, kind := range []string{"settings", "keybindings", "tasks", "mcp", "snippets"} {
		if strategy.Content(kind) == "unmanaged" {
			unmanaged[kind] = true
		}
	}
	if len(unmanaged) == 0 {
		unmanaged = nil
	}
	return cookbook.Inheritance{Settings: strategy.Settings == "default", Keybindings: strategy.Keybindings == "default", Tasks: strategy.Tasks == "default", MCP: strategy.MCP == "default", Snippets: strategy.Snippets == "default", Unmanaged: unmanaged}
}

type PoolUpdater interface {
	Update(context.Context, string, runtimelock.Snapshot, *converge.Report)
}

type Result struct {
	Operations converge.Report
	Before     Verification
	After      Verification
	Fresh      runtimelock.Snapshot
}

type Service struct {
	Pool       converge.ArtifactResolver
	PoolUpdate PoolUpdater
	Locks      runtimelock.Store
}

func (s Service) Recover(ctx context.Context, prepared Prepared, targetPath string, runtime runtimeio.Runtime) (Result, error) {
	return s.RecoverAt(ctx, prepared, targetPath, filepath.Join(targetPath, ".ext"), runtime)
}

// RecoverAt separates the Lock/diagnostic workspace from the physical
// Extension area. Deactivation uses this to reconstruct extensions at their
// final host path, avoiding location-sensitive Platform metadata drift.
func (s Service) RecoverAt(ctx context.Context, prepared Prepared, targetPath, extensionArea string, runtime runtimeio.Runtime) (Result, error) {
	result := Result{Before: prepared.Before}
	if err := requireEmptyExtensions(extensionArea); err != nil {
		return result, err
	}
	if err := copyProvenance(prepared.RecipePath, filepath.Join(targetPath, ".meta", "recipe.yaml")); err != nil {
		return result, err
	}
	result.Operations = converge.Plan(ctx, runtime, prepared.Plan, s.Pool, false)
	if err := result.Operations.Error(); err != nil {
		return result, err
	}
	fresh, err := s.Locks.Refresh(ctx, targetPath, prepared.RecipePath, runtime, prepared.Plan)
	if err != nil {
		return result, err
	}
	result.Fresh = fresh
	result.After = Compare(prepared.Plan, prepared.Source, fresh)
	if s.PoolUpdate != nil && prepared.Plan.ExtensionPool == "refresh" {
		s.PoolUpdate.Update(ctx, prepared.Plan.Platform, fresh, &result.Operations)
	}
	return result, nil
}

func copyProvenance(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read Recovery provenance: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create Recovery provenance directory: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(target), ".recipe-*")
	if err != nil {
		return fmt.Errorf("create Recovery provenance staging: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("write Recovery provenance: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close Recovery provenance: %w", err)
	}
	if err := os.Rename(path, target); err != nil {
		return fmt.Errorf("publish Recovery provenance: %w", err)
	}
	return nil
}

func requireEmptyExtensions(path string) error {
	entries, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}
	if err != nil {
		return fmt.Errorf("inspect Recovery Extension area: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("Recovery Extension area must be empty: %s", path)
	}
	return nil
}

func compareRecipeAndSource(definition recipe.Recipe, source runtimelock.Snapshot) Verification {
	result := Verification{}
	recipeProfiles := stringSet(definition.Profile)
	sourceProfiles := map[string]runtimelock.ScopeSnapshot{}
	for _, profile := range source.Profiles {
		sourceProfiles[profile.Name] = profile
	}
	for _, name := range sortedSetDifference(recipeProfiles, keys(sourceProfiles)) {
		result.Differences = append(result.Differences, Difference{Phase: "before", Kind: "profile-missing-from-lock", Scope: name, Expected: "present", Actual: "missing", Risk: "Profile state may be empty or lost"})
	}
	for _, name := range sortedSetDifference(keys(sourceProfiles), recipeProfiles) {
		result.Differences = append(result.Differences, Difference{Phase: "before", Kind: "profile-missing-from-recipe", Scope: name, Expected: "selected", Actual: "not selected", Risk: "Profile will not be recovered and may be lost"})
	}
	for _, name := range definition.Profile {
		observed, ok := sourceProfiles[name]
		if !ok {
			continue
		}
		expected := inheritance(definition.ProfileContent(name))
		if !reflect.DeepEqual(expected, observed.Inheritance) {
			result.Differences = append(result.Differences, Difference{Phase: "before", Kind: "inheritance", Scope: name, Expected: jsonValue(expected), Actual: jsonValue(observed.Inheritance), Risk: "Profile content ownership may change"})
		}
	}
	return result
}

func Compare(plan cookbook.Plan, source, recovered runtimelock.Snapshot) Verification {
	result := Verification{}
	expectedProfiles := map[string]runtimelock.ScopeSnapshot{}
	sourceProfiles := snapshotProfiles(source)
	for _, profile := range plan.Profiles {
		if value, ok := sourceProfiles[profile.Name]; ok {
			expectedProfiles[profile.Name] = value
		} else {
			expectedProfiles[profile.Name] = planSnapshot(profile)
		}
	}
	actualProfiles := snapshotProfiles(recovered)
	for _, name := range sortedSetDifference(keys(expectedProfiles), keys(actualProfiles)) {
		result.Differences = append(result.Differences, Difference{Phase: "after", Kind: "profile-missing", Scope: name, Expected: "present", Actual: "missing", Risk: "Recovered Runtime is missing Profile state"})
	}
	for _, name := range sortedSetDifference(keys(actualProfiles), keys(expectedProfiles)) {
		result.Differences = append(result.Differences, Difference{Phase: "after", Kind: "unexpected-profile", Scope: name, Expected: "absent", Actual: "present", Risk: "Recovered Runtime contains unexpected Profile state"})
	}
	compareScope(&result, "", source.Default, recovered.Default)
	for _, name := range sortedKeys(expectedProfiles) {
		actual, ok := actualProfiles[name]
		if ok {
			compareScope(&result, name, expectedProfiles[name], actual)
		}
	}
	return result
}

func compareScope(result *Verification, scope string, expected, actual runtimelock.ScopeSnapshot) {
	if !reflect.DeepEqual(expected.Settings, actual.Settings) {
		result.Differences = append(result.Differences, Difference{Phase: "after", Kind: "settings", Scope: scope, Expected: jsonValue(expected.Settings), Actual: jsonValue(actual.Settings), Risk: "Settings may be replaced or lost"})
	}
	if !reflect.DeepEqual(expected.Inheritance, actual.Inheritance) {
		result.Differences = append(result.Differences, Difference{Phase: "after", Kind: "inheritance", Scope: scope, Expected: jsonValue(expected.Inheritance), Actual: jsonValue(actual.Inheritance), Risk: "Profile content ownership may change"})
	}
	for _, artifact := range []struct {
		name             string
		expected, actual any
	}{
		{"keybindings", expected.Keybindings, actual.Keybindings}, {"tasks", expected.Tasks, actual.Tasks}, {"mcp", expected.MCP, actual.MCP}, {"snippets", expected.Snippets, actual.Snippets},
	} {
		equal := reflect.DeepEqual(artifact.expected, artifact.actual)
		if artifact.name == "tasks" {
			equal = runtimeartifact.TasksEqual(expected.Tasks, actual.Tasks)
		}
		if !equal {
			result.Differences = append(result.Differences, Difference{Phase: "after", Kind: artifact.name, Scope: scope, Expected: jsonValue(artifact.expected), Actual: jsonValue(artifact.actual), Risk: "Runtime Artifact may be replaced or lost"})
		}
	}
	expectedIDs, actualIDs := extensionIDs(expected.Extensions), extensionIDs(actual.Extensions)
	if !reflect.DeepEqual(expectedIDs, actualIDs) {
		result.Differences = append(result.Differences, Difference{Phase: "after", Kind: "extensions", Scope: scope, Expected: strings.Join(expectedIDs, ","), Actual: strings.Join(actualIDs, ","), Risk: "Extension functionality or state may be unavailable"})
	}
}

func snapshotProfiles(snapshot runtimelock.Snapshot) map[string]runtimelock.ScopeSnapshot {
	result := map[string]runtimelock.ScopeSnapshot{}
	for _, profile := range snapshot.Profiles {
		result[profile.Name] = profile
	}
	return result
}
func planSnapshot(plan cookbook.ScopePlan) runtimelock.ScopeSnapshot {
	extensions := make([]runtimeio.Extension, len(plan.Extensions))
	for i, id := range plan.Extensions {
		extensions[i] = runtimeio.Extension{ID: id}
	}
	return runtimelock.ScopeSnapshot{Name: plan.Name, Settings: plan.Settings, Keybindings: plan.Keybindings, Tasks: plan.Tasks, MCP: plan.MCP, Snippets: plan.Snippets, Extensions: extensions, Inheritance: plan.Inheritance}
}
func extensionIDs(extensions []runtimeio.Extension) []string {
	result := make([]string, len(extensions))
	for i, extension := range extensions {
		result[i] = extension.ID
	}
	sort.Strings(result)
	return result
}
func stringSet(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}
func keys[T any](values map[string]T) map[string]bool {
	result := map[string]bool{}
	for key := range values {
		result[key] = true
	}
	return result
}
func sortedSetDifference(left, right map[string]bool) []string {
	var result []string
	for value := range left {
		if !right[value] {
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
func sortedKeys[T any](values map[string]T) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func jsonValue(value any) string { data, _ := json.Marshal(value); return string(data) }
