package codevenv

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestSupportedPlatformsIncludesCursor(t *testing.T) {
	if !slices.Contains(SupportedPlatforms(), "cursor") {
		t.Fatal("SupportedPlatforms() does not include cursor")
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
