package codevenv

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func SupportedPlatforms() []string {
	return []string{"code", "kiro"}
}

type HostPaths struct {
	UserData   string
	User       string
	Extensions string
}

func ResolveHostPaths(platformName string) (HostPaths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return HostPaths{}, fmt.Errorf("resolve home directory: %w", err)
	}
	var dataName, extensionName string
	switch platformName {
	case "code":
		dataName, extensionName = "Code", ".vscode"
	case "kiro":
		dataName, extensionName = "Kiro", ".kiro"
	default:
		return HostPaths{}, fmt.Errorf("host paths are not configured for platform: %s", platformName)
	}
	var data string
	switch runtime.GOOS {
	case "darwin":
		data = filepath.Join(home, "Library", "Application Support", dataName)
	case "windows":
		data = filepath.Join(home, "AppData", "Roaming", dataName)
	default:
		return HostPaths{}, fmt.Errorf("CodeVenv host integration is unsupported on %s", runtime.GOOS)
	}
	return HostPaths{UserData: data, User: filepath.Join(data, "User"), Extensions: filepath.Join(home, extensionName, "extensions")}, nil
}

func hostOSName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}
