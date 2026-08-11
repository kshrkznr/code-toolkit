package converge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
)

type ArtifactResolver interface {
	ResolveCandidates(platform, extensionID string) ([]ArtifactCandidate, error)
}

type ArtifactCandidate struct {
	Path       string
	Repository string
	Primary    bool
}

type MarketplacePolicy struct {
	Platform string
	Allowed  bool
	Pool     ArtifactResolver
	Force    bool
}

func Extensions(ctx context.Context, runtime runtimeio.Runtime, scope runtimeio.Scope, desired []string, policy MarketplacePolicy, report *Report) {
	installed, err := runtime.Extensions(ctx, scope)
	if err != nil {
		report.Add(Operation{Scope: scope.Name, Action: "observe extension", Status: Failed, Err: err})
		return
	}
	installedByID := map[string]runtimeio.Extension{}
	installedFolded := map[string]string{}
	for _, extension := range installed {
		installedByID[extension.ID] = extension
		installedFolded[strings.ToLower(extension.ID)] = extension.ID
	}
	desiredSet := map[string]struct{}{}
	desiredFolded := map[string]string{}
	conflicts := map[string]bool{}
	for _, id := range desired {
		desiredSet[id] = struct{}{}
		folded := strings.ToLower(id)
		if other, ok := desiredFolded[folded]; ok && other != id {
			conflicts[id], conflicts[other] = true, true
			report.Add(Operation{Scope: scope.Name, Action: "extension case conflict", Subject: id + " != " + other, Status: Failed, Err: fmt.Errorf("correct the Ingredient extension IDs")})
		} else {
			desiredFolded[folded] = id
		}
		if actual, ok := installedFolded[folded]; ok && actual != id {
			conflicts[id], conflicts[actual] = true, true
			report.Add(Operation{Scope: scope.Name, Action: "extension case conflict", Subject: id + " != " + actual, Status: Failed, Err: fmt.Errorf("correct the Ingredient extension ID")})
		}
	}
	for _, id := range sortedKeys(desiredSet) {
		if conflicts[id] {
			continue
		}
		if _, ok := installedByID[id]; ok {
			report.Add(Operation{Scope: scope.Name, Action: "retain extension", Subject: id, Status: Completed})
			continue
		}
		var primary, secondary string
		if policy.Pool != nil {
			candidates, err := policy.Pool.ResolveCandidates(policy.Platform, id)
			if err != nil {
				report.Add(Operation{Scope: scope.Name, Action: "resolve extension", Subject: id, Status: Failed, Err: err})
				continue
			}
			for _, candidate := range candidates {
				if candidate.Primary && primary == "" {
					primary = candidate.Path
				} else if !candidate.Primary && secondary == "" {
					secondary = candidate.Path
				}
			}
			if primary == "" && secondary == "" && !policy.Allowed {
				report.Add(Operation{Scope: scope.Name, Action: "resolve extension", Subject: id, Status: Unresolved, Err: fmt.Errorf("Marketplace disabled and Pool artifact unavailable")})
				continue
			}
		} else if !policy.Allowed {
			report.Add(Operation{Scope: scope.Name, Action: "resolve extension", Subject: id, Status: Unresolved, Err: fmt.Errorf("Marketplace disabled and no Pool configured")})
			continue
		}
		var installErr error
		switch {
		case primary != "":
			installErr = runtime.InstallExtension(ctx, scope, primary)
		case policy.Allowed:
			installErr = runtime.InstallExtension(ctx, scope, id)
			if installErr != nil && secondary != "" {
				marketplaceErr := installErr
				installErr = runtime.InstallExtension(ctx, scope, secondary)
				if installErr != nil {
					installErr = errors.Join(fmt.Errorf("Platform Repository: %w", marketplaceErr), fmt.Errorf("secondary Pool artifact: %w", installErr))
				}
			}
		case secondary != "":
			installErr = runtime.InstallExtension(ctx, scope, secondary)
		}
		if installErr == nil {
			report.Add(Operation{Scope: scope.Name, Action: "install extension", Subject: id, Status: Completed})
		} else {
			status := Failed
			if policy.Force {
				status = Unresolved
			}
			report.Add(Operation{Scope: scope.Name, Action: "install extension", Subject: id, Status: status, Err: installErr})
		}
	}
	installedIDs := make([]string, 0, len(installedByID))
	for id := range installedByID {
		installedIDs = append(installedIDs, id)
	}
	sort.Strings(installedIDs)
	for _, id := range installedIDs {
		if conflicts[id] {
			continue
		}
		if _, ok := desiredSet[id]; ok {
			continue
		}
		if err := runtime.UninstallExtension(ctx, scope, id); err != nil {
			report.Add(Operation{Scope: scope.Name, Action: "uninstall extension", Subject: id, Status: Failed, Err: err})
		} else {
			report.Add(Operation{Scope: scope.Name, Action: "uninstall extension", Subject: id, Status: Completed})
		}
	}
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

type Pool struct{ Root string }

func (p Pool) ResolveCandidates(platform, extensionID string) ([]ArtifactCandidate, error) {
	repositories := platformRepositories(platform)
	var candidates []ArtifactCandidate
	for index, repository := range repositories {
		matches, err := poolArtifacts(filepath.Join(p.Root, repository), extensionID)
		if err != nil {
			return nil, fmt.Errorf("search Extension Pool: %w", err)
		}
		sort.Strings(matches)
		if len(matches) > 0 {
			candidates = append(candidates, ArtifactCandidate{Path: matches[len(matches)-1], Repository: repository, Primary: index == 0})
		}
	}
	return candidates, nil
}

func poolArtifactName(extension runtimeio.Extension) string {
	return strings.ToLower(extension.ID) + "-" + extension.Version + ".vsix"
}

// poolArtifacts reads legacy mixed-case names while all new Pool writes use a
// lower-case Extension ID as their storage key.
func poolArtifacts(directory, extensionID string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	prefix := strings.ToLower(extensionID) + "-"
	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(strings.ToLower(name), prefix) || !strings.HasSuffix(strings.ToLower(name), ".vsix") {
			continue
		}
		matches = append(matches, filepath.Join(directory, name))
	}
	sort.Strings(matches)
	return matches, nil
}

func platformRepositories(platform string) []string {
	switch platform {
	case "codium":
		return []string{"open-vsx", "visual-studio-marketplace"}
	case "kiro":
		return []string{"open-vsx", "visual-studio-marketplace"}
	case "cursor":
		return []string{"cursor-marketplace", "visual-studio-marketplace"}
	case "devin-desktop":
		return []string{"windsurf-marketplace", "visual-studio-marketplace"}
	default:
		return []string{"visual-studio-marketplace"}
	}
}

// platformDownloadRepositories may be narrower than the local Pool search.
// VSCodium can try an already-present secondary VSIX, but CTK must not acquire
// a new Visual Studio Marketplace artifact on its behalf.
func platformDownloadRepositories(platform string) []string {
	if platform == "codium" {
		return []string{"open-vsx"}
	}
	return platformRepositories(platform)
}
