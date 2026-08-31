// Package completion provides CTK's static command-description surface for
// subcommand help and shell completion. Normal CTK command dispatch remains
// owned by cmd/ctk.
package completion

import (
	"fmt"
	"io"
	"strings"

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

// IsHelpRequest reports whether args request help for a subcommand. Both
// "ctk help <command>" and "ctk <command> --help" use the same static command
// description. Arguments after -- belong to the launched Platform and are not
// CTK help requests.
func IsHelpRequest(args []string) bool {
	if len(args) < 2 {
		return false
	}
	if args[0] == "help" {
		return isHelpCommandRequest(args)
	}
	for _, arg := range args[1:] {
		if arg == "--" {
			return false
		}
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

// Help renders static subcommand help without resolving a CTK Workspace or
// loading the packaged Documentation Bundle.
func Help(output, diagnostics io.Writer, args []string) error {
	if isHelpCommandRequest(args) {
		args = append(append([]string(nil), args[1:]...), "--help")
	}
	root := commandTree()
	root.SetArgs(args)
	root.SetOut(output)
	root.SetErr(diagnostics)
	root.SilenceErrors = true
	root.SilenceUsage = true
	_, err := root.ExecuteC()
	return err
}

func isHelpCommandRequest(args []string) bool {
	if len(args) < 2 || args[0] != "help" {
		return false
	}
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			return false
		}
	}
	return true
}

func commandTree() *cobra.Command {
	root := node("ctk", "CTK Cookbook and Runtime integration")
	root.CompletionOptions.DisableDefaultCmd = true

	activate := documentedNode(
		"activate [platform]",
		"Import and manage a Platform's default Runtime",
		"When platform is omitted, CTK selects an available Platform interactively.",
		"Knowledge.integration.code-venv.md#platform-activation",
	)
	activate.Flags().Bool("force", false, "continue when the Platform is running")

	build := documentedNode(
		"build [recipe-or-archive]",
		"Build a new Distribution",
		"When the source is omitted, CTK selects a Recipe or Archive interactively.",
		"Knowledge.core.build-lifecycle.md",
	)
	build.Flags().Bool("force", false, "replace conflicting output")
	build.Flags().Bool("keep-staging", false, "retain staging output")
	choiceFlag(build, "on-conflict", []string{"suffix", "abort"})

	apply := documentedNode("apply [recipe-or-archive] [dist]", "Converge an existing Distribution", "", "Knowledge.core.build-lifecycle.md")
	apply.Flags().Bool("force", false, "continue when the Platform is running")

	archive := documentedNode("archive [dist]", "Preserve an offline-reconstructable Runtime", "", "Knowledge.core.persistence-lifecycle.md")
	choiceFlag(archive, "on-conflict", []string{"suffix", "replace", "abort"})

	lock := documentedNode("lock [dist]", "Observe a Distribution into its Lock", "", "Knowledge.core.persistence-lifecycle.md")

	freeze := documentedNode("freeze", "Generate or commit a Freeze Draft", "Choose draft or commit explicitly.", "Knowledge.contract.workbench.md")
	freezeDraft := documentedNode("draft [dist]", "Generate a Freeze Draft Workbench", "", "Knowledge.contract.workbench.md")
	choiceFlag(freezeDraft, "on-conflict", []string{"abort", "replace"})
	freezeCommit := documentedNode("commit", "Commit present Draft Artifacts into the Cookbook", "", "Knowledge.contract.workbench.md#freeze-commit-boundary")
	freezeCommit.Flags().Bool("force", false, "replace conflicting Cookbook content")
	freeze.AddCommand(freezeDraft, freezeCommit)

	view := documentedNode("view [source]", "View a Distribution, Recipe, or Ingredient", "", "Knowledge.contract.workbench.md#view-and-sync")
	choiceFlag(view, "on-conflict", []string{"abort", "replace"})
	for _, child := range []*cobra.Command{
		documentedNode("dist [dist]", "View a Distribution Inventory", "", "Knowledge.contract.workbench.md#view-and-sync"),
		documentedNode("recipe [recipe]", "View a resolved Recipe Inventory", "", "Knowledge.contract.workbench.md#view-and-sync"),
		documentedNode("ingredient [all|layer|layer.name]", "View Ingredient content", "", "Knowledge.contract.workbench.md#view-and-sync"),
	} {
		choiceFlag(child, "on-conflict", []string{"abort", "replace"})
		view.AddCommand(child)
	}

	syncCommand := documentedNode("sync [left] [right]", "Compare Distribution or Recipe completed states", "", "Knowledge.contract.workbench.md#view-and-sync")
	choiceFlag(syncCommand, "on-conflict", []string{"abort", "replace"})

	deactivate := documentedNode("deactivate [platform]", "Restore a Platform's imported default Runtime", "", "Knowledge.integration.code-venv.md#deactivation")
	deactivate.Flags().Bool("force", false, "continue when the Platform is running")
	deactivate.Flags().Bool("force-empty", false, "restore an empty Runtime when origin is unavailable")

	launch := documentedNode("launch [dist] -- [args...]", "Temporarily launch a Distribution", "Arguments after -- are forwarded to the Platform command.", "Knowledge.integration.code-venv.md#launch")

	workbench := documentedNode("workbench [draft|inspect] [viewpoint]", "Open a Draft or Inspect Workbench", "Choose draft or inspect explicitly, or omit it to select interactively.", "Knowledge.contract.workbench.md")
	workbenchDraft := documentedNode("draft", "Open the Draft Workbench", "", "Knowledge.contract.workbench.md")
	workbenchInspect := documentedNode("inspect [viewpoint]", "Open an Inspect Workbench", "", "Knowledge.contract.workbench.md")
	workbench.PersistentFlags().String("editor", "", "editor command")
	workbench.AddCommand(workbenchDraft, workbenchInspect)

	docs := node("docs [node]", "Navigate documentation packaged with this binary")
	docs.Long = `Navigate documentation packaged with this binary.

With no operation, docs shows the Concept and resolver Bootstrap. Use resolve
to find a reference, toc to inspect its headings, and show to read it. Resolve
searches identities, paths, Node aliases, titles, and headings, not bodies.
Place --source before the operation to select an explicit local repository.`
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

	initCommand := documentedNode("init <path>", "Create an optional CTK Workspace footing", "", "Knowledge.integration.workspace.md")
	initCommand.Flags().Bool("exclude-sample", false, "create only the minimum Workspace directories")
	completionCommand := node("completion <shell>", "Generate a static shell-completion script")
	completionCommand.ValidArgs = append([]string(nil), shells...)

	root.AddCommand(
		activate, build, apply, archive, lock, freeze, view, syncCommand,
		node("list", "List Distributions"),
		documentedNode("current [platform]", "Show selected Runtime(s)", "", "Knowledge.integration.code-venv.md#runtime-selection"),
		deactivate,
		documentedNode("use [dist]", "Select a Runtime for its active Platform", "", "Knowledge.integration.code-venv.md#runtime-selection"),
		launch, workbench, docs,
		node("select", "Select a command interactively"),
		node("version", "Show binary version and build provenance"),
		node("help [command]", "Show root or command help"),
		initCommand, completionCommand,
	)
	return root
}

func documentedNode(use, description, detail, reference string) *cobra.Command {
	command := node(use, description)
	sections := []string{description}
	if detail != "" {
		sections = append(sections, detail)
	}
	sections = append(sections, "Related documentation:\n  ctk docs show "+reference)
	command.Long = strings.Join(sections, "\n\n")
	return command
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
	command.Flags().String(name, "", "conflict policy ("+strings.Join(values, "|")+")")
	_ = command.RegisterFlagCompletionFunc(name, func(*cobra.Command, []string, string) ([]cobra.Completion, cobra.ShellCompDirective) {
		result := make([]cobra.Completion, len(values))
		for index, value := range values {
			result[index] = cobra.Completion(value)
		}
		return result, cobra.ShellCompDirectiveNoFileComp
	})
}
