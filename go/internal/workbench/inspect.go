package workbench

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/distribution"
	"code-toolkit/internal/flatformat"
	"code-toolkit/internal/recipe"
	"code-toolkit/internal/runtimeio"
	"code-toolkit/internal/runtimelock"
	"code-toolkit/internal/settings"
)

type CompletedSource struct {
	Kind, Name, RecipePath string
	Plan                   cookbook.Plan
	Snapshot               runtimelock.Snapshot
}

func (s Service) RecipeSource(path string) (CompletedSource, error) {
	plan, err := (cookbook.Repository{Root: filepath.Join(s.CookbookRoot, "ingredient")}).Resolve(path)
	if err != nil {
		return CompletedSource{}, err
	}
	return CompletedSource{Kind: "recipe", Name: plan.Name, RecipePath: path, Plan: plan, Snapshot: snapshotFromPlan(plan)}, nil
}

func (s Service) DistributionSource(ctx context.Context, dist distribution.Distribution) (CompletedSource, error) {
	recipePath := filepath.Join(dist.Path, ".meta", "recipe.yaml")
	plan := observationPlan(dist.Recipe, recipePath)
	snapshot, _, err := s.snapshot(ctx, dist, recipePath, plan)
	if err != nil {
		return CompletedSource{}, err
	}
	return CompletedSource{Kind: "dist", Name: dist.Name, RecipePath: recipePath, Plan: planFromSnapshot(plan, snapshot), Snapshot: snapshot}, nil
}

func planFromSnapshot(identity cookbook.Plan, snapshot runtimelock.Snapshot) cookbook.Plan {
	convert := func(scope runtimelock.ScopeSnapshot) cookbook.ScopePlan {
		extensions := make([]string, 0, len(scope.Extensions))
		for _, extension := range scope.Extensions {
			extensions = append(extensions, extension.ID)
		}
		return cookbook.ScopePlan{Name: scope.Name, Settings: scope.Settings, Keybindings: scope.Keybindings, Tasks: scope.Tasks, MCP: scope.MCP, Snippets: scope.Snippets, Extensions: extensions, Inheritance: scope.Inheritance}
	}
	plan := identity
	plan.Default = convert(snapshot.Default)
	plan.Profiles = nil
	for _, scope := range snapshot.Profiles {
		plan.Profiles = append(plan.Profiles, convert(scope))
	}
	return plan
}

func snapshotFromPlan(plan cookbook.Plan) runtimelock.Snapshot {
	convert := func(scope cookbook.ScopePlan) runtimelock.ScopeSnapshot {
		extensions := make([]runtimeio.Extension, 0, len(scope.Extensions))
		for _, id := range scope.Extensions {
			extensions = append(extensions, runtimeio.Extension{ID: id})
		}
		value := scope.Settings
		if scope.Name != "" && scope.Inheritance.Settings {
			value = settings.Document{}
		}
		return runtimelock.ScopeSnapshot{Name: scope.Name, Settings: value, Keybindings: scope.Keybindings, Tasks: scope.Tasks, MCP: scope.MCP, Snippets: scope.Snippets, Extensions: extensions, Inheritance: scope.Inheritance}
	}
	snapshot := runtimelock.Snapshot{RecipeName: plan.Name, Platform: plan.Platform, ObservedAt: time.Now(), Default: convert(plan.Default)}
	for _, scope := range plan.Profiles {
		snapshot.Profiles = append(snapshot.Profiles, convert(scope))
	}
	return snapshot
}

func (s Service) GenerateRecipeView(source CompletedSource, conflict string) (Result, error) {
	target := filepath.Join(s.CookbookRoot, "inspect", "recipe."+safeName(source.Name))
	staging, err := s.inspectStaging(target, conflict, ".view-recipe-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(staging)
	settingsText, settingsResult, err := renderSettings(nil, source.Snapshot)
	if err != nil {
		return Result{}, err
	}
	extensionsText, extensionsResult := renderExtensions(nil, source.Snapshot)
	runtimeArtifacts, runtimeResults, err := renderRuntimeArtifacts(nil, source.Snapshot)
	if err != nil {
		return Result{}, err
	}
	recipeData, err := os.ReadFile(source.RecipePath)
	if err != nil {
		return Result{}, err
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	source.RecipePath = s.displayPath(source.RecipePath)
	artifacts := map[string][]byte{"settings.draft.md": []byte(settingsText), "extensions.draft.md": []byte(extensionsText), "recipe.draft.yaml": recipeData, "summary.md": []byte(inspectSummary("Recipe View", source, settingsResult, extensionsResult, runtimeResults, now))}
	for name, data := range runtimeArtifacts {
		artifacts[name] = data
	}
	if err := writeArtifacts(staging, artifacts); err != nil {
		return Result{}, err
	}
	if err := publish(staging, target, false); err != nil {
		return Result{}, err
	}
	return Result{Path: target, Snapshot: source.Snapshot}, nil
}

func (s Service) GenerateIngredientView(query, conflict string) (Result, error) {
	paths, err := ingredientResources(filepath.Join(s.CookbookRoot, "ingredient"), query)
	if err != nil {
		return Result{}, err
	}
	if len(paths) == 0 {
		return Result{}, fmt.Errorf("Ingredient Resource not found: %s", query)
	}
	target := filepath.Join(s.CookbookRoot, "inspect", "ingredient."+safeName(query))
	staging, err := s.inspectStaging(target, conflict, ".view-ingredient-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(staging)
	var settingsOut, extensionsOut strings.Builder
	artifactOut := map[string]*strings.Builder{}
	for _, kind := range []string{"keybindings", "tasks", "mcp", "snippets"} {
		builder := &strings.Builder{}
		fmt.Fprintf(builder, "# %s View\n\n## Inventory\n", strings.ToUpper(kind[:1])+kind[1:])
		artifactOut[kind] = builder
	}
	settingsOut.WriteString("# Settings View\n\n## Inventory\n")
	extensionsOut.WriteString("# Extensions View\n\n## Inventory\n")
	settingsCount, extensionCount := 0, 0
	root := filepath.Join(s.CookbookRoot, "ingredient")
	resourceNames := make([]string, 0, len(paths))
	for _, path := range paths {
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		resourceNames = append(resourceNames, relative)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return Result{}, readErr
		}
		base := filepath.Base(path)
		if isSettingsName(base) {
			document, parseErr := settings.Parse(data)
			if parseErr != nil {
				return Result{}, fmt.Errorf("parse %s: %w", relative, parseErr)
			}
			encoded, encodeErr := flatformat.Encode(map[string]any(document))
			if encodeErr != nil {
				return Result{}, encodeErr
			}
			fmt.Fprintf(&settingsOut, "\n### %s\n\n```diff\n", relative)
			for _, line := range strings.Split(strings.TrimSuffix(string(encoded), "\n"), "\n") {
				if line != "" {
					settingsOut.WriteString("+ " + line + "\n")
					settingsCount++
				}
			}
			settingsOut.WriteString("```\n")
		} else if isExtensionName(base) {
			fmt.Fprintf(&extensionsOut, "\n### %s\n\n```diff\n", relative)
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				id := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
				if id != "" {
					extensionsOut.WriteString("+ " + id + "\n")
					extensionCount++
				}
			}
			extensionsOut.WriteString("```\n")
		} else if isRuntimeArtifactName(base) || strings.Contains(relative, "/snippets/") {
			kind := "snippets"
			for _, candidate := range []string{"keybindings", "tasks", "mcp"} {
				if strings.Contains(base, candidate) {
					kind = candidate
					break
				}
			}
			fmt.Fprintf(artifactOut[kind], "\n### %s\n\n```diff\n", relative)
			for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
				artifactOut[kind].WriteString("+ " + line + "\n")
			}
			artifactOut[kind].WriteString("```\n")
		}
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	identities := resolvedIngredientIdentities(s.CookbookRoot, resourceNames)
	artifacts := map[string][]byte{"summary.md": []byte(ingredientSummary(query, s.displayPath(root), resourceNames, identities, settingsCount, extensionCount, now))}
	if settingsCount > 0 {
		artifacts["settings.draft.md"] = []byte(settingsOut.String())
	}
	if extensionCount > 0 {
		artifacts["extensions.draft.md"] = []byte(extensionsOut.String())
	}
	for kind, output := range artifactOut {
		if strings.Contains(output.String(), "### ") {
			artifacts[kind+".draft.md"] = []byte(output.String())
		}
	}
	if err := writeArtifacts(staging, artifacts); err != nil {
		return Result{}, err
	}
	if err := publish(staging, target, false); err != nil {
		return Result{}, err
	}
	return Result{Path: target}, nil
}

func ingredientSummary(query, root string, resources, identities []string, settingsCount, extensionCount int, generated time.Time) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# Ingredient View Summary\n\n> Source: `%s`\n\n## Result\n\n- Resources: %d\n- Settings assignments: %d\n- Extensions: %d\n\n## Resolved Ingredient Resources\n\n", query, len(resources), settingsCount, extensionCount)
	writeMarkdownList(&output, resources)
	output.WriteString("\n## Resolved Ingredient\n\n")
	writeMarkdownList(&output, identities)
	output.WriteString("\nThese identities are Recipe authoring hints inferred from Resource names; they do not imply Recipe selection.\n")
	fmt.Fprintf(&output, "\n## Resolution\n\n- Query scope: `%s`\n- Ingredient root: `%s`\n- Resolution: raw Resource Inventory; Recipe selection and Settings merge are not applied\n- Generated: `%s`\n", ingredientQueryScope(query), root, generated.Format(time.RFC3339))
	return output.String()
}

func resolvedIngredientIdentities(cookbookRoot string, resources []string) []string {
	variants := map[string]bool{}
	paths, _ := filepath.Glob(filepath.Join(cookbookRoot, "recipe", "*.yaml"))
	for _, path := range paths {
		definition, err := recipe.Load(path)
		if err == nil {
			variants[definition.OS], variants[definition.Platform] = true, true
		}
	}
	set := map[string]bool{}
	for _, resource := range resources {
		parts := strings.Split(filepath.ToSlash(resource), "/")
		var layer, name string
		if len(parts) >= 3 {
			layer, name = parts[0], parts[1]
		} else if len(parts) == 2 {
			layer, name = parts[0], ingredientNameStem(parts[1], variants)
		} else if first, rest, ok := strings.Cut(parts[0], "."); ok {
			layer, name = first, ingredientNameStem(rest, variants)
		}
		if layer != "" && name != "" {
			set[layer+"."+name] = true
		}
	}
	return sortedKeys(set)
}
func ingredientNameStem(name string, variants map[string]bool) string {
	if index := strings.Index(name, ".snippets."); index >= 0 {
		name = name[:index]
	}
	for _, suffix := range []string{".extensions", ".settings.jsonc", ".settings.json", ".keybindings.jsonc", ".keybindings.json", ".tasks.jsonc", ".tasks.json", ".mcp.jsonc", ".mcp.json"} {
		name = strings.TrimSuffix(name, suffix)
	}
	parts := strings.Split(name, ".")
	if len(parts) > 1 && variants[parts[len(parts)-1]] {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, ".")
}
func ingredientQueryScope(query string) string {
	if query == "all" {
		return "all"
	}
	if strings.Contains(query, ".") {
		return "ingredient"
	}
	return "layer"
}

func (s Service) GenerateSync(left, right CompletedSource, conflict string) (Result, error) {
	target := filepath.Join(s.CookbookRoot, "inspect", "sync."+safeName(left.Kind+"-"+left.Name)+"."+safeName(right.Kind+"-"+right.Name))
	staging, err := s.inspectStaging(target, conflict, ".sync-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(staging)
	settingsText, settingsResult, err := renderSettings(&left.Plan, right.Snapshot)
	if err != nil {
		return Result{}, err
	}
	extensionsText, extensionsResult := renderExtensions(&left.Plan, right.Snapshot)
	runtimeArtifacts, runtimeResults, err := renderRuntimeArtifacts(&left.Plan, right.Snapshot)
	if err != nil {
		return Result{}, err
	}
	leftRecipe, _ := os.ReadFile(left.RecipePath)
	rightRecipe, _ := os.ReadFile(right.RecipePath)
	recipeStatus := "SAME"
	recipeDiff := "No Recipe differences.\n"
	if string(leftRecipe) != string(rightRecipe) {
		recipeStatus = "DIFFERENT"
		recipeDiff = fencedDiff(leftRecipe, rightRecipe)
	}
	summary := fmt.Sprintf("# Sync Summary\n\n> Left: `%s:%s`  \n> Right: `%s:%s`\n\n## Result\n\n- Recipe: %s\n- Settings: %s (+%d/-%d/~%d)\n- Extensions: %s (+%d/-%d/~%d)\n\n## Recipe Difference\n\n%s", left.Kind, left.Name, right.Kind, right.Name, recipeStatus, settingsResult.Status, settingsResult.Counts.Added, settingsResult.Counts.Removed, settingsResult.Counts.Changed, extensionsResult.Status, extensionsResult.Counts.Added, extensionsResult.Counts.Removed, extensionsResult.Counts.Changed, recipeDiff)
	for _, name := range sortedArtifactNames(runtimeResults) {
		summary += fmt.Sprintf("- %s: %s\n", strings.ToUpper(name[:1])+name[1:], runtimeResults[name].Status)
	}
	artifacts := map[string][]byte{"summary.md": []byte(summary), "recipe.draft.yaml": rightRecipe}
	for name, data := range runtimeArtifacts {
		artifacts[name] = data
	}
	if settingsResult.Status != "SAME" {
		artifacts["settings.draft.md"] = []byte(settingsText)
	}
	if extensionsResult.Status != "SAME" {
		artifacts["extensions.draft.md"] = []byte(extensionsText)
	}
	if err := writeArtifacts(staging, artifacts); err != nil {
		return Result{}, err
	}
	if err := publish(staging, target, false); err != nil {
		return Result{}, err
	}
	return Result{Path: target, Snapshot: right.Snapshot}, nil
}

func (s Service) inspectStaging(target, conflict, prefix string) (string, error) {
	if _, err := os.Lstat(target); err == nil && conflict != "replace" {
		return "", fmt.Errorf("Workbench already exists: %s", target)
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return "", err
	}
	return os.MkdirTemp(s.CookbookRoot, prefix)
}
func writeArtifacts(root string, artifacts map[string][]byte) error {
	for name, data := range artifacts {
		if err := os.WriteFile(filepath.Join(root, name), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
func inspectSummary(title string, source CompletedSource, settingsResult, extensionsResult artifactResult, runtimeResults map[string]artifactResult, generated time.Time) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# %s Summary\n\n> Source: `%s`  \n> Recipe: `%s`  \n> OS: `%s`  \n> Platform: `%s`\n\n", title, source.Name, source.Plan.Name, source.Plan.OS, source.Plan.Platform)
	output.WriteString("## Result\n\n| Artifact | Status | Added | Removed | Changed | Review |\n| --- | --- | ---: | ---: | ---: | --- |\n")
	output.WriteString("| Recipe | INVENTORY | - | - | - | [recipe.draft.yaml](recipe.draft.yaml) |\n")
	fmt.Fprintf(&output, "| Settings | INVENTORY | %d | 0 | 0 | [settings.draft.md](settings.draft.md) |\n", settingsResult.Counts.Added)
	fmt.Fprintf(&output, "| Extensions | INVENTORY | %d | 0 | 0 | [extensions.draft.md](extensions.draft.md) |\n", extensionsResult.Counts.Added)
	for _, name := range sortedArtifactNames(runtimeResults) {
		result := runtimeResults[name]
		fmt.Fprintf(&output, "| %s | INVENTORY | %d | 0 | %d | [%s.draft.md](%s.draft.md) |\n", strings.ToUpper(name[:1])+name[1:], result.Counts.Added, result.Counts.Changed, name, name)
	}
	output.WriteString("\n## Profiles\n\n| Profile | Settings | Extensions | Inherits Settings |\n| --- | ---: | ---: | --- |\n")
	fmt.Fprintf(&output, "| default | %d | %d | - |\n", len(source.Plan.Default.Settings), len(source.Plan.Default.Extensions))
	for _, scope := range source.Plan.Profiles {
		fmt.Fprintf(&output, "| %s | %d | %d | %t |\n", scope.Name, len(scope.Settings), len(scope.Extensions), scope.Inheritance.Settings)
	}
	extensions := map[string]bool{}
	for _, id := range source.Plan.Default.Extensions {
		extensions[id] = true
	}
	for _, scope := range source.Plan.Profiles {
		for _, id := range scope.Extensions {
			extensions[id] = true
		}
	}
	output.WriteString("\n## Extensions Used by Recipe\n\n")
	writeMarkdownList(&output, sortedKeys(extensions))
	output.WriteString("\n## Resolved Ingredient Resources\n\n")
	resources := resolvedResources(source.Plan)
	if len(resources) == 0 {
		output.WriteString("- None\n")
	} else {
		for _, item := range resources {
			fmt.Fprintf(&output, "- %s\n", item.Name)
		}
	}
	fmt.Fprintf(&output, "\n## Resolution\n\n- Recipe source: `%s`\n- Resolver: Cookbook Resolver with declared Go Merge Rules\n- Generated: `%s`\n", source.RecipePath, generated.Format(time.RFC3339))
	return output.String()
}

type resolvedResource struct {
	Name, Layer, Ingredient, Variant string
	Scopes                           []string
}

func resolvedResources(plan cookbook.Plan) []resolvedResource {
	type value struct {
		source cookbook.Source
		scopes map[string]bool
	}
	found := map[string]*value{}
	collect := func(scope string, sources []cookbook.Source) {
		for _, source := range sources {
			key := source.Path
			if found[key] == nil {
				found[key] = &value{source: source, scopes: map[string]bool{}}
			}
			found[key].scopes[scope] = true
		}
	}
	collect("default", plan.Default.Sources)
	for _, scope := range plan.Profiles {
		collect(scope.Name, scope.Sources)
	}
	keys := make([]string, 0, len(found))
	for key := range found {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]resolvedResource, 0, len(keys))
	for _, key := range keys {
		item := found[key]
		scopes := sortedKeys(item.scopes)
		result = append(result, resolvedResource{Name: ingredientResourceName(item.source.Path), Layer: item.source.Layer, Ingredient: item.source.Ingredient, Variant: item.source.Variant, Scopes: scopes})
	}
	return result
}
func ingredientResourceName(path string) string {
	slash := filepath.ToSlash(path)
	if index := strings.LastIndex(slash, "/ingredient/"); index >= 0 {
		return slash[index+len("/ingredient/"):]
	}
	return filepath.Base(path)
}
func safeName(value string) string {
	value = strings.ReplaceAll(filepath.ToSlash(value), "/", "-")
	value = strings.ReplaceAll(value, "..", "-")
	return value
}
func isSettingsName(name string) bool {
	return name == "settings.json" || name == "settings.jsonc" || strings.Contains(name, "settings.json")
}
func isExtensionName(name string) bool {
	return name == "extensions" || strings.HasSuffix(name, ".extensions")
}
func isRuntimeArtifactName(name string) bool {
	for _, kind := range []string{"keybindings", "tasks", "mcp", "snippets"} {
		if name == kind+".json" || name == kind+".jsonc" || strings.Contains(name, "."+kind+".") {
			return true
		}
	}
	return strings.HasSuffix(name, ".code-snippets")
}
func ingredientResources(root, query string) ([]string, error) {
	if query == "" {
		return nil, fmt.Errorf("Ingredient query is required")
	}
	parts := strings.SplitN(query, ".", 2)
	if query != "all" && (parts[0] == "" || (len(parts) == 2 && parts[1] == "")) {
		return nil, fmt.Errorf("invalid Ingredient query: %s", query)
	}
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		slash := filepath.ToSlash(relative)
		matched := query == "all"
		if !matched && len(parts) == 1 {
			matched = strings.HasPrefix(slash, parts[0]+".") || strings.HasPrefix(slash, parts[0]+"/")
		}
		if !matched && len(parts) == 2 {
			matched = strings.HasPrefix(slash, query+".") || strings.HasPrefix(slash, parts[0]+"/"+parts[1]+".") || strings.HasPrefix(slash, parts[0]+"/"+parts[1]+"/")
		}
		if matched && (isSettingsName(entry.Name()) || isExtensionName(entry.Name()) || isRuntimeArtifactName(entry.Name()) || strings.Contains(slash, "/snippets/")) {
			paths = append(paths, path)
		}
		return nil
	})
	if os.IsNotExist(err) {
		err = nil
	}
	sort.Strings(paths)
	return paths, err
}
