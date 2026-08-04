package runtimelock

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

const FormatVersion = 1

type Snapshot struct {
	FormatVersion int             `json:"formatVersion"`
	RecipeName    string          `json:"recipeName"`
	Platform      string          `json:"platform"`
	ObservedAt    time.Time       `json:"observedAt"`
	Default       ScopeSnapshot   `json:"default"`
	Profiles      []ScopeSnapshot `json:"profiles"`
}

type ScopeSnapshot struct {
	Name        string                   `json:"name,omitempty"`
	Settings    settings.Document        `json:"settings"`
	Extensions  []runtimeio.Extension    `json:"extensions"`
	Inheritance cookbook.Inheritance     `json:"inheritance,omitempty"`
	Keybindings runtimeartifact.Array    `json:"keybindings"`
	Tasks       runtimeartifact.Object   `json:"tasks"`
	MCP         runtimeartifact.Object   `json:"mcp"`
	Snippets    runtimeartifact.Snippets `json:"snippets"`
}

type Collector struct{ Now func() time.Time }

func (c Collector) Collect(ctx context.Context, runtime runtimeio.Runtime, plan cookbook.Plan) (Snapshot, error) {
	scopes, err := runtime.Scopes(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot := Snapshot{FormatVersion: FormatVersion, RecipeName: plan.Name, Platform: plan.Platform, ObservedAt: time.Now()}
	if c.Now != nil {
		snapshot.ObservedAt = c.Now()
	}
	plans := map[string]cookbook.ScopePlan{"": plan.Default}
	for _, scope := range plan.Profiles {
		plans[scope.Name] = scope
	}
	seen := map[string]bool{}
	for _, scope := range scopes {
		if seen[scope.Name] {
			return Snapshot{}, fmt.Errorf("duplicate Runtime scope %q", scope.Name)
		}
		seen[scope.Name] = true
		scopePlan, selected := plans[scope.Name]
		value, err := c.collectPlannedScope(ctx, runtime, scope, scopePlan, selected)
		if err != nil {
			return Snapshot{}, err
		}
		if scope.IsDefault() {
			snapshot.Default = value
		} else {
			snapshot.Profiles = append(snapshot.Profiles, value)
		}
	}
	if !seen[""] {
		return Snapshot{}, fmt.Errorf("default Runtime scope not found")
	}
	sort.Slice(snapshot.Profiles, func(i, j int) bool { return snapshot.Profiles[i].Name < snapshot.Profiles[j].Name })
	if err := Validate(snapshot, plan); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func (c Collector) collectPlannedScope(ctx context.Context, runtime runtimeio.Runtime, scope runtimeio.Scope, plan cookbook.ScopePlan, selected bool) (ScopeSnapshot, error) {
	result := ScopeSnapshot{Name: scope.Name}
	if !selected {
		return c.collectScope(ctx, runtime, scope)
	}
	var err error
	if !plan.Inheritance.Unmanaged["settings"] {
		result.Settings, err = runtime.ReadSettings(ctx, scope)
		if err != nil {
			return result, err
		}
		if result.Settings == nil {
			result.Settings = settings.Document{}
		}
	}
	if !plan.Inheritance.Unmanaged["extensions"] {
		result.Extensions, err = runtime.Extensions(ctx, scope)
		if err != nil {
			return result, err
		}
		if result.Extensions == nil {
			result.Extensions = []runtimeio.Extension{}
		}
	}
	result.Inheritance, err = runtime.ReadInheritance(ctx, scope)
	if err != nil {
		return result, err
	}
	result.Inheritance.Unmanaged = plan.Inheritance.Unmanaged
	if artifacts, ok := runtime.(runtimeio.ArtifactRuntime); ok {
		if err := c.collectPlanArtifacts(ctx, artifacts, plan, scope, &result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (c Collector) collectPlanArtifacts(ctx context.Context, runtime runtimeio.ArtifactRuntime, plan cookbook.ScopePlan, scope runtimeio.Scope, snapshot *ScopeSnapshot) error {
	managed := func(kind string, inherited bool, value any) bool {
		if plan.Inheritance.Unmanaged[kind] || !scope.IsDefault() && inherited || value == nil {
			return false
		}
		reflected := reflect.ValueOf(value)
		switch reflected.Kind() {
		case reflect.Map, reflect.Slice, reflect.Pointer, reflect.Interface:
			return !reflected.IsNil()
		default:
			return true
		}
	}
	var err error
	if managed("keybindings", plan.Inheritance.Keybindings, plan.Keybindings) {
		snapshot.Keybindings, err = runtime.ReadKeybindings(ctx, scope)
		if err != nil {
			return fmt.Errorf("observe %q Keybindings: %w", scope.Name, err)
		}
	}
	if managed("tasks", plan.Inheritance.Tasks, plan.Tasks) {
		snapshot.Tasks, err = runtime.ReadTasks(ctx, scope)
		if err != nil {
			return fmt.Errorf("observe %q Tasks: %w", scope.Name, err)
		}
	}
	if managed("mcp", plan.Inheritance.MCP, plan.MCP) {
		snapshot.MCP, err = runtime.ReadMCP(ctx, scope)
		if err != nil {
			return fmt.Errorf("observe %q MCP: %w", scope.Name, err)
		}
	}
	if managed("snippets", plan.Inheritance.Snippets, plan.Snippets) {
		snapshot.Snippets, err = runtime.ReadSnippets(ctx, scope)
		if err != nil {
			return fmt.Errorf("observe %q Snippets: %w", scope.Name, err)
		}
	}
	return nil
}

// Observe captures every scope exposed by a Runtime before a Recipe Plan is
// available. Callers must validate the resulting Snapshot against the Plan
// they derive from this observation before trusting it for recovery.
func (c Collector) Observe(ctx context.Context, runtime runtimeio.Runtime, recipeName, platform string) (Snapshot, error) {
	scopes, err := runtime.Scopes(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("observe Runtime scopes: %w", err)
	}
	seen := map[string]bool{}
	snapshot := Snapshot{FormatVersion: FormatVersion, RecipeName: recipeName, Platform: platform}
	if c.Now != nil {
		snapshot.ObservedAt = c.Now()
	} else {
		snapshot.ObservedAt = time.Now()
	}
	for _, scope := range scopes {
		if seen[scope.Name] {
			return Snapshot{}, fmt.Errorf("duplicate Runtime scope %q", scope.Name)
		}
		seen[scope.Name] = true
		value, err := c.collectScope(ctx, runtime, scope)
		if err != nil {
			return Snapshot{}, err
		}
		if scope.IsDefault() {
			snapshot.Default = value
		} else {
			snapshot.Profiles = append(snapshot.Profiles, value)
		}
	}
	if !seen[""] {
		return Snapshot{}, fmt.Errorf("default Runtime scope not found")
	}
	sort.Slice(snapshot.Profiles, func(i, j int) bool { return snapshot.Profiles[i].Name < snapshot.Profiles[j].Name })
	return snapshot, nil
}

func (c Collector) collectScope(ctx context.Context, runtime runtimeio.Runtime, scope runtimeio.Scope) (ScopeSnapshot, error) {
	settingsValue, err := runtime.ReadSettings(ctx, scope)
	if err != nil {
		return ScopeSnapshot{}, fmt.Errorf("observe %q Settings: %w", scope.Name, err)
	}
	if settingsValue == nil {
		settingsValue = settings.Document{}
	}
	extensions, err := runtime.Extensions(ctx, scope)
	if err != nil {
		return ScopeSnapshot{}, fmt.Errorf("observe %q Extensions: %w", scope.Name, err)
	}
	if extensions == nil {
		extensions = []runtimeio.Extension{}
	}
	inheritance, err := runtime.ReadInheritance(ctx, scope)
	if err != nil {
		return ScopeSnapshot{}, fmt.Errorf("observe %q inheritance: %w", scope.Name, err)
	}
	return ScopeSnapshot{Name: scope.Name, Settings: settingsValue, Extensions: extensions, Inheritance: inheritance}, nil
}

func Validate(snapshot Snapshot, plan cookbook.Plan) error {
	if err := ValidateStructure(snapshot, plan.Name, plan.Platform); err != nil {
		return err
	}
	profiles := map[string]ScopeSnapshot{}
	for _, profile := range snapshot.Profiles {
		profiles[profile.Name] = profile
	}
	for _, required := range plan.Profiles {
		observed, ok := profiles[required.Name]
		if !ok {
			return fmt.Errorf("required Profile %q not observed", required.Name)
		}
		if !required.Inheritance.Unmanaged["settings"] && !required.Inheritance.Settings && observed.Settings == nil {
			return fmt.Errorf("required Profile %q Settings not observed", required.Name)
		}
		if !required.Inheritance.Unmanaged["extensions"] && observed.Extensions == nil {
			return fmt.Errorf("required Profile %q Extensions not observed", required.Name)
		}
		if err := validateManagedArtifacts(required, observed); err != nil {
			return err
		}
	}
	if plan.Default.Settings != nil && !plan.Default.Inheritance.Unmanaged["settings"] && snapshot.Default.Settings == nil {
		return fmt.Errorf("default Settings not observed")
	}
	if !plan.Default.Inheritance.Unmanaged["extensions"] && snapshot.Default.Extensions == nil {
		return fmt.Errorf("default Extensions not observed")
	}
	return nil
}

func validateManagedArtifacts(plan cookbook.ScopePlan, observed ScopeSnapshot) error {
	profile := plan.Name != ""
	checks := []struct {
		name         string
		inherited    bool
		planned, got any
	}{
		{"Keybindings", plan.Inheritance.Keybindings, plan.Keybindings, observed.Keybindings}, {"Tasks", plan.Inheritance.Tasks, plan.Tasks, observed.Tasks}, {"MCP", plan.Inheritance.MCP, plan.MCP, observed.MCP}, {"Snippets", plan.Inheritance.Snippets, plan.Snippets, observed.Snippets},
	}
	for _, check := range checks {
		planned := reflect.ValueOf(check.planned)
		if plan.Inheritance.Unmanaged[strings.ToLower(check.name)] || profile && check.inherited || !planned.IsValid() || planned.IsNil() {
			continue
		}
		value := reflect.ValueOf(check.got)
		if !value.IsValid() || value.IsNil() {
			return fmt.Errorf("required Profile %q %s not observed", plan.Name, check.name)
		}
	}
	return nil
}

func ValidateStructure(snapshot Snapshot, recipeName, platform string) error {
	if snapshot.FormatVersion != FormatVersion {
		return fmt.Errorf("unsupported Lock format version %d", snapshot.FormatVersion)
	}
	if snapshot.RecipeName != recipeName || snapshot.Platform != platform {
		return fmt.Errorf("Lock provenance does not match Runtime Plan")
	}
	profiles := map[string]ScopeSnapshot{}
	for _, profile := range snapshot.Profiles {
		if profile.Name == "" {
			return fmt.Errorf("Lock contains an unnamed Profile")
		}
		if _, exists := profiles[profile.Name]; exists {
			return fmt.Errorf("Lock contains duplicate Profile %q", profile.Name)
		}
		profiles[profile.Name] = profile
	}
	if err := validateExtensions("default", snapshot.Default.Extensions); err != nil {
		return err
	}
	for _, profile := range snapshot.Profiles {
		if err := validateExtensions(profile.Name, profile.Extensions); err != nil {
			return err
		}
	}
	return nil
}

func validateExtensions(scope string, extensions []runtimeio.Extension) error {
	seen := map[string]bool{}
	for _, extension := range extensions {
		if extension.ID == "" {
			return fmt.Errorf("Lock contains an empty Extension ID in %q", scope)
		}
		if seen[extension.ID] {
			return fmt.Errorf("Lock contains duplicate Extension %q in %q", extension.ID, scope)
		}
		seen[extension.ID] = true
	}
	return nil
}

func Read(root string, plan cookbook.Plan) (Snapshot, string, error) {
	manifestPath := filepath.Join(root, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if os.IsNotExist(err) {
		return Snapshot{}, "", fmt.Errorf("Go trusted Lock not found: %s", manifestPath)
	}
	if err != nil {
		return Snapshot{}, "", fmt.Errorf("read trusted Lock: %w", err)
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, "", fmt.Errorf("parse trusted Lock: %w", err)
	}
	if err := ValidateStructure(snapshot, plan.Name, plan.Platform); err != nil {
		return Snapshot{}, "", fmt.Errorf("validate trusted Lock: %w", err)
	}
	recipePath := filepath.Join(root, "recipe.yaml")
	if info, err := os.Stat(recipePath); err != nil || info.IsDir() {
		return Snapshot{}, "", fmt.Errorf("trusted Lock Recipe not found: %s", recipePath)
	}
	return snapshot, recipePath, nil
}

type Store struct {
	Attempts int
	Delay    func(int)
}

// Seal publishes a previously observed Snapshot only after it satisfies the
// derived Plan. It is used by activation to establish trusted source state
// before any host path is changed.
func (s Store) Seal(distPath, recipePath string, snapshot Snapshot, plan cookbook.Plan) error {
	if err := Validate(snapshot, plan); err != nil {
		return fmt.Errorf("validate source Lock: %w", err)
	}
	return s.publish(distPath, recipePath, snapshot)
}

func (s Store) Refresh(ctx context.Context, distPath, recipePath string, runtime runtimeio.Runtime, plan cookbook.Plan) (Snapshot, error) {
	attempts := s.Attempts
	if attempts <= 0 {
		attempts = 3
	}
	var snapshot Snapshot
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		snapshot, err = (Collector{}).Collect(ctx, runtime, plan)
		if err == nil {
			err = s.publish(distPath, recipePath, snapshot)
			if err == nil {
				return snapshot, nil
			}
		}
		if attempt < attempts {
			if s.Delay != nil {
				s.Delay(attempt)
			} else {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
		}
	}
	return Snapshot{}, fmt.Errorf("Lock failed after %d attempts: %w", attempts, err)
}

func (s Store) Reuse(distPath string, plan cookbook.Plan) error {
	root := filepath.Join(distPath, ".lock")
	_, _, err := Read(root, plan)
	return err
}

func (s Store) publish(distPath, recipePath string, snapshot Snapshot) error {
	staging, err := os.MkdirTemp(distPath, ".lock.staging-")
	if err != nil {
		return fmt.Errorf("create Lock staging: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := writeSnapshot(staging, recipePath, snapshot); err != nil {
		return err
	}
	final := filepath.Join(distPath, ".lock")
	backup := filepath.Join(distPath, ".lock.previous")
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(final); err == nil {
		if err := os.Rename(final, backup); err != nil {
			return fmt.Errorf("preserve previous Lock: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect previous Lock: %w", err)
	}
	if err := os.Rename(staging, final); err != nil {
		_ = os.Rename(backup, final)
		return fmt.Errorf("publish Lock: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove previous Lock: %w", err)
	}
	return nil
}

func writeSnapshot(root, recipePath string, snapshot Snapshot) error {
	manifest, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal Lock: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), append(manifest, '\n'), 0o644); err != nil {
		return fmt.Errorf("write Lock manifest: %w", err)
	}
	recipeData, err := os.ReadFile(recipePath)
	if err != nil {
		return fmt.Errorf("read Lock Recipe: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "recipe.yaml"), recipeData, 0o644); err != nil {
		return fmt.Errorf("write Lock Recipe: %w", err)
	}
	if err := writeScope(root, "runtime", snapshot.Default); err != nil {
		return err
	}
	for _, profile := range snapshot.Profiles {
		if err := writeScope(root, profile.Name, profile); err != nil {
			return err
		}
	}
	return nil
}

func writeScope(root, name string, snapshot ScopeSnapshot) error {
	settingsName := name + ".settings.jsonc"
	if name == "runtime" {
		settingsName = "settings.jsonc"
	}
	settingsData, err := settings.Marshal(snapshot.Settings)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, settingsName), settingsData, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", settingsName, err)
	}
	extensionName := name + ".extensions.lock"
	file, err := os.Create(filepath.Join(root, extensionName))
	if err != nil {
		return fmt.Errorf("create %s: %w", extensionName, err)
	}
	for _, extension := range snapshot.Extensions {
		value := extension.ID
		if extension.Version != "" {
			value += "@" + extension.Version
		}
		if _, err := fmt.Fprintln(file, value); err != nil {
			file.Close()
			return fmt.Errorf("write %s: %w", extensionName, err)
		}
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", extensionName, err)
	}
	return nil
}
