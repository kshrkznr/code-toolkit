// Package workspaceinit creates the optional filesystem footing for a CTK
// Workspace without selecting or persisting that Workspace.
package workspaceinit

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed assets
var assets embed.FS

type Options struct {
	IncludeSample bool
}

type Result struct {
	Path      string
	Created   []string
	Unchanged []string
}

// Initialize creates the minimum discoverable Workspace at target. Existing
// files with different content are rejected before any content is written.
func Initialize(target string, options Options) (Result, error) {
	if strings.TrimSpace(target) == "" {
		return Result{}, fmt.Errorf("Workspace path is required")
	}
	absolute, err := filepath.Abs(target)
	if err != nil {
		return Result{}, fmt.Errorf("resolve Workspace path %s: %w", target, err)
	}
	if info, err := os.Stat(absolute); err == nil && !info.IsDir() {
		return Result{}, fmt.Errorf("Workspace path is not a directory: %s", absolute)
	} else if err != nil && !os.IsNotExist(err) {
		return Result{}, fmt.Errorf("inspect Workspace path %s: %w", absolute, err)
	}

	result := Result{Path: absolute}
	files := map[string][]byte{}
	if options.IncludeSample {
		err := fs.WalkDir(assets, "assets", func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			content, err := assets.ReadFile(path)
			if err != nil {
				return err
			}
			relative := strings.TrimPrefix(filepath.ToSlash(path), "assets/")
			files[relative] = content
			return nil
		})
		if err != nil {
			return Result{}, fmt.Errorf("read embedded Workspace sample: %w", err)
		}
	}

	paths := make([]string, 0, len(files))
	for relative := range files {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	var conflicts []string
	for _, relative := range paths {
		path := filepath.Join(absolute, filepath.FromSlash(relative))
		content, err := os.ReadFile(path)
		switch {
		case err == nil && bytes.Equal(content, files[relative]):
			result.Unchanged = append(result.Unchanged, relative)
		case err == nil:
			conflicts = append(conflicts, relative)
		case os.IsNotExist(err):
		default:
			conflicts = append(conflicts, relative)
		}
	}
	if len(conflicts) > 0 {
		return Result{}, fmt.Errorf("Workspace initialization would overwrite existing content: %s", strings.Join(conflicts, ", "))
	}

	for _, relative := range []string{"cookbook/recipe", "cookbook/ingredient"} {
		if err := os.MkdirAll(filepath.Join(absolute, filepath.FromSlash(relative)), 0o755); err != nil {
			return Result{}, fmt.Errorf("create Workspace directory %s: %w", relative, err)
		}
	}
	for _, relative := range paths {
		path := filepath.Join(absolute, filepath.FromSlash(relative))
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Result{}, fmt.Errorf("create Workspace sample directory for %s: %w", relative, err)
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return Result{}, fmt.Errorf("write Workspace sample %s: %w", relative, err)
		}
		if _, err := file.Write(files[relative]); err != nil {
			file.Close()
			return Result{}, fmt.Errorf("write Workspace sample %s: %w", relative, err)
		}
		if err := file.Close(); err != nil {
			return Result{}, fmt.Errorf("write Workspace sample %s: %w", relative, err)
		}
		result.Created = append(result.Created, relative)
	}
	return result, nil
}
