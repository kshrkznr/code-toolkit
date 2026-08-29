package docbundle

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ExportResult struct {
	Path          string
	ContentSHA256 string
	Documents     int
}

func (bundle *Bundle) Export(target string) (ExportResult, error) {
	absolute, err := filepath.Abs(target)
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolve Documentation Export target: %w", err)
	}
	parent := filepath.Dir(absolute)
	parentInfo, err := os.Lstat(parent)
	if err != nil {
		return ExportResult{}, fmt.Errorf("inspect Documentation Export parent: %w", err)
	}
	if !parentInfo.IsDir() || parentInfo.Mode()&os.ModeSymlink != 0 {
		return ExportResult{}, fmt.Errorf("Documentation Export parent must be a directory without symlink substitution: %s", parent)
	}

	targetExisted, err := validateExportTarget(absolute)
	if err != nil {
		return ExportResult{}, err
	}
	staging, err := os.MkdirTemp(parent, "."+filepath.Base(absolute)+".ctk-export-")
	if err != nil {
		return ExportResult{}, fmt.Errorf("create Documentation Export staging directory: %w", err)
	}
	stagingPublished := false
	defer func() {
		if !stagingPublished {
			_ = os.RemoveAll(staging)
		}
	}()

	entries := bundle.exportEntries()
	if err := writeExportTree(staging, entries); err != nil {
		return ExportResult{}, err
	}
	if err := verifyExportTree(staging, entries); err != nil {
		return ExportResult{}, err
	}
	if err := os.Chmod(staging, 0o755); err != nil {
		return ExportResult{}, fmt.Errorf("set Documentation Export root mode: %w", err)
	}

	claimedAbsentTarget := false
	if !targetExisted {
		if err := os.Mkdir(absolute, 0o700); err != nil {
			return ExportResult{}, fmt.Errorf("claim absent Documentation Export target: %w", err)
		}
		claimedAbsentTarget = true
	}
	if err := publishExport(staging, absolute); err != nil {
		if claimedAbsentTarget {
			_ = os.Remove(absolute)
		}
		return ExportResult{}, err
	}
	stagingPublished = true
	return ExportResult{Path: absolute, ContentSHA256: bundle.manifest.ContentSHA256, Documents: len(bundle.manifest.Documents)}, nil
}

func (bundle *Bundle) exportEntries() map[string][]byte {
	entries := make(map[string][]byte, len(bundle.documents)+2)
	entries[ManifestPath] = bytes.Clone(bundle.manifestRaw)
	entries[bundle.manifest.Bootstrap.Path] = bytes.Clone(bundle.bootstrap)
	for documentPath, content := range bundle.documents {
		entries[documentPath] = bytes.Clone(content)
	}
	return entries
}

func validateExportTarget(target string) (bool, error) {
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect Documentation Export target: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("Documentation Export target must be absent or an empty directory without symlink substitution: %s", target)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return false, fmt.Errorf("inspect Documentation Export target contents: %w", err)
	}
	if len(entries) != 0 {
		return false, fmt.Errorf("Documentation Export target is not empty: %s", target)
	}
	return true, nil
}

func writeExportTree(root string, entries map[string][]byte) error {
	paths := sortedExportPaths(entries)
	for _, relative := range paths {
		target := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create Documentation Export directory for %s: %w", relative, err)
		}
		file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return fmt.Errorf("create Documentation Export file %s: %w", relative, err)
		}
		_, writeErr := file.Write(entries[relative])
		modeErr := file.Chmod(0o644)
		closeErr := file.Close()
		if writeErr != nil {
			return fmt.Errorf("write Documentation Export file %s: %w", relative, writeErr)
		}
		if modeErr != nil {
			return fmt.Errorf("set Documentation Export file mode %s: %w", relative, modeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close Documentation Export file %s: %w", relative, closeErr)
		}
	}
	directories := map[string]bool{}
	for relative := range entries {
		for directory := filepath.Dir(filepath.Join(root, filepath.FromSlash(relative))); directory != root; directory = filepath.Dir(directory) {
			directories[directory] = true
		}
	}
	for directory := range directories {
		if err := os.Chmod(directory, 0o755); err != nil {
			return fmt.Errorf("set Documentation Export directory mode %s: %w", directory, err)
		}
	}
	return nil
}

func verifyExportTree(root string, expected map[string][]byte) error {
	seen := map[string]bool{}
	casePaths := map[string]string{}
	expectedDirectories := map[string]bool{}
	for relative := range expected {
		for directory := filepath.ToSlash(filepath.Dir(relative)); directory != "."; directory = filepath.ToSlash(filepath.Dir(directory)) {
			expectedDirectories[directory] = true
		}
	}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root {
			return nil
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Documentation Export contains a symlink: %s", relative)
		}
		if entry.IsDir() {
			if !expectedDirectories[relative] {
				return fmt.Errorf("Documentation Export contains an unexpected directory: %s", relative)
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("Documentation Export contains a special file: %s", relative)
		}
		expectedContent, ok := expected[relative]
		if !ok {
			return fmt.Errorf("Documentation Export contains an unexpected file: %s", relative)
		}
		folded := strings.ToLower(relative)
		if previous, exists := casePaths[folded]; exists {
			return fmt.Errorf("case-insensitive Documentation Export collision: %s and %s", previous, relative)
		}
		content, err := os.ReadFile(current)
		if err != nil {
			return err
		}
		if !bytes.Equal(content, expectedContent) {
			return fmt.Errorf("Documentation Export content mismatch: %s", relative)
		}
		casePaths[folded] = relative
		seen[relative] = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("verify Documentation Export: %w", err)
	}
	for relative := range expected {
		if !seen[relative] {
			return fmt.Errorf("verify Documentation Export: missing file %s", relative)
		}
	}
	return nil
}

func publishExport(staging, target string) error {
	parent := filepath.Dir(target)
	reservation, err := os.MkdirTemp(parent, "."+filepath.Base(target)+".ctk-empty-")
	if err != nil {
		return fmt.Errorf("reserve Documentation Export publication path: %w", err)
	}
	if err := os.Remove(reservation); err != nil {
		return fmt.Errorf("prepare Documentation Export publication reservation: %w", err)
	}
	if err := os.Rename(target, reservation); err != nil {
		return fmt.Errorf("reserve Documentation Export target: %w", err)
	}
	restore := true
	defer func() {
		if restore {
			_ = os.Rename(reservation, target)
		}
	}()
	entries, err := os.ReadDir(reservation)
	if err != nil {
		return fmt.Errorf("verify reserved Documentation Export target: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("Documentation Export target changed while publishing: %s", target)
	}
	if err := os.Rename(staging, target); err != nil {
		return fmt.Errorf("publish Documentation Export: %w", err)
	}
	restore = false
	_ = os.Remove(reservation)
	return nil
}

func sortedExportPaths(entries map[string][]byte) []string {
	paths := make([]string, 0, len(entries))
	for relative := range entries {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	return paths
}
