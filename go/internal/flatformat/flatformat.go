package flatformat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Assignment struct {
	Path  string
	Value string
}

type SegmentKind int

const (
	Property SegmentKind = iota
	Index
	UnionAnonymous
	UnionNamed
)

type Segment struct {
	Kind  SegmentKind
	Name  string
	Index int
}

// DecodePath parses a CTK Flat path. Union selectors are Workbench-only.
func DecodePath(path string) ([]Segment, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	if path == "[]" {
		return nil, nil
	}
	encoded := strings.ReplaceAll(path, "[*]", `"__ctk_union_anonymous__"`)
	for {
		start := strings.Index(encoded, "[@")
		if start < 0 {
			break
		}
		end := strings.Index(encoded[start:], "]")
		if end < 0 {
			return nil, fmt.Errorf("invalid named selector in %q", path)
		}
		end += start
		name := encoded[start+2 : end]
		quoted, _ := json.Marshal("__ctk_union_named__" + name)
		encoded = encoded[:start] + string(quoted) + encoded[end+1:]
	}
	var raw []any
	if err := json.Unmarshal([]byte(encoded), &raw); err != nil {
		return nil, fmt.Errorf("invalid path %q: %w", path, err)
	}
	segments := make([]Segment, 0, len(raw))
	for _, item := range raw {
		switch value := item.(type) {
		case string:
			switch {
			case value == "__ctk_union_anonymous__":
				segments = append(segments, Segment{Kind: UnionAnonymous})
			case strings.HasPrefix(value, "__ctk_union_named__"):
				segments = append(segments, Segment{Kind: UnionNamed, Name: strings.TrimPrefix(value, "__ctk_union_named__")})
			default:
				segments = append(segments, Segment{Kind: Property, Name: value})
			}
		case []any:
			segments = append(segments, Segment{Kind: Index, Index: int(value[0].(float64))})
		}
	}
	return segments, nil
}

func Flatten(value any) ([]Assignment, error) {
	var result []Assignment
	if err := flatten(nil, value, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func Encode(value any) ([]byte, error) {
	assignments, err := Flatten(value)
	if err != nil {
		return nil, err
	}
	var output strings.Builder
	for _, assignment := range assignments {
		fmt.Fprintf(&output, "%s=%s\n", assignment.Path, assignment.Value)
	}
	return []byte(output.String()), nil
}

func Parse(data []byte) ([]Assignment, error) {
	var result []Assignment
	seen := map[string]map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\r"))
		if text == "" {
			continue
		}
		separator := assignmentSeparator(text)
		if separator < 0 {
			return nil, fmt.Errorf("line %d: assignment separator not found", line)
		}
		path, value := strings.TrimSpace(text[:separator]), strings.TrimSpace(text[separator+1:])
		if err := validatePath(path); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		var parsed any
		decoder := json.NewDecoder(strings.NewReader(value))
		decoder.UseNumber()
		if err := decoder.Decode(&parsed); err != nil {
			return nil, fmt.Errorf("line %d: invalid JSON value: %w", line, err)
		}
		canonical, err := marshalJSON(parsed)
		if err != nil {
			return nil, fmt.Errorf("line %d: marshal value: %w", line, err)
		}
		if seen[path] == nil {
			seen[path] = map[string]bool{}
		}
		if seen[path][string(canonical)] {
			if strings.Contains(path, "[*]") {
				continue
			}
			return nil, fmt.Errorf("line %d: duplicate path %s", line, path)
		}
		if len(seen[path]) > 0 && !strings.Contains(path, "[*]") {
			return nil, fmt.Errorf("line %d: conflicting path %s", line, path)
		}
		seen[path][string(canonical)] = true
		result = append(result, Assignment{Path: path, Value: string(canonical)})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func flatten(path []any, value any, result *[]Assignment) error {
	encodedPath, err := encodePath(path)
	if err != nil {
		return err
	}
	switch typed := value.(type) {
	case map[string]any:
		*result = append(*result, Assignment{Path: encodedPath, Value: "{}"})
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := flatten(appendPath(path, key), typed[key], result); err != nil {
				return err
			}
		}
	case []any:
		*result = append(*result, Assignment{Path: encodedPath, Value: "[]"})
		for index, item := range typed {
			if err := flatten(appendPath(path, []int{index}), item, result); err != nil {
				return err
			}
		}
	default:
		encoded, err := marshalJSON(typed)
		if err != nil {
			return fmt.Errorf("encode %s: %w", encodedPath, err)
		}
		*result = append(*result, Assignment{Path: encodedPath, Value: string(encoded)})
	}
	return nil
}

func appendPath(path []any, value any) []any {
	result := make([]any, len(path), len(path)+1)
	copy(result, path)
	return append(result, value)
}

func encodePath(path []any) (string, error) {
	parts := make([]string, 0, len(path))
	for _, component := range path {
		switch typed := component.(type) {
		case string:
			value, err := marshalJSON(typed)
			if err != nil {
				return "", err
			}
			parts = append(parts, string(value))
		case []int:
			parts = append(parts, fmt.Sprintf("[%d]", typed[0]))
		default:
			return "", fmt.Errorf("unsupported path component %T", component)
		}
	}
	return "[" + strings.Join(parts, ", ") + "]", nil
}

func marshalJSON(value any) ([]byte, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(output.Bytes(), []byte("\n")), nil
}

func assignmentSeparator(line string) int {
	inString, escaped := false, false
	for index, character := range line {
		if escaped {
			escaped = false
			continue
		}
		if character == '\\' && inString {
			escaped = true
			continue
		}
		if character == '"' {
			inString = !inString
			continue
		}
		if character == '=' && !inString {
			return index
		}
	}
	return -1
}

func validatePath(path string) error {
	if path == "[]" {
		return nil
	}
	if !strings.HasPrefix(path, "[") || !strings.HasSuffix(path, "]") {
		return fmt.Errorf("invalid path %q", path)
	}
	jsonPath := strings.ReplaceAll(path, "[*]", `"__ctk_wildcard__"`)
	for {
		start := strings.Index(jsonPath, "[@")
		if start < 0 {
			break
		}
		end := strings.Index(jsonPath[start:], "]")
		if end < 0 || end == 2 {
			return fmt.Errorf("invalid named selector in %q", path)
		}
		end += start
		name := jsonPath[start+2 : end]
		if strings.ContainsAny(name, `"\\,[]`) {
			return fmt.Errorf("invalid named selector in %q", path)
		}
		quoted, _ := json.Marshal("__ctk_named_" + name)
		jsonPath = jsonPath[:start] + string(quoted) + jsonPath[end+1:]
	}
	var components []any
	if err := json.Unmarshal([]byte(jsonPath), &components); err != nil {
		return fmt.Errorf("invalid path %q: %w", path, err)
	}
	for _, component := range components {
		switch typed := component.(type) {
		case string:
		case []any:
			if len(typed) != 1 {
				return fmt.Errorf("invalid array selector in %q", path)
			}
			index, ok := typed[0].(float64)
			if !ok || index < 0 || index != float64(int(index)) {
				return fmt.Errorf("invalid array index in %q", path)
			}
		default:
			return fmt.Errorf("invalid path component in %q", path)
		}
	}
	return nil
}
