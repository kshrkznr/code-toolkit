package settings

import (
	"code-toolkit/internal/mergerules"
	"testing"
)

func TestParseJSONCAndMerge(t *testing.T) {
	first, err := Parse([]byte(`{"object":{"left":1},"array":[1],"value":"old"}`))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Parse([]byte(`{/* comment */"object":{"right":2},"array":[2,],"value":null,}`))
	if err != nil {
		t.Fatal(err)
	}
	got := Merge(first, second)
	object := got["object"].(map[string]any)
	if object["left"] != float64(1) || object["right"] != float64(2) {
		t.Fatalf("object = %#v", object)
	}
	if got["value"] != nil {
		t.Fatalf("value = %#v", got["value"])
	}
	array := got["array"].([]any)
	if len(array) != 1 || array[0] != float64(2) {
		t.Fatalf("array = %#v", array)
	}
}

func TestMergeWithRulesUnionsByCanonicalValueInFirstOccurrenceOrder(t *testing.T) {
	rules := mergerules.Rules{Union: map[string]bool{mergerules.Key([]string{"values"}): true}}
	merged, err := MergeWithRules(rules,
		Document{"values": []any{float64(1), map[string]any{"a": float64(1)}}},
		Document{"values": []any{map[string]any{"a": float64(1)}, float64(2)}},
	)
	if err != nil {
		t.Fatal(err)
	}
	values := merged["values"].([]any)
	if len(values) != 3 || values[0] != float64(1) || values[2] != float64(2) {
		t.Fatalf("values = %#v", values)
	}
}

func TestMergeWithRulesRejectsNestedNonArrayOnFirstDocument(t *testing.T) {
	rules := mergerules.Rules{Union: map[string]bool{mergerules.Key([]string{"outer", "values"}): true}}
	if _, err := MergeWithRules(rules, Document{"outer": map[string]any{"values": "wrong"}}); err == nil {
		t.Fatal("expected declared Union path type error")
	}
}
