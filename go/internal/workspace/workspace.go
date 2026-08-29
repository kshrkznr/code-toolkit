package workspace

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

const configPath = ".config/workspace.yaml"

// Paths resolves the locations owned or selected by one CTK Workspace.
type Paths struct {
	Root           string
	CookbookSource string
	Workbench      string
	Dist           string
	Archive        string
	Pool           string
}

type configuration struct {
	Paths pathConfiguration `yaml:"paths"`
}

type pathConfiguration struct {
	CookbookSource string `yaml:"cookbook-source"`
	Dist           string `yaml:"dist"`
}

// HasMarker reports whether path can be considered during Workspace discovery.
// Configuration is validated by Load after the nearest marker is selected.
func HasMarker(path string) bool {
	if isFile(filepath.Join(path, filepath.FromSlash(configPath))) {
		return true
	}
	return isDir(filepath.Join(path, "cookbook", "recipe")) && isDir(filepath.Join(path, "cookbook", "ingredient"))
}

// Load resolves optional location overrides and validates Cookbook Source before
// any lifecycle is allowed to mutate Host or Workspace state.
func Load(root string) (Paths, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve CTK Workspace: %w", err)
	}
	result := Paths{
		Root:           absoluteRoot,
		CookbookSource: filepath.Join(absoluteRoot, "cookbook"),
		Workbench:      filepath.Join(absoluteRoot, "cookbook"),
		Dist:           filepath.Join(absoluteRoot, "dist"),
		Archive:        filepath.Join(absoluteRoot, "archive"),
		Pool:           filepath.Join(absoluteRoot, ".vsix"),
	}

	path := filepath.Join(absoluteRoot, filepath.FromSlash(configPath))
	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return Paths{}, fmt.Errorf("read Workspace configuration %s: %w", path, err)
		}
	} else {
		defer file.Close()
		decoder := yaml.NewDecoder(file)
		decoder.KnownFields(true)
		var config configuration
		if err := decoder.Decode(&config); err != nil && !errors.Is(err, io.EOF) {
			return Paths{}, fmt.Errorf("parse Workspace configuration %s: %w", path, err)
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			if err == nil {
				err = fmt.Errorf("multiple YAML documents are not supported")
			}
			return Paths{}, fmt.Errorf("parse Workspace configuration %s: %w", path, err)
		}
		if config.Paths.CookbookSource != "" {
			result.CookbookSource = resolve(absoluteRoot, config.Paths.CookbookSource)
		}
		if config.Paths.Dist != "" {
			result.Dist = resolve(absoluteRoot, config.Paths.Dist)
		}
	}

	if !isDir(filepath.Join(result.CookbookSource, "recipe")) || !isDir(filepath.Join(result.CookbookSource, "ingredient")) {
		return Paths{}, fmt.Errorf("Cookbook Source must contain recipe and ingredient directories: %s", result.CookbookSource)
	}
	return result, nil
}

func resolve(root, path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return absolute
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
