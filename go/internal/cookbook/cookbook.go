package cookbook

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kshrkznr/code-toolkit/go/internal/mergerules"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type Repository struct{ Root string }

type Source struct {
	Layer      string
	Ingredient string
	Variant    string
	Path       string
}

type Plan struct {
	RecipePath           string
	Name                 string
	OS                   string
	Platform             string
	ExtensionMarketplace bool
	LockMode             string
	Default              ScopePlan
	Profiles             []ScopePlan
}

type ScopePlan struct {
	Name        string
	Settings    settings.Document
	Keybindings runtimeartifact.Array
	Tasks       runtimeartifact.Object
	MCP         runtimeartifact.Object
	Snippets    runtimeartifact.Snippets
	Extensions  []string
	Inheritance Inheritance
	Sources     []Source
}

type Inheritance struct {
	Settings    bool            `json:"settings"`
	Keybindings bool            `json:"keybindings"`
	Tasks       bool            `json:"tasks"`
	MCP         bool            `json:"mcp"`
	Snippets    bool            `json:"snippets"`
	Unmanaged   map[string]bool `json:"unmanaged,omitempty"`
}

func (r Repository) Resolve(recipePath string) (Plan, error) {
	definition, err := recipe.Load(recipePath)
	if err != nil {
		return Plan{}, err
	}
	if mode := definition.LockMode(); mode != "refresh" && mode != "reuse" && mode != "ask" {
		return Plan{}, fmt.Errorf("unsupported lock-mode %q", mode)
	}
	rules, err := mergerules.Load(filepath.Dir(r.Root))
	if err != nil {
		return Plan{}, err
	}

	runtimeExtensions, runtimeExtensionSources, err := r.extensions("runtime", definition.Runtime)
	if err != nil {
		return Plan{}, err
	}

	defaultDocuments := []settings.Document{}
	defaultSources := []Source{}
	appendSettings := func(layer, ingredient string) error {
		documents, sources, err := r.settingVariants(layer, ingredient, definition.OS, definition.Platform)
		if err != nil {
			return err
		}
		defaultDocuments = append(defaultDocuments, documents...)
		defaultSources = append(defaultSources, sources...)
		return nil
	}
	if err := appendSettings("os", definition.OS); err != nil {
		return Plan{}, err
	}
	if err := appendSettings("platform", definition.Platform); err != nil {
		return Plan{}, err
	}
	for _, id := range runtimeExtensions {
		if err := appendSettings("extension", id); err != nil {
			return Plan{}, err
		}
	}
	for _, ingredient := range definition.Runtime {
		if err := appendSettings("runtime", ingredient); err != nil {
			return Plan{}, err
		}
	}

	profileExtensions := make(map[string][]string, len(definition.Profile))
	profileExtensionSources := make(map[string][]Source, len(definition.Profile))
	for _, name := range definition.Profile {
		extensions, sources, err := r.extensions("profile", []string{name})
		if err != nil {
			return Plan{}, err
		}
		profileExtensions[name], profileExtensionSources[name] = extensions, sources
		if definition.ProfileContent(name).Settings == "default" {
			for _, id := range extensions {
				if err := appendSettings("extension", id); err != nil {
					return Plan{}, err
				}
			}
			if err := appendSettings("profile", name); err != nil {
				return Plan{}, err
			}
		}
	}
	baseArtifactRefs := []ingredientRef{{"os", definition.OS}, {"platform", definition.Platform}}
	for _, id := range runtimeExtensions {
		baseArtifactRefs = append(baseArtifactRefs, ingredientRef{"extension", id})
	}
	for _, name := range definition.Runtime {
		baseArtifactRefs = append(baseArtifactRefs, ingredientRef{"runtime", name})
	}
	defaultArtifactRefs := func(kind string) ([]ingredientRef, error) {
		refs := append([]ingredientRef(nil), baseArtifactRefs...)
		for _, name := range definition.Profile {
			strategy := definition.ProfileContent(name).Content(kind)
			if strategy != "default" && strategy != "profile" && strategy != "unmanaged" {
				return nil, fmt.Errorf("unsupported profile %q %s strategy %q", name, kind, strategy)
			}
			if strategy == "default" {
				for _, id := range profileExtensions[name] {
					refs = append(refs, ingredientRef{"extension", id})
				}
				refs = append(refs, ingredientRef{"profile", name})
			}
		}
		return refs, nil
	}

	defaultExtensions := runtimeExtensions
	if definition.DefaultExtensionMode() == "clean" {
		defaultExtensions = nil
	}
	if mode := definition.DefaultExtensionMode(); mode != "clean" && mode != "runtime" && mode != "unmanaged" {
		return Plan{}, fmt.Errorf("unsupported default-profile.extensions %q", mode)
	}
	if definition.DefaultExtensionMode() == "unmanaged" {
		defaultExtensions = nil
	}
	plan := Plan{
		RecipePath: recipePath, Name: definition.Name, OS: definition.OS, Platform: definition.Platform,
		ExtensionMarketplace: definition.ExtensionMarketplace(), LockMode: definition.LockMode(),
		Default: ScopePlan{
			Name: "", Extensions: append([]string(nil), defaultExtensions...), Sources: append(defaultSources, runtimeExtensionSources...),
			Inheritance: Inheritance{Unmanaged: map[string]bool{}},
		},
	}
	if definition.DefaultExtensionMode() == "unmanaged" {
		plan.Default.Inheritance.Unmanaged["extensions"] = true
	}
	for _, kind := range []string{"keybindings", "tasks", "mcp"} {
		mode := definition.DefaultContent(kind)
		if mode != "runtime" && mode != "clean" && mode != "unmanaged" {
			return Plan{}, fmt.Errorf("unsupported default-profile.%s %q", kind, mode)
		}
		if mode == "unmanaged" {
			plan.Default.Inheritance.Unmanaged[kind] = true
			continue
		}
		var value runtimeartifact.Value
		var sources []Source
		if mode == "runtime" {
			refs, err := defaultArtifactRefs(kind)
			if err != nil {
				return Plan{}, err
			}
			value, sources, err = r.resolveArtifact(kind, definition.OS, definition.Platform, refs)
			if err != nil {
				return Plan{}, err
			}
		}
		switch kind {
		case "keybindings":
			if value == nil {
				plan.Default.Keybindings = runtimeartifact.Array{}
			} else {
				plan.Default.Keybindings = value.(runtimeartifact.Array)
			}
		case "tasks":
			if value == nil {
				plan.Default.Tasks = runtimeartifact.Object{}
			} else {
				plan.Default.Tasks = value.(runtimeartifact.Object)
			}
		case "mcp":
			if value == nil {
				plan.Default.MCP = runtimeartifact.Object{}
			} else {
				plan.Default.MCP = value.(runtimeartifact.Object)
			}
		}
		plan.Default.Sources = append(plan.Default.Sources, sources...)
	}
	snippetMode := definition.DefaultContent("snippets")
	if snippetMode != "runtime" && snippetMode != "clean" && snippetMode != "unmanaged" {
		return Plan{}, fmt.Errorf("unsupported default-profile.snippets %q", snippetMode)
	}
	if snippetMode != "unmanaged" {
		plan.Default.Snippets = runtimeartifact.Snippets{}
		if snippetMode == "runtime" {
			refs, err := defaultArtifactRefs("snippets")
			if err != nil {
				return Plan{}, err
			}
			plan.Default.Snippets, defaultSources, err = r.resolveSnippets(definition.OS, definition.Platform, refs)
			if err != nil {
				return Plan{}, err
			}
			plan.Default.Sources = append(plan.Default.Sources, defaultSources...)
		}
	} else {
		plan.Default.Inheritance.Unmanaged["snippets"] = true
	}
	plan.Default.Settings, err = settings.MergeWithRules(rules, defaultDocuments...)
	if err != nil {
		return Plan{}, err
	}
	resolvedDefaultSettings := plan.Default.Settings
	defaultSettingsMode := definition.DefaultContent("settings")
	if defaultSettingsMode != "runtime" && defaultSettingsMode != "clean" && defaultSettingsMode != "unmanaged" {
		return Plan{}, fmt.Errorf("unsupported default-profile.settings %q", defaultSettingsMode)
	}
	if defaultSettingsMode == "clean" {
		plan.Default.Settings = settings.Document{}
	}
	if defaultSettingsMode == "unmanaged" {
		plan.Default.Settings = nil
		plan.Default.Inheritance.Unmanaged["settings"] = true
	}
	for _, name := range definition.Profile {
		strategy := definition.ProfileContent(name)
		documents := []settings.Document{resolvedDefaultSettings}
		sources := append([]Source(nil), plan.Default.Sources...)
		if strategy.Settings == "profile" {
			for _, id := range profileExtensions[name] {
				resolved, found, err := r.settingVariants("extension", id, definition.OS, definition.Platform)
				if err != nil {
					return Plan{}, err
				}
				documents, sources = append(documents, resolved...), append(sources, found...)
			}
			resolved, found, err := r.settingVariants("profile", name, definition.OS, definition.Platform)
			if err != nil {
				return Plan{}, err
			}
			documents, sources = append(documents, resolved...), append(sources, found...)
		} else if strategy.Settings == "unmanaged" {
			documents = nil
		} else if strategy.Settings != "default" {
			return Plan{}, fmt.Errorf("unsupported profile %q settings strategy %q", name, strategy.Settings)
		}
		extensions := uniqueSorted(append(append([]string{}, runtimeExtensions...), profileExtensions[name]...))
		merged, err := settings.MergeWithRules(rules, documents...)
		if err != nil {
			return Plan{}, err
		}
		profilePlan := ScopePlan{
			Name: name, Settings: merged, Extensions: extensions,
			Inheritance: Inheritance{Settings: strategy.Settings == "default", Keybindings: strategy.Keybindings == "default", Tasks: strategy.Tasks == "default", MCP: strategy.MCP == "default", Snippets: strategy.Snippets == "default", Unmanaged: map[string]bool{}},
			Sources:     append(sources, profileExtensionSources[name]...),
		}
		if strategy.Settings == "unmanaged" {
			profilePlan.Settings = nil
			profilePlan.Inheritance.Unmanaged["settings"] = true
		}
		for _, kind := range []string{"keybindings", "tasks", "mcp"} {
			mode := strategy.Content(kind)
			if mode == "unmanaged" {
				profilePlan.Inheritance.Unmanaged[kind] = true
				continue
			}
			if mode == "default" {
				continue
			}
			refs := []ingredientRef{}
			for _, id := range profileExtensions[name] {
				refs = append(refs, ingredientRef{"extension", id})
			}
			refs = append(refs, ingredientRef{"profile", name})
			value, found, err := r.resolveArtifact(kind, definition.OS, definition.Platform, refs)
			if err != nil {
				return Plan{}, err
			}
			switch kind {
			case "keybindings":
				profilePlan.Keybindings = runtimeartifact.AppendArrays(plan.Default.Keybindings, value.(runtimeartifact.Array))
			case "tasks":
				profilePlan.Tasks, err = runtimeartifact.MergeTasks(plan.Default.Tasks, value.(runtimeartifact.Object))
			case "mcp":
				profilePlan.MCP = runtimeartifact.MergeMCP(plan.Default.MCP, value.(runtimeartifact.Object))
			}
			if err != nil {
				return Plan{}, err
			}
			profilePlan.Sources = append(profilePlan.Sources, found...)
		}
		if strategy.Snippets != "default" && strategy.Snippets != "profile" && strategy.Snippets != "unmanaged" {
			return Plan{}, fmt.Errorf("unsupported profile %q snippets strategy %q", name, strategy.Snippets)
		}
		if strategy.Snippets == "unmanaged" {
			profilePlan.Inheritance.Unmanaged["snippets"] = true
		} else if strategy.Snippets == "profile" {
			refs := []ingredientRef{}
			for _, id := range profileExtensions[name] {
				refs = append(refs, ingredientRef{"extension", id})
			}
			refs = append(refs, ingredientRef{"profile", name})
			local, found, err := r.resolveSnippets(definition.OS, definition.Platform, refs)
			if err != nil {
				return Plan{}, err
			}
			profilePlan.Snippets = runtimeartifact.Snippets{}
			for filename, document := range plan.Default.Snippets {
				profilePlan.Snippets[filename] = runtimeartifact.MergeSnippets(document)
			}
			for filename, document := range local {
				profilePlan.Snippets[filename] = runtimeartifact.MergeSnippets(profilePlan.Snippets[filename], document)
			}
			profilePlan.Sources = append(profilePlan.Sources, found...)
		}
		if len(profilePlan.Inheritance.Unmanaged) == 0 {
			profilePlan.Inheritance.Unmanaged = nil
		}
		plan.Profiles = append(plan.Profiles, profilePlan)
	}
	return plan, nil
}

func (r Repository) settingVariants(layer, ingredient, osName, platform string) ([]settings.Document, []Source, error) {
	variants := []struct{ name, suffix string }{{"", ""}, {osName, osName + "."}, {platform, platform + "."}}
	var documents []settings.Document
	var sources []Source
	for _, variant := range variants {
		path, err := r.one(settingsCandidates(r.Root, layer, ingredient, variant.suffix), layer, ingredient, "settings "+variant.name)
		if err != nil {
			return nil, nil, err
		}
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", path, err)
		}
		document, err := settings.Parse(data)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %s: %w", path, err)
		}
		documents = append(documents, document)
		sources = append(sources, Source{Layer: layer, Ingredient: ingredient, Variant: variant.name, Path: path})
	}
	return documents, sources, nil
}

func (r Repository) extensions(layer string, ingredients []string) ([]string, []Source, error) {
	var ids []string
	var sources []Source
	for _, ingredient := range ingredients {
		path, err := r.one(extensionCandidates(r.Root, layer, ingredient), layer, ingredient, "extensions")
		if err != nil {
			return nil, nil, err
		}
		if path == "" {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, nil, fmt.Errorf("open %s: %w", path, err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if id := strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\r")); id != "" {
				ids = append(ids, id)
			}
		}
		scanErr, closeErr := scanner.Err(), file.Close()
		if scanErr != nil {
			return nil, nil, fmt.Errorf("read %s: %w", path, scanErr)
		}
		if closeErr != nil {
			return nil, nil, fmt.Errorf("close %s: %w", path, closeErr)
		}
		sources = append(sources, Source{Layer: layer, Ingredient: ingredient, Path: path})
	}
	return uniqueSorted(ids), sources, nil
}

func (r Repository) one(candidates []string, layer, ingredient, resource string) (string, error) {
	var matches []string
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			matches = append(matches, candidate)
			continue
		}
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect %s: %w", candidate, err)
		}
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous %s resource for %s.%s: %s", resource, layer, ingredient, strings.Join(matches, ", "))
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return "", nil
}

func settingsCandidates(root, layer, ingredient, suffix string) []string {
	var result []string
	for _, extension := range []string{"json", "jsonc"} {
		name := suffix + "settings." + extension
		result = append(result,
			filepath.Join(root, layer+"."+ingredient+"."+name),
			filepath.Join(root, layer, ingredient+"."+name),
			filepath.Join(root, layer, ingredient, name),
		)
	}
	return result
}

func extensionCandidates(root, layer, ingredient string) []string {
	return []string{
		filepath.Join(root, layer+"."+ingredient+".extensions"),
		filepath.Join(root, layer, ingredient+".extensions"),
		filepath.Join(root, layer, ingredient, "extensions"),
	}
}

func jsonArtifactCandidates(root, layer, ingredient, suffix, kind string) []string {
	var result []string
	for _, extension := range []string{"json", "jsonc"} {
		name := suffix + kind + "." + extension
		result = append(result,
			filepath.Join(root, layer+"."+ingredient+"."+name),
			filepath.Join(root, layer, ingredient+"."+name),
			filepath.Join(root, layer, ingredient, name),
		)
	}
	return result
}

func (r Repository) artifactVariants(layer, ingredient, osName, platform, kind string) ([]runtimeartifact.Value, []Source, error) {
	variants := []struct{ name, suffix string }{{"", ""}, {osName, osName + "."}, {platform, platform + "."}}
	var values []runtimeartifact.Value
	var sources []Source
	for _, variant := range variants {
		path, err := r.one(jsonArtifactCandidates(r.Root, layer, ingredient, variant.suffix, kind), layer, ingredient, kind+" "+variant.name)
		if err != nil {
			return nil, nil, err
		}
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", path, err)
		}
		value, err := runtimeartifact.Parse(data)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %s: %w", path, err)
		}
		values = append(values, value)
		sources = append(sources, Source{Layer: layer, Ingredient: ingredient, Variant: variant.name, Path: path})
	}
	return values, sources, nil
}

type ingredientRef struct{ layer, name string }

func (r Repository) resolveArtifact(kind, osName, platform string, refs []ingredientRef) (runtimeartifact.Value, []Source, error) {
	var values []runtimeartifact.Value
	var sources []Source
	for _, ref := range refs {
		found, source, err := r.artifactVariants(ref.layer, ref.name, osName, platform, kind)
		if err != nil {
			return nil, nil, err
		}
		values, sources = append(values, found...), append(sources, source...)
	}
	switch kind {
	case "keybindings":
		var arrays []runtimeartifact.Array
		for _, value := range values {
			array, ok := value.([]any)
			if !ok {
				return nil, nil, fmt.Errorf("Keybindings root must be an array")
			}
			arrays = append(arrays, runtimeartifact.Array(array))
		}
		return runtimeartifact.AppendArrays(arrays...), sources, nil
	case "tasks":
		var objects []runtimeartifact.Object
		for _, value := range values {
			object, ok := value.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("Tasks root must be an object")
			}
			objects = append(objects, runtimeartifact.Object(object))
		}
		merged, err := runtimeartifact.MergeTasks(objects...)
		return merged, sources, err
	case "mcp":
		var objects []runtimeartifact.Object
		for _, value := range values {
			object, ok := value.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("MCP root must be an object")
			}
			objects = append(objects, runtimeartifact.Object(object))
		}
		return runtimeartifact.MergeMCP(objects...), sources, nil
	default:
		return nil, nil, fmt.Errorf("unsupported Artifact %q", kind)
	}
}

func (r Repository) resolveSnippets(osName, platform string, refs []ingredientRef) (runtimeartifact.Snippets, []Source, error) {
	result := runtimeartifact.Snippets{}
	var sources []Source
	for _, ref := range refs {
		for _, variant := range []struct{ name, suffix string }{{"", ""}, {osName, osName + "."}, {platform, platform + "."}} {
			groups := [][]string{}
			flat, _ := filepath.Glob(filepath.Join(r.Root, ref.layer+"."+ref.name+"."+variant.suffix+"snippets.*"))
			groups = append(groups, flat)
			middle, _ := filepath.Glob(filepath.Join(r.Root, ref.layer, ref.name+"."+variant.suffix+"snippets.*"))
			groups = append(groups, middle)
			directory := filepath.Join(r.Root, ref.layer, ref.name, variant.suffix+"snippets")
			var nested []string
			if entries, err := os.ReadDir(directory); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						nested = append(nested, filepath.Join(directory, entry.Name()))
					}
				}
			} else if !os.IsNotExist(err) {
				return nil, nil, err
			}
			groups = append(groups, nested)
			used := 0
			for _, group := range groups {
				if len(group) != 0 {
					used++
				}
			}
			if used > 1 {
				return nil, nil, fmt.Errorf("ambiguous snippets resource for %s.%s", ref.layer, ref.name)
			}
			for index, group := range groups {
				for _, path := range group {
					base := filepath.Base(path)
					if index == 0 {
						base = strings.TrimPrefix(base, ref.layer+"."+ref.name+"."+variant.suffix+"snippets.")
					}
					if index == 1 {
						base = strings.TrimPrefix(base, ref.name+"."+variant.suffix+"snippets.")
					}
					if !strings.HasSuffix(base, ".json") && !strings.HasSuffix(base, ".jsonc") && !strings.HasSuffix(base, ".code-snippets") {
						continue
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return nil, nil, err
					}
					document, err := runtimeartifact.ParseObject(data)
					if err != nil {
						return nil, nil, fmt.Errorf("parse %s: %w", path, err)
					}
					result[base] = runtimeartifact.MergeSnippets(result[base], document)
					sources = append(sources, Source{Layer: ref.layer, Ingredient: ref.name, Variant: variant.name, Path: path})
				}
			}
		}
	}
	return result, sources, nil
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		seen[value] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
