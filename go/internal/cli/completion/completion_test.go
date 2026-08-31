package completion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func TestHelpShowsCommandSyntaxOptionsAndDocumentation(t *testing.T) {
	for _, test := range []struct {
		name     string
		args     []string
		expected []string
		excluded []string
	}{
		{
			name: "activate",
			args: []string{"activate", "--help"},
			expected: []string{
				"Import and manage a Platform's default Runtime",
				"Usage:\n  ctk activate [platform] [flags]",
				"--force",
				"ctk docs show Knowledge.integration.code-venv.md#platform-activation",
			},
		},
		{
			name: "build options",
			args: []string{"build", "-h"},
			expected: []string{
				"Usage:\n  ctk build [recipe-or-archive] [flags]",
				"--force",
				"--keep-staging",
				"conflict policy (suffix|abort)",
				"ctk docs show Knowledge.core.build-lifecycle.md",
			},
		},
		{
			name: "nested command",
			args: []string{"freeze", "draft", "--help"},
			expected: []string{
				"Usage:\n  ctk freeze draft [dist] [flags]",
				"conflict policy (abort|replace)",
				"ctk docs show Knowledge.contract.workbench.md",
			},
		},
		{
			name: "view writes Inspect Artifacts",
			args: []string{"view", "--help"},
			expected: []string{
				"Generate a Workspace-local Inspect Inventory",
				"Writes disposable Artifacts under the Workspace Inspect Workbench.",
				"Usage:\n  ctk view [source] [flags]",
			},
		},
		{
			name: "help command syntax",
			args: []string{"help", "freeze", "draft"},
			expected: []string{
				"Usage:\n  ctk freeze draft [dist] [flags]",
				"conflict policy (abort|replace)",
			},
		},
		{
			name:     "simple command",
			args:     []string{"list", "--help"},
			expected: []string{"List Distributions", "Usage:\n  ctk list [flags]"},
			excluded: []string{"Related documentation:"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output, diagnostics bytes.Buffer
			if err := Help(&output, &diagnostics, test.args); err != nil {
				t.Fatalf("Help(%v): %v\n%s", test.args, err, diagnostics.String())
			}
			for _, expected := range test.expected {
				if !strings.Contains(output.String(), expected) {
					t.Fatalf("Help(%v) does not contain %q:\n%s", test.args, expected, output.String())
				}
			}
			for _, excluded := range test.excluded {
				if strings.Contains(output.String(), excluded) {
					t.Fatalf("Help(%v) unexpectedly contains %q:\n%s", test.args, excluded, output.String())
				}
			}
		})
	}
}

func TestHelpRequestStopsAtForwardedArguments(t *testing.T) {
	for _, test := range []struct {
		args []string
		want bool
	}{
		{args: []string{"activate", "--help"}, want: true},
		{args: []string{"freeze", "draft", "-h"}, want: true},
		{args: []string{"help", "activate"}, want: true},
		{args: []string{"help", "freeze", "draft"}, want: true},
		{args: []string{"help", "--help"}, want: false},
		{args: []string{"--help"}, want: false},
		{args: []string{"launch", "sample", "--", "--help"}, want: false},
	} {
		if got := IsHelpRequest(test.args); got != test.want {
			t.Fatalf("IsHelpRequest(%v) = %t, want %t", test.args, got, test.want)
		}
	}
}

func TestEverySubcommandRendersHelp(t *testing.T) {
	root := commandTree()
	var visit func(*cobra.Command)
	visit = func(command *cobra.Command) {
		for _, child := range command.Commands() {
			path := strings.Fields(child.CommandPath())
			args := append(append([]string(nil), path[1:]...), "--help")
			t.Run(strings.Join(path[1:], "/"), func(t *testing.T) {
				var output, diagnostics bytes.Buffer
				if err := Help(&output, &diagnostics, args); err != nil {
					t.Fatalf("Help(%v): %v\n%s", args, err, diagnostics.String())
				}
				if !strings.Contains(output.String(), "Usage:\n  "+child.UseLine()) {
					t.Fatalf("Help(%v) is missing usage %q:\n%s", args, child.UseLine(), output.String())
				}
			})
			visit(child)
		}
	}
	visit(root)
}
