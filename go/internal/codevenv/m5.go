package codevenv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/platform"
	"code-toolkit/internal/recipe"
	"code-toolkit/internal/recovery"
	"code-toolkit/internal/runtimeio"
	"code-toolkit/internal/runtimelock"
	"go.yaml.in/yaml/v3"
)

type RuntimeFactory func(platform, userDataDir, extensionsDir string) (runtimeio.Runtime, error)

type SafetyGate func(recovery.Verification) (bool, error)

type Service struct {
	DistRoot   string
	Runtime    RuntimeFactory
	HostPaths  func(string) (HostPaths, error)
	Stopper    platform.ProcessStopper
	Recovery   recovery.Service
	Locks      runtimelock.Store
	SafetyGate SafetyGate
}

func (s Service) hostPaths(platformName string) (HostPaths, error) {
	if s.HostPaths != nil {
		return s.HostPaths(platformName)
	}
	return ResolveHostPaths(platformName)
}

type Health struct {
	Platform       string
	State          string
	Current        string
	UserState      string
	ExtensionState string
}

type ActivateOptions struct{ Force bool }
type DeactivateOptions struct {
	Force      bool
	ForceEmpty bool
}

type ActivateResult struct {
	Platform string
	Current  string
	NoOp     bool
	Forced   bool
}

type DeactivateResult struct {
	Platform string
	Forced   bool
	Empty    bool
}

func (s Service) CheckHealth(platformName string, paths HostPaths) (Health, error) {
	health := Health{Platform: platformName}
	current := filepath.Join(s.DistRoot, "current."+platformName)
	currentTarget, currentExists, err := linkTarget(current)
	if err != nil {
		return health, fmt.Errorf("inspect current Selection: %w", err)
	}
	health.UserState = pathState(paths.User, "")
	health.ExtensionState = pathState(paths.Extensions, "")
	if !currentExists {
		if health.UserState == "link" || health.ExtensionState == "link" {
			health.State = "unhealthy"
			return health, nil
		}
		health.State = "inactive"
		return health, nil
	}
	health.Current = filepath.Base(filepath.Clean(currentTarget))
	userExpected := filepath.Join(current, ".data", "User")
	extExpected := filepath.Join(current, ".ext")
	health.UserState = pathState(paths.User, userExpected)
	health.ExtensionState = pathState(paths.Extensions, extExpected)
	if health.UserState == "linked" && health.ExtensionState == "linked" {
		health.State = "healthy"
	} else {
		health.State = "unhealthy"
	}
	return health, nil
}

func (s Service) Activate(ctx context.Context, platformName string, options ActivateOptions) (result ActivateResult, err error) {
	result.Platform = platformName
	paths, err := s.hostPaths(platformName)
	if err != nil {
		return result, err
	}
	if err := os.MkdirAll(s.DistRoot, 0o755); err != nil {
		return result, fmt.Errorf("create Distribution root: %w", err)
	}
	if err := s.recoverInterrupted(platformName); err != nil {
		return result, err
	}
	health, err := s.CheckHealth(platformName, paths)
	if err != nil {
		return result, err
	}
	if health.State == "healthy" {
		result.Current, result.NoOp = health.Current, true
		return result, nil
	}
	if health.State != "inactive" {
		return result, unhealthyError(health)
	}
	if s.Runtime == nil || s.Stopper == nil {
		return result, fmt.Errorf("CodeVenv service is not configured")
	}
	unlock, err := acquireLifecycleLock(s.DistRoot, platformName)
	if err != nil {
		return result, err
	}
	defer unlock()
	if err := s.Stopper.StopForSelection(ctx, platformName); err != nil {
		return result, err
	}
	hostRuntime, err := s.Runtime(platformName, paths.UserData, paths.Extensions)
	if err != nil {
		return result, err
	}
	originName := "origin." + platformName
	snapshot, err := (runtimelock.Collector{}).Observe(ctx, hostRuntime, originName, platformName)
	if err != nil {
		return result, err
	}
	definition, plan := observedRecipe(snapshot)
	sourceRoot, err := os.MkdirTemp(s.DistRoot, ".activate-source-"+platformName+"-")
	if err != nil {
		return result, fmt.Errorf("create activation source: %w", err)
	}
	keepSource := false
	defer func() {
		if !keepSource {
			_ = os.RemoveAll(sourceRoot)
		}
	}()
	recipePath := filepath.Join(sourceRoot, ".meta", "recipe.yaml")
	if err := writeRecipe(recipePath, definition); err != nil {
		return result, err
	}
	if err := s.Locks.Seal(sourceRoot, recipePath, snapshot, plan); err != nil {
		return result, err
	}
	prepared, err := recovery.Prepare(filepath.Join(sourceRoot, ".lock"))
	if err != nil {
		return result, err
	}
	staging, err := os.MkdirTemp(s.DistRoot, ".origin-"+platformName+"-")
	if err != nil {
		return result, fmt.Errorf("create origin staging: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()
	for _, directory := range []string{".data", ".ext", ".meta"} {
		if err := os.MkdirAll(filepath.Join(staging, directory), 0o755); err != nil {
			return result, err
		}
	}
	if err := copyTree(paths.User, filepath.Join(staging, ".data", "User")); err != nil {
		return result, fmt.Errorf("copy Host User into origin: %w", err)
	}
	originRuntime, err := s.Runtime(platformName, filepath.Join(staging, ".data"), filepath.Join(staging, ".ext"))
	if err != nil {
		return result, err
	}
	recovered, err := s.Recovery.Recover(ctx, prepared, staging, originRuntime)
	if err != nil {
		return result, err
	}
	verification := combineVerification(recovered.Before, recovered.After)
	forced, err := s.approve(verification, options.Force)
	if err != nil {
		return result, err
	}
	result.Forced = forced

	origin := filepath.Join(s.DistRoot, originName)
	current := filepath.Join(s.DistRoot, "current."+platformName)
	suffix := fmt.Sprintf(".ctk-%d-%d", time.Now().UnixNano(), os.Getpid())
	j := journal{Operation: "activate", Platform: platformName, Phase: "prepared", Current: current, Origin: origin, HostUser: paths.User, HostExtensions: paths.Extensions}
	jPath := journalPath(s.DistRoot, platformName)
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := rollbackJournal(j); rollbackErr != nil {
			err = errors.Join(err, rollbackErr)
			keepSource = true
			return
		}
		_ = os.Remove(jPath)
	}()
	j.OriginBackup, err = plannedBackup(origin, suffix)
	if err != nil {
		return result, err
	}
	j.Phase = "origin-backup-planned"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := executeBackup(origin, j.OriginBackup); err != nil {
		return result, err
	}
	if err := os.Rename(staging, origin); err != nil {
		return result, fmt.Errorf("publish origin: %w", err)
	}
	j.Phase = "origin-published"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	j.HostUserBackup, err = plannedBackup(paths.User, suffix)
	if err != nil {
		return result, err
	}
	j.HostExtBackup, err = plannedBackup(paths.Extensions, suffix)
	if err != nil {
		return result, err
	}
	j.Phase = "host-backups-planned"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := executeBackup(paths.User, j.HostUserBackup); err != nil {
		return result, err
	}
	if err := executeBackup(paths.Extensions, j.HostExtBackup); err != nil {
		return result, err
	}
	j.Phase = "host-backed-up"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := createManagedLink(origin, current); err != nil {
		return result, fmt.Errorf("activate current Selection: %w", err)
	}
	j.Phase = "current-linked"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := createManagedLink(filepath.Join(current, ".data", "User"), paths.User); err != nil {
		return result, fmt.Errorf("redirect Host User: %w", err)
	}
	if err := createManagedLink(filepath.Join(current, ".ext"), paths.Extensions); err != nil {
		return result, fmt.Errorf("redirect Host Extensions: %w", err)
	}
	j.Phase = "host-linked"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	finalHealth, err := s.CheckHealth(platformName, paths)
	if err != nil || finalHealth.State != "healthy" {
		return result, fmt.Errorf("validate activated Platform: state=%s: %w", finalHealth.State, err)
	}
	if forced {
		if diagnosticErr := preserveDiagnostics(s.DistRoot, "activate", platformName, sourceRoot, filepath.Join(origin, ".lock"), verification); diagnosticErr != nil {
			return result, diagnosticErr
		}
	}
	j.Phase = "verified"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := os.Remove(jPath); err != nil {
		return result, err
	}
	committed = true
	_ = os.RemoveAll(j.HostUserBackup)
	_ = os.RemoveAll(j.HostExtBackup)
	_ = os.RemoveAll(j.OriginBackup)
	result.Current = originName
	return result, nil
}

func (s Service) Deactivate(ctx context.Context, platformName string, options DeactivateOptions) (result DeactivateResult, err error) {
	result.Platform = platformName
	paths, err := s.hostPaths(platformName)
	if err != nil {
		return result, err
	}
	if err := s.recoverInterrupted(platformName); err != nil {
		return result, err
	}
	health, err := s.CheckHealth(platformName, paths)
	if err != nil {
		return result, err
	}
	if health.State == "inactive" {
		return result, fmt.Errorf("platform is not active: %s", platformName)
	}
	if options.ForceEmpty {
		return s.deactivateEmpty(ctx, platformName, paths)
	}
	if health.State != "healthy" && !options.Force {
		return result, unhealthyError(health)
	}
	if s.Runtime == nil || s.Stopper == nil {
		return result, fmt.Errorf("CodeVenv service is not configured")
	}
	current := filepath.Join(s.DistRoot, "current."+platformName)
	currentTarget, exists, err := linkTarget(current)
	if err != nil || !exists {
		return result, fmt.Errorf("trusted current Selection is unavailable; forced recovery is unsafe")
	}
	origin := filepath.Join(s.DistRoot, "origin."+platformName)
	prepared, err := recovery.Prepare(filepath.Join(origin, ".lock"))
	if err != nil {
		return result, fmt.Errorf("load trusted origin: %w", err)
	}
	unlock, err := acquireLifecycleLock(s.DistRoot, platformName)
	if err != nil {
		return result, err
	}
	defer unlock()
	if err := s.Stopper.StopForSelection(ctx, platformName, filepath.Join(currentTarget, ".data")); err != nil {
		return result, err
	}
	suffix := fmt.Sprintf(".ctk-%d-%d", time.Now().UnixNano(), os.Getpid())
	j := journal{Operation: "deactivate", Platform: platformName, Phase: "prepared", Current: current, CurrentTarget: currentTarget, Origin: origin, HostUser: paths.User, HostExtensions: paths.Extensions}
	jPath := journalPath(s.DistRoot, platformName)
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := rollbackJournal(j); rollbackErr != nil {
			err = errors.Join(err, rollbackErr)
			return
		}
		_ = os.Remove(jPath)
	}()
	j.HostUserBackup, err = plannedBackup(paths.User, suffix)
	if err != nil {
		return result, err
	}
	j.HostExtBackup, err = plannedBackup(paths.Extensions, suffix)
	if err != nil {
		return result, err
	}
	j.Phase = "host-backups-planned"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := executeBackup(paths.User, j.HostUserBackup); err != nil {
		return result, err
	}
	if err := executeBackup(paths.Extensions, j.HostExtBackup); err != nil {
		return result, err
	}
	j.Phase = "host-links-removed"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := copyTree(filepath.Join(origin, ".data", "User"), paths.User); err != nil {
		return result, fmt.Errorf("restore Host User: %w", err)
	}
	if err := os.MkdirAll(paths.Extensions, 0o755); err != nil {
		return result, fmt.Errorf("create Host Extensions: %w", err)
	}
	j.Phase = "host-physical"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	hostRuntime, err := s.Runtime(platformName, paths.UserData, paths.Extensions)
	if err != nil {
		return result, err
	}
	workspace, err := os.MkdirTemp(s.DistRoot, ".deactivate-"+platformName+"-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(workspace)
	recovered, err := s.Recovery.RecoverAt(ctx, prepared, workspace, paths.Extensions, hostRuntime)
	if err != nil {
		return result, err
	}
	verification := combineVerification(recovered.Before, recovered.After)
	forced, err := s.approve(verification, options.Force)
	if err != nil {
		return result, err
	}
	result.Forced = forced || health.State != "healthy"
	if result.Forced {
		if err := preserveDiagnostics(s.DistRoot, "deactivate", platformName, filepath.Join(origin, ".lock"), filepath.Join(workspace, ".lock"), verification); err != nil {
			return result, err
		}
	}
	if err := os.Remove(current); err != nil {
		return result, fmt.Errorf("remove current Selection: %w", err)
	}
	j.Phase = "verified"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := os.Remove(jPath); err != nil {
		return result, err
	}
	committed = true
	_ = os.RemoveAll(j.HostUserBackup)
	_ = os.RemoveAll(j.HostExtBackup)
	return result, nil
}

func (s Service) deactivateEmpty(ctx context.Context, platformName string, paths HostPaths) (result DeactivateResult, err error) {
	result = DeactivateResult{Platform: platformName, Forced: true, Empty: true}
	if s.Stopper == nil {
		return result, fmt.Errorf("CodeVenv service is not configured")
	}
	current := filepath.Join(s.DistRoot, "current."+platformName)
	currentTarget, currentExists, err := linkTarget(current)
	if err != nil {
		return result, fmt.Errorf("inspect current Selection: %w", err)
	}
	unlock, err := acquireLifecycleLock(s.DistRoot, platformName)
	if err != nil {
		return result, err
	}
	defer unlock()
	var stopPaths []string
	if currentExists {
		stopPaths = append(stopPaths, filepath.Join(currentTarget, ".data"))
	}
	if err := s.Stopper.StopForSelection(ctx, platformName, stopPaths...); err != nil {
		return result, err
	}
	suffix := fmt.Sprintf(".ctk-%d-%d", time.Now().UnixNano(), os.Getpid())
	j := journal{
		Operation: "deactivate", Platform: platformName, Phase: "prepared",
		Current: current, CurrentTarget: currentTarget, Origin: filepath.Join(s.DistRoot, "origin."+platformName),
		HostUser: paths.User, HostExtensions: paths.Extensions,
	}
	jPath := journalPath(s.DistRoot, platformName)
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := rollbackJournal(j); rollbackErr != nil {
			err = errors.Join(err, rollbackErr)
			return
		}
		_ = os.Remove(jPath)
	}()
	j.HostUserBackup, err = plannedBackup(paths.User, suffix)
	if err != nil {
		return result, err
	}
	j.HostExtBackup, err = plannedBackup(paths.Extensions, suffix)
	if err != nil {
		return result, err
	}
	j.Phase = "host-backups-planned"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := executeBackup(paths.User, j.HostUserBackup); err != nil {
		return result, err
	}
	if err := executeBackup(paths.Extensions, j.HostExtBackup); err != nil {
		return result, err
	}
	j.Phase = "host-links-removed"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := os.MkdirAll(paths.User, 0o755); err != nil {
		return result, fmt.Errorf("create empty Host User: %w", err)
	}
	if err := os.MkdirAll(paths.Extensions, 0o755); err != nil {
		return result, fmt.Errorf("create empty Host Extensions: %w", err)
	}
	j.Phase = "host-physical"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if currentExists {
		if err := os.Remove(current); err != nil {
			return result, fmt.Errorf("remove current Selection: %w", err)
		}
	}
	j.Phase = "verified"
	if err := writeJournal(jPath, j); err != nil {
		return result, err
	}
	if err := os.Remove(jPath); err != nil {
		return result, err
	}
	committed = true
	_ = os.RemoveAll(j.HostUserBackup)
	_ = os.RemoveAll(j.HostExtBackup)
	return result, nil
}

func observedRecipe(snapshot runtimelock.Snapshot) (recipe.Recipe, cookbook.Plan) {
	marketplace := true
	definition := recipe.Recipe{Name: snapshot.RecipeName, OS: hostOSName(), Platform: snapshot.Platform, Config: recipe.Config{DistStrategy: recipe.DistStrategy{ExtensionMarketplace: &marketplace, LockMode: "refresh", DefaultProfile: recipe.DefaultProfile{Extensions: "runtime"}}, ProfileStrategy: map[string]recipe.ProfileStrategy{}}}
	plan := cookbook.Plan{Name: snapshot.RecipeName, OS: definition.OS, Platform: snapshot.Platform, ExtensionMarketplace: true, LockMode: "refresh"}
	profiles := append([]runtimelock.ScopeSnapshot(nil), snapshot.Profiles...)
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	for _, profile := range profiles {
		definition.Profile = append(definition.Profile, profile.Name)
		strategy := recipe.ProfileStrategy{Settings: ownership(profile.Inheritance.Settings), Keybindings: ownership(profile.Inheritance.Keybindings), Tasks: ownership(profile.Inheritance.Tasks), MCP: ownership(profile.Inheritance.MCP), Snippets: ownership(profile.Inheritance.Snippets)}
		definition.Config.ProfileStrategy[profile.Name] = strategy
		plan.Profiles = append(plan.Profiles, cookbook.ScopePlan{Name: profile.Name, Inheritance: profile.Inheritance})
	}
	return definition, plan
}

func ownership(inherited bool) string {
	if inherited {
		return "default"
	}
	return "profile"
}

func writeRecipe(path string, definition recipe.Recipe) error {
	data, err := yaml.Marshal(definition)
	if err != nil {
		return fmt.Errorf("encode temporary Recipe: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write temporary Recipe: %w", err)
	}
	return nil
}

func combineVerification(values ...recovery.Verification) recovery.Verification {
	result := recovery.Verification{}
	for _, value := range values {
		result.Differences = append(result.Differences, value.Differences...)
	}
	return result
}

func (s Service) approve(verification recovery.Verification, force bool) (bool, error) {
	if verification.Matches() {
		return false, nil
	}
	if force {
		return true, nil
	}
	if s.SafetyGate == nil {
		return false, fmt.Errorf("semantic verification found %d difference(s); rerun with --force", len(verification.Differences))
	}
	approved, err := s.SafetyGate(verification)
	if err != nil {
		return false, err
	}
	if !approved {
		return false, fmt.Errorf("activation safety gate declined")
	}
	return true, nil
}

func unhealthyError(health Health) error {
	return fmt.Errorf("unhealthy %s activation: current=%q Host User=%s Host Extensions=%s (run: ctk-go deactivate %s --force)", health.Platform, health.Current, health.UserState, health.ExtensionState, health.Platform)
}

func linkTarget(path string) (string, bool, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	target, err := os.Readlink(path)
	if err != nil {
		return "", false, fmt.Errorf("expected managed link: %s: %w", path, err)
	}
	return absoluteLinkTarget(path, target), true, nil
}

func pathState(path, expected string) string {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "missing"
	}
	if err != nil {
		return "unreadable"
	}
	target, err := os.Readlink(path)
	if err != nil {
		return "physical"
	}
	if expected == "" {
		return "link"
	}
	if samePath(absoluteLinkTarget(path, target), expected) {
		return "linked"
	}
	return "unexpected-link"
}

func createManagedLink(target, link string) error {
	if _, err := os.Lstat(link); err == nil {
		return fmt.Errorf("path already exists: %s", link)
	} else if !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("managed link target is unavailable: %s: %w", target, err)
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	return createSelectionLink(target, link)
}

func acquireLifecycleLock(root, platformName string) (func(), error) {
	path := filepath.Join(root, ".codevenv."+platformName+".lock")
	return acquireProcessLock(path, "CodeVenv lifecycle already in progress for platform: "+platformName)
}

func (s Service) recoverInterrupted(platformName string) error {
	path := journalPath(s.DistRoot, platformName)
	j, err := readJournal(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unfinished CodeVenv transaction requires manual inspection: %w", err)
	}
	if j.Platform != platformName {
		return fmt.Errorf("CodeVenv journal Platform mismatch: %s", path)
	}
	paths, err := s.hostPaths(platformName)
	if err != nil {
		return err
	}
	if err := validateJournal(s.DistRoot, paths, j); err != nil {
		return fmt.Errorf("CodeVenv journal is unsafe to recover: %w", err)
	}
	if j.Phase == "verified" {
		_ = os.RemoveAll(j.HostUserBackup)
		_ = os.RemoveAll(j.HostExtBackup)
		_ = os.RemoveAll(j.OriginBackup)
		return os.Remove(path)
	}
	if err := rollbackJournal(j); err != nil {
		return fmt.Errorf("recover unfinished CodeVenv transaction %s: %w", path, err)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("complete transaction recovery: %w", err)
	}
	return nil
}

func rollbackJournal(j journal) error {
	var result error
	restore := func(path, backup string, removeWhenAbsent bool) {
		if backup == "" {
			if removeWhenAbsent {
				if err := os.RemoveAll(path); err != nil {
					result = errors.Join(result, err)
				}
			}
			return
		}
		if _, err := os.Lstat(backup); err != nil {
			if !os.IsNotExist(err) {
				result = errors.Join(result, err)
			}
			return
		}
		if err := os.RemoveAll(path); err != nil {
			result = errors.Join(result, err)
			return
		}
		if err := os.Rename(backup, path); err != nil {
			result = errors.Join(result, err)
		}
	}
	hostTouched := j.Phase != "prepared" && j.Phase != "origin-backup-planned" && j.Phase != "origin-published"
	restore(j.HostUser, j.HostUserBackup, hostTouched)
	restore(j.HostExtensions, j.HostExtBackup, hostTouched)
	if j.Operation == "activate" {
		_ = os.Remove(j.Current)
		if j.OriginBackup != "" {
			restore(j.Origin, j.OriginBackup, true)
		} else if j.Phase != "prepared" {
			if err := os.RemoveAll(j.Origin); err != nil {
				result = errors.Join(result, err)
			}
		}
	} else if j.Operation == "deactivate" && j.CurrentTarget != "" {
		if _, err := os.Lstat(j.Current); os.IsNotExist(err) {
			if linkErr := createManagedLink(j.CurrentTarget, j.Current); linkErr != nil {
				result = errors.Join(result, linkErr)
			}
		} else if err != nil {
			result = errors.Join(result, err)
		}
	}
	return result
}

func validateJournal(root string, paths HostPaths, j journal) error {
	expectedCurrent := filepath.Join(root, "current."+j.Platform)
	expectedOrigin := filepath.Join(root, "origin."+j.Platform)
	for label, pair := range map[string][2]string{
		"current": {j.Current, expectedCurrent}, "origin": {j.Origin, expectedOrigin},
		"Host User": {j.HostUser, paths.User}, "Host Extensions": {j.HostExtensions, paths.Extensions},
	} {
		if filepath.Clean(pair[0]) != filepath.Clean(pair[1]) {
			return fmt.Errorf("%s path mismatch", label)
		}
	}
	for label, pair := range map[string][2]string{
		"origin backup": {j.OriginBackup, j.Origin}, "Host User backup": {j.HostUserBackup, j.HostUser}, "Host Extensions backup": {j.HostExtBackup, j.HostExtensions},
	} {
		if pair[0] != "" && !strings.HasPrefix(filepath.Clean(pair[0]), filepath.Clean(pair[1])+".ctk-") {
			return fmt.Errorf("%s path is outside its managed boundary", label)
		}
	}
	if j.CurrentTarget != "" {
		relative, err := filepath.Rel(root, j.CurrentTarget)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("current target is outside Distribution root")
		}
	}
	return nil
}

func preserveDiagnostics(root, operation, platformName, sourceLock, freshLock string, verification recovery.Verification) error {
	directory := filepath.Join(root, ".diagnostics", fmt.Sprintf("%s-%s-%d", operation, platformName, time.Now().UnixNano()))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	if err := copyTree(sourceLock, filepath.Join(directory, "source-lock")); err != nil {
		return err
	}
	if err := copyTree(freshLock, filepath.Join(directory, "fresh-lock")); err != nil {
		return err
	}
	data, err := json.MarshalIndent(verification, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "verification.json"), append(data, '\n'), 0o644)
}
