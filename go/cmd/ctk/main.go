package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	ctkarchive "github.com/kshrkznr/code-toolkit/go/internal/archive"
	"github.com/kshrkznr/code-toolkit/go/internal/buildinfo"
	"github.com/kshrkznr/code-toolkit/go/internal/cli/selector"
	"github.com/kshrkznr/code-toolkit/go/internal/codevenv"
	"github.com/kshrkznr/code-toolkit/go/internal/converge"
	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/docbundle"
	"github.com/kshrkznr/code-toolkit/go/internal/launcher"
	"github.com/kshrkznr/code-toolkit/go/internal/lifecycle"
	"github.com/kshrkznr/code-toolkit/go/internal/platform"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/recovery"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio/vscode"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
	"github.com/kshrkznr/code-toolkit/go/internal/workbench"
	ctkworkspace "github.com/kshrkznr/code-toolkit/go/internal/workspace"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		if !errors.Is(err, selector.ErrCancelled) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(exitCode(err))
	}
}

func run(args []string) error {
	if len(args) == 0 {
		args = []string{"select"}
	}
	if handled, err := runSelfDescription(args); handled {
		return err
	}

	root, err := projectRoot()
	if err != nil {
		return err
	}
	paths, err := ctkworkspace.Load(root)
	if err != nil {
		return err
	}
	distDir := paths.Dist
	cookbookDir := paths.CookbookSource
	workbenchDir := paths.Workbench
	archiveDir := paths.Archive
	poolDir := paths.Pool
	nativeSelector := selector.New()
	service := lifecycleService(poolDir, cookbookDir, nativeSelector)

	switch args[0] {
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("usage: ctk list")
		}
		names, err := distribution.List(distDir)
		if err != nil {
			return err
		}
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	case "current":
		if len(args) > 2 {
			return fmt.Errorf("usage: ctk current [platform]")
		}
		platform := ""
		if len(args) == 2 {
			platform = args[1]
		}
		selections, err := codevenv.Current(distDir, platform)
		if err != nil {
			return err
		}
		if platform != "" {
			fmt.Println(selections[platform])
			return nil
		}
		if len(selections) == 0 {
			fmt.Println("none")
			return nil
		}
		for _, name := range codevenv.Platforms(selections) {
			fmt.Printf("%s: %s\n", name, selections[name])
		}
		return nil
	case "select":
		if len(args) != 1 {
			return fmt.Errorf("usage: ctk select")
		}
		command, err := nativeSelector.Select("Select command", []string{"activate", "apply", "archive", "build", "current", "deactivate", "docs", "freeze commit", "freeze draft", "help", "launch", "list", "lock", "sync", "use", "version", "view", "workbench"})
		if err != nil {
			return err
		}
		return run(strings.Fields(command))
	case "use":
		if len(args) > 2 {
			return fmt.Errorf("usage: ctk use [dist]")
		}
		name, err := selectDistribution(distDir, args[1:], nativeSelector)
		if err != nil {
			return err
		}
		dist, err := distribution.Load(distDir, name)
		if err != nil {
			return err
		}
		result, err := codevenv.Use(context.Background(), distDir, dist, platform.NewProcessStopper())
		if err != nil {
			return err
		}
		if result.Changed {
			fmt.Printf("[changed] %s: %s\n", result.Platform, result.Current)
		} else {
			fmt.Printf("[current] %s: %s\n", result.Platform, result.Current)
		}
		return nil
	case "activate":
		platformName, force, err := parsePlatformForce(args[1:], false)
		if err != nil {
			return fmt.Errorf("usage: ctk activate [platform] [--force]: %w", err)
		}
		if platformName == "" {
			platformName, err = selectAvailablePlatform(nativeSelector)
			if err != nil {
				return err
			}
		}
		result, err := codevenvService(poolDir, distDir, nativeSelector).Activate(context.Background(), platformName, codevenv.ActivateOptions{Force: force})
		if err != nil {
			return err
		}
		if result.NoOp {
			fmt.Printf("[active] %s: %s\n", result.Platform, result.Current)
		} else if result.Forced {
			fmt.Printf("[activated:forced] %s: %s\n", result.Platform, result.Current)
		} else {
			fmt.Printf("[activated] %s: %s\n", result.Platform, result.Current)
		}
		return nil
	case "deactivate":
		platformName, force, forceEmpty, err := parseDeactivateArgs(args[1:])
		if err != nil {
			return fmt.Errorf("usage: ctk deactivate [platform] [--force|--force-empty]: %w", err)
		}
		if platformName == "" {
			selections, err := codevenv.Current(distDir, "")
			if err != nil {
				return err
			}
			platformName, err = nativeSelector.Select("Select Platform to deactivate", codevenv.Platforms(selections))
			if err != nil {
				return err
			}
		}
		result, err := codevenvService(poolDir, distDir, nativeSelector).Deactivate(context.Background(), platformName, codevenv.DeactivateOptions{Force: force, ForceEmpty: forceEmpty})
		if err != nil {
			return err
		}
		if result.Empty {
			fmt.Printf("[deactivated:empty] %s\n", result.Platform)
		} else if result.Forced {
			fmt.Printf("[deactivated:forced] %s\n", result.Platform)
		} else {
			fmt.Printf("[deactivated] %s\n", result.Platform)
		}
		return printActivePlatforms(os.Stdout, distDir)
	case "launch":
		distArg, forward, err := parseLaunchArgs(args[1:])
		if err != nil {
			return err
		}
		name, err := selectDistribution(distDir, nonEmptyArg(distArg), nativeSelector)
		if err != nil {
			return err
		}
		runtimeLauncher := launcher.New()
		dist, err := distribution.LoadForLaunch(distDir, name, launcher.OverrideName(runtimeLauncher.GOOS))
		if err != nil {
			return err
		}
		return runtimeLauncher.Launch(context.Background(), dist, forward)
	case "workbench":
		return openWorkbench(workbenchDir, args[1:], nativeSelector)
	case "build":
		options, err := parseBuildArgs(args[1:])
		if err != nil {
			return err
		}
		sourceKind, sourcePath, err := selectRuntimeSource(archiveDir, cookbookDir, options.recipe, nativeSelector)
		if err != nil {
			return err
		}
		var definition recipe.Recipe
		var bundle ctkarchive.Bundle
		if sourceKind == "archive" {
			bundle, err = ctkarchive.Load(sourcePath)
			definition = bundle.Recipe
		} else {
			sourcePath, err = selectRecipe(cookbookDir, sourcePath, nativeSelector)
			if err == nil {
				definition, err = recipe.Load(sourcePath)
			}
		}
		if err != nil {
			return err
		}
		name := definition.Name
		available, err := lifecycle.NextAvailableName(distDir, name)
		if err != nil {
			return err
		}
		if available != name {
			mode := options.onConflict
			if mode == "" && isTerminal() {
				choice, err := nativeSelector.Select("Distribution already exists", []string{"Abort", "Build as " + available})
				if err != nil {
					return err
				}
				if choice == "Abort" {
					return selector.ErrCancelled
				}
				mode = "suffix"
			}
			if mode == "" {
				mode = "suffix"
			}
			if mode == "abort" {
				return fmt.Errorf("Distribution already exists: %s", name)
			}
			name = available
		}
		var result lifecycle.Result
		if sourceKind == "archive" {
			if options.force {
				return fmt.Errorf("--force is available only for Recipe Build")
			}
			result, err = service.BuildArchive(context.Background(), bundle, distDir, name, options.keepStaging)
		} else {
			result, err = service.Build(context.Background(), sourcePath, distDir, name, options.keepStaging, options.force)
		}
		result.Report.Print(os.Stderr)
		if err != nil {
			if result.StagingPath != "" && options.keepStaging {
				fmt.Fprintf(os.Stderr, "[staging retained] %s\n", result.StagingPath)
			}
			return err
		}
		fmt.Printf("[built] %s\n", result.Distribution.Name)
		for _, warning := range result.Warnings {
			fmt.Fprintln(os.Stderr, "[warn]", warning)
		}
		return nil
	case "apply":
		options, err := parseApplyArgs(args[1:])
		if err != nil {
			return err
		}
		sourceArg := options.source
		sourceKind, sourcePath, err := selectRuntimeSource(archiveDir, cookbookDir, sourceArg, nativeSelector)
		if err != nil {
			return err
		}
		var definition recipe.Recipe
		var bundle ctkarchive.Bundle
		if sourceKind == "archive" {
			bundle, err = ctkarchive.Load(sourcePath)
			definition = bundle.Recipe
		} else {
			sourcePath, err = selectRecipe(cookbookDir, sourcePath, nativeSelector)
			if err == nil {
				definition, err = recipe.Load(sourcePath)
			}
		}
		if err != nil {
			return err
		}
		distName := ""
		if options.dist != "" {
			distName = options.dist
		} else {
			distName, err = selectIdentityDistribution(distDir, definition, nativeSelector)
			if err != nil {
				return err
			}
		}
		dist, err := distribution.Load(distDir, distName)
		if err != nil {
			return err
		}
		var result lifecycle.Result
		if sourceKind == "archive" {
			if options.force {
				return fmt.Errorf("--force is available only for Recipe Apply")
			}
			result, err = service.ApplyArchive(context.Background(), bundle, dist)
		} else {
			result, err = service.Apply(context.Background(), sourcePath, dist, options.force)
		}
		result.Report.Print(os.Stderr)
		if err != nil {
			return err
		}
		fmt.Printf("[applied] %s\n", result.Distribution.Name)
		for _, warning := range result.Warnings {
			fmt.Fprintln(os.Stderr, "[warn]", warning)
		}
		return nil
	case "archive":
		options, err := parseArchiveArgs(args[1:])
		if err != nil {
			return err
		}
		name, err := selectDistribution(distDir, nonEmptyArg(options.dist), nativeSelector)
		if err != nil {
			return err
		}
		dist, err := distribution.Load(distDir, name)
		if err != nil {
			return err
		}
		mode := options.onConflict
		archiveRoot := archiveDir
		if pathIsDir(filepath.Join(archiveRoot, dist.Name)) && mode == "" && isTerminal() {
			available, _ := ctkarchive.NextAvailableName(archiveRoot, dist.Name)
			choice, selectErr := nativeSelector.Select("Archive already exists", []string{"Abort", "Replace", "Archive as " + available})
			if selectErr != nil {
				return selectErr
			}
			switch choice {
			case "Abort":
				return selector.ErrCancelled
			case "Replace":
				mode = "replace"
			default:
				mode = "suffix"
			}
		}
		result, err := archiveService(poolDir, cookbookDir, nativeSelector).Create(context.Background(), archiveRoot, dist, ctkarchive.Options{OnConflict: mode})
		if err != nil {
			return err
		}
		for _, warning := range result.Warnings {
			fmt.Fprintln(os.Stderr, "[warn]", warning)
		}
		fmt.Printf("[archived] %s\n", result.Bundle.Manifest.Name)
		return nil
	case "lock":
		if len(args) > 2 {
			return fmt.Errorf("usage: ctk lock [dist]")
		}
		name, err := selectDistribution(distDir, args[1:], nativeSelector)
		if err != nil {
			return err
		}
		dist, err := distribution.Load(distDir, name)
		if err != nil {
			return err
		}
		report, err := service.Lock(context.Background(), dist)
		report.Print(os.Stderr)
		if err != nil {
			return err
		}
		fmt.Printf("[locked] %s\n", dist.Name)
		return nil
	case "freeze":
		if len(args) < 2 {
			return fmt.Errorf("usage: ctk freeze draft [dist] [--on-conflict abort|replace] | freeze commit [--force]")
		}
		if args[1] == "commit" {
			force := false
			for _, arg := range args[2:] {
				if arg == "--force" {
					force = true
				} else {
					return fmt.Errorf("usage: ctk freeze commit [--force]")
				}
			}
			result, err := (workbench.Service{WorkspaceRoot: root, CookbookRoot: cookbookDir, WorkbenchRoot: workbenchDir}).Commit(force)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "[operations] completed=%d unresolved=%d failed=0\n", result.Completed, result.Unresolved)
			fmt.Printf("[committed] files=%d\n", len(result.Files))
			return nil
		}
		if args[1] != "draft" {
			return fmt.Errorf("usage: ctk freeze draft [dist] [--on-conflict abort|replace] | freeze commit [--force]")
		}
		options, err := parseWorkbenchArgs(args[2:])
		if err != nil {
			return err
		}
		return generateWorkbench(context.Background(), root, distDir, cookbookDir, workbenchDir, workbench.FreezeDraft, options, nativeSelector)
	case "view":
		viewType, viewArgs := "", args[1:]
		if len(viewArgs) > 0 && (viewArgs[0] == "dist" || viewArgs[0] == "recipe" || viewArgs[0] == "ingredient") {
			viewType, viewArgs = viewArgs[0], viewArgs[1:]
		}
		options, err := parseWorkbenchArgs(viewArgs)
		if err != nil {
			return err
		}
		if viewType == "" && options.dist == "" {
			viewType, err = nativeSelector.Select("Select View source type", []string{"dist", "recipe", "ingredient"})
			if err != nil {
				return err
			}
		}
		return generateTypedView(context.Background(), root, distDir, cookbookDir, workbenchDir, viewType, options, nativeSelector)
	case "sync":
		options, inputs, err := parseSyncArgs(args[1:])
		if err != nil {
			return err
		}
		service := workbenchService(root, cookbookDir, workbenchDir, nativeSelector)
		var left, right workbench.CompletedSource
		if len(inputs) > 0 {
			left, err = resolveCompletedSource(context.Background(), service, distDir, cookbookDir, inputs[0])
		} else {
			left, err = selectCompletedSource(context.Background(), service, distDir, cookbookDir, "left", nativeSelector)
		}
		if err != nil {
			return err
		}
		if len(inputs) > 1 {
			right, err = resolveCompletedSource(context.Background(), service, distDir, cookbookDir, inputs[1])
		} else {
			right, err = selectCompletedSource(context.Background(), service, distDir, cookbookDir, "right", nativeSelector)
		}
		if err != nil {
			return err
		}
		conflict, err := inspectConflict(filepath.Join(workbenchDir, "inspect", "sync."+sourceLabel(left)+"."+sourceLabel(right)), options.onConflict, "Sync", nativeSelector)
		if err != nil {
			return err
		}
		result, err := service.GenerateSync(left, right, conflict)
		if err != nil {
			return err
		}
		fmt.Printf("[sync] %s\n", result.Path)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runSelfDescription(args []string) (bool, error) {
	switch args[0] {
	case "help", "-h", "--help":
		if len(args) != 1 {
			return true, fmt.Errorf("usage: ctk help")
		}
		usage()
		return true, nil
	case "version", "--version":
		if len(args) != 1 {
			return true, fmt.Errorf("usage: ctk version")
		}
		fmt.Println(buildinfo.String())
		return true, nil
	case "docs":
		if len(args) == 2 && slices.Contains([]string{"help", "-h", "--help"}, args[1]) {
			return true, writeDocsUsage(os.Stdout)
		}
		bundle, err := executableDocumentationBundle()
		if err != nil {
			return true, err
		}
		return true, runDocs(os.Stdout, bundle, args[1:])
	default:
		return false, nil
	}
}

func executableDocumentationBundle() (*docbundle.Bundle, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable Documentation Bundle: %w", err)
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return nil, fmt.Errorf("resolve executable Documentation Bundle symlinks: %w", err)
	}
	bundle, err := docbundle.OpenExecutable(executable)
	if err != nil {
		return nil, fmt.Errorf("load packaged documentation: %w", err)
	}
	return bundle, nil
}

func runDocs(output io.Writer, bundle *docbundle.Bundle, args []string) error {
	if len(args) == 0 {
		_, err := output.Write(bundle.Bootstrap())
		return err
	}
	switch args[0] {
	case "help", "-h", "--help":
		if len(args) != 1 {
			return fmt.Errorf("usage: ctk docs help")
		}
		return writeDocsUsage(output)
	case "nodes":
		if len(args) != 1 {
			return fmt.Errorf("usage: ctk docs nodes")
		}
		for _, document := range bundle.Manifest().Documents {
			for _, alias := range document.Aliases {
				if _, err := fmt.Fprintf(output, "%s\t%s\t%s\n", alias, document.Path, document.Title); err != nil {
					return err
				}
			}
		}
		return nil
	case "resolve":
		if len(args) < 2 {
			return fmt.Errorf("usage: ctk docs resolve <terms...>")
		}
		candidates := bundle.Resolve(args[1:])
		if len(candidates) == 0 {
			return fmt.Errorf("no bundled documentation metadata matches: %s; try fewer terms or report a missing documentation route", strings.Join(args[1:], " "))
		}
		if _, err := fmt.Fprintln(output, "IDENTITY\tPATH\tTITLE\tMATCHED"); err != nil {
			return err
		}
		limit := min(len(candidates), 10)
		for _, candidate := range candidates[:limit] {
			identity := candidate.Identity
			if identity == "" {
				identity = "-"
			}
			if _, err := fmt.Fprintf(output, "%s\t%s\t%s\t%s\n", identity, candidate.Path, candidate.Title, strings.Join(candidate.Matched, ",")); err != nil {
				return err
			}
		}
		if len(candidates) > limit {
			_, err := fmt.Fprintf(output, "# showing %d of %d candidates; add terms to narrow\n", limit, len(candidates))
			return err
		}
		return nil
	case "show":
		if len(args) != 2 {
			return fmt.Errorf("usage: ctk docs show <canonical-identity-or-path>[#heading]")
		}
		content, err := bundle.Show(args[1])
		if err != nil {
			return fmt.Errorf("%w; find bundled references with: ctk docs resolve <terms...>; full repository: %s", err, bundle.RepositoryReference())
		}
		_, err = output.Write(content)
		return err
	case "export":
		return fmt.Errorf("ctk docs export is not implemented")
	default:
		if len(args) != 1 {
			return fmt.Errorf("usage: ctk docs [<node>|resolve <terms...>|show <reference>]")
		}
		content, err := bundle.ShowNode(args[0])
		if err != nil {
			return err
		}
		_, err = output.Write(content)
		return err
	}
}

func writeDocsUsage(output io.Writer) error {
	_, err := fmt.Fprint(output, `Usage:
  ctk docs                         Show the Concept and resolver Bootstrap
  ctk docs nodes                   List short navigation Node aliases
  ctk docs <node>                  Show one Node README (for example: core)
  ctk docs resolve <terms...>      Rank matching bundled documents
  ctk docs show <reference>        Show an identity or path, optionally #heading

Resolve output is tab-separated. Copy its IDENTITY or PATH into docs show.
Resolve searches identity, path, Node alias, title, and headings, not bodies.
Repository-only material is linked at this binary's exact tag or commit.
`)
	return err
}

func printActivePlatforms(output io.Writer, distDir string) error {
	selections, err := codevenv.Current(distDir, "")
	if err != nil {
		return err
	}
	platforms := codevenv.Platforms(selections)
	if len(platforms) == 0 {
		_, err = fmt.Fprintln(output, "[active] none")
		return err
	}
	for _, platformName := range platforms {
		if _, err := fmt.Fprintf(output, "[active] %s: %s\n", platformName, selections[platformName]); err != nil {
			return err
		}
	}
	label := "Platform"
	if len(platforms) != 1 {
		label = "Platforms"
	}
	if _, err := fmt.Fprintf(output, "[hint] deactivate remaining %s individually:\n", label); err != nil {
		return err
	}
	for _, platformName := range platforms {
		if _, err := fmt.Fprintf(output, "  ctk deactivate %s\n", platformName); err != nil {
			return err
		}
	}
	return nil
}

type workbenchOptions struct{ dist, onConflict string }

type archiveOptions struct{ dist, onConflict string }

func parseArchiveArgs(args []string) (archiveOptions, error) {
	options := archiveOptions{}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--on-conflict":
			index++
			if index >= len(args) {
				return options, fmt.Errorf("--on-conflict requires suffix, replace, or abort")
			}
			options.onConflict = args[index]
			if options.onConflict != "suffix" && options.onConflict != "replace" && options.onConflict != "abort" {
				return options, fmt.Errorf("invalid Archive conflict mode: %s", options.onConflict)
			}
		default:
			if strings.HasPrefix(args[index], "-") {
				return options, fmt.Errorf("unknown Archive option: %s", args[index])
			}
			if options.dist != "" {
				return options, fmt.Errorf("usage: ctk archive [dist] [--on-conflict suffix|replace|abort]")
			}
			options.dist = args[index]
		}
	}
	return options, nil
}

func parseSyncArgs(args []string) (workbenchOptions, []string, error) {
	options := workbenchOptions{}
	inputs := []string{}
	for index := 0; index < len(args); index++ {
		if args[index] == "--on-conflict" {
			index++
			if index >= len(args) || (args[index] != "abort" && args[index] != "replace") {
				return options, nil, fmt.Errorf("--on-conflict requires abort or replace")
			}
			options.onConflict = args[index]
		} else if strings.HasPrefix(args[index], "-") {
			return options, nil, fmt.Errorf("unknown Sync option: %s", args[index])
		} else {
			inputs = append(inputs, args[index])
		}
	}
	if len(inputs) > 2 {
		return options, nil, fmt.Errorf("usage: ctk sync [left] [right] [--on-conflict abort|replace]")
	}
	return options, inputs, nil
}

func parseWorkbenchArgs(args []string) (workbenchOptions, error) {
	var options workbenchOptions
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--on-conflict":
			index++
			if index >= len(args) {
				return options, fmt.Errorf("--on-conflict requires abort or replace")
			}
			options.onConflict = args[index]
			if options.onConflict != "abort" && options.onConflict != "replace" {
				return options, fmt.Errorf("invalid --on-conflict: %s", options.onConflict)
			}
		default:
			if strings.HasPrefix(args[index], "-") {
				return options, fmt.Errorf("unknown Workbench option: %s", args[index])
			}
			if options.dist != "" {
				return options, fmt.Errorf("multiple Distributions provided")
			}
			options.dist = args[index]
		}
	}
	return options, nil
}

func generateWorkbench(ctx context.Context, root, distDir, cookbookDir, workbenchDir string, kind workbench.Kind, options workbenchOptions, selection selector.Selector) error {
	name, err := selectDistribution(distDir, nonEmptyArg(options.dist), selection)
	if err != nil {
		return err
	}
	dist, err := distribution.Load(distDir, name)
	if err != nil {
		return err
	}
	directory := filepath.Join("inspect", "dist."+dist.Name)
	if kind == workbench.FreezeDraft {
		directory = "draft"
	}
	conflict := options.onConflict
	if _, err := os.Lstat(filepath.Join(workbenchDir, directory)); err == nil && conflict == "" {
		if !isTerminal() {
			conflict = "abort"
		} else {
			choice, err := selection.Select(string(kind)+" Workbench already exists", []string{"Abort", "Replace"})
			if err != nil {
				return err
			}
			if choice == "Abort" {
				return selector.ErrCancelled
			}
			conflict = "replace"
		}
	}
	service := workbenchService(root, cookbookDir, workbenchDir, selection)
	result, err := service.Generate(ctx, kind, dist, conflict)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] %s\n", strings.ToLower(strings.ReplaceAll(string(kind), " ", "-")), result.Path)
	return nil
}

func nonEmptyArg(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func parseLaunchArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, nil
	}
	dist := args[0]
	forward := args[1:]
	if dist == "--" {
		dist = ""
	} else if len(forward) > 0 && forward[0] == "--" {
		forward = forward[1:]
	}
	return dist, forward, nil
}

type openWorkbenchOptions struct {
	kind, viewpoint, editor string
}

func parseOpenWorkbenchArgs(args []string) (openWorkbenchOptions, error) {
	var options openWorkbenchOptions
	for index := 0; index < len(args); index++ {
		if args[index] == "--editor" {
			index++
			if index >= len(args) || args[index] == "" {
				return options, fmt.Errorf("--editor requires a command")
			}
			options.editor = args[index]
			continue
		}
		if strings.HasPrefix(args[index], "--") {
			return options, fmt.Errorf("unknown workbench option: %s", args[index])
		}
		if options.kind == "" {
			options.kind = args[index]
		} else if options.viewpoint == "" {
			options.viewpoint = args[index]
		} else {
			return options, fmt.Errorf("too many workbench arguments")
		}
	}
	if options.kind != "" && options.kind != "draft" && options.kind != "inspect" {
		return options, fmt.Errorf("workbench must be draft or inspect: %s", options.kind)
	}
	if options.kind == "draft" && options.viewpoint != "" {
		return options, fmt.Errorf("draft Workbench does not accept a viewpoint")
	}
	return options, nil
}

func openWorkbench(cookbookDir string, args []string, selection selector.Selector) error {
	options, err := parseOpenWorkbenchArgs(args)
	if err != nil {
		return fmt.Errorf("usage: ctk workbench [draft|inspect] [viewpoint] [--editor command]: %w", err)
	}
	if options.kind == "" {
		var kinds []string
		if pathIsDir(filepath.Join(cookbookDir, "draft")) {
			kinds = append(kinds, "draft")
		}
		if values, _ := inspectWorkbenches(cookbookDir); len(values) > 0 {
			kinds = append(kinds, "inspect")
		}
		options.kind, err = selection.Select("Select Workbench", kinds)
		if err != nil {
			return err
		}
	}

	target := filepath.Join(cookbookDir, "draft")
	if options.kind == "inspect" {
		viewpoints, listErr := inspectWorkbenches(cookbookDir)
		if listErr != nil {
			return listErr
		}
		if options.viewpoint == "" {
			options.viewpoint, err = selection.Select("Select Inspect Workbench", viewpoints)
			if err != nil {
				return err
			}
		}
		if !slices.Contains(viewpoints, options.viewpoint) {
			return fmt.Errorf("Inspect Workbench not found: %s", options.viewpoint)
		}
		target = filepath.Join(cookbookDir, "inspect", options.viewpoint)
	}
	if !pathIsDir(target) {
		return fmt.Errorf("Workbench not found: %s", target)
	}

	editor := options.editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		if _, lookErr := exec.LookPath("code"); lookErr == nil {
			editor = "code"
		} else {
			editor = "vim"
		}
	}
	command := exec.Command(editor, target)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("open %s Workbench with %s: %w", options.kind, editor, err)
	}
	return nil
}

func inspectWorkbenches(cookbookDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(cookbookDir, "inspect"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list Inspect Workbenches: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func generateTypedView(ctx context.Context, root, distDir, cookbookDir, workbenchDir, explicit string, options workbenchOptions, selection selector.Selector) error {
	service := workbenchService(root, cookbookDir, workbenchDir, selection)
	kind, value, err := detectViewSource(distDir, cookbookDir, explicit, options.dist)
	if err != nil {
		return err
	}
	switch kind {
	case "dist":
		name := value
		if name == "" {
			name, err = selectDistribution(distDir, nil, selection)
			if err != nil {
				return err
			}
		}
		if info, statErr := os.Stat(value); value != "" && statErr == nil && info.IsDir() {
			distDir, name = filepath.Dir(value), filepath.Base(value)
		}
		dist, loadErr := distribution.Load(distDir, name)
		if loadErr != nil {
			return loadErr
		}
		options.dist = dist.Name
		return generateWorkbench(ctx, root, distDir, cookbookDir, workbenchDir, workbench.View, options, selection)
	case "recipe":
		path, selectErr := selectRecipe(cookbookDir, value, selection)
		if selectErr != nil {
			return selectErr
		}
		source, sourceErr := service.RecipeSource(path)
		if sourceErr != nil {
			return sourceErr
		}
		target := filepath.Join(workbenchDir, "inspect", "recipe."+source.Name)
		conflict, conflictErr := inspectConflict(target, options.onConflict, "Recipe View", selection)
		if conflictErr != nil {
			return conflictErr
		}
		result, generateErr := service.GenerateRecipeView(source, conflict)
		if generateErr != nil {
			return generateErr
		}
		fmt.Printf("[view] %s\n", result.Path)
		return nil
	case "ingredient":
		query := value
		if query == "" {
			candidates, candidateErr := ingredientCandidates(cookbookDir)
			if candidateErr != nil {
				return candidateErr
			}
			scope, selectErr := selection.Select("Select Ingredient view scope", []string{"all", "layer", "ingredient"})
			if selectErr != nil {
				return selectErr
			}
			switch scope {
			case "all":
				query = "all"
			case "layer":
				layers := ingredientLayers(candidates)
				query, err = selection.Select("Select Ingredient layer", layers)
				if err != nil {
					return err
				}
			case "ingredient":
				query, err = selection.Select("Select Ingredient", candidates)
				if err != nil {
					return err
				}
			}
		}
		target := filepath.Join(workbenchDir, "inspect", "ingredient."+strings.ReplaceAll(query, "/", "-"))
		conflict, conflictErr := inspectConflict(target, options.onConflict, "Ingredient View", selection)
		if conflictErr != nil {
			return conflictErr
		}
		result, generateErr := service.GenerateIngredientView(query, conflict)
		if generateErr != nil {
			return generateErr
		}
		fmt.Printf("[view] %s\n", result.Path)
		return nil
	default:
		return fmt.Errorf("unsupported View source %q", kind)
	}
}

func detectViewSource(distDir, cookbookDir, explicit, value string) (string, string, error) {
	if explicit != "" {
		return explicit, value, nil
	}
	if value == "" {
		return "dist", "", nil
	}
	if info, err := os.Stat(value); err == nil {
		absolute, _ := filepath.Abs(value)
		if info.IsDir() {
			if _, metaErr := os.Stat(filepath.Join(absolute, ".meta", "recipe.yaml")); metaErr == nil {
				return "dist", absolute, nil
			}
			relative, relErr := filepath.Rel(filepath.Join(cookbookDir, "ingredient"), absolute)
			if relErr == nil && !strings.HasPrefix(relative, "..") {
				parts := strings.Split(filepath.ToSlash(relative), "/")
				if len(parts) >= 2 {
					return "ingredient", parts[0] + "." + parts[1], nil
				}
			}
			return "", "", fmt.Errorf("cannot infer View directory type: %s", value)
		}
		if strings.HasSuffix(strings.ToLower(value), ".yaml") {
			return "recipe", absolute, nil
		}
		return "", "", fmt.Errorf("cannot infer View file type: %s", value)
	}
	distExists := pathIsDir(filepath.Join(distDir, value))
	recipeValue := value
	if filepath.Ext(recipeValue) == "" {
		recipeValue += ".yaml"
	}
	recipeExists := pathIsFile(filepath.Join(cookbookDir, "recipe", recipeValue))
	if distExists && recipeExists {
		return "", "", fmt.Errorf("ambiguous View source %q; use view dist or view recipe", value)
	}
	if distExists {
		return "dist", value, nil
	}
	if recipeExists {
		return "recipe", value, nil
	}
	if strings.Contains(value, ".") {
		return "ingredient", value, nil
	}
	return "", "", fmt.Errorf("View source not found: %s", value)
}

func resolveCompletedSource(ctx context.Context, service workbench.Service, distDir, cookbookDir, input string) (workbench.CompletedSource, error) {
	kind, value, err := detectViewSource(distDir, cookbookDir, "", input)
	if err != nil {
		return workbench.CompletedSource{}, err
	}
	switch kind {
	case "dist":
		name := value
		if pathIsDir(value) {
			distDir, name = filepath.Dir(value), filepath.Base(value)
		}
		dist, err := distribution.Load(distDir, name)
		if err != nil {
			return workbench.CompletedSource{}, err
		}
		return service.DistributionSource(ctx, dist)
	case "recipe":
		path, err := selectRecipe(cookbookDir, value, selector.New())
		if err != nil {
			return workbench.CompletedSource{}, err
		}
		return service.RecipeSource(path)
	default:
		return workbench.CompletedSource{}, fmt.Errorf("sync does not accept Ingredient source: %s", input)
	}
}

func selectCompletedSource(ctx context.Context, service workbench.Service, distDir, cookbookDir, side string, selection selector.Selector) (workbench.CompletedSource, error) {
	kind, err := selection.Select("Select "+side+" Sync source type", []string{"dist", "recipe"})
	if err != nil {
		return workbench.CompletedSource{}, err
	}
	if kind == "dist" {
		name, err := selectDistribution(distDir, nil, selection)
		if err != nil {
			return workbench.CompletedSource{}, err
		}
		dist, err := distribution.Load(distDir, name)
		if err != nil {
			return workbench.CompletedSource{}, err
		}
		return service.DistributionSource(ctx, dist)
	}
	path, err := selectRecipe(cookbookDir, "", selection)
	if err != nil {
		return workbench.CompletedSource{}, err
	}
	return service.RecipeSource(path)
}
func inspectConflict(target, current, label string, selection selector.Selector) (string, error) {
	if _, err := os.Lstat(target); err == nil && current == "" {
		if !isTerminal() {
			return "abort", nil
		}
		choice, selectErr := selection.Select(label+" Workbench already exists", []string{"Abort", "Replace"})
		if selectErr != nil {
			return "", selectErr
		}
		if choice == "Abort" {
			return "", selector.ErrCancelled
		}
		return "replace", nil
	}
	return current, nil
}
func sourceLabel(source workbench.CompletedSource) string {
	return strings.ReplaceAll(source.Kind+"-"+source.Name, "/", "-")
}
func pathIsDir(path string) bool  { info, err := os.Stat(path); return err == nil && info.IsDir() }
func pathIsFile(path string) bool { info, err := os.Stat(path); return err == nil && !info.IsDir() }

func ingredientCandidates(cookbookDir string) ([]string, error) {
	set := map[string]bool{}
	variants := map[string]bool{}
	paths, err := filepath.Glob(filepath.Join(cookbookDir, "recipe", "*.yaml"))
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		definition, loadErr := recipe.Load(path)
		if loadErr != nil {
			continue
		}
		set["os."+definition.OS], set["platform."+definition.Platform] = true, true
		variants[definition.OS], variants[definition.Platform] = true, true
		for _, name := range definition.Runtime {
			set["runtime."+name] = true
		}
		for _, name := range definition.Profile {
			set["profile."+name] = true
		}
	}
	ingredientRoot := filepath.Join(cookbookDir, "ingredient")
	_ = filepath.WalkDir(ingredientRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		relative, relErr := filepath.Rel(ingredientRoot, path)
		if relErr != nil {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		var layer, name string
		if len(parts) >= 3 {
			layer, name = parts[0], parts[1]
		} else if len(parts) == 2 {
			layer = parts[0]
			name = ingredientStem(parts[1], variants)
		} else {
			first, rest, ok := strings.Cut(parts[0], ".")
			if ok {
				layer, name = first, ingredientStem(rest, variants)
			}
		}
		if layer != "" && name != "" {
			set[layer+"."+name] = true
		}
		return nil
	})
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values, nil
}

func ingredientStem(name string, variants map[string]bool) string {
	for _, suffix := range []string{".extensions", ".settings.jsonc", ".settings.json"} {
		name = strings.TrimSuffix(name, suffix)
	}
	tokens := strings.Split(name, ".")
	if len(tokens) > 1 && variants[tokens[len(tokens)-1]] {
		tokens = tokens[:len(tokens)-1]
	}
	return strings.Join(tokens, ".")
}

func ingredientLayers(identities []string) []string {
	set := map[string]bool{}
	for _, identity := range identities {
		if layer, _, ok := strings.Cut(identity, "."); ok {
			set[layer] = true
		}
	}
	layers := make([]string, 0, len(set))
	for layer := range set {
		layers = append(layers, layer)
	}
	sort.Strings(layers)
	return layers
}

func selectRuntimeSource(archiveRoot, cookbookDir, value string, selection selector.Selector) (string, string, error) {
	if value != "" {
		recipeCandidate := value
		if filepath.Ext(recipeCandidate) == "" {
			recipeCandidate = filepath.Join(cookbookDir, "recipe", value+".yaml")
		}
		archiveCandidate := value
		if !pathIsDir(archiveCandidate) {
			archiveCandidate = filepath.Join(archiveRoot, value)
		}
		isRecipe := pathIsFile(recipeCandidate)
		isArchive := pathIsFile(filepath.Join(archiveCandidate, "manifest.json"))
		if isRecipe && isArchive {
			return "", "", fmt.Errorf("ambiguous Runtime source %q", value)
		}
		if isArchive {
			return "archive", archiveCandidate, nil
		}
		if isRecipe {
			return "recipe", recipeCandidate, nil
		}
		if pathIsFile(value) && strings.HasSuffix(strings.ToLower(value), ".yaml") {
			return "recipe", value, nil
		}
		return "", "", fmt.Errorf("Runtime source not found: %s", value)
	}
	kind, err := selection.Select("Select Runtime source type", []string{"recipe", "archive"})
	if err != nil {
		return "", "", err
	}
	if kind == "recipe" {
		path, err := selectRecipe(cookbookDir, "", selection)
		return kind, path, err
	}
	names, err := listArchives(archiveRoot)
	if err != nil {
		return "", "", err
	}
	name, err := selection.Select("Select Archive", names)
	if err != nil {
		return "", "", err
	}
	return kind, filepath.Join(archiveRoot, name), nil
}
func listArchives(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && pathIsFile(filepath.Join(root, entry.Name(), "manifest.json")) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
func selectIdentityDistribution(root string, definition recipe.Recipe, selection selector.Selector) (string, error) {
	names, err := distribution.List(root)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, name := range names {
		dist, loadErr := distribution.Load(root, name)
		if loadErr == nil && dist.Recipe.Name == definition.Name && dist.Recipe.OS == definition.OS && dist.Recipe.Platform == definition.Platform {
			matches = append(matches, name)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no Distribution matches %s + %s + %s; use build", definition.Name, definition.OS, definition.Platform)
	case 1:
		return matches[0], nil
	default:
		if !isTerminal() {
			return "", fmt.Errorf("multiple Distributions match %s + %s + %s", definition.Name, definition.OS, definition.Platform)
		}
		return selection.Select("Select matching Distribution", matches)
	}
}

type buildOptions struct {
	recipe, onConflict string
	keepStaging        bool
	force              bool
}

func parseBuildArgs(args []string) (buildOptions, error) {
	var options buildOptions
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--keep-staging":
			options.keepStaging = true
		case "--force":
			options.force = true
		case "--on-conflict":
			index++
			if index >= len(args) {
				return options, fmt.Errorf("--on-conflict requires suffix or abort")
			}
			options.onConflict = args[index]
			if options.onConflict != "suffix" && options.onConflict != "abort" {
				return options, fmt.Errorf("invalid --on-conflict: %s", options.onConflict)
			}
		default:
			if strings.HasPrefix(args[index], "-") {
				return options, fmt.Errorf("unknown build option: %s", args[index])
			}
			if options.recipe != "" {
				return options, fmt.Errorf("usage: ctk build [recipe] [--on-conflict suffix|abort] [--keep-staging] [--force]")
			}
			options.recipe = args[index]
		}
	}
	return options, nil
}

type applyOptions struct {
	source, dist string
	force        bool
}

func parseApplyArgs(args []string) (applyOptions, error) {
	var options applyOptions
	for _, arg := range args {
		if arg == "--force" {
			options.force = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return options, fmt.Errorf("unknown apply option: %s", arg)
		}
		if options.source == "" {
			options.source = arg
		} else if options.dist == "" {
			options.dist = arg
		} else {
			return options, fmt.Errorf("usage: ctk apply [recipe-or-archive] [dist] [--force]")
		}
	}
	return options, nil
}

func lifecycleService(poolRoot, cookbookDir string, selection selector.Selector) lifecycle.Service {
	stopper := platform.NewProcessStopper()
	factory := func(dist distribution.Distribution) (runtimeio.Runtime, error) {
		definition, err := platform.Lookup(dist.Recipe.Platform)
		if err != nil {
			return nil, err
		}
		command, err := exec.LookPath(definition.Command)
		if err != nil {
			return nil, fmt.Errorf("platform command not found: %s", definition.Command)
		}
		dataDir := filepath.Join(dist.Path, ".data")
		return vscode.Adapter{
			Command: command, UserDataDir: dataDir, ExtensionsDir: filepath.Join(dist.Path, ".ext"),
			StopForDatabaseWrite: func(ctx context.Context) error { return stopper.StopRuntime(ctx, dist.Recipe.Platform, dataDir) },
		}, nil
	}
	return lifecycle.Service{
		Cookbook: cookbook.Repository{Root: filepath.Join(cookbookDir, "ingredient")},
		Runtime:  factory, Pool: converge.Pool{Root: poolRoot},
		PoolUpdate: converge.PoolUpdater{Root: poolRoot}, Locks: runtimelock.Store{},
		ChooseLock: func() (string, error) {
			choice, err := selection.Select("Lock after mutation", []string{"Abort", "Refresh", "Reuse"})
			if err != nil {
				return "", err
			}
			return strings.ToLower(choice), nil
		},
	}
}

func archiveService(poolRoot, cookbookDir string, selection selector.Selector) ctkarchive.Service {
	stopper := platform.NewProcessStopper()
	factory := func(dist distribution.Distribution) (runtimeio.Runtime, error) {
		definition, err := platform.Lookup(dist.Recipe.Platform)
		if err != nil {
			return nil, err
		}
		command, err := exec.LookPath(definition.Command)
		if err != nil {
			return nil, fmt.Errorf("platform command not found: %s", definition.Command)
		}
		dataDir := filepath.Join(dist.Path, ".data")
		return vscode.Adapter{Command: command, UserDataDir: dataDir, ExtensionsDir: filepath.Join(dist.Path, ".ext"), StopForDatabaseWrite: func(ctx context.Context) error { return stopper.StopRuntime(ctx, dist.Recipe.Platform, dataDir) }}, nil
	}
	return ctkarchive.Service{Cookbook: cookbook.Repository{Root: filepath.Join(cookbookDir, "ingredient")}, Runtime: factory, Locks: runtimelock.Store{}, Pool: converge.PoolUpdater{Root: poolRoot}, ChooseLock: func() (string, error) {
		choice, err := selection.Select("Lock for Archive", []string{"Abort", "Refresh", "Reuse"})
		if err != nil {
			return "", err
		}
		return strings.ToLower(choice), nil
	}}
}

func workbenchService(root, cookbookDir, workbenchDir string, selection selector.Selector) workbench.Service {
	stopper := platform.NewProcessStopper()
	factory := func(dist distribution.Distribution) (runtimeio.Runtime, error) {
		definition, err := platform.Lookup(dist.Recipe.Platform)
		if err != nil {
			return nil, err
		}
		command, err := exec.LookPath(definition.Command)
		if err != nil {
			return nil, fmt.Errorf("platform command not found: %s", definition.Command)
		}
		dataDir := filepath.Join(dist.Path, ".data")
		return vscode.Adapter{
			Command: command, UserDataDir: dataDir, ExtensionsDir: filepath.Join(dist.Path, ".ext"),
			StopForDatabaseWrite: func(ctx context.Context) error { return stopper.StopRuntime(ctx, dist.Recipe.Platform, dataDir) },
		}, nil
	}
	return workbench.Service{
		WorkspaceRoot: root, CookbookRoot: cookbookDir, WorkbenchRoot: workbenchDir, Runtime: factory, Locks: runtimelock.Store{},
		ChooseLock: func() (string, error) {
			choice, err := selection.Select("Lock for Workbench", []string{"Abort", "Refresh", "Reuse"})
			if err != nil {
				return "", err
			}
			return strings.ToLower(choice), nil
		},
	}
}

func codevenvService(poolRoot, distRoot string, selection selector.Selector) codevenv.Service {
	stopper := platform.NewProcessStopper()
	factory := func(platformName, userDataDir, extensionsDir string) (runtimeio.Runtime, error) {
		definition, err := platform.Lookup(platformName)
		if err != nil {
			return nil, err
		}
		command, err := exec.LookPath(definition.Command)
		if err != nil {
			return nil, fmt.Errorf("platform command not found: %s", definition.Command)
		}
		return vscode.Adapter{
			Command: command, UserDataDir: userDataDir, ExtensionsDir: extensionsDir,
			StopForDatabaseWrite: func(ctx context.Context) error { return stopper.StopRuntime(ctx, platformName, userDataDir) },
		}, nil
	}
	pool := converge.Pool{Root: poolRoot}
	poolUpdate := converge.PoolUpdater{Root: poolRoot}
	locks := runtimelock.Store{}
	gate := codevenv.SafetyGate(nil)
	if isTerminal() {
		gate = func(verification recovery.Verification) (bool, error) {
			fmt.Fprintf(os.Stderr, "[verification] differences=%d\n", len(verification.Differences))
			for _, difference := range verification.Differences {
				fmt.Fprintf(os.Stderr, "- %s/%s %s: %s\n", difference.Phase, difference.Kind, difference.Scope, difference.Risk)
			}
			choice, err := selection.Select("Semantic verification differs", []string{"Abort", "Force"})
			if err != nil {
				return false, err
			}
			return choice == "Force", nil
		}
	}
	return codevenv.Service{
		DistRoot: distRoot, Runtime: factory, Stopper: stopper, Locks: locks, SafetyGate: gate,
		Recovery: recovery.Service{Pool: pool, PoolUpdate: poolUpdate, Locks: locks},
	}
}

func parsePlatformForce(args []string, required bool) (string, bool, error) {
	platformName := ""
	force := false
	for _, value := range args {
		switch value {
		case "--force":
			force = true
		default:
			if strings.HasPrefix(value, "-") {
				return "", false, fmt.Errorf("unknown option: %s", value)
			}
			if platformName != "" {
				return "", false, fmt.Errorf("multiple Platforms provided")
			}
			platformName = value
		}
	}
	if required && platformName == "" {
		return "", false, fmt.Errorf("Platform is required")
	}
	return platformName, force, nil
}

func parseDeactivateArgs(args []string) (string, bool, bool, error) {
	platformName := ""
	force, forceEmpty := false, false
	for _, value := range args {
		switch value {
		case "--force":
			force = true
		case "--force-empty":
			forceEmpty = true
		default:
			if strings.HasPrefix(value, "-") {
				return "", false, false, fmt.Errorf("unknown option: %s", value)
			}
			if platformName != "" {
				return "", false, false, fmt.Errorf("multiple Platforms provided")
			}
			platformName = value
		}
	}
	if force && forceEmpty {
		return "", false, false, fmt.Errorf("--force and --force-empty are mutually exclusive")
	}
	return platformName, force, forceEmpty, nil
}

func selectAvailablePlatform(selection selector.Selector) (string, error) {
	var candidates []string
	for _, platformName := range codevenv.SupportedPlatforms() {
		definition, err := platform.Lookup(platformName)
		if err == nil {
			_, err = exec.LookPath(definition.Command)
		}
		if err == nil {
			candidates = append(candidates, platformName)
		}
	}
	return selection.Select("Select Platform to activate", candidates)
}

func selectRecipe(cookbookDir, value string, selection selector.Selector) (string, error) {
	if value != "" {
		if info, err := os.Stat(value); err == nil && !info.IsDir() {
			return filepath.Abs(value)
		}
		candidate := filepath.Join(cookbookDir, "recipe", value)
		if filepath.Ext(candidate) == "" {
			candidate += ".yaml"
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
		return "", fmt.Errorf("Recipe not found: %s", value)
	}
	paths, err := filepath.Glob(filepath.Join(cookbookDir, "recipe", "*.yaml"))
	if err != nil {
		return "", err
	}
	names := make([]string, len(paths))
	for index, path := range paths {
		names[index] = filepath.Base(path)
	}
	name, err := selection.Select("Select Recipe", names)
	if err != nil {
		return "", err
	}
	return filepath.Join(cookbookDir, "recipe", name), nil
}

func selectRecipeDistribution(distDir, recipeName string, selection selector.Selector) (string, error) {
	names, err := distribution.List(distDir)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, name := range names {
		dist, err := distribution.Load(distDir, name)
		if err == nil && dist.Recipe.Name == recipeName {
			matches = append(matches, name)
		}
	}
	return selection.Select("Select Distribution for "+recipeName, matches)
}

func isTerminal() bool {
	stdin, stdinErr := os.Stdin.Stat()
	stdout, stdoutErr := os.Stdout.Stat()
	return stdinErr == nil && stdoutErr == nil && stdin.Mode()&os.ModeCharDevice != 0 && stdout.Mode()&os.ModeCharDevice != 0
}

func selectDistribution(distDir string, args []string, selection selector.Selector) (string, error) {
	if len(args) > 0 && args[0] != "--" {
		return args[0], nil
	}
	names, err := distribution.List(distDir)
	if err != nil {
		return "", err
	}
	return selection.Select("Select Distribution", names)
}

func exitCode(err error) int {
	if errors.Is(err, selector.ErrCancelled) {
		return 130
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func projectRoot() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("resolve executable symlinks: %w", err)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	return findProjectRoot(os.Getenv("CTK_HOME"), workingDirectory, executable)
}

func findProjectRoot(configured, workingDirectory, executable string) (string, error) {
	if configured != "" {
		path, err := filepath.Abs(configured)
		if err != nil || !isProjectRoot(path) {
			return "", fmt.Errorf("CTK_HOME is not a CTK workspace: %s", configured)
		}
		return path, nil
	}
	for path := workingDirectory; ; path = filepath.Dir(path) {
		if isProjectRoot(path) {
			return path, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
	}
	executableRoot := filepath.Dir(filepath.Dir(executable))
	if isProjectRoot(executableRoot) {
		return executableRoot, nil
	}
	return "", fmt.Errorf("CTK workspace not found; run inside a workspace or set CTK_HOME")
}

func isProjectRoot(path string) bool {
	return ctkworkspace.HasMarker(path)
}

func usage() {
	fmt.Println(`Usage: ctk <command>

Commands:
  activate [platform] [--force]
                      Import and manage a Platform's default Runtime
  build [recipe-or-archive] [--force]
                      Build a new Distribution
  apply [recipe-or-archive] [dist] [--force]
                      Converge an existing Distribution
  archive [dist] [--on-conflict suffix|replace|abort]
                      Preserve an offline-reconstructable Runtime
  lock [dist]         Observe a Distribution into its Lock
  freeze draft [dist] Generate a Freeze Draft Workbench
  freeze commit       Commit present Draft Artifacts into the Cookbook
  view [source]       Auto-detect and view a Distribution, Recipe, or Ingredient
  view dist [dist]    View a Distribution Inventory explicitly
  view recipe [recipe]
                      View a resolved Recipe Inventory explicitly
  view ingredient [all|layer|layer.name]
                      View all, one layer, or one Ingredient explicitly
  sync [left] [right] Compare Distribution or Recipe completed states
  list                List Distributions
  current [platform]  Show selected Runtime(s)
  deactivate [platform] [--force|--force-empty]
                      Restore a Platform's imported default Runtime
  use [dist]          Select a Runtime for its active Platform
  launch [dist] [--] [args...]
                      Temporarily launch a Distribution
  workbench [draft|inspect] [viewpoint] [--editor command]
                      Open a Draft or Inspect Workbench
  docs [<node>|resolve <terms...>|show <reference>]
                      Navigate documentation packaged with this binary
  select              Select a command interactively
  version             Show binary version and build provenance
  help                Show this help`)
}
