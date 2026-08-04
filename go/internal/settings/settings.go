package settings

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kshrkznr/code-toolkit/go/internal/mergerules"

	"github.com/tailscale/hujson"
)

type Document map[string]any

func Parse(data []byte) (Document, error) {
	standard, err := hujson.Standardize(data)
	if err != nil {
		return nil, fmt.Errorf("standardize JSONC: %w", err)
	}
	var document Document
	if err := json.Unmarshal(standard, &document); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}
	if document == nil {
		document = Document{}
	}
	return document, nil
}

func Merge(documents ...Document) Document {
	result, _ := MergeWithRules(mergerules.Rules{Union: map[string]bool{}}, documents...)
	return result
}

func MergeWithRules(rules mergerules.Rules, documents ...Document) (Document, error) {
	result := Document{}
	for _, document := range documents {
		if err := mergeObjectWithRules(result, document, nil, rules); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func mergeObjectWithRules(target, source map[string]any, path []string, rules mergerules.Rules) error {
	for key, value := range source {
		currentPath := append(append([]string(nil), path...), key)
		if rules.Union[mergerules.Key(currentPath)] {
			sourceArray, ok := value.([]any)
			if !ok {
				return fmt.Errorf("Merge Rule path %v resolves to non-array", currentPath)
			}
			targetValue, present := target[key]
			targetArray, exists := targetValue.([]any)
			if present && !exists {
				return fmt.Errorf("Merge Rule path %v resolves to non-array", currentPath)
			}
			for _, item := range sourceArray {
				if !containsCanonical(targetArray, item) {
					targetArray = append(targetArray, clone(item))
				}
			}
			target[key] = targetArray
			continue
		}
		sourceObject, sourceIsObject := value.(map[string]any)
		if sourceIsObject {
			targetObject, targetIsObject := target[key].(map[string]any)
			if !targetIsObject {
				targetObject = map[string]any{}
				target[key] = targetObject
			}
			if err := mergeObjectWithRules(targetObject, sourceObject, currentPath, rules); err != nil {
				return err
			}
			continue
		}
		target[key] = clone(value)
	}
	return nil
}

func containsCanonical(values []any, candidate any) bool {
	want, _ := json.Marshal(candidate)
	for _, value := range values {
		got, _ := json.Marshal(value)
		if bytes.Equal(got, want) {
			return true
		}
	}
	return false
}

func mergeObject(target, source map[string]any) {
	for key, value := range source {
		sourceObject, sourceIsObject := value.(map[string]any)
		targetObject, targetIsObject := target[key].(map[string]any)
		if sourceIsObject && targetIsObject {
			mergeObject(targetObject, sourceObject)
			continue
		}
		target[key] = clone(value)
	}
}

func clone(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		copy := map[string]any{}
		mergeObject(copy, typed)
		return copy
	case []any:
		copy := make([]any, len(typed))
		for i := range typed {
			copy[i] = clone(typed[i])
		}
		return copy
	default:
		return value
	}
}

func Marshal(document Document) ([]byte, error) {
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal settings: %w", err)
	}
	return append(data, '\n'), nil
}
