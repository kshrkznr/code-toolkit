package distribution

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"code-toolkit/internal/recipe"
)

type Distribution struct {
	Name   string
	Path   string
	Recipe recipe.Recipe
}

func List(distDir string) ([]string, error) {
	entries, err := os.ReadDir(distDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list distributions: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name()[0] != '.' && !strings.HasPrefix(entry.Name(), "current.") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func Recipe(distDir string) (recipe.Recipe, error) {
	return recipe.Load(filepath.Join(distDir, ".meta", "recipe.yaml"))
}

func Load(root, name string) (Distribution, error) {
	path, err := resolveDirectory(root, name)
	if err != nil {
		return Distribution{}, err
	}
	return loadRuntime(path)
}

// LoadForLaunch accepts a platform-appropriate Launch Override as a
// launch-only input. Without an Override it requires a complete Runtime
// Distribution through the same validation as Load.
func LoadForLaunch(root, name, overrideName string) (Distribution, error) {
	path, err := resolveDirectory(root, name)
	if err != nil {
		return Distribution{}, err
	}
	if overrideName != "" {
		if info, statErr := os.Stat(filepath.Join(path, overrideName)); statErr == nil && !info.IsDir() {
			return Distribution{Name: filepath.Base(path), Path: path}, nil
		}
	}
	return loadRuntime(path)
}

func resolveDirectory(root, name string) (string, error) {
	path, err := resolvePath(root, name)
	if err != nil {
		return "", fmt.Errorf("resolve distribution %s: %w", name, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("load distribution %s: %w", name, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("distribution is not a directory: %s", path)
	}
	return path, nil
}

func loadRuntime(path string) (Distribution, error) {
	metadata, err := Recipe(path)
	if err != nil {
		return Distribution{}, err
	}
	if metadata.Platform == "" {
		return Distribution{}, fmt.Errorf("recipe.platform missing: %s", filepath.Join(path, ".meta", "recipe.yaml"))
	}
	for _, required := range []string{".data", ".ext"} {
		value, err := os.Stat(filepath.Join(path, required))
		if err != nil || !value.IsDir() {
			return Distribution{}, fmt.Errorf("distribution runtime directory missing: %s", filepath.Join(path, required))
		}
	}
	return Distribution{Name: filepath.Base(path), Path: path, Recipe: metadata}, nil
}

// resolvePath resolves a bare value as a name below root. Relative values with
// path components remain relative to the working directory, and absolute paths
// are used as provided.
func resolvePath(root, value string) (string, error) {
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	clean := filepath.Clean(value)
	if filepath.Dir(clean) == "." {
		return filepath.Abs(filepath.Join(root, clean))
	}
	return filepath.Abs(clean)
}
