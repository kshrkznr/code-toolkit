package workbench

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/flatformat"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
)

type Kind string

const (
	FreezeDraft Kind = "Freeze Draft"
	View        Kind = "View"
)

type RuntimeFactory func(distribution.Distribution) (runtimeio.Runtime, error)

type Service struct {
	WorkspaceRoot string
	CookbookRoot  string
	WorkbenchRoot string
	Runtime       RuntimeFactory
	Locks         runtimelock.Store
	ChooseLock    func() (string, error)
	Now           func() time.Time
}

type Result struct {
	Path     string
	Snapshot runtimelock.Snapshot
}

func (s Service) displayPath(path string) string {
	workspaceRoot := s.WorkspaceRoot
	if workspaceRoot == "" {
		workspaceRoot = filepath.Dir(filepath.Clean(s.workbenchRoot()))
	}
	workspaceRoot, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return path
	}
	sourcePath := path
	if !filepath.IsAbs(sourcePath) {
		sourcePath, err = filepath.Abs(sourcePath)
		if err != nil {
			return path
		}
	}
	relative, err := filepath.Rel(workspaceRoot, filepath.Clean(sourcePath))
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return path
	}
	if relative == "." {
		return "$CTK_HOME"
	}
	return "$CTK_HOME/" + filepath.ToSlash(relative)
}

func (s Service) cookbookRoot() string { return s.CookbookRoot }

func (s Service) workbenchRoot() string {
	if s.WorkbenchRoot != "" {
		return s.WorkbenchRoot
	}
	return s.CookbookRoot
}

type Counts struct{ Added, Removed, Changed int }

type artifactResult struct {
	Status string
	Counts Counts
}

func (s Service) Generate(ctx context.Context, kind Kind, dist distribution.Distribution, conflict string) (Result, error) {
	directory := filepath.Join("inspect", "dist."+dist.Name)
	if kind == FreezeDraft {
		directory = "draft"
	}
	target := filepath.Join(s.workbenchRoot(), directory)
	if _, err := os.Lstat(target); err == nil && conflict != "replace" {
		return Result{}, fmt.Errorf("Workbench already exists: %s", target)
	} else if err != nil && !os.IsNotExist(err) {
		return Result{}, fmt.Errorf("inspect Workbench: %w", err)
	}

	recipePath := filepath.Join(dist.Path, ".meta", "recipe.yaml")
	var plan cookbook.Plan
	if kind == FreezeDraft {
		repository := cookbook.Repository{Root: filepath.Join(s.cookbookRoot(), "ingredient")}
		var err error
		plan, err = repository.Resolve(recipePath)
		if err != nil {
			return Result{}, err
		}
	} else {
		plan = observationPlan(dist.Recipe, recipePath)
	}
	snapshot, effectiveMode, err := s.snapshot(ctx, dist, recipePath, plan)
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return Result{}, fmt.Errorf("create Workbench parent: %w", err)
	}
	stagingPrefix := ".draft.staging-"
	if kind == View {
		stagingPrefix = ".view-dist." + dist.Name + ".staging-"
	}
	staging, err := os.MkdirTemp(s.workbenchRoot(), stagingPrefix)
	if err != nil {
		return Result{}, fmt.Errorf("create Workbench staging: %w", err)
	}
	defer os.RemoveAll(staging)

	recipeStatus, recipeDiff := "INVENTORY", ""
	if kind == FreezeDraft {
		_, recipeStatus, recipeDiff = findCurrentRecipe(s.cookbookRoot(), dist.Recipe, recipePath)
	}
	comparisonPlan := &plan
	if kind == View {
		comparisonPlan = nil
	}
	settingsText, settingsResult, err := renderSettings(comparisonPlan, snapshot)
	if err != nil {
		return Result{}, err
	}
	extensionsText, extensionsResult := renderExtensions(comparisonPlan, snapshot)
	runtimeArtifacts, runtimeResults, err := renderRuntimeArtifacts(comparisonPlan, snapshot)
	if err != nil {
		return Result{}, err
	}
	var used, available []string
	if kind == FreezeDraft {
		var inventoryErr error
		used, available, inventoryErr = ingredientInventory(filepath.Join(s.cookbookRoot(), "ingredient"), plan, dist.Recipe)
		if inventoryErr != nil {
			return Result{}, inventoryErr
		}
		settingsText = insertInventory(settingsText, used, available, "settings")
		extensionsText = insertInventory(extensionsText, used, available, "extensions")
	}

	summary := s.renderSummary(kind, dist, plan, snapshot, effectiveMode, recipeStatus, recipeDiff, settingsResult, extensionsResult, runtimeResults, used)
	artifacts := map[string][]byte{"summary.md": []byte(summary)}
	for name, data := range runtimeArtifacts {
		artifacts[name] = data
	}
	if kind == View || settingsResult.Status != "SAME" {
		artifacts["settings.draft.md"] = []byte(settingsText)
	}
	if kind == View || extensionsResult.Status != "SAME" {
		artifacts["extensions.draft.md"] = []byte(extensionsText)
	}
	for name, data := range artifacts {
		if err := os.WriteFile(filepath.Join(staging, name), data, 0o644); err != nil {
			return Result{}, fmt.Errorf("write %s: %w", name, err)
		}
	}
	recipeData, err := os.ReadFile(recipePath)
	if err != nil {
		return Result{}, fmt.Errorf("read Recipe provenance: %w", err)
	}
	if err := os.WriteFile(filepath.Join(staging, "recipe.draft.yaml"), recipeData, 0o644); err != nil {
		return Result{}, fmt.Errorf("write Recipe Draft: %w", err)
	}
	if err := publish(staging, target, kind == FreezeDraft); err != nil {
		return Result{}, err
	}
	return Result{Path: target, Snapshot: snapshot}, nil
}

func observationPlan(definition recipe.Recipe, recipePath string) cookbook.Plan {
	plan := cookbook.Plan{
		RecipePath: recipePath, Name: definition.Name, OS: definition.OS,
		Platform: definition.Platform, LockMode: definition.LockMode(),
	}
	for _, name := range definition.Profile {
		plan.Profiles = append(plan.Profiles, cookbook.ScopePlan{Name: name})
	}
	return plan
}

func (s Service) snapshot(ctx context.Context, dist distribution.Distribution, recipePath string, plan cookbook.Plan) (runtimelock.Snapshot, string, error) {
	mode := plan.LockMode
	if mode == "ask" {
		if s.ChooseLock == nil {
			return runtimelock.Snapshot{}, mode, fmt.Errorf("lock-mode ask requires an interactive selector")
		}
		selected, err := s.ChooseLock()
		if err != nil {
			return runtimelock.Snapshot{}, mode, err
		}
		mode = selected
	}
	switch mode {
	case "refresh":
		runtime, err := s.Runtime(dist)
		if err != nil {
			return runtimelock.Snapshot{}, mode, err
		}
		snapshot, err := s.Locks.Refresh(ctx, dist.Path, recipePath, runtime, plan)
		return snapshot, mode, err
	case "reuse":
		snapshot, _, err := runtimelock.Read(filepath.Join(dist.Path, ".lock"), plan)
		return snapshot, mode, err
	case "abort":
		return runtimelock.Snapshot{}, mode, fmt.Errorf("Lock observation declined")
	default:
		return runtimelock.Snapshot{}, mode, fmt.Errorf("invalid lock-mode %q", mode)
	}
}

func renderSettings(plan *cookbook.Plan, snapshot runtimelock.Snapshot) (string, artifactResult, error) {
	var output strings.Builder
	if plan == nil {
		output.WriteString("# Settings View\n\n## Inventory\n")
	} else {
		output.WriteString("# Settings Draft\n\n## Inventory\n\n## Difference\n")
	}
	result := artifactResult{Status: "SAME"}
	planScopes := map[string]cookbook.ScopePlan{}
	if plan != nil {
		planScopes[""] = plan.Default
		for _, scope := range plan.Profiles {
			planScopes[scope.Name] = scope
		}
	}
	observed := append([]runtimelock.ScopeSnapshot{snapshot.Default}, snapshot.Profiles...)
	for _, scope := range observed {
		current := cookbook.ScopePlan{}
		if value, ok := planScopes[scope.Name]; ok {
			current = value
		}
		currentSettings := current.Settings
		if scope.Name != "" && current.Inheritance.Settings {
			currentSettings = nil
		}
		before, err := flatformat.Flatten(map[string]any(currentSettings))
		if err != nil {
			return "", result, err
		}
		after, err := flatformat.Flatten(map[string]any(scope.Settings))
		if err != nil {
			return "", result, err
		}
		lines, counts := assignmentDiff(before, after)
		result.Counts = addCounts(result.Counts, counts)
		if len(lines) == 0 {
			continue
		}
		result.Status = "DIFFERENT"
		name := "runtime.draft.settings.json"
		if scope.Name != "" {
			name = "profile." + scope.Name + ".draft.settings.json"
		}
		fmt.Fprintf(&output, "\n### %s\n", name)
		groups := groupAssignments(lines)
		for _, group := range sortedKeys(groups) {
			fmt.Fprintf(&output, "\n#### %s\n\n```diff\n%s```\n", group, strings.Join(groups[group], ""))
		}
	}
	if result.Status == "SAME" {
		output.WriteString("\nNo Settings differences.\n")
	}
	return output.String(), result, nil
}

func renderExtensions(plan *cookbook.Plan, snapshot runtimelock.Snapshot) (string, artifactResult) {
	var output strings.Builder
	if plan == nil {
		output.WriteString("# Extensions View\n\n## Inventory\n")
	} else {
		output.WriteString("# Extensions Draft\n\n## Inventory\n\n## Difference\n")
	}
	result := artifactResult{Status: "SAME"}
	planScopes := map[string][]string{}
	if plan != nil {
		planScopes[""] = plan.Default.Extensions
		for _, scope := range plan.Profiles {
			planScopes[scope.Name] = scope.Extensions
		}
	}
	observed := append([]runtimelock.ScopeSnapshot{snapshot.Default}, snapshot.Profiles...)
	for _, scope := range observed {
		current := append([]string(nil), planScopes[scope.Name]...)
		after := make([]string, 0, len(scope.Extensions))
		for _, extension := range scope.Extensions {
			after = append(after, extension.ID)
		}
		lines, counts := stringDiff(current, after)
		result.Counts = addCounts(result.Counts, counts)
		if len(lines) == 0 {
			continue
		}
		result.Status = "DIFFERENT"
		name := "runtime.draft.extensions"
		if scope.Name != "" {
			name = "profile." + scope.Name + ".draft.extensions"
		}
		fmt.Fprintf(&output, "\n### %s\n\n```diff\n%s```\n", name, strings.Join(lines, ""))
	}
	if result.Status == "SAME" {
		output.WriteString("\nNo Extension differences.\n")
	}
	return output.String(), result
}

func assignmentDiff(before, after []flatformat.Assignment) ([]string, Counts) {
	left, right := map[string]string{}, map[string]string{}
	for _, item := range before {
		left[item.Path] = item.Value
	}
	for _, item := range after {
		right[item.Path] = item.Value
	}
	keys := unionKeys(left, right)
	var lines []string
	var counts Counts
	for _, key := range keys {
		old, oldOK := left[key]
		next, nextOK := right[key]
		switch {
		case !oldOK:
			counts.Added++
			lines = append(lines, fmt.Sprintf("+ %s=%s\n", key, next))
		case !nextOK:
			counts.Removed++
			lines = append(lines, fmt.Sprintf("- %s=%s\n", key, old))
		case old != next:
			counts.Changed++
			lines = append(lines, fmt.Sprintf("- %s=%s\n+ %s=%s\n", key, old, key, next))
		}
	}
	return lines, counts
}

func stringDiff(before, after []string) ([]string, Counts) {
	left, right := map[string]bool{}, map[string]bool{}
	for _, value := range before {
		left[value] = true
	}
	for _, value := range after {
		right[value] = true
	}
	keys := unionKeys(left, right)
	var lines []string
	var counts Counts
	for _, key := range keys {
		if !left[key] {
			counts.Added++
			lines = append(lines, "+ "+key+"\n")
		}
		if !right[key] {
			counts.Removed++
			lines = append(lines, "- "+key+"\n")
		}
	}
	return lines, counts
}

func groupAssignments(lines []string) map[string][]string {
	result := map[string][]string{}
	for _, line := range lines {
		value := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "+"), "-"))
		path := strings.TrimSpace(strings.SplitN(value, "=", 2)[0])
		group := "root"
		if strings.HasPrefix(path, "[\"") {
			var key string
			decoder := json.NewDecoder(strings.NewReader(path[1:]))
			if decoder.Decode(&key) == nil {
				group = key
			}
		}
		result[group] = append(result[group], line)
	}
	return result
}

func insertInventory(document string, used, available []string, kind string) string {
	var inventory strings.Builder
	inventory.WriteString("## Inventory: Used by Recipe\n\n")
	writeFilteredList(&inventory, used, kind)
	inventory.WriteString("\n## Inventory: Available but Unused\n\n")
	writeFilteredList(&inventory, available, kind)
	return strings.Replace(document, "## Inventory\n", inventory.String(), 1)
}

func writeFilteredList(output *strings.Builder, values []string, kind string) {
	written := false
	for _, value := range values {
		if kind == "settings" && !strings.Contains(value, "settings.") {
			continue
		}
		if kind == "extensions" && !strings.HasSuffix(value, "extensions") {
			continue
		}
		fmt.Fprintf(output, "### %s\n\n", value)
		written = true
	}
	if !written {
		output.WriteString("- None\n")
	}
}

func ingredientInventory(root string, plan cookbook.Plan, definition recipe.Recipe) ([]string, []string, error) {
	direct := map[string]bool{
		"os\x00" + definition.OS:             true,
		"platform\x00" + definition.Platform: true,
	}
	for _, name := range definition.Runtime {
		direct["runtime\x00"+name] = true
	}
	for _, name := range definition.Profile {
		direct["profile\x00"+name] = true
	}
	usedSet := map[string]bool{}
	for _, scope := range append([]cookbook.ScopePlan{plan.Default}, plan.Profiles...) {
		for _, source := range scope.Sources {
			if !direct[source.Layer+"\x00"+source.Ingredient] {
				continue
			}
			relative, err := filepath.Rel(root, source.Path)
			if err == nil {
				usedSet[filepath.ToSlash(relative)] = true
			}
		}
	}
	var all []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		all = append(all, filepath.ToSlash(relative))
		return nil
	})
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return nil, nil, err
	}
	used := sortedSet(usedSet)
	var available []string
	for _, value := range all {
		if !usedSet[value] {
			available = append(available, value)
		}
	}
	sort.Strings(available)
	return used, available, nil
}

func (s Service) renderSummary(kind Kind, dist distribution.Distribution, plan cookbook.Plan, snapshot runtimelock.Snapshot, mode, recipeStatus, recipeDiff string, settingsResult, extensionsResult artifactResult, runtimeResults map[string]artifactResult, used []string) string {
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	var output strings.Builder
	fmt.Fprintf(&output, "# %s Summary\n\n> Source: `%s`  \n> Recipe: `%s`  \n> OS: `%s`  \n> Platform: `%s`\n\n", kind, dist.Name, plan.Name, plan.OS, plan.Platform)
	output.WriteString("## Result\n\n| Artifact | Status | Added | Removed | Changed | Review |\n| --- | --- | ---: | ---: | ---: | --- |\n")
	if kind == View {
		recipeStatus = "INVENTORY"
	}
	fmt.Fprintf(&output, "| Recipe | %s | - | - | - | [recipe.draft.yaml](recipe.draft.yaml) |\n", recipeStatus)
	fmt.Fprintf(&output, "| Settings | %s | %d | %d | %d | %s |\n", settingsResult.Status, settingsResult.Counts.Added, settingsResult.Counts.Removed, settingsResult.Counts.Changed, artifactReview(kind, settingsResult.Status, "settings.draft.md"))
	fmt.Fprintf(&output, "| Extensions | %s | %d | %d | %d | %s |\n", extensionsResult.Status, extensionsResult.Counts.Added, extensionsResult.Counts.Removed, extensionsResult.Counts.Changed, artifactReview(kind, extensionsResult.Status, "extensions.draft.md"))
	for _, name := range sortedArtifactNames(runtimeResults) {
		result := runtimeResults[name]
		fmt.Fprintf(&output, "| %s | %s | %d | %d | %d | %s |\n", strings.ToUpper(name[:1])+name[1:], result.Status, result.Counts.Added, result.Counts.Removed, result.Counts.Changed, artifactReview(kind, result.Status, name+".draft.md"))
	}
	if kind == FreezeDraft {
		fmt.Fprintf(&output, "\n## Recipe Difference\n\nStatus: `%s`\n\n%s", recipeStatus, recipeDiff)
	} else {
		output.WriteString("\n## View\n\nThe Distribution Lock is rendered against an empty reference. Every observed item is Inventory, not a Cookbook comparison.\n")
	}
	output.WriteString("\n## Profiles\n\n| Profile | Settings | Extensions | Inherits Settings |\n| --- | ---: | ---: | --- |\n")
	fmt.Fprintf(&output, "| default | %d | %d | - |\n", len(snapshot.Default.Settings), len(snapshot.Default.Extensions))
	for _, scope := range snapshot.Profiles {
		fmt.Fprintf(&output, "| %s | %d | %d | %t |\n", scope.Name, len(scope.Settings), len(scope.Extensions), scope.Inheritance.Settings)
	}
	output.WriteString("\n## Extension Versions\n\n| Extension | Version |\n| --- | --- |\n")
	writeUniqueVersions(&output, snapshot)
	if kind == FreezeDraft {
		output.WriteString("\n## Ingredient Context\n\n### Used by Recipe\n\n")
		writeMarkdownList(&output, used)
	}
	fmt.Fprintf(&output, "\n## Observation\n\n- Workbench: `%s`\n- Lock mode: `%s`\n- Observed: `%s`\n- Generated: `%s`\n", kind, mode, snapshot.ObservedAt.Format(time.RFC3339), now.Format(time.RFC3339))
	return output.String()
}

func artifactReview(kind Kind, status, name string) string {
	if kind == FreezeDraft && status == "SAME" {
		return "-"
	}
	return fmt.Sprintf("[%s](%s)", name, name)
}

func writeUniqueVersions(output *strings.Builder, snapshot runtimelock.Snapshot) {
	versions := map[string]map[string]bool{}
	collect := func(extensions []runtimeio.Extension) {
		for _, extension := range extensions {
			if versions[extension.ID] == nil {
				versions[extension.ID] = map[string]bool{}
			}
			versions[extension.ID][extension.Version] = true
		}
	}
	collect(snapshot.Default.Extensions)
	for _, scope := range snapshot.Profiles {
		collect(scope.Extensions)
	}
	for _, id := range sortedKeys(versions) {
		values := sortedKeys(versions[id])
		for index, version := range values {
			values[index] = "`" + version + "`"
		}
		fmt.Fprintf(output, "| `%s` | %s |\n", id, strings.Join(values, ", "))
	}
}

func findCurrentRecipe(cookbookRoot string, source recipe.Recipe, provenance string) (string, string, string) {
	data, err := os.ReadFile(provenance)
	if err != nil {
		return "", "UNAVAILABLE", fmt.Sprintf("Recipe provenance could not be read: %v\n", err)
	}
	paths, _ := filepath.Glob(filepath.Join(cookbookRoot, "recipe", "*.yaml"))
	var matches []string
	for _, path := range paths {
		definition, err := recipe.Load(path)
		if err == nil && definition.Name == source.Name && definition.OS == source.OS && definition.Platform == source.Platform {
			matches = append(matches, path)
		}
	}
	if len(matches) == 0 {
		return "", "MISSING", fencedDiff(nil, data)
	}
	if len(matches) > 1 {
		return "", "UNAVAILABLE", fmt.Sprintf("Recipe identity is ambiguous: %s\n", strings.Join(matches, ", "))
	}
	current := matches[0]
	before, err := os.ReadFile(current)
	if err != nil {
		return current, "UNAVAILABLE", fmt.Sprintf("Current Recipe could not be read: %v\n", err)
	}
	if string(before) == string(data) {
		return current, "SAME", "No Recipe differences.\n"
	}
	return current, "DIFFERENT", fencedDiff(before, data)
}

func fencedDiff(before, after []byte) string {
	var output strings.Builder
	oldLines := contentLines(before)
	newLines := contentLines(after)
	output.WriteString("```diff\n--- current Cookbook Recipe\n+++ observed Distribution Recipe\n")
	fmt.Fprintf(&output, "@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines))
	for _, line := range oldLines {
		if line != "" {
			output.WriteString("- " + line + "\n")
		}
	}
	for _, line := range newLines {
		if line != "" {
			output.WriteString("+ " + line + "\n")
		}
	}
	output.WriteString("```\n")
	return output.String()
}

func contentLines(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func publish(staging, target string, preserveOld bool) error {
	backup := target + ".old"
	if preserveOld {
		_ = os.RemoveAll(backup)
	} else {
		backup += ".staging"
		_ = os.RemoveAll(backup)
	}
	if _, err := os.Lstat(target); err == nil {
		if err := os.Rename(target, backup); err != nil {
			return fmt.Errorf("preserve Workbench: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(staging, target); err != nil {
		_ = os.Rename(backup, target)
		return fmt.Errorf("publish Workbench: %w", err)
	}
	if !preserveOld {
		_ = os.RemoveAll(backup)
	}
	return nil
}

func addCounts(left, right Counts) Counts {
	return Counts{left.Added + right.Added, left.Removed + right.Removed, left.Changed + right.Changed}
}
func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func unionKeys[V any](left, right map[string]V) []string {
	values := map[string]bool{}
	for key := range left {
		values[key] = true
	}
	for key := range right {
		values[key] = true
	}
	return sortedKeys(values)
}
func sortedSet(values map[string]bool) []string { return sortedKeys(values) }
func writeMarkdownList(output *strings.Builder, values []string) {
	if len(values) == 0 {
		output.WriteString("- None\n")
		return
	}
	for _, value := range values {
		fmt.Fprintf(output, "- %s\n", value)
	}
}
