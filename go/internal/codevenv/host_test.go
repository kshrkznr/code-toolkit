package codevenv

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestSupportedPlatformsIncludesObservedAdapters(t *testing.T) {
	for _, platform := range []string{"codium", "cursor", "devin-desktop"} {
		if !slices.Contains(SupportedPlatforms(), platform) {
			t.Fatalf("SupportedPlatforms() does not include %s", platform)
		}
	}
}

func TestResolveVSCodiumHostPaths(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("CodeVenv host integration is only available on macOS and Windows")
	}

	paths, err := ResolveHostPaths("codium")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(paths.UserData), "/VSCodium") {
		t.Fatalf("UserData = %q, want VSCodium data directory", paths.UserData)
	}
	if paths.User != filepath.Join(paths.UserData, "User") {
		t.Fatalf("User = %q, want User under UserData", paths.User)
	}
	if !strings.HasSuffix(filepath.ToSlash(paths.Extensions), "/.vscode-oss/extensions") {
		t.Fatalf("Extensions = %q, want .vscode-oss/extensions", paths.Extensions)
	}
}

func TestResolveCursorHostPaths(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("CodeVenv host integration is only available on macOS and Windows")
	}

	paths, err := ResolveHostPaths("cursor")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(paths.UserData), "/Cursor") {
		t.Fatalf("UserData = %q, want Cursor data directory", paths.UserData)
	}
	if paths.User != filepath.Join(paths.UserData, "User") {
		t.Fatalf("User = %q, want User under UserData", paths.User)
	}
	if !strings.HasSuffix(filepath.ToSlash(paths.Extensions), "/.cursor/extensions") {
		t.Fatalf("Extensions = %q, want .cursor/extensions", paths.Extensions)
	}
}

func TestResolveDevinDesktopHostPaths(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("CodeVenv host integration is only available on macOS and Windows")
	}

	paths, err := ResolveHostPaths("devin-desktop")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(paths.UserData), "/Devin") {
		t.Fatalf("UserData = %q, want Devin data directory", paths.UserData)
	}
	if paths.User != filepath.Join(paths.UserData, "User") {
		t.Fatalf("User = %q, want User under UserData", paths.User)
	}
	if !strings.HasSuffix(filepath.ToSlash(paths.Extensions), "/.devin/extensions") {
		t.Fatalf("Extensions = %q, want .devin/extensions", paths.Extensions)
	}
}
