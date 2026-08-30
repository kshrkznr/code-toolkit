// Package completion provides CTK's static shell-completion surface. It uses
// Cobra only as a completion generator; normal CTK command dispatch remains
// owned by cmd/ctk.
package completion

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var shells = []string{"bash", "zsh", "fish", "powershell"}

func Generate(output io.Writer, shell string) error {
	root := commandTree()
	switch shell {
	case "bash":
		return root.GenBashCompletionV2(output, false)
	case "zsh":
		return root.GenZshCompletionNoDesc(output)
	case "fish":
		return root.GenFishCompletion(output, false)
	case "powershell":
		return root.GenPowerShellCompletion(output)
	default:
		return fmt.Errorf("unsupported shell %q; expected bash, zsh, fish, or powershell", shell)
	}
}

// Complete handles Cobra's hidden completion protocol without resolving a CTK
// Workspace. Generated scripts invoke this path for static candidates.
func Complete(output, diagnostics io.Writer, args []string) error {
	root := commandTree()
	root.SetArgs(args)
	root.SetOut(output)
	root.SetErr(diagnostics)
	root.SilenceErrors = true
	root.SilenceUsage = true
	_, err := root.ExecuteC()
	return err
}

func commandTree() *cobra.Command {
	root := node("ctk", "CTK Cookbook and Runtime integration")
	root.CompletionOptions.DisableDefaultCmd = true

	activate := node("activate [platform]", "Import and manage a Platform's default Runtime")
	activate.Flags().Bool("force", false, "continue when the Platform is running")

	build := node("build [recipe-or-archive]", "Build a new Distribution")
	build.Flags().Bool("force", false, "replace conflicting output")
	build.Flags().Bool("keep-staging", false, "retain staging output")
	choiceFlag(build, "on-conflict", []string{"suffix", "abort"})

	apply := node("apply [recipe-or-archive] [dist]", "Converge an existing Distribution")
	apply.Flags().Bool("force", false, "continue when the Platform is running")

	archive := node("archive [dist]", "Preserve an offline-reconstructable Runtime")
	choiceFlag(archive, "on-conflict", []string{"suffix", "replace", "abort"})

	lock := node("lock [dist]", "Observe a Distribution into its Lock")

	freeze := node("freeze", "Generate or commit a Freeze Draft")
	freezeDraft := node("draft [dist]", "Generate a Freeze Draft Workbench")
	choiceFlag(freezeDraft, "on-conflict", []string{"abort", "replace"})
	freezeCommit := node("commit", "Commit present Draft Artifacts into the Cookbook")
	freezeCommit.Flags().Bool("force", false, "replace conflicting Cookbook content")
	freeze.AddCommand(freezeDraft, freezeCommit)

	view := node("view [source]", "View a Distribution, Recipe, or Ingredient")
	choiceFlag(view, "on-conflict", []string{"abort", "replace"})
	for _, child := range []*cobra.Command{
		node("dist [dist]", "View a Distribution Inventory"),
		node("recipe [recipe]", "View a resolved Recipe Inventory"),
		node("ingredient [all|layer|layer.name]", "View Ingredient content"),
	} {
		choiceFlag(child, "on-conflict", []string{"abort", "replace"})
		view.AddCommand(child)
	}

	syncCommand := node("sync [left] [right]", "Compare Distribution or Recipe completed states")
	choiceFlag(syncCommand, "on-conflict", []string{"abort", "replace"})

	deactivate := node("deactivate [platform]", "Restore a Platform's imported default Runtime")
	deactivate.Flags().Bool("force", false, "continue when the Platform is running")
	deactivate.Flags().Bool("force-empty", false, "restore an empty Runtime when origin is unavailable")

	launch := node("launch [dist] -- [args...]", "Temporarily launch a Distribution")

	workbench := node("workbench", "Open a Draft or Inspect Workbench")
	workbenchDraft := node("draft", "Open the Draft Workbench")
	workbenchInspect := node("inspect [viewpoint]", "Open an Inspect Workbench")
	workbenchDraft.Flags().String("editor", "", "editor command")
	workbenchInspect.Flags().String("editor", "", "editor command")
	workbench.AddCommand(workbenchDraft, workbenchInspect)

	docs := node("docs", "Navigate documentation packaged with this binary")
	docs.Flags().String("source", "", "explicit local documentation repository")
	docsShow := node("show <reference>", "Show a documentation reference")
	docsShow.Flags().String("depth", "", "heading depth or inclusive range")
	docs.AddCommand(
		node("help", "Show documentation help"),
		node("status", "Show Documentation Bundle provenance"),
		node("nodes", "List documentation nodes"),
		node("resolve <terms...>", "Resolve documentation terms"),
		node("toc <reference>", "Show a document table of contents"),
		docsShow,
		node("export <directory>", "Export packaged documentation"),
	)

	initCommand := node("init <path>", "Create an optional CTK Workspace footing")
	initCommand.Flags().Bool("exclude-sample", false, "create only the minimum Workspace directories")
	completionCommand := node("completion <shell>", "Generate a static shell-completion script")
	completionCommand.ValidArgs = append([]string(nil), shells...)

	root.AddCommand(
		activate, build, apply, archive, lock, freeze, view, syncCommand,
		node("list", "List Distributions"),
		node("current [platform]", "Show selected Runtime(s)"),
		deactivate,
		node("use [dist]", "Select a Runtime for its active Platform"),
		launch, workbench, docs,
		node("select", "Select a command interactively"),
		node("version", "Show binary version and build provenance"),
		node("help", "Show help"),
		initCommand, completionCommand,
	)
	return root
}

func node(use, description string) *cobra.Command {
	return &cobra.Command{
		Use:               use,
		Short:             description,
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		Run:               func(*cobra.Command, []string) {},
	}
}

func choiceFlag(command *cobra.Command, name string, values []string) {
	command.Flags().String(name, "", "conflict policy")
	_ = command.RegisterFlagCompletionFunc(name, func(*cobra.Command, []string, string) ([]cobra.Completion, cobra.ShellCompDirective) {
		result := make([]cobra.Completion, len(values))
		for index, value := range values {
			result[index] = cobra.Completion(value)
		}
		return result, cobra.ShellCompDirectiveNoFileComp
	})
}
