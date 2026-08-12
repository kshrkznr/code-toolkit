package codevenv

import (
	"fmt"
	"os"
	"runtime"

	ctkplatform "github.com/kshrkznr/code-toolkit/go/internal/platform"
)

func SupportedPlatforms() []string {
	return ctkplatform.Identities()
}

type HostPaths = ctkplatform.HostPaths

func ResolveHostPaths(platformName string) (HostPaths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return HostPaths{}, fmt.Errorf("resolve home directory: %w", err)
	}
	return ctkplatform.ResolveHostPaths(platformName, runtime.GOOS, home)
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
