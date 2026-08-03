package workbench

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/runtimeartifact"
	"code-toolkit/internal/runtimelock"
)

func renderRuntimeArtifacts(plan *cookbook.Plan, snapshot runtimelock.Snapshot) (map[string][]byte, map[string]artifactResult, error) {
	files := map[string][]byte{}
	results := map[string]artifactResult{}
	types := []struct {
		name      string
		planned   func(cookbook.ScopePlan) any
		observed  func(runtimelock.ScopeSnapshot) any
		inherited func(cookbook.Inheritance) bool
	}{
		{"keybindings", func(s cookbook.ScopePlan) any { return s.Keybindings }, func(s runtimelock.ScopeSnapshot) any { return s.Keybindings }, func(i cookbook.Inheritance) bool { return i.Keybindings }},
		{"tasks", func(s cookbook.ScopePlan) any { return s.Tasks }, func(s runtimelock.ScopeSnapshot) any { return s.Tasks }, func(i cookbook.Inheritance) bool { return i.Tasks }},
		{"mcp", func(s cookbook.ScopePlan) any { return s.MCP }, func(s runtimelock.ScopeSnapshot) any { return s.MCP }, func(i cookbook.Inheritance) bool { return i.MCP }},
	}
	planScopes := map[string]cookbook.ScopePlan{}
	if plan != nil {
		planScopes[""] = plan.Default
		for _, scope := range plan.Profiles {
			planScopes[scope.Name] = scope
		}
	}
	observed := append([]runtimelock.ScopeSnapshot{snapshot.Default}, snapshot.Profiles...)
	for _, kind := range types {
		var output strings.Builder
		title := strings.ToUpper(kind.name[:1]) + kind.name[1:]
		if plan == nil {
			fmt.Fprintf(&output, "# %s View\n\n## Inventory\n", title)
		} else {
			fmt.Fprintf(&output, "# %s Draft\n\n## Inventory\n\n## Difference\n", title)
		}
		result := artifactResult{Status: "SAME"}
		for _, scope := range observed {
			before := any(nil)
			if plan != nil {
				current := planScopes[scope.Name]
				if current.Inheritance.Unmanaged[kind.name] || scope.Name != "" && kind.inherited(current.Inheritance) {
					continue
				}
				before = kind.planned(current)
			}
			after := kind.observed(scope)
			if plan == nil && isNilArtifact(after) {
				continue
			}
			equal := reflect.DeepEqual(before, after)
			if kind.name == "tasks" {
				left, leftOK := before.(runtimeartifact.Object)
				right, rightOK := after.(runtimeartifact.Object)
				equal = leftOK && rightOK && runtimeartifact.TasksEqual(left, right)
			}
			if plan != nil && equal {
				continue
			}
			result.Status = "DIFFERENT"
			result.Counts.Changed++
			name := "runtime.draft." + kind.name + ".json"
			if scope.Name != "" {
				name = "profile." + scope.Name + ".draft." + kind.name + ".json"
			}
			fmt.Fprintf(&output, "\n### %s\n\n```diff\n", name)
			if plan != nil && !isNilArtifact(before) {
				writeJSONDiff(&output, '-', before)
			}
			if !isNilArtifact(after) {
				writeJSONDiff(&output, '+', after)
			}
			output.WriteString("```\n")
		}
		results[kind.name] = result
		if plan == nil || result.Status != "SAME" {
			files[kind.name+".draft.md"] = []byte(output.String())
		}
	}
	snippetText, snippetResult := renderSnippetArtifact(plan, snapshot, planScopes)
	results["snippets"] = snippetResult
	if plan == nil || snippetResult.Status != "SAME" {
		files["snippets.draft.md"] = []byte(snippetText)
	}
	return files, results, nil
}

func renderSnippetArtifact(plan *cookbook.Plan, snapshot runtimelock.Snapshot, planScopes map[string]cookbook.ScopePlan) (string, artifactResult) {
	var output strings.Builder
	if plan == nil {
		output.WriteString("# Snippets View\n\n## Inventory\n")
	} else {
		output.WriteString("# Snippets Draft\n\n## Inventory\n\n## Difference\n")
	}
	result := artifactResult{Status: "SAME"}
	for _, scope := range append([]runtimelock.ScopeSnapshot{snapshot.Default}, snapshot.Profiles...) {
		before := map[string]any{}
		if plan != nil {
			current := planScopes[scope.Name]
			if current.Inheritance.Unmanaged["snippets"] || scope.Name != "" && current.Inheritance.Snippets {
				continue
			}
			for name, value := range current.Snippets {
				before[name] = value
			}
		}
		names := map[string]bool{}
		for name := range before {
			names[name] = true
		}
		for name := range scope.Snippets {
			names[name] = true
		}
		ordered := make([]string, 0, len(names))
		for name := range names {
			ordered = append(ordered, name)
		}
		sort.Strings(ordered)
		for _, filename := range ordered {
			old, oldOK := before[filename]
			after, newOK := scope.Snippets[filename]
			if plan != nil && oldOK == newOK && reflect.DeepEqual(old, after) {
				continue
			}
			result.Status = "DIFFERENT"
			result.Counts.Changed++
			target := "runtime.draft.snippets." + filename
			if scope.Name != "" {
				target = "profile." + scope.Name + ".draft.snippets." + filename
			}
			fmt.Fprintf(&output, "\n### %s\n\n```diff\n", target)
			if oldOK {
				writeJSONDiff(&output, '-', old)
			}
			if newOK {
				writeJSONDiff(&output, '+', after)
			}
			output.WriteString("```\n")
		}
	}
	return output.String(), result
}

func writeJSONDiff(output *strings.Builder, prefix byte, value any) {
	data, _ := json.MarshalIndent(value, "", "  ")
	for _, line := range strings.Split(string(data), "\n") {
		output.WriteByte(prefix)
		output.WriteByte(' ')
		output.WriteString(line)
		output.WriteByte('\n')
	}
}

func isNilArtifact(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return (v.Kind() == reflect.Map || v.Kind() == reflect.Slice) && v.IsNil()
}

func sortedArtifactNames(values map[string]artifactResult) []string {
	result := make([]string, 0, len(values))
	for name := range values {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}
