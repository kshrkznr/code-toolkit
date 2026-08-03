package codevenv

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Current(distDir, platform string) (map[string]string, error) {
	if platform != "" {
		value, err := currentPlatform(distDir, platform)
		if err != nil {
			return nil, err
		}
		return map[string]string{platform: value}, nil
	}

	matches, err := filepath.Glob(filepath.Join(distDir, "current.*"))
	if err != nil {
		return nil, fmt.Errorf("find current platforms: %w", err)
	}
	result := make(map[string]string)
	for _, match := range matches {
		target, exists, err := linkTarget(match)
		if err != nil {
			return nil, fmt.Errorf("inspect current platform %s: %w", match, err)
		}
		if !exists {
			continue
		}
		name := strings.TrimPrefix(filepath.Base(match), "current.")
		if name == "" {
			continue
		}
		result[name] = filepath.Base(filepath.Clean(target))
	}
	return result, nil
}

func Platforms(selections map[string]string) []string {
	names := make([]string, 0, len(selections))
	for name := range selections {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func currentPlatform(distDir, platform string) (string, error) {
	link := filepath.Join(distDir, "current."+platform)
	target, err := os.Readlink(link)
	if os.IsNotExist(err) {
		return "none", nil
	}
	if err != nil {
		return "", fmt.Errorf("read current platform %s: %w", platform, err)
	}
	return filepath.Base(filepath.Clean(target)), nil
}
