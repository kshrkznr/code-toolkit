package docbundle

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type MatchStatus string

type DirtyStatus string

const (
	Match        MatchStatus = "match"
	Mismatch     MatchStatus = "mismatch"
	Unknown      MatchStatus = "unknown"
	Clean        DirtyStatus = "clean"
	Dirty        DirtyStatus = "dirty"
	DirtyUnknown DirtyStatus = "unknown"
)

type SourceStatus struct {
	Kind                     string
	Path                     string
	Version                  string
	Revision                 string
	Tag                      string
	Repository               string
	DefinitionSHA256         string
	ContentSHA256            string
	ComparisonContentSHA256  string
	PackagedRevision         string
	PackagedDefinitionSHA256 string
	PackagedContentSHA256    string
	RevisionMatch            MatchStatus
	DefinitionMatch          MatchStatus
	ContentMatch             MatchStatus
	SelectedPathDirty        DirtyStatus
	SelectedDirtyPaths       []string
	RepositoryDirty          DirtyStatus
}

type LocalSource struct {
	Bundle *Bundle
	Status SourceStatus
}

func PackagedSourceStatus(bundle *Bundle) SourceStatus {
	manifest := bundle.Manifest()
	return SourceStatus{
		Kind:             "packaged",
		Version:          manifest.Version,
		Revision:         manifest.Revision,
		Tag:              manifest.Tag,
		Repository:       manifest.Repository,
		DefinitionSHA256: manifest.DefinitionSHA256,
		ContentSHA256:    manifest.ContentSHA256,
	}
}

func OpenLocal(repositoryRoot string, packaged *Bundle) (LocalSource, error) {
	if packaged == nil {
		return LocalSource{}, fmt.Errorf("packaged Documentation Bundle is required for local source comparison")
	}
	absolute, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return LocalSource{}, fmt.Errorf("resolve local documentation source: %w", err)
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return LocalSource{}, fmt.Errorf("inspect local documentation source %s: %w", absolute, err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return LocalSource{}, fmt.Errorf("local documentation source must be a directory without symlink substitution: %s", absolute)
	}

	packagedManifest := packaged.Manifest()
	comparisonMetadata := Metadata{
		Version:  packagedManifest.Version,
		Revision: packagedManifest.Revision,
		Tag:      packagedManifest.Tag,
	}
	comparison, err := Generate(absolute, comparisonMetadata)
	if err != nil {
		return LocalSource{}, fmt.Errorf("generate local documentation comparison: %w", err)
	}

	revision, gitKnown := localGitRevision(absolute)
	localMetadata := Metadata{Version: packagedManifest.Version, Revision: revision}
	if gitKnown && revision == packagedManifest.Revision {
		localMetadata.Tag = packagedManifest.Tag
	}
	localResult := comparison
	if localMetadata != comparisonMetadata {
		localResult, err = Generate(absolute, localMetadata)
		if err != nil {
			return LocalSource{}, fmt.Errorf("generate local Documentation Bundle: %w", err)
		}
	}
	localBundle, err := Open(localResult.Archive)
	if err != nil {
		return LocalSource{}, fmt.Errorf("open local Documentation Bundle: %w", err)
	}

	status := SourceStatus{
		Kind:                     "local",
		Path:                     absolute,
		Version:                  localResult.Manifest.Version,
		Revision:                 revision,
		Tag:                      localResult.Manifest.Tag,
		Repository:               localResult.Manifest.Repository,
		DefinitionSHA256:         comparison.Manifest.DefinitionSHA256,
		ContentSHA256:            localResult.Manifest.ContentSHA256,
		ComparisonContentSHA256:  comparison.Manifest.ContentSHA256,
		PackagedRevision:         packagedManifest.Revision,
		PackagedDefinitionSHA256: packagedManifest.DefinitionSHA256,
		PackagedContentSHA256:    packagedManifest.ContentSHA256,
		DefinitionMatch:          matchStatus(comparison.Manifest.DefinitionSHA256 == packagedManifest.DefinitionSHA256),
		ContentMatch:             matchStatus(comparison.Manifest.ContentSHA256 == packagedManifest.ContentSHA256),
		RevisionMatch:            Unknown,
		SelectedPathDirty:        DirtyUnknown,
		RepositoryDirty:          DirtyUnknown,
	}
	if gitKnown {
		if packagedManifest.Revision != "" && packagedManifest.Revision != "unknown" {
			status.RevisionMatch = matchStatus(revision == packagedManifest.Revision)
		}
		status.SelectedDirtyPaths = selectedDirtyPaths(absolute, localResult.Documents)
		if len(status.SelectedDirtyPaths) == 0 {
			status.SelectedPathDirty = Clean
		} else {
			status.SelectedPathDirty = Dirty
		}
		if dirty, known := localGitDirty(absolute); known {
			status.RepositoryDirty = Clean
			if dirty {
				status.RepositoryDirty = Dirty
			}
		}
	}
	return LocalSource{Bundle: localBundle, Status: status}, nil
}

func matchStatus(matches bool) MatchStatus {
	if matches {
		return Match
	}
	return Mismatch
}

func localGitRevision(root string) (string, bool) {
	topLevelOutput, err := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", false
	}
	topLevel := strings.TrimSpace(string(topLevelOutput))
	rootInfo, rootErr := os.Stat(root)
	topLevelInfo, topLevelErr := os.Stat(topLevel)
	if rootErr != nil || topLevelErr != nil || !os.SameFile(rootInfo, topLevelInfo) {
		return "", false
	}
	output, err := exec.Command("git", "-C", root, "rev-parse", "--verify", "HEAD").Output()
	if err != nil {
		return "", false
	}
	revision := strings.TrimSpace(string(output))
	return revision, revision != ""
}

func selectedDirtyPaths(root string, documents map[string][]byte) []string {
	dirty := []string{}
	for _, documentPath := range sortedKeys(documents) {
		committed, err := exec.Command("git", "-C", root, "cat-file", "blob", "HEAD:"+documentPath).Output()
		if err != nil || !bytes.Equal(committed, documents[documentPath]) {
			dirty = append(dirty, documentPath)
		}
	}
	sort.Strings(dirty)
	return dirty
}

func localGitDirty(root string) (bool, bool) {
	output, err := exec.Command("git", "-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all").Output()
	if err != nil {
		return false, false
	}
	return len(output) > 0, true
}
