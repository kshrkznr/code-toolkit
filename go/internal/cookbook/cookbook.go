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
	ExtensionPool        string
	LockMode             string
	Default              ScopePlan
	Profiles             []ScopePlan
}

type ScopePlan struct {
	Name             string
	Settings         settings.Document
	Keybindings      runtimeartifact.Array
	Tasks            runtimeartifact.Object
	MCP              runtimeartifact.Object
	Snippets         runtimeartifact.Snippets
	Extensions       []string
	ExtensionOrigins map[string][]Source
	ExtensionSets    []ExtensionSetReference
	Inheritance      Inheritance
	Sources          []Source
}

type ExtensionSetReference struct {
	Name        string
	Declaration Source
}

type extensionSelection struct {
	IDs     []string
	Sources []Source
	Origins map[string][]Source
	SetRefs []ExtensionSetReference
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
	if mode := definition.ExtensionPoolMode(); mode != "reuse" && mode != "refresh" {
		return Plan{}, fmt.Errorf("unsupported extension-pool %q", mode)
	}
	rules, err := mergerules.Load(filepath.Dir(r.Root))
	if err != nil {
		return Plan{}, err
	}

	runtimeSelection, err := r.extensions("runtime", definition.Runtime)
	if err != nil {
		return Plan{}, err
	}
	runtimeExtensions := runtimeSelection.IDs

	defaultSettingsMode := definition.DefaultContent("settings")
	if defaultSettingsMode != "runtime" && defaultSettingsMode != "clean" && defaultSettingsMode != "unmanaged" {
		return Plan{}, fmt.Errorf("unsupported default-profile.settings %q", defaultSettingsMode)
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
	defaultSettingExtensions := map[string]bool{}
	defaultSettingSets := map[string]bool{}
	if defaultSettingsMode == "runtime" {
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
			defaultSettingExtensions[id] = true
		}
		for _, set := range uniqueExtensionSetReferences(runtimeSelection.SetRefs, defaultSettingSets) {
			if err := appendSettings("extension-set", set.Name); err != nil {
				return Plan{}, err
			}
		}
		for _, ingredient := range definition.Runtime {
			if err := appendSettings("runtime", ingredient); err != nil {
				return Plan{}, err
			}
		}
	}

	profileSelections := make(map[string]extensionSelection, len(definition.Profile))
	for _, name := range definition.Profile {
		selection, err := r.extensions("profile", []string{name})
		if err != nil {
			return Plan{}, err
		}
		profileSelections[name] = selection
		if defaultSettingsMode == "runtime" && definition.ProfileContent(name).Settings == "default" {
			for _, id := range selection.IDs {
				if defaultSettingExtensions[id] {
					continue
				}
				defaultSettingExtensions[id] = true
				if err := appendSettings("extension", id); err != nil {
					return Plan{}, err
				}
			}
			for _, set := range uniqueExtensionSetReferences(selection.SetRefs, defaultSettingSets) {
				if err := appendSettings("extension-set", set.Name); err != nil {
					return Plan{}, err
				}
			}
			if err := appendSettings("profile", name); err != nil {
				return Plan{}, err
			}
		}
	}
	baseArtifactRefs := []ingredientRef{{"os", definition.OS}, {"platform", definition.Platform}}
	baseArtifactExtensions := map[string]bool{}
	for _, id := range runtimeExtensions {
		baseArtifactRefs = append(baseArtifactRefs, ingredientRef{"extension", id})
		baseArtifactExtensions[id] = true
	}
	baseArtifactSets := map[string]bool{}
	for _, set := range uniqueExtensionSetReferences(runtimeSelection.SetRefs, baseArtifactSets) {
		baseArtifactRefs = append(baseArtifactRefs, ingredientRef{"extension-set", set.Name})
	}
	for _, name := range definition.Runtime {
		baseArtifactRefs = append(baseArtifactRefs, ingredientRef{"runtime", name})
	}
	defaultArtifactRefs := func(kind string) ([]ingredientRef, map[string]bool, map[string]bool, error) {
		refs := append([]ingredientRef(nil), baseArtifactRefs...)
		extensionIDs := copyStringSet(baseArtifactExtensions)
		setNames := copyStringSet(baseArtifactSets)
		for _, name := range definition.Profile {
			strategy := definition.ProfileContent(name).Content(kind)
			if strategy != "default" && strategy != "profile" && strategy != "unmanaged" {
				return nil, nil, nil, fmt.Errorf("unsupported profile %q %s strategy %q", name, kind, strategy)
			}
			if strategy == "default" {
				selection := profileSelections[name]
				for _, id := range selection.IDs {
					if extensionIDs[id] {
						continue
					}
					extensionIDs[id] = true
					refs = append(refs, ingredientRef{"extension", id})
				}
				for _, set := range uniqueExtensionSetReferences(selection.SetRefs, setNames) {
					refs = append(refs, ingredientRef{"extension-set", set.Name})
				}
				refs = append(refs, ingredientRef{"profile", name})
			}
		}
		return refs, extensionIDs, setNames, nil
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
		ExtensionMarketplace: definition.ExtensionMarketplace(), ExtensionPool: definition.ExtensionPoolMode(), LockMode: definition.LockMode(),
		Default: ScopePlan{
			Name: "", Extensions: append([]string(nil), defaultExtensions...), ExtensionOrigins: selectExtensionOrigins(runtimeSelection.Origins, defaultExtensions), ExtensionSets: append([]ExtensionSetReference(nil), runtimeSelection.SetRefs...), Sources: append(defaultSources, runtimeSelection.Sources...),
			Inheritance: Inheritance{Unmanaged: map[string]bool{}},
		},
	}
	if definition.DefaultExtensionMode() == "unmanaged" {
		plan.Default.Inheritance.Unmanaged["extensions"] = true
	}
	defaultArtifactExtensions := map[string]map[string]bool{}
	defaultArtifactSets := map[string]map[string]bool{}
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
			refs, extensionIDs, setNames, err := defaultArtifactRefs(kind)
			if err != nil {
				return Plan{}, err
			}
			defaultArtifactExtensions[kind] = extensionIDs
			defaultArtifactSets[kind] = setNames
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
			refs, extensionIDs, setNames, err := defaultArtifactRefs("snippets")
			if err != nil {
				return Plan{}, err
			}
			defaultArtifactExtensions["snippets"] = extensionIDs
			defaultArtifactSets["snippets"] = setNames
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
	if defaultSettingsMode == "clean" {
		plan.Default.Settings = settings.Document{}
	}
	if defaultSettingsMode == "unmanaged" {
		plan.Default.Settings = nil
		plan.Default.Inheritance.Unmanaged["settings"] = true
	}
	for _, name := range definition.Profile {
		strategy := definition.ProfileContent(name)
		profileSelection := profileSelections[name]
		var documents []settings.Document
		if defaultSettingsMode == "runtime" {
			documents = []settings.Document{resolvedDefaultSettings}
		}
		sources := append([]Source(nil), plan.Default.Sources...)
		if strategy.Settings == "profile" {
			for _, id := range profileSelection.IDs {
				if defaultSettingExtensions[id] {
					continue
				}
				resolved, found, err := r.settingVariants("extension", id, definition.OS, definition.Platform)
				if err != nil {
					return Plan{}, err
				}
				documents, sources = append(documents, resolved...), append(sources, found...)
			}
			profileSettingSets := copyStringSet(defaultSettingSets)
			for _, set := range uniqueExtensionSetReferences(profileSelection.SetRefs, profileSettingSets) {
				resolved, found, err := r.settingVariants("extension-set", set.Name, definition.OS, definition.Platform)
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
		extensions := uniqueSorted(append(append([]string{}, runtimeExtensions...), profileSelection.IDs...))
		merged, err := settings.MergeWithRules(rules, documents...)
		if err != nil {
			return Plan{}, err
		}
		if strategy.Settings == "default" {
			merged = plan.Default.Settings
		}
		profilePlan := ScopePlan{
			Name: name, Settings: merged, Extensions: extensions, ExtensionOrigins: selectExtensionOrigins(mergeExtensionOrigins(runtimeSelection.Origins, profileSelection.Origins), extensions), ExtensionSets: mergeExtensionSetReferences(runtimeSelection.SetRefs, profileSelection.SetRefs),
			Inheritance: Inheritance{Settings: strategy.Settings == "default", Keybindings: strategy.Keybindings == "default", Tasks: strategy.Tasks == "default", MCP: strategy.MCP == "default", Snippets: strategy.Snippets == "default", Unmanaged: map[string]bool{}},
			Sources:     append(sources, profileSelection.Sources...),
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
			extensionIDs := copyStringSet(defaultArtifactExtensions[kind])
			for _, id := range profileSelection.IDs {
				if extensionIDs[id] {
					continue
				}
				extensionIDs[id] = true
				refs = append(refs, ingredientRef{"extension", id})
			}
			setNames := copyStringSet(defaultArtifactSets[kind])
			for _, set := range uniqueExtensionSetReferences(profileSelection.SetRefs, setNames) {
				refs = append(refs, ingredientRef{"extension-set", set.Name})
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
			extensionIDs := copyStringSet(defaultArtifactExtensions["snippets"])
			for _, id := range profileSelection.IDs {
				if extensionIDs[id] {
					continue
				}
				extensionIDs[id] = true
				refs = append(refs, ingredientRef{"extension", id})
			}
			setNames := copyStringSet(defaultArtifactSets["snippets"])
			for _, set := range uniqueExtensionSetReferences(profileSelection.SetRefs, setNames) {
				refs = append(refs, ingredientRef{"extension-set", set.Name})
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

func (r Repository) extensions(layer string, ingredients []string) (extensionSelection, error) {
	var ids []string
	var sources []Source
	origins := map[string][]Source{}
	var setRefs []ExtensionSetReference
	seenSets := map[string]bool{}
	for _, ingredient := range ingredients {
		path, err := r.one(extensionCandidates(r.Root, layer, ingredient), layer, ingredient, "extensions")
		if err != nil {
			return extensionSelection{}, err
		}
		if path == "" {
			continue
		}
		lines, err := extensionLines(path)
		if err != nil {
			return extensionSelection{}, err
		}
		owner := Source{Layer: layer, Ingredient: ingredient, Path: path}
		sources = append(sources, owner)
		for _, line := range lines {
			if !strings.HasPrefix(line, "set:") {
				ids = append(ids, line)
				addExtensionOrigin(origins, line, owner)
				continue
			}
			name := strings.TrimPrefix(line, "set:")
			if !validExtensionSetName(name) {
				return extensionSelection{}, fmt.Errorf("invalid Extension Set declaration %q in %s: name must match [A-Za-z0-9][A-Za-z0-9._-]*", line, path)
			}
			if !seenSets[name] {
				seenSets[name] = true
				setRefs = append(setRefs, ExtensionSetReference{Name: name, Declaration: owner})
			}
			members, source, err := r.extensionSet(name)
			if err != nil {
				return extensionSelection{}, err
			}
			ids = append(ids, members...)
			if source != nil {
				sources = append(sources, *source)
				for _, member := range members {
					addExtensionOrigin(origins, member, *source)
				}
			}
		}
	}
	return extensionSelection{IDs: uniqueSorted(ids), Sources: sources, Origins: origins, SetRefs: setRefs}, nil
}

func uniqueExtensionSetReferences(values []ExtensionSetReference, seen map[string]bool) []ExtensionSetReference {
	var result []ExtensionSetReference
	for _, value := range values {
		if seen[value.Name] {
			continue
		}
		seen[value.Name] = true
		result = append(result, value)
	}
	return result
}

func mergeExtensionSetReferences(values ...[]ExtensionSetReference) []ExtensionSetReference {
	seen := map[string]bool{}
	var result []ExtensionSetReference
	for _, refs := range values {
		for _, ref := range refs {
			key := ref.Name + "\x00" + ref.Declaration.Path
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, ref)
		}
	}
	return result
}

func copyStringSet(value map[string]bool) map[string]bool {
	result := map[string]bool{}
	for name := range value {
		result[name] = true
	}
	return result
}

func addExtensionOrigin(origins map[string][]Source, id string, source Source) {
	for _, existing := range origins[id] {
		if existing == source {
			return
		}
	}
	origins[id] = append(origins[id], source)
}

func mergeExtensionOrigins(values ...map[string][]Source) map[string][]Source {
	result := map[string][]Source{}
	for _, origins := range values {
		for id, sources := range origins {
			for _, source := range sources {
				addExtensionOrigin(result, id, source)
			}
		}
	}
	return result
}

func selectExtensionOrigins(origins map[string][]Source, ids []string) map[string][]Source {
	if len(ids) == 0 {
		return nil
	}
	result := make(map[string][]Source, len(ids))
	for _, id := range ids {
		result[id] = append([]Source(nil), origins[id]...)
	}
	return result
}

func (r Repository) extensionSet(name string) ([]string, *Source, error) {
	path, err := r.one(extensionCandidates(r.Root, "extension-set", name), "extension-set", name, "extensions")
	if err != nil {
		return nil, nil, err
	}
	if path == "" {
		return nil, nil, nil
	}
	lines, err := extensionLines(path)
	if err != nil {
		return nil, nil, err
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "set:") {
			return nil, nil, fmt.Errorf("nested Extension Set declaration %q in %s is not allowed", line, path)
		}
	}
	return lines, &Source{Layer: "extension-set", Ingredient: name, Path: path}, nil
}

func extensionLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line := strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\r")); line != "" {
			lines = append(lines, line)
		}
	}
	scanErr, closeErr := scanner.Err(), file.Close()
	if scanErr != nil {
		return nil, fmt.Errorf("read %s: %w", path, scanErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close %s: %w", path, closeErr)
	}
	return lines, nil
}

func validExtensionSetName(name string) bool {
	if name == "" {
		return false
	}
	for index := 0; index < len(name); index++ {
		character := name[index]
		if (character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			continue
		}
		if index > 0 && (character == '.' || character == '_' || character == '-') {
			continue
		}
		return false
	}
	return true
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
