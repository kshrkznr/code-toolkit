package archive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/converge"
	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
)

const FormatVersion = 1

type FileRecord struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type Manifest struct {
	FormatVersion      int                   `json:"formatVersion"`
	Name               string                `json:"name"`
	SourceDistribution string                `json:"sourceDistribution"`
	RecipeName         string                `json:"recipeName"`
	OS                 string                `json:"os"`
	Platform           string                `json:"platform"`
	CreatedAt          time.Time             `json:"createdAt"`
	Extensions         []runtimeio.Extension `json:"extensions"`
	LaunchOverrides    []string              `json:"launchOverrides,omitempty"`
	Files              []FileRecord          `json:"files"`
}
type Bundle struct {
	Path     string
	Manifest Manifest
	Recipe   recipe.Recipe
	Snapshot runtimelock.Snapshot
}
type RuntimeFactory func(distribution.Distribution) (runtimeio.Runtime, error)
type Service struct {
	Cookbook   cookbook.Repository
	Runtime    RuntimeFactory
	Locks      runtimelock.Store
	ChooseLock func() (string, error)
	Pool       converge.PoolUpdater
	Now        func() time.Time
}
type Options struct{ OnConflict string }
type Result struct {
	Bundle   Bundle
	Warnings []string
}

func (s Service) Create(ctx context.Context, root string, dist distribution.Distribution, options Options) (Result, error) {
	for _, required := range []string{".data", ".ext"} {
		info, err := os.Stat(filepath.Join(dist.Path, required))
		if err != nil || !info.IsDir() {
			return Result{}, fmt.Errorf("Distribution runtime directory missing: %s", filepath.Join(dist.Path, required))
		}
	}
	recipePath := filepath.Join(dist.Path, ".meta", "recipe.yaml")
	if info, err := os.Stat(recipePath); err != nil || info.IsDir() {
		return Result{}, fmt.Errorf("Distribution Recipe missing: %s", recipePath)
	}
	plan, err := s.Cookbook.Resolve(recipePath)
	if err != nil {
		return Result{}, err
	}
	snapshot, err := s.selectLock(ctx, dist, recipePath, plan)
	if err != nil {
		return Result{}, err
	}
	extensions, err := uniqueExtensions(snapshot)
	if err != nil {
		return Result{}, err
	}
	name := dist.Name
	target := filepath.Join(root, name)
	mode := options.OnConflict
	if mode == "" {
		mode = "suffix"
	}
	if _, err := os.Stat(target); err == nil {
		switch mode {
		case "suffix":
			name, err = nextName(root, name)
			if err != nil {
				return Result{}, err
			}
			target = filepath.Join(root, name)
		case "replace":
		case "abort":
			return Result{}, fmt.Errorf("Archive already exists: %s", target)
		default:
			return Result{}, fmt.Errorf("invalid Archive conflict mode %q", mode)
		}
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return Result{}, err
	}
	staging, err := os.MkdirTemp(root, ".archive-"+name+"-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(staging)
	if err := copyTree(filepath.Join(dist.Path, ".lock"), filepath.Join(staging, "lock")); err != nil {
		return Result{}, fmt.Errorf("copy trusted Lock: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(staging, "vsix"), 0755); err != nil {
		return Result{}, err
	}
	for _, extension := range extensions {
		artifact, ensureErr := s.Pool.ResolveExact(plan.Platform, extension)
		if ensureErr != nil && plan.ExtensionPool == "refresh" {
			artifact, ensureErr = s.Pool.EnsureExact(ctx, plan.Platform, extension)
		}
		if ensureErr != nil {
			if plan.ExtensionPool == "refresh" {
				return Result{}, fmt.Errorf("resolve Archive VSIX with extension-pool refresh: %w", ensureErr)
			}
			return Result{}, fmt.Errorf("resolve Archive VSIX: %w; set config.dist-strategy.extension-pool to refresh to permit Repository download", ensureErr)
		}
		if err := copyFile(artifact, filepath.Join(staging, "vsix", artifactName(extension))); err != nil {
			return Result{}, err
		}
	}
	result := Result{}
	var overrideNames []string
	for _, override := range []string{"run.sh", "run.cmd"} {
		source := filepath.Join(dist.Path, override)
		if info, statErr := os.Stat(source); statErr == nil && !info.IsDir() {
			destination := filepath.Join(staging, "launch-override", override)
			if err := copyFile(source, destination); err != nil {
				return Result{}, err
			}
			result.Warnings = append(result.Warnings, "Launch Override preserved but will not be restored: "+override)
			overrideNames = append(overrideNames, override)
		}
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	manifest := Manifest{FormatVersion: FormatVersion, Name: name, SourceDistribution: dist.Name, RecipeName: plan.Name, OS: plan.OS, Platform: plan.Platform, CreatedAt: now, Extensions: extensions}
	manifest.LaunchOverrides = overrideNames
	manifest.Files, err = hashTree(staging)
	if err != nil {
		return Result{}, err
	}
	if err := writeManifest(staging, manifest); err != nil {
		return Result{}, err
	}
	if _, err := Load(staging); err != nil {
		return Result{}, fmt.Errorf("validate Archive staging: %w", err)
	}
	if err := publish(staging, target, mode == "replace"); err != nil {
		return Result{}, err
	}
	bundle, err := Load(target)
	if err != nil {
		return Result{}, err
	}
	result.Bundle = bundle
	return result, nil
}

func (s Service) selectLock(ctx context.Context, dist distribution.Distribution, recipePath string, plan cookbook.Plan) (runtimelock.Snapshot, error) {
	mode := plan.LockMode
	if mode == "ask" {
		if s.ChooseLock == nil {
			return runtimelock.Snapshot{}, fmt.Errorf("lock-mode ask requires selector")
		}
		selected, err := s.ChooseLock()
		if err != nil {
			return runtimelock.Snapshot{}, err
		}
		mode = selected
	}
	switch mode {
	case "refresh":
		runtime, err := s.Runtime(dist)
		if err != nil {
			return runtimelock.Snapshot{}, err
		}
		return s.Locks.Refresh(ctx, dist.Path, recipePath, runtime, plan)
	case "reuse":
		snapshot, _, err := runtimelock.Read(filepath.Join(dist.Path, ".lock"), plan)
		return snapshot, err
	case "abort":
		return runtimelock.Snapshot{}, fmt.Errorf("Archive Lock declined")
	default:
		return runtimelock.Snapshot{}, fmt.Errorf("invalid lock-mode %q", mode)
	}
}

func Load(root string) (Bundle, error) {
	data, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return Bundle{}, fmt.Errorf("read Archive manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Bundle{}, fmt.Errorf("parse Archive manifest: %w", err)
	}
	if manifest.FormatVersion != FormatVersion {
		return Bundle{}, fmt.Errorf("unsupported Archive format version %d", manifest.FormatVersion)
	}
	recorded := map[string]bool{}
	for _, record := range manifest.Files {
		if filepath.IsAbs(record.Path) || strings.HasPrefix(filepath.Clean(record.Path), "..") {
			return Bundle{}, fmt.Errorf("unsafe Archive path %q", record.Path)
		}
		path := filepath.Join(root, filepath.FromSlash(record.Path))
		info, statErr := os.Lstat(path)
		if statErr != nil || !info.Mode().IsRegular() {
			return Bundle{}, fmt.Errorf("Archive entry is not a regular file: %s", record.Path)
		}
		hash, size, err := hashFile(path)
		if err != nil {
			return Bundle{}, fmt.Errorf("validate Archive file %s: %w", record.Path, err)
		}
		if hash != record.SHA256 || size != record.Size {
			return Bundle{}, fmt.Errorf("Archive integrity mismatch: %s", record.Path)
		}
		if recorded[record.Path] {
			return Bundle{}, fmt.Errorf("duplicate Archive file record: %s", record.Path)
		}
		recorded[record.Path] = true
	}
	for _, extension := range manifest.Extensions {
		if extension.ID == "" || extension.Version == "" || strings.ContainsAny(extension.ID+extension.Version, `/\\`) {
			return Bundle{}, fmt.Errorf("invalid archived Extension identity: %s@%s", extension.ID, extension.Version)
		}
		path := filepath.Join(root, "vsix", artifactName(extension))
		relative := filepath.ToSlash(filepath.Join("vsix", artifactName(extension)))
		if !recorded[relative] {
			return Bundle{}, fmt.Errorf("Archive file record missing: %s", relative)
		}
		if err := converge.ValidateVSIXExact(path, extension); err != nil {
			return Bundle{}, fmt.Errorf("validate %s: %w", filepath.Base(path), err)
		}
	}
	recipePath := filepath.Join(root, "lock", "recipe.yaml")
	definition, err := recipe.Load(recipePath)
	if err != nil {
		return Bundle{}, err
	}
	if definition.Name != manifest.RecipeName || definition.OS != manifest.OS || definition.Platform != manifest.Platform {
		return Bundle{}, fmt.Errorf("Archive Recipe identity mismatch")
	}
	plan := cookbook.Plan{Name: manifest.RecipeName, OS: manifest.OS, Platform: manifest.Platform}
	snapshot, _, err := runtimelock.Read(filepath.Join(root, "lock"), plan)
	if err != nil {
		return Bundle{}, err
	}
	for _, required := range []string{"lock/manifest.json", "lock/recipe.yaml"} {
		if !recorded[required] {
			return Bundle{}, fmt.Errorf("Archive file record missing: %s", required)
		}
	}
	expectedExtensions, err := uniqueExtensions(snapshot)
	if err != nil {
		return Bundle{}, err
	}
	if !sameExtensions(expectedExtensions, manifest.Extensions) {
		return Bundle{}, fmt.Errorf("Archive Extension manifest does not match trusted Lock")
	}
	for _, override := range manifest.LaunchOverrides {
		if override != "run.sh" && override != "run.cmd" {
			return Bundle{}, fmt.Errorf("invalid archived Launch Override: %s", override)
		}
		relative := "launch-override/" + override
		if !recorded[relative] {
			return Bundle{}, fmt.Errorf("Archive file record missing: %s", relative)
		}
	}
	return Bundle{Path: root, Manifest: manifest, Recipe: definition, Snapshot: snapshot}, nil
}

func Plan(bundle Bundle) cookbook.Plan {
	convert := func(scope runtimelock.ScopeSnapshot) cookbook.ScopePlan {
		ids := make([]string, 0, len(scope.Extensions))
		for _, extension := range scope.Extensions {
			ids = append(ids, extension.ID)
		}
		return cookbook.ScopePlan{Name: scope.Name, Settings: scope.Settings, Keybindings: scope.Keybindings, Tasks: scope.Tasks, MCP: scope.MCP, Snippets: scope.Snippets, Extensions: ids, Inheritance: scope.Inheritance}
	}
	plan := cookbook.Plan{Name: bundle.Manifest.RecipeName, OS: bundle.Manifest.OS, Platform: bundle.Manifest.Platform, ExtensionMarketplace: false, ExtensionPool: "reuse", LockMode: "refresh", Default: convert(bundle.Snapshot.Default)}
	for _, scope := range bundle.Snapshot.Profiles {
		plan.Profiles = append(plan.Profiles, convert(scope))
	}
	return plan
}

func Verify(expected, actual runtimelock.Snapshot) error {
	if expected.RecipeName != actual.RecipeName || expected.Platform != actual.Platform {
		return fmt.Errorf("Archive verification provenance mismatch")
	}
	normalizeTasks := func(snapshot runtimelock.Snapshot) runtimelock.Snapshot {
		snapshot.Default.Tasks = runtimeartifact.NormalizeTasksForComparison(snapshot.Default.Tasks)
		snapshot.Profiles = append([]runtimelock.ScopeSnapshot(nil), snapshot.Profiles...)
		for index := range snapshot.Profiles {
			snapshot.Profiles[index].Tasks = runtimeartifact.NormalizeTasksForComparison(snapshot.Profiles[index].Tasks)
		}
		return snapshot
	}
	expected = normalizeTasks(expected)
	actual = normalizeTasks(actual)
	expectedData, _ := json.Marshal(struct {
		Default  runtimelock.ScopeSnapshot
		Profiles []runtimelock.ScopeSnapshot
	}{expected.Default, expected.Profiles})
	actualData, _ := json.Marshal(struct {
		Default  runtimelock.ScopeSnapshot
		Profiles []runtimelock.ScopeSnapshot
	}{actual.Default, actual.Profiles})
	if string(expectedData) != string(actualData) {
		return fmt.Errorf("Archive reconstruction verification mismatch")
	}
	return nil
}

type Resolver struct{ Root string }

func (r Resolver) ResolveExact(extension runtimeio.Extension) (string, error) {
	path := filepath.Join(r.Root, "vsix", artifactName(extension))
	if err := converge.ValidateVSIXExact(path, extension); err != nil {
		return "", err
	}
	return path, nil
}
func artifactName(extension runtimeio.Extension) string {
	return extension.ID + "-" + extension.Version + ".vsix"
}
func uniqueExtensions(snapshot runtimelock.Snapshot) ([]runtimeio.Extension, error) {
	seen := map[string]runtimeio.Extension{}
	collect := func(values []runtimeio.Extension) error {
		for _, extension := range values {
			if extension.ID == "" || extension.Version == "" {
				return fmt.Errorf("Archive requires versioned Extension observation: %s", extension.ID)
			}
			seen[extension.ID+"@"+extension.Version] = extension
		}
		return nil
	}
	if err := collect(snapshot.Default.Extensions); err != nil {
		return nil, err
	}
	for _, scope := range snapshot.Profiles {
		if err := collect(scope.Extensions); err != nil {
			return nil, err
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]runtimeio.Extension, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result, nil
}
func sameExtensions(left, right []runtimeio.Extension) bool {
	if len(left) != len(right) {
		return false
	}
	leftKeys := make([]string, len(left))
	rightKeys := make([]string, len(right))
	for index, value := range left {
		leftKeys[index] = value.ID + "@" + value.Version
	}
	for index, value := range right {
		rightKeys[index] = value.ID + "@" + value.Version
	}
	sort.Strings(leftKeys)
	sort.Strings(rightKeys)
	return strings.Join(leftKeys, "\x00") == strings.Join(rightKeys, "\x00")
}

func writeManifest(root string, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "manifest.json"), append(data, '\n'), 0644)
}
func hashTree(root string) ([]FileRecord, error) {
	var result []FileRecord
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || path == filepath.Join(root, "manifest.json") {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		hash, size, err := hashFile(path)
		if err != nil {
			return err
		}
		result = append(result, FileRecord{Path: filepath.ToSlash(relative), SHA256: hash, Size: size})
		return nil
	})
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, err
}
func hashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil)), size, err
}
func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(source, path)
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}
func copyFile(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
func nextName(root, name string) (string, error) {
	for index := 1; ; index++ {
		candidate := fmt.Sprintf("%s.%d", name, index)
		if _, err := os.Stat(filepath.Join(root, candidate)); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
}
func NextAvailableName(root, name string) (string, error) { return nextName(root, name) }
func publish(staging, target string, replace bool) error {
	backup := target + ".previous"
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(target); err == nil {
		if !replace {
			return fmt.Errorf("Archive already exists: %s", target)
		}
		if err := os.Rename(target, backup); err != nil {
			return err
		}
	}
	if err := os.Rename(staging, target); err != nil {
		_ = os.Rename(backup, target)
		return err
	}
	return os.RemoveAll(backup)
}
