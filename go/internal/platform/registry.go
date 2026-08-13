package platform

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/kshrkznr/code-toolkit/go/internal/repository"
)

type PathBase string

const (
	PathBaseHome                   PathBase = "home"
	PathBaseApplicationSupport     PathBase = "application-support"
	PathBaseRoamingApplicationData PathBase = "roaming-application-data"
)

const FilterSameNameRoot = "same-name-root"

var builtInFilters = map[string]struct{}{FilterSameNameRoot: {}}

type PathDefinition struct {
	Base PathBase
	Path string
}

type HostDefinition struct {
	UserData   PathDefinition
	User       string
	Extensions PathDefinition
}

type ProcessDefinition struct {
	Identities        []string
	AdditionalFilters []string
}

type OSDefinition struct {
	Host    HostDefinition
	Process ProcessDefinition
}

type PoolRepository struct {
	RepositoryID    string
	DownloadEnabled bool
}

type Definition struct {
	Identity         string
	Command          string
	OS               map[string]OSDefinition
	PoolRepositories []PoolRepository
}

type HostPaths struct {
	UserData   string
	User       string
	Extensions string
}

var builtIns, builtInOrder = mustRegistry([]Definition{
	definition("code", "code", "Code", ".vscode",
		[]string{"Visual Studio Code.app/Contents/MacOS/Code", "Visual Studio Code.app/Contents/MacOS/Electron"}, "Code.exe",
		pool("visual-studio-marketplace")),
	definition("codium", "codium", "VSCodium", ".vscode-oss",
		[]string{"VSCodium.app/Contents/MacOS/Electron", "VSCodium.app/Contents/MacOS/VSCodium"}, "VSCodium.exe",
		pool("open-vsx", "visual-studio-marketplace")),
	definition("kiro", "kiro", "Kiro", ".kiro",
		[]string{"Kiro.app/Contents/MacOS/Electron"}, "Kiro.exe",
		pool("open-vsx", "visual-studio-marketplace")),
	definition("cursor", "cursor", "Cursor", ".cursor",
		[]string{"Cursor.app/Contents/MacOS/Cursor"}, "Cursor.exe",
		pool("cursor-marketplace", "visual-studio-marketplace")),
	definition("devin-desktop", "devin-desktop", "Devin", ".devin",
		[]string{"Devin.app/Contents/MacOS/Devin"}, "Devin.exe",
		pool("windsurf-marketplace", "visual-studio-marketplace")),
})

func definition(identity, command, dataName, extensionName string, darwinProcess []string, windowsProcess string, repositories []PoolRepository) Definition {
	return Definition{
		Identity: identity,
		Command:  command,
		OS: map[string]OSDefinition{
			"darwin": {
				Host: HostDefinition{
					UserData:   PathDefinition{Base: PathBaseApplicationSupport, Path: dataName},
					User:       "User",
					Extensions: PathDefinition{Base: PathBaseHome, Path: path.Join(extensionName, "extensions")},
				},
				Process: ProcessDefinition{Identities: darwinProcess},
			},
			"windows": {
				Host: HostDefinition{
					UserData:   PathDefinition{Base: PathBaseRoamingApplicationData, Path: dataName},
					User:       "User",
					Extensions: PathDefinition{Base: PathBaseHome, Path: path.Join(extensionName, "extensions")},
				},
				Process: ProcessDefinition{Identities: []string{windowsProcess}, AdditionalFilters: []string{FilterSameNameRoot}},
			},
		},
		PoolRepositories: repositories,
	}
}

func pool(ids ...string) []PoolRepository {
	result := make([]PoolRepository, 0, len(ids))
	for _, id := range ids {
		result = append(result, PoolRepository{RepositoryID: id, DownloadEnabled: true})
	}
	return result
}

func mustRegistry(definitions []Definition) (map[string]Definition, []string) {
	registry := make(map[string]Definition, len(definitions))
	order := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		if err := validate(definition); err != nil {
			panic(err)
		}
		if _, exists := registry[definition.Identity]; exists {
			panic(fmt.Sprintf("duplicate Platform identity: %s", definition.Identity))
		}
		registry[definition.Identity] = definition
		order = append(order, definition.Identity)
	}
	return registry, order
}

func validate(definition Definition) error {
	if definition.Identity == "" || definition.Command == "" {
		return fmt.Errorf("Platform identity and command are required")
	}
	for _, goos := range []string{"darwin", "windows"} {
		osDefinition, ok := definition.OS[goos]
		if !ok {
			return fmt.Errorf("Platform %s has no %s definition", definition.Identity, goos)
		}
		if osDefinition.Host.UserData.Path == "" || osDefinition.Host.User == "" || osDefinition.Host.Extensions.Path == "" || len(osDefinition.Process.Identities) == 0 {
			return fmt.Errorf("Platform %s has an incomplete %s definition", definition.Identity, goos)
		}
		for _, filter := range osDefinition.Process.AdditionalFilters {
			if _, ok := builtInFilters[filter]; !ok {
				return fmt.Errorf("Platform %s references unknown process filter %q", definition.Identity, filter)
			}
		}
	}
	if len(definition.PoolRepositories) == 0 {
		return fmt.Errorf("Platform %s has no Extension Pool repositories", definition.Identity)
	}
	seenRepositories := map[string]struct{}{}
	for _, candidate := range definition.PoolRepositories {
		if _, exists := seenRepositories[candidate.RepositoryID]; exists {
			return fmt.Errorf("Platform %s contains duplicate Extension repository %q", definition.Identity, candidate.RepositoryID)
		}
		seenRepositories[candidate.RepositoryID] = struct{}{}
		repositoryDefinition, err := repository.Lookup(candidate.RepositoryID)
		if err != nil {
			return fmt.Errorf("Platform %s: %w", definition.Identity, err)
		}
		if candidate.DownloadEnabled && repositoryDefinition.Connector == "" {
			return fmt.Errorf("Platform %s enables download without a connector for Extension repository %q", definition.Identity, candidate.RepositoryID)
		}
	}
	return nil
}

func Identities() []string {
	return append([]string(nil), builtInOrder...)
}

func Lookup(identity string) (Definition, error) {
	definition, ok := builtIns[identity]
	if !ok {
		return Definition{}, fmt.Errorf("Platform is not configured: %s", identity)
	}
	return clone(definition), nil
}

func clone(definition Definition) Definition {
	result := definition
	result.OS = make(map[string]OSDefinition, len(definition.OS))
	for goos, osDefinition := range definition.OS {
		osDefinition.Process.Identities = append([]string(nil), osDefinition.Process.Identities...)
		osDefinition.Process.AdditionalFilters = append([]string(nil), osDefinition.Process.AdditionalFilters...)
		result.OS[goos] = osDefinition
	}
	result.PoolRepositories = append([]PoolRepository(nil), definition.PoolRepositories...)
	return result
}

func ResolveHostPaths(identity, goos, home string) (HostPaths, error) {
	definition, err := Lookup(identity)
	if err != nil {
		return HostPaths{}, fmt.Errorf("host paths are not configured for platform: %s", identity)
	}
	osDefinition, ok := definition.OS[goos]
	if !ok {
		return HostPaths{}, fmt.Errorf("CodeVenv host integration is unsupported on %s", goos)
	}
	userData, err := resolvePath(osDefinition.Host.UserData, goos, home)
	if err != nil {
		return HostPaths{}, err
	}
	extensions, err := resolvePath(osDefinition.Host.Extensions, goos, home)
	if err != nil {
		return HostPaths{}, err
	}
	return HostPaths{UserData: userData, User: joinPath(goos, userData, osDefinition.Host.User), Extensions: extensions}, nil
}

func resolvePath(definition PathDefinition, goos, home string) (string, error) {
	var base string
	switch definition.Base {
	case "":
		if !isAbsolutePath(goos, definition.Path) {
			return "", fmt.Errorf("Host path without a base must be absolute: %s", definition.Path)
		}
		return cleanPath(goos, definition.Path), nil
	case PathBaseHome:
		base = home
	case PathBaseApplicationSupport:
		if goos != "darwin" {
			return "", fmt.Errorf("path base %q is unsupported on %s", definition.Base, goos)
		}
		base = joinPath(goos, home, "Library", "Application Support")
	case PathBaseRoamingApplicationData:
		if goos != "windows" {
			return "", fmt.Errorf("path base %q is unsupported on %s", definition.Base, goos)
		}
		base = joinPath(goos, home, "AppData", "Roaming")
	default:
		return "", fmt.Errorf("unknown Host path base %q", definition.Base)
	}
	if isAbsolutePath(goos, definition.Path) {
		return "", fmt.Errorf("Host path with base %q must be relative: %s", definition.Base, definition.Path)
	}
	return joinPath(goos, base, definition.Path), nil
}

func cleanPath(goos, value string) string {
	if goos == "windows" {
		return filepath.Clean(value)
	}
	return path.Clean(value)
}

func joinPath(goos string, elements ...string) string {
	if goos == "windows" {
		return filepath.Join(elements...)
	}
	return path.Join(elements...)
}

func isAbsolutePath(goos, value string) bool {
	if goos == "windows" {
		return strings.HasPrefix(value, `\\`) || len(value) >= 3 && ((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) && value[1] == ':' && (value[2] == '\\' || value[2] == '/')
	}
	return strings.HasPrefix(value, "/")
}
