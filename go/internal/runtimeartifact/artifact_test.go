package runtimeartifact

import "testing"

func TestOpaqueComposition(t *testing.T) {
	a, err := ParseArray([]byte(`[{"key":"a"},]`))
	if err != nil {
		t.Fatal(err)
	}
	if got := AppendArrays(a, a); len(got) != 2 {
		t.Fatalf("keybindings = %#v", got)
	}
	tasks, err := MergeTasks(Object{"version": "2.0.0", "tasks": []any{map[string]any{"label": "same"}}}, Object{"version": "2.0.0", "tasks": []any{map[string]any{"label": "same"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks["tasks"].([]any)) != 2 {
		t.Fatalf("tasks = %#v", tasks)
	}
}

func TestMCPServerLaterWinsAndInputsAppend(t *testing.T) {
	got := MergeMCP(Object{"servers": map[string]any{"one": map[string]any{"url": "old"}}, "inputs": []any{"a"}}, Object{"servers": map[string]any{"one": map[string]any{"url": "new"}}, "inputs": []any{"a"}})
	if len(got["inputs"].([]any)) != 2 {
		t.Fatalf("inputs = %#v", got)
	}
	one := got["servers"].(map[string]any)["one"].(map[string]any)
	if one["url"] != "new" {
		t.Fatalf("server = %#v", one)
	}
}

func TestTasksEqualTreatsFixedEmptyEnvelopeAsSemanticEmpty(t *testing.T) {
	empty := Object{}
	envelope := Object{"version": "2.0.0", "tasks": []any{}, "inputs": []any{}}
	if !TasksEqual(empty, envelope) {
		t.Fatal("empty Tasks envelope should equal an absent document")
	}
	withTask := Object{"version": "2.0.0", "tasks": []any{map[string]any{"label": "build"}}, "inputs": []any{}}
	withoutEmptyInputs := Object{"version": "2.0.0", "tasks": []any{map[string]any{"label": "build"}}}
	if !TasksEqual(withTask, withoutEmptyInputs) {
		t.Fatal("an empty optional array should not change Tasks meaning")
	}
	if TasksEqual(empty, Object{"version": "1.0.0", "tasks": []any{}}) {
		t.Fatal("a non-default Tasks version must remain observable")
	}
	if TasksEqual(empty, Object{"version": "2.0.0", "windows": map[string]any{}}) {
		t.Fatal("opaque root content must remain observable")
	}
}
