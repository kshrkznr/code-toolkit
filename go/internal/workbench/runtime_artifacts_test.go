package workbench

import (
	"testing"

	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/runtimeartifact"
	"code-toolkit/internal/runtimelock"
)

func TestTasksDraftIgnoresSemanticEmptyEnvelope(t *testing.T) {
	plan := cookbook.Plan{Default: cookbook.ScopePlan{Tasks: runtimeartifact.Object{
		"version": "2.0.0", "tasks": []any{}, "inputs": []any{},
	}}}
	snapshot := runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Tasks: runtimeartifact.Object{}}}
	files, results, err := renderRuntimeArtifacts(&plan, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if results["tasks"].Status != "SAME" {
		t.Fatalf("Tasks status = %s, want SAME", results["tasks"].Status)
	}
	if _, exists := files["tasks.draft.md"]; exists {
		t.Fatal("semantic empty Tasks produced a Draft")
	}
}
