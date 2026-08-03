package runtimeartifact

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/tailscale/hujson"
)

type Value = any
type Array []Value
type Object map[string]Value
type Snippets map[string]Object

func Parse(data []byte) (Value, error) {
	standard, err := hujson.Standardize(data)
	if err != nil {
		return nil, err
	}
	var value Value
	if err := json.Unmarshal(standard, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func ParseArray(data []byte) (Array, error) {
	value, err := Parse(data)
	if err != nil {
		return nil, err
	}
	array, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("Artifact root must be an array")
	}
	return Array(array), nil
}

func ParseObject(data []byte) (Object, error) {
	value, err := Parse(data)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Artifact root must be an object")
	}
	return Object(object), nil
}

func Marshal(value Value) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func AppendArrays(values ...Array) Array {
	var result Array
	for _, value := range values {
		result = append(result, value...)
	}
	if result == nil {
		return Array{}
	}
	return result
}

func MergeSnippets(values ...Object) Object {
	result := Object{}
	for _, value := range values {
		for name, definition := range value {
			result[name] = definition
		}
	}
	return result
}

func MergeMCP(values ...Object) Object {
	result := Object{}
	servers := Object{}
	var inputs Array
	for _, value := range values {
		for key, item := range value {
			switch key {
			case "servers":
				if object, ok := item.(map[string]any); ok {
					for name, server := range object {
						servers[name] = server
					}
				} else {
					result[key] = item
				}
			case "inputs":
				if array, ok := item.([]any); ok {
					inputs = append(inputs, array...)
				} else {
					result[key] = item
				}
			default:
				result[key] = item
			}
		}
	}
	if len(servers) != 0 {
		result["servers"] = map[string]any(servers)
	}
	if inputs != nil {
		result["inputs"] = []any(inputs)
	}
	return result
}

func MergeTasks(values ...Object) (Object, error) {
	result := Object{}
	var tasks, inputs Array
	version := "2.0.0"
	for _, value := range values {
		for key, item := range value {
			switch key {
			case "version":
				text, ok := item.(string)
				if !ok || version != text {
					return nil, fmt.Errorf("incompatible Tasks version")
				}
				version = text
			case "tasks":
				array, ok := item.([]any)
				if !ok {
					return nil, fmt.Errorf("Tasks tasks must be an array")
				}
				tasks = append(tasks, array...)
			case "inputs":
				array, ok := item.([]any)
				if !ok {
					return nil, fmt.Errorf("Tasks inputs must be an array")
				}
				inputs = append(inputs, array...)
			default:
				result[key] = item
			}
		}
	}
	result["version"] = version
	result["tasks"] = []any(tasks)
	result["inputs"] = []any(inputs)
	return result, nil
}

// TasksEqual treats representationally different empty Tasks envelopes as one
// semantic empty state. Runtime observation itself remains unchanged.
func TasksEqual(left, right Object) bool {
	return reflect.DeepEqual(NormalizeTasksForComparison(left), NormalizeTasksForComparison(right))
}

func NormalizeTasksForComparison(value Object) Object {
	if emptyTasksDocument(value) {
		return Object{}
	}
	result := Object{}
	for key, item := range value {
		if (key == "tasks" || key == "inputs") && emptyArray(item) {
			continue
		}
		result[key] = item
	}
	return result
}

func emptyTasksDocument(value Object) bool {
	for key, item := range value {
		switch key {
		case "version":
			if item != "2.0.0" {
				return false
			}
		case "tasks", "inputs":
			if !emptyArray(item) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func emptyArray(value any) bool {
	array, ok := value.([]any)
	return ok && len(array) == 0
}
