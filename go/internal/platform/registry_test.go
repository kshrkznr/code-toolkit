package platform

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuiltInRegistryIsComplete(t *testing.T) {
	wantIdentities := []string{"code", "codium", "kiro", "cursor", "devin-desktop"}
	if got := Identities(); !reflect.DeepEqual(got, wantIdentities) {
		t.Fatalf("Identities() = %v, want %v", got, wantIdentities)
	}

	want := map[string]struct {
		command      string
		data         string
		extensions   string
		darwin       []string
		windows      string
		repositories []string
	}{
		"code":          {"code", "Code", ".vscode", []string{"Visual Studio Code.app/Contents/MacOS/Code", "Visual Studio Code.app/Contents/MacOS/Electron"}, "Code.exe", []string{"visual-studio-marketplace"}},
		"codium":        {"codium", "VSCodium", ".vscode-oss", []string{"VSCodium.app/Contents/MacOS/Electron", "VSCodium.app/Contents/MacOS/VSCodium"}, "VSCodium.exe", []string{"open-vsx", "visual-studio-marketplace"}},
		"kiro":          {"kiro", "Kiro", ".kiro", []string{"Kiro.app/Contents/MacOS/Electron"}, "Kiro.exe", []string{"open-vsx", "visual-studio-marketplace"}},
		"cursor":        {"cursor", "Cursor", ".cursor", []string{"Cursor.app/Contents/MacOS/Cursor"}, "Cursor.exe", []string{"cursor-marketplace", "visual-studio-marketplace"}},
		"devin-desktop": {"devin-desktop", "Devin", ".devin", []string{"Devin.app/Contents/MacOS/Devin"}, "Devin.exe", []string{"windsurf-marketplace", "visual-studio-marketplace"}},
	}

	for _, identity := range wantIdentities {
		expected := want[identity]
		definition, err := Lookup(identity)
		if err != nil {
			t.Fatalf("Lookup(%q): %v", identity, err)
		}
		if definition.Command != expected.command {
			t.Errorf("%s command = %q, want %q", identity, definition.Command, expected.command)
		}
		if got := definition.OS["darwin"].Process.Identities; !reflect.DeepEqual(got, expected.darwin) {
			t.Errorf("%s darwin process identities = %v, want %v", identity, got, expected.darwin)
		}
		windows := definition.OS["windows"].Process
		if !reflect.DeepEqual(windows.Identities, []string{expected.windows}) || !reflect.DeepEqual(windows.AdditionalFilters, []string{FilterSameNameRoot}) {
			t.Errorf("%s windows process = %#v", identity, windows)
		}
		var repositories []string
		for _, candidate := range definition.PoolRepositories {
			if !candidate.DownloadEnabled {
				t.Errorf("%s repository %s unexpectedly disables existing download behavior", identity, candidate.RepositoryID)
			}
			repositories = append(repositories, candidate.RepositoryID)
		}
		if !reflect.DeepEqual(repositories, expected.repositories) {
			t.Errorf("%s repositories = %v, want %v", identity, repositories, expected.repositories)
		}

		for _, test := range []struct {
			goos     string
			dataBase string
		}{
			{"darwin", filepath.Join("home", "Library", "Application Support")},
			{"windows", filepath.Join("home", "AppData", "Roaming")},
		} {
			paths, err := ResolveHostPaths(identity, test.goos, "home")
			if err != nil {
				t.Errorf("ResolveHostPaths(%q, %q): %v", identity, test.goos, err)
				continue
			}
			if paths.UserData != filepath.Join(test.dataBase, expected.data) || paths.User != filepath.Join(test.dataBase, expected.data, "User") || paths.Extensions != filepath.Join("home", expected.extensions, "extensions") {
				t.Errorf("ResolveHostPaths(%q, %q) = %#v", identity, test.goos, paths)
			}
		}
	}
}

func TestLookupReturnsIndependentDefinition(t *testing.T) {
	definition, err := Lookup("cursor")
	if err != nil {
		t.Fatal(err)
	}
	definition.OS["windows"] = OSDefinition{}
	definition.PoolRepositories[0].RepositoryID = "changed"

	again, err := Lookup("cursor")
	if err != nil {
		t.Fatal(err)
	}
	if len(again.OS["windows"].Process.Identities) == 0 || again.PoolRepositories[0].RepositoryID != "cursor-marketplace" {
		t.Fatalf("Lookup returned mutable Registry storage: %#v", again)
	}
}

func TestLookupRejectsUnknownPlatform(t *testing.T) {
	if _, err := Lookup("unknown"); err == nil {
		t.Fatal("expected unknown Platform error")
	}
}

func TestValidateRejectsUnregisteredReferences(t *testing.T) {
	base, err := Lookup("code")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*Definition)
	}{
		{"unknown filter", func(definition *Definition) {
			windows := definition.OS["windows"]
			windows.Process.AdditionalFilters = []string{"unknown"}
			definition.OS["windows"] = windows
		}},
		{"unknown repository", func(definition *Definition) {
			definition.PoolRepositories = []PoolRepository{{RepositoryID: "unknown"}}
		}},
		{"duplicate repository", func(definition *Definition) {
			definition.PoolRepositories = []PoolRepository{{RepositoryID: "visual-studio-marketplace"}, {RepositoryID: "visual-studio-marketplace"}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := clone(base)
			test.mutate(&definition)
			if err := validate(definition); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestResolvePathAcceptsOnlyDeclaredForms(t *testing.T) {
	if got, err := resolvePath(PathDefinition{Path: "/opt/editor/data"}, "darwin", "/home/user"); err != nil || got != "/opt/editor/data" {
		t.Fatalf("absolute path = %q, %v", got, err)
	}
	if _, err := resolvePath(PathDefinition{Path: "relative/data"}, "darwin", "/home/user"); err == nil {
		t.Fatal("expected relative path without base to fail")
	}
	if _, err := resolvePath(PathDefinition{Base: PathBaseHome, Path: "/absolute/data"}, "darwin", "/home/user"); err == nil {
		t.Fatal("expected absolute path with base to fail")
	}
	if got, err := resolvePath(PathDefinition{Path: `C:\Editor\Data`}, "windows", `C:\Users\sample`); err != nil || got != `C:\Editor\Data` {
		t.Fatalf("Windows absolute path = %q, %v", got, err)
	}
}
