package completion

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateSupportedShells(t *testing.T) {
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			var output bytes.Buffer
			if err := Generate(&output, shell); err != nil {
				t.Fatal(err)
			}
			if output.Len() == 0 || !strings.Contains(strings.ToLower(output.String()), "ctk") {
				t.Fatalf("completion output is missing CTK registration:\n%s", output.String())
			}
		})
	}
	if err := Generate(&bytes.Buffer{}, "pwsh"); err == nil {
		t.Fatal("unsupported alias pwsh was accepted")
	}
}

func TestCompleteReturnsStaticCommandsAndOptions(t *testing.T) {
	for _, test := range []struct {
		args     []string
		expected []string
	}{
		{args: []string{"__complete", ""}, expected: []string{"init", "completion", "docs"}},
		{args: []string{"__complete", "init", "--"}, expected: []string{"--exclude-sample"}},
		{args: []string{"__complete", "completion", ""}, expected: []string{"bash", "zsh", "fish", "powershell"}},
		{args: []string{"__complete", "docs", ""}, expected: []string{"help", "status", "nodes", "resolve", "toc", "show", "export"}},
	} {
		var output, diagnostics bytes.Buffer
		if err := Complete(&output, &diagnostics, test.args); err != nil {
			t.Fatalf("Complete(%v): %v\n%s", test.args, err, diagnostics.String())
		}
		for _, expected := range test.expected {
			if !strings.Contains(output.String(), expected) {
				t.Fatalf("Complete(%v) does not contain %q:\n%s", test.args, expected, output.String())
			}
		}
	}
}
