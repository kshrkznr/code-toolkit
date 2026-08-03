package mergerules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"go.yaml.in/yaml/v3"
)

type File struct {
	FormatVersion int `yaml:"format-version"`
	Settings      struct {
		Union [][]string `yaml:"union"`
	} `yaml:"settings"`
}

type Rules struct{ Union map[string]bool }

func Load(cookbookRoot string) (Rules, error) {
	path := filepath.Join(cookbookRoot, "kitchen-notes", "go.merge-rules.yaml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Rules{Union: map[string]bool{}}, nil
	}
	if err != nil {
		return Rules{}, fmt.Errorf("read Merge Rules: %w", err)
	}
	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return Rules{}, fmt.Errorf("parse Merge Rules: %w", err)
	}
	if file.FormatVersion != 1 {
		return Rules{}, fmt.Errorf("unsupported Merge Rules format-version %d", file.FormatVersion)
	}
	rules := Rules{Union: map[string]bool{}}
	for _, path := range file.Settings.Union {
		if len(path) == 0 {
			return Rules{}, fmt.Errorf("empty Merge Rule path")
		}
		rules.Union[key(path)] = true
	}
	return rules, nil
}

func Add(data []byte, paths [][]string) ([]byte, error) {
	var file File
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &file); err != nil {
			return nil, err
		}
		if file.FormatVersion != 1 {
			return nil, fmt.Errorf("unsupported Merge Rules format-version %d", file.FormatVersion)
		}
	} else {
		file.FormatVersion = 1
	}
	seen := map[string][]string{}
	for _, path := range append(file.Settings.Union, paths...) {
		if len(path) > 0 {
			seen[key(path)] = path
		}
	}
	keys := make([]string, 0, len(seen))
	for value := range seen {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	file.Settings.Union = nil
	for _, value := range keys {
		file.Settings.Union = append(file.Settings.Union, seen[value])
	}
	return yaml.Marshal(file)
}

func key(path []string) string {
	var result string
	for _, value := range path {
		result += fmt.Sprintf("%d:%s", len(value), value)
	}
	return result
}
func Key(path []string) string { return key(path) }
