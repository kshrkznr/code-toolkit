package lifecycle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	ctkarchive "github.com/kshrkznr/code-toolkit/go/internal/archive"
	"github.com/kshrkznr/code-toolkit/go/internal/converge"
	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/directlauncher"
	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
)

type RuntimeFactory func(distribution.Distribution) (runtimeio.Runtime, error)

type Service struct {
	Cookbook   cookbook.Repository
	Runtime    RuntimeFactory
	Pool       converge.ArtifactResolver
	PoolUpdate interface {
		Update(context.Context, string, runtimelock.Snapshot, *converge.Report)
	}
	Locks      runtimelock.Store
	ChooseLock func() (string, error)
}

type Result struct {
	Distribution distribution.Distribution
	Report       converge.Report
	StagingPath  string
	Warnings     []string
}

func (s Service) BuildArchive(ctx context.Context, bundle ctkarchive.Bundle, distRoot, name string, keepStaging bool) (result Result, err error) {
	plan := ctkarchive.Plan(bundle)
	final := filepath.Join(distRoot, name)
	if _, err := os.Lstat(final); err == nil {
		return result, fmt.Errorf("Distribution already exists: %s", name)
	} else if !os.IsNotExist(err) {
		return result, err
	}
	if err = os.MkdirAll(distRoot, 0755); err != nil {
		return result, err
	}
	staging, err := os.MkdirTemp(distRoot, ".build-archive-")
	if err != nil {
		return result, err
	}
	result.StagingPath = staging
	defer func() {
		if err != nil && keepStaging {
			return
		}
		_ = os.RemoveAll(staging)
	}()
	for _, directory := range []string{".data", ".ext", ".meta"} {
		if err = os.MkdirAll(filepath.Join(staging, directory), 0755); err != nil {
			return result, err
		}
	}
	if err = copyAtomic(filepath.Join(bundle.Path, "lock", "recipe.yaml"), filepath.Join(staging, ".meta", "recipe.yaml")); err != nil {
		return result, err
	}
	dist := distribution.Distribution{Name: name, Path: staging, Recipe: bundle.Recipe}
	runtime, err := s.Runtime(dist)
	if err != nil {
		return result, err
	}
	result.Report = converge.ArchiveSnapshot(ctx, runtime, plan, bundle.Snapshot, ctkarchive.Resolver{Root: bundle.Path})
	if reportErr := result.Report.Error(); reportErr != nil {
		return result, reportErr
	}
	fresh, err := (runtimelock.Collector{}).Collect(ctx, runtime, plan)
	if err != nil {
		return result, err
	}
	if err = ctkarchive.Verify(bundle.Snapshot, fresh); err != nil {
		return result, err
	}
	if err = s.Locks.Seal(staging, filepath.Join(bundle.Path, "lock", "recipe.yaml"), fresh, plan); err != nil {
		return result, err
	}
	if err = generateDirectLauncher(dist, &result); err != nil {
		return result, err
	}
	if err = os.Rename(staging, final); err != nil {
		return result, err
	}
	result.StagingPath = ""
	result.Distribution, err = distribution.Load(distRoot, name)
	for _, override := range bundle.Manifest.LaunchOverrides {
		result.Warnings = append(result.Warnings, "archived Launch Override was not restored: "+override)
	}
	return result, err
}

func (s Service) ApplyArchive(ctx context.Context, bundle ctkarchive.Bundle, dist distribution.Distribution) (result Result, err error) {
	plan := ctkarchive.Plan(bundle)
	result.Distribution = dist
	if dist.Recipe.Name != plan.Name || dist.Recipe.OS != plan.OS || dist.Recipe.Platform != plan.Platform {
		return result, fmt.Errorf("Archive identity does not match Distribution")
	}
	runtime, err := s.Runtime(dist)
	if err != nil {
		return result, err
	}
	result.Report = converge.ArchiveSnapshot(ctx, runtime, plan, bundle.Snapshot, ctkarchive.Resolver{Root: bundle.Path})
	if reportErr := result.Report.Error(); reportErr != nil {
		return result, reportErr
	}
	if err = copyAtomic(filepath.Join(bundle.Path, "lock", "recipe.yaml"), filepath.Join(dist.Path, ".meta", "recipe.yaml")); err != nil {
		return result, err
	}
	fresh, err := (runtimelock.Collector{}).Collect(ctx, runtime, plan)
	if err != nil {
		return result, err
	}
	if err = ctkarchive.Verify(bundle.Snapshot, fresh); err != nil {
		return result, err
	}
	if err = s.Locks.Seal(dist.Path, filepath.Join(bundle.Path, "lock", "recipe.yaml"), fresh, plan); err != nil {
		return result, err
	}
	if err = generateDirectLauncher(distribution.Distribution{Name: dist.Name, Path: dist.Path, Recipe: bundle.Recipe}, &result); err != nil {
		return result, err
	}
	result.Distribution, err = distribution.Load(filepath.Dir(dist.Path), dist.Path)
	for _, override := range bundle.Manifest.LaunchOverrides {
		result.Warnings = append(result.Warnings, "archived Launch Override was not restored: "+override)
	}
	return result, err
}

func (s Service) Apply(ctx context.Context, recipePath string, dist distribution.Distribution, forceExtensions bool) (Result, error) {
	plan, err := s.Cookbook.Resolve(recipePath)
	if err != nil {
		return Result{}, err
	}
	if plan.Platform != dist.Recipe.Platform {
		return Result{}, fmt.Errorf("Recipe platform %q does not match Distribution platform %q", plan.Platform, dist.Recipe.Platform)
	}
	runtime, err := s.Runtime(dist)
	if err != nil {
		return Result{}, err
	}
	report := converge.Plan(ctx, runtime, plan, s.Pool, forceExtensions)
	result := Result{Distribution: dist, Report: report}
	if err := report.Error(); err != nil {
		return result, err
	}
	if err := copyAtomic(recipePath, filepath.Join(dist.Path, ".meta", "recipe.yaml")); err != nil {
		return result, err
	}
	snapshot, err := s.finishLock(ctx, dist.Path, recipePath, runtime, plan)
	if err != nil {
		return result, err
	}
	if s.PoolUpdate != nil && plan.ExtensionPool == "refresh" {
		s.PoolUpdate.Update(ctx, plan.Platform, snapshot, &result.Report)
	}
	if err := generateDirectLauncher(distribution.Distribution{Name: dist.Name, Path: dist.Path, Recipe: recipe.Recipe{Name: plan.Name, OS: plan.OS, Platform: plan.Platform}}, &result); err != nil {
		return result, err
	}
	metadata, err := distribution.Load(filepath.Dir(dist.Path), dist.Path)
	if err == nil {
		result.Distribution = metadata
	}
	return result, err
}

func (s Service) Lock(ctx context.Context, dist distribution.Distribution) (converge.Report, error) {
	report := converge.Report{}
	recipePath := filepath.Join(dist.Path, ".meta", "recipe.yaml")
	plan, err := s.Cookbook.Resolve(recipePath)
	if err != nil {
		return report, err
	}
	runtime, err := s.Runtime(dist)
	if err != nil {
		return report, err
	}
	snapshot, err := s.Locks.Refresh(ctx, dist.Path, recipePath, runtime, plan)
	if err == nil && s.PoolUpdate != nil && plan.ExtensionPool == "refresh" {
		s.PoolUpdate.Update(ctx, plan.Platform, snapshot, &report)
	}
	return report, err
}

func (s Service) Build(ctx context.Context, recipePath, distRoot, name string, keepStaging, forceExtensions bool) (result Result, err error) {
	plan, err := s.Cookbook.Resolve(recipePath)
	if err != nil {
		return Result{}, err
	}
	final := filepath.Join(distRoot, name)
	if _, err := os.Lstat(final); err == nil {
		return Result{}, fmt.Errorf("Distribution already exists: %s", name)
	} else if !os.IsNotExist(err) {
		return Result{}, err
	}
	if err := os.MkdirAll(distRoot, 0o755); err != nil {
		return Result{}, fmt.Errorf("create Distribution root: %w", err)
	}
	staging, err := os.MkdirTemp(distRoot, ".build-")
	if err != nil {
		return Result{}, fmt.Errorf("create Build staging: %w", err)
	}
	result.StagingPath = staging
	defer func() {
		if err != nil && keepStaging {
			return
		}
		_ = os.RemoveAll(staging)
	}()
	for _, directory := range []string{".data", ".ext", ".meta"} {
		if err = os.MkdirAll(filepath.Join(staging, directory), 0o755); err != nil {
			return result, fmt.Errorf("create Runtime directory: %w", err)
		}
	}
	if err = copyAtomic(recipePath, filepath.Join(staging, ".meta", "recipe.yaml")); err != nil {
		return result, err
	}
	dist := distribution.Distribution{Name: name, Path: staging, Recipe: recipe.Recipe{Name: plan.Name, OS: plan.OS, Platform: plan.Platform}}
	runtime, err := s.Runtime(dist)
	if err != nil {
		return result, err
	}
	result.Report = converge.Plan(ctx, runtime, plan, s.Pool, forceExtensions)
	if reportErr := result.Report.Error(); reportErr != nil {
		return result, reportErr
	}
	var snapshot runtimelock.Snapshot
	if snapshot, err = s.finishLock(ctx, staging, recipePath, runtime, plan); err != nil {
		return result, err
	}
	if s.PoolUpdate != nil && plan.ExtensionPool == "refresh" {
		s.PoolUpdate.Update(ctx, plan.Platform, snapshot, &result.Report)
	}
	if err = generateDirectLauncher(dist, &result); err != nil {
		return result, err
	}
	if err = os.Rename(staging, final); err != nil {
		return result, fmt.Errorf("publish Distribution %s: %w", name, err)
	}
	result.StagingPath = ""
	result.Distribution, err = distribution.Load(distRoot, name)
	return result, err
}

func generateDirectLauncher(dist distribution.Distribution, result *Result) error {
	launcherResult, err := directlauncher.Generate(dist)
	if err != nil {
		return err
	}
	if launcherResult.Warning != "" {
		result.Warnings = append(result.Warnings, launcherResult.Warning)
	}
	return nil
}

func (s Service) finishLock(ctx context.Context, distPath, recipePath string, runtime runtimeio.Runtime, plan cookbook.Plan) (runtimelock.Snapshot, error) {
	mode := plan.LockMode
	if mode == "ask" {
		if s.ChooseLock == nil {
			return runtimelock.Snapshot{}, fmt.Errorf("lock-mode ask requires an interactive selector")
		}
		selected, err := s.ChooseLock()
		if err != nil {
			return runtimelock.Snapshot{}, err
		}
		mode = selected
	}
	switch mode {
	case "refresh":
		return s.Locks.Refresh(ctx, distPath, recipePath, runtime, plan)
	case "reuse":
		if err := s.Locks.Reuse(distPath, plan); err != nil {
			return runtimelock.Snapshot{}, err
		}
		snapshot, _, err := runtimelock.Read(filepath.Join(distPath, ".lock"), plan)
		return snapshot, err
	case "abort":
		return runtimelock.Snapshot{}, fmt.Errorf("Lock update declined")
	default:
		return runtimelock.Snapshot{}, fmt.Errorf("invalid lock-mode %q", mode)
	}
}

func copyAtomic(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read provenance: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create provenance directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".recipe-*")
	if err != nil {
		return fmt.Errorf("create provenance staging: %w", err)
	}
	path := temporary.Name()
	defer os.Remove(path)
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write provenance: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close provenance: %w", err)
	}
	if err := os.Rename(path, target); err != nil {
		return fmt.Errorf("publish provenance: %w", err)
	}
	return nil
}

func NextAvailableName(distRoot, base string) (string, error) {
	for suffix := 0; ; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s.%d", base, suffix)
		}
		if _, err := os.Lstat(filepath.Join(distRoot, candidate)); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("inspect Distribution name %s: %w", candidate, err)
		}
	}
}
