package workbench

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"code-toolkit/internal/flatformat"
	"code-toolkit/internal/mergerules"
	"code-toolkit/internal/recipe"
	"code-toolkit/internal/runtimeartifact"
	"code-toolkit/internal/settings"
)

type CommitResult struct {
	Completed, Unresolved int
	Files                 []string
}
type patchLine struct {
	Add  bool
	Text string
	Line int
}

func (s Service) Commit(force bool) (CommitResult, error) {
	draft := filepath.Join(s.CookbookRoot, "draft")
	result := CommitResult{}
	writes := map[string][]byte{}
	unionPaths := map[string][]string{}

	for _, artifact := range []string{"settings.draft.md", "extensions.draft.md"} {
		path := filepath.Join(draft, artifact)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, err
		}
		sections, err := parseDifference(data)
		if err != nil {
			return result, fmt.Errorf("parse %s: %w", artifact, err)
		}
		for target, lines := range sections {
			path, err := ingredientTarget(filepath.Join(s.CookbookRoot, "ingredient"), target, artifact)
			if err != nil {
				return result, err
			}
			current, readErr := os.ReadFile(path)
			if readErr != nil && !os.IsNotExist(readErr) {
				return result, readErr
			}
			var output []byte
			if artifact == "settings.draft.md" {
				if hasJSONComments(current) && !force {
					return result, fmt.Errorf("%s contains JSONC comments; rerun with --force to accept comment loss", path)
				}
				output, err = applySettingsPatch(current, lines, &result, unionPaths)
			} else {
				output, err = applyExtensionPatch(current, lines, &result)
			}
			if err != nil {
				return result, fmt.Errorf("commit %s: %w", target, err)
			}
			writes[path] = output
		}
	}
	for _, kind := range []string{"keybindings", "tasks", "mcp", "snippets"} {
		artifact := kind + ".draft.md"
		data, err := os.ReadFile(filepath.Join(draft, artifact))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, err
		}
		sections, err := parseDifference(data)
		if err != nil {
			return result, fmt.Errorf("parse %s: %w", artifact, err)
		}
		for target, lines := range sections {
			path, err := runtimeArtifactTarget(filepath.Join(s.CookbookRoot, "ingredient"), target, kind)
			if err != nil {
				return result, err
			}
			current, readErr := os.ReadFile(path)
			if readErr != nil && !os.IsNotExist(readErr) {
				return result, readErr
			}
			if hasJSONComments(current) && !force {
				return result, fmt.Errorf("%s contains JSONC comments; rerun with --force to accept comment loss", path)
			}
			var source strings.Builder
			for _, line := range lines {
				if line.Add {
					source.WriteString(line.Text)
					source.WriteByte('\n')
				}
			}
			if source.Len() == 0 {
				if kind == "keybindings" {
					source.WriteString("[]")
				} else {
					source.WriteString("{}")
				}
			}
			var value any
			if kind == "keybindings" {
				value, err = runtimeartifact.ParseArray([]byte(source.String()))
			} else {
				value, err = runtimeartifact.ParseObject([]byte(source.String()))
			}
			if err != nil {
				return result, fmt.Errorf("commit %s: %w", target, err)
			}
			output, err := runtimeartifact.Marshal(value)
			if err != nil {
				return result, err
			}
			writes[path] = output
			result.Completed++
		}
	}

	recipeDraft := filepath.Join(draft, "recipe.draft.yaml")
	if data, err := os.ReadFile(recipeDraft); err == nil {
		target, err := resolveRecipeTarget(filepath.Join(s.CookbookRoot, "recipe"), recipeDraft)
		if err != nil {
			return result, err
		}
		writes[target] = data
		result.Completed++
	} else if !os.IsNotExist(err) {
		return result, err
	}

	if len(unionPaths) > 0 {
		path := filepath.Join(s.CookbookRoot, "kitchen-notes", "go.merge-rules.yaml")
		current, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return result, err
		}
		paths := make([][]string, 0, len(unionPaths))
		for _, value := range unionPaths {
			paths = append(paths, value)
		}
		data, err := mergerules.Add(current, paths)
		if err != nil {
			return result, fmt.Errorf("update Merge Rules: %w", err)
		}
		writes[path] = data
	}
	if len(writes) == 0 {
		return result, fmt.Errorf("no Commit Artifacts found in %s", draft)
	}
	if err := publishFiles(writes); err != nil {
		return result, err
	}
	for path := range writes {
		result.Files = append(result.Files, path)
	}
	sort.Strings(result.Files)
	return result, nil
}

func runtimeArtifactTarget(root, name, kind string) (string, error) {
	if name == "" || filepath.IsAbs(name) || strings.ContainsAny(name, "\x00\r\n") {
		return "", fmt.Errorf("invalid Commit target %q", name)
	}
	clean := filepath.Clean(name)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("Commit target escapes Ingredient root: %s", name)
	}
	base := filepath.Base(clean)
	valid := base == kind+".json" || base == kind+".jsonc" || strings.Contains(base, "."+kind+".json") || strings.Contains(base, "."+kind+".jsonc")
	if kind == "snippets" {
		valid = strings.Contains(base, ".snippets.") && (strings.HasSuffix(base, ".json") || strings.HasSuffix(base, ".jsonc") || strings.HasSuffix(base, ".code-snippets"))
	}
	if !valid {
		return "", fmt.Errorf("invalid %s target %q", kind, name)
	}
	return filepath.Join(root, clean), nil
}

func parseDifference(data []byte) (map[string][]patchLine, error) {
	sections := map[string][]patchLine{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	inDifference, inFence, target := false, false, ""
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSuffix(scanner.Text(), "\r")
		if text == "## Difference" {
			inDifference, target = true, ""
			continue
		}
		if strings.HasPrefix(text, "## ") && text != "## Difference" {
			inDifference, target = false, ""
			continue
		}
		if !inDifference {
			continue
		}
		if strings.HasPrefix(text, "### ") && !strings.HasPrefix(text, "#### ") {
			target = strings.TrimSpace(strings.TrimPrefix(text, "### "))
			continue
		}
		if strings.HasPrefix(text, "```") {
			inFence = !inFence
			continue
		}
		if !inFence || target == "" || len(text) < 2 || (text[0] != '+' && text[0] != '-') {
			continue
		}
		sections[target] = append(sections[target], patchLine{Add: text[0] == '+', Text: strings.TrimSpace(text[1:]), Line: line})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sections, nil
}

func ingredientTarget(root, name, artifact string) (string, error) {
	if name == "" || filepath.IsAbs(name) || strings.ContainsAny(name, "\x00\r\n") {
		return "", fmt.Errorf("invalid Commit target %q", name)
	}
	clean := filepath.Clean(name)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("Commit target escapes Ingredient root: %s", name)
	}
	base := filepath.Base(clean)
	if artifact == "settings.draft.md" {
		valid := base == "settings.json" || base == "settings.jsonc" || strings.HasSuffix(base, ".settings.json") || strings.HasSuffix(base, ".settings.jsonc")
		if !valid {
			return "", fmt.Errorf("invalid Settings target %q", name)
		}
	} else if base != "extensions" && !strings.HasSuffix(base, ".extensions") {
		return "", fmt.Errorf("invalid Extensions target %q", name)
	}
	return filepath.Join(root, clean), nil
}

func applySettingsPatch(current []byte, lines []patchLine, result *CommitResult, unions map[string][]string) ([]byte, error) {
	document := settings.Document{}
	if len(bytes.TrimSpace(current)) > 0 {
		parsed, err := settings.Parse(current)
		if err != nil {
			return nil, err
		}
		document = parsed
	}
	type operation struct {
		patchLine
		assignment flatformat.Assignment
		path       []flatformat.Segment
		value      any
	}
	var removes, adds []operation
	additions := map[string]string{}
	for _, line := range lines {
		assignments, err := flatformat.Parse([]byte(line.Text + "\n"))
		if err != nil || len(assignments) != 1 {
			if err == nil {
				err = fmt.Errorf("expected one assignment")
			}
			return nil, fmt.Errorf("line %d: %w", line.Line, err)
		}
		segments, err := flatformat.DecodePath(assignments[0].Path)
		if err != nil {
			return nil, err
		}
		var value any
		decoder := json.NewDecoder(strings.NewReader(assignments[0].Value))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		op := operation{line, assignments[0], segments, value}
		if line.Add {
			if previous, exists := additions[assignments[0].Path]; exists && previous != assignments[0].Value && !strings.Contains(assignments[0].Path, "[*]") {
				return nil, fmt.Errorf("line %d: incompatible additions for %s", line.Line, assignments[0].Path)
			}
			additions[assignments[0].Path] = assignments[0].Value
			adds = append(adds, op)
		} else {
			removes = append(removes, op)
		}
	}
	sort.SliceStable(removes, func(i, j int) bool {
		if len(removes[i].path) != len(removes[j].path) {
			return len(removes[i].path) > len(removes[j].path)
		}
		left, right := removes[i].path, removes[j].path
		if len(left) > 0 && left[len(left)-1].Kind == flatformat.Index && right[len(right)-1].Kind == flatformat.Index && sameParent(left, right) {
			return left[len(left)-1].Index > right[len(right)-1].Index
		}
		return false
	})
	sort.SliceStable(adds, func(i, j int) bool { return len(adds[i].path) < len(adds[j].path) })
	root := any(map[string]any(document))
	for _, op := range removes {
		matched, err := removeValue(&root, op.path, op.value)
		if err != nil {
			return nil, err
		}
		if matched {
			result.Completed++
		} else {
			result.Unresolved++
		}
	}
	named := map[string]int{}
	for _, op := range adds {
		if err := setValue(&root, op.path, op.value, named, unions); err != nil {
			return nil, err
		}
		result.Completed++
	}
	object, ok := root.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Settings root must be an object")
	}
	return settings.Marshal(settings.Document(object))
}

func sameParent(left, right []flatformat.Segment) bool {
	if len(left) != len(right) {
		return false
	}
	for index := 0; index < len(left)-1; index++ {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func applyExtensionPatch(current []byte, lines []patchLine, result *CommitResult) ([]byte, error) {
	values := []string{}
	seen := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(current))
	for scanner.Scan() {
		value := strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\r"))
		if value != "" && !seen[value] {
			seen[value] = true
			values = append(values, value)
		}
	}
	for _, line := range lines {
		id := strings.TrimSpace(line.Text)
		if id == "" {
			return nil, fmt.Errorf("line %d: empty Extension ID", line.Line)
		}
		if line.Add {
			if !seen[id] {
				seen[id] = true
				values = append(values, id)
			}
			result.Completed++
		} else if seen[id] {
			delete(seen, id)
			result.Completed++
		} else {
			result.Unresolved++
		}
	}
	output := []string{}
	for _, value := range values {
		if seen[value] {
			output = append(output, value)
		}
	}
	sort.Strings(output)
	if len(output) == 0 {
		return []byte{}, nil
	}
	return []byte(strings.Join(output, "\n") + "\n"), nil
}

func removeValue(root *any, path []flatformat.Segment, expected any) (bool, error) {
	actual, ok := getValue(*root, path)
	if !ok || !semanticEqual(actual, expected) {
		return false, nil
	}
	return deleteValue(root, path), nil
}
func semanticEqual(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}

func getValue(value any, path []flatformat.Segment) (any, bool) {
	for _, segment := range path {
		switch segment.Kind {
		case flatformat.Property:
			object, ok := value.(map[string]any)
			if !ok {
				return nil, false
			}
			value, ok = object[segment.Name]
			if !ok {
				return nil, false
			}
		case flatformat.Index:
			array, ok := value.([]any)
			if !ok || segment.Index >= len(array) {
				return nil, false
			}
			value = array[segment.Index]
		default:
			return nil, false
		}
	}
	return value, true
}
func deleteValue(root *any, path []flatformat.Segment) bool {
	if len(path) == 0 {
		return false
	}
	parent, ok := getValue(*root, path[:len(path)-1])
	if !ok {
		return false
	}
	last := path[len(path)-1]
	if last.Kind == flatformat.Property {
		delete(parent.(map[string]any), last.Name)
		return true
	}
	if last.Kind == flatformat.Index {
		array := parent.([]any)
		if last.Index >= len(array) {
			return false
		}
		array = append(array[:last.Index], array[last.Index+1:]...)
		return setRegular(root, path[:len(path)-1], array) == nil
	}
	return false
}

func setValue(root *any, path []flatformat.Segment, value any, named map[string]int, unions map[string][]string) error {
	for index, segment := range path {
		if segment.Kind == flatformat.UnionAnonymous || segment.Kind == flatformat.UnionNamed {
			props := propertyPath(path[:index])
			if props == nil {
				return fmt.Errorf("Union Rule path must contain only properties")
			}
			unions[mergerules.Key(props)] = props
			parent, ok := getValue(*root, path[:index])
			if !ok {
				if err := setRegular(root, path[:index], []any{}); err != nil {
					return err
				}
				parent, _ = getValue(*root, path[:index])
			}
			array, ok := parent.([]any)
			if !ok {
				return fmt.Errorf("Union path %v is not an array", props)
			}
			if segment.Kind == flatformat.UnionAnonymous {
				if index != len(path)-1 {
					return fmt.Errorf("cannot traverse beneath [*]")
				}
				array = append(array, value)
			} else {
				key := mergerules.Key(props) + "@" + segment.Name
				position, exists := named[key]
				if !exists {
					position = len(array)
					named[key] = position
					array = append(array, map[string]any{})
				}
				childPath := append([]flatformat.Segment{{Kind: flatformat.Index, Index: position}}, path[index+1:]...)
				temp := any(array)
				if err := setRegular(&temp, childPath, value); err != nil {
					return err
				}
				array = temp.([]any)
			}
			return setRegular(root, path[:index], array)
		}
	}
	return setRegular(root, path, value)
}
func propertyPath(path []flatformat.Segment) []string {
	result := []string{}
	for _, s := range path {
		if s.Kind != flatformat.Property {
			return nil
		}
		result = append(result, s.Name)
	}
	return result
}

func setRegular(root *any, path []flatformat.Segment, value any) error {
	if len(path) == 0 {
		*root = value
		return nil
	}
	current := root
	for index, segment := range path {
		last := index == len(path)-1
		switch segment.Kind {
		case flatformat.Property:
			object, ok := (*current).(map[string]any)
			if !ok {
				return fmt.Errorf("property %q parent is not object", segment.Name)
			}
			if last {
				object[segment.Name] = value
				return nil
			}
			child, exists := object[segment.Name]
			if !exists {
				if path[index+1].Kind == flatformat.Index {
					child = []any{}
				} else {
					child = map[string]any{}
				}
				object[segment.Name] = child
			}
			current = &child
			defer func(o map[string]any, k string, p *any) { o[k] = *p }(object, segment.Name, current)
		case flatformat.Index:
			array, ok := (*current).([]any)
			if !ok {
				return fmt.Errorf("index parent is not array")
			}
			for len(array) <= segment.Index {
				array = append(array, nil)
			}
			if last {
				array[segment.Index] = value
				*current = array
				return nil
			}
			child := array[segment.Index]
			if child == nil {
				if path[index+1].Kind == flatformat.Index {
					child = []any{}
				} else {
					child = map[string]any{}
				}
			}
			array[segment.Index] = child
			*current = array
			current = &array[segment.Index]
		default:
			return fmt.Errorf("unexpected Union selector")
		}
	}
	return nil
}

func resolveRecipeTarget(root, draft string) (string, error) {
	definition, err := recipe.Load(draft)
	if err != nil {
		return "", err
	}
	paths, _ := filepath.Glob(filepath.Join(root, "*.yaml"))
	matches := []string{}
	for _, path := range paths {
		value, err := recipe.Load(path)
		if err != nil {
			continue
		}
		if value.Name == definition.Name && value.OS == definition.OS && value.Platform == definition.Platform {
			matches = append(matches, path)
		}
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous Recipe identity %s/%s/%s", definition.Name, definition.OS, definition.Platform)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return filepath.Join(root, fmt.Sprintf("%s.%s.%s.yaml", definition.Name, definition.OS, definition.Platform)), nil
}

func hasJSONComments(data []byte) bool {
	inString, escaped := false, false
	for i := 0; i < len(data)-1; i++ {
		c := data[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inString {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if !inString && c == '/' && (data[i+1] == '/' || data[i+1] == '*') {
			return true
		}
	}
	return false
}

func publishFiles(writes map[string][]byte) error {
	type item struct {
		target, temp, backup string
		existed, published   bool
	}
	items := []item{}
	for target, data := range writes {
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		file, err := os.CreateTemp(filepath.Dir(target), ".ctk-commit-")
		if err != nil {
			return err
		}
		temp := file.Name()
		if _, err = file.Write(data); err != nil {
			file.Close()
			os.Remove(temp)
			return err
		}
		if err = file.Close(); err != nil {
			os.Remove(temp)
			return err
		}
		_, statErr := os.Stat(target)
		items = append(items, item{target: target, temp: temp, backup: temp + ".old", existed: statErr == nil})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].target < items[j].target })
	rollback := func() {
		for i := len(items) - 1; i >= 0; i-- {
			entry := &items[i]
			if entry.published {
				os.Remove(entry.target)
			}
			if entry.existed {
				os.Rename(entry.backup, entry.target)
			}
			os.Remove(entry.temp)
		}
	}
	for i := range items {
		entry := &items[i]
		if entry.existed {
			if err := os.Rename(entry.target, entry.backup); err != nil {
				rollback()
				return err
			}
		}
		if err := os.Rename(entry.temp, entry.target); err != nil {
			rollback()
			return err
		}
		entry.published = true
	}
	for _, entry := range items {
		os.Remove(entry.backup)
	}
	return nil
}
