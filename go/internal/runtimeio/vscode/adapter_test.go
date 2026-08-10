package vscode

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type call struct {
	command string
	args    []string
}
type fakeRunner struct {
	output []byte
	calls  []call
	onRun  func()
}

func (f *fakeRunner) Run(_ context.Context, command string, args []string) ([]byte, error) {
	f.calls = append(f.calls, call{command, append([]string(nil), args...)})
	if f.onRun != nil {
		f.onRun()
	}
	return f.output, nil
}

func TestEnsureProfileOwnsLaunchStopAndVerification(t *testing.T) {
	root := t.TempDir()
	storage := filepath.Join(root, "data", "User", "globalStorage", "storage.json")
	runner := &fakeRunner{onRun: func() { mustWrite(t, storage, `{"userDataProfiles":[{"name":"work","location":"abc"}]}`) }}
	stops := 0
	adapter := Adapter{
		Command: "code", UserDataDir: filepath.Join(root, "data"), ExtensionsDir: filepath.Join(root, "ext"), Runner: runner,
		StopForDatabaseWrite: func(context.Context) error { stops++; return nil },
	}
	if err := adapter.EnsureProfile(context.Background(), "work"); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 2 || stops != 1 {
		t.Fatalf("calls=%v stops=%d", runner.calls, stops)
	}
	if got := runner.calls[1].args; !reflect.DeepEqual(got, []string{"--user-data-dir", filepath.Join(root, "data"), "--extensions-dir", filepath.Join(root, "ext"), "--list-extensions"}) {
		t.Fatalf("settle args = %v", got)
	}
}

func TestEnsureProfileWaitsForPersistenceBeforeStop(t *testing.T) {
	root := t.TempDir()
	storage := filepath.Join(root, "data", "User", "globalStorage", "storage.json")
	persisted := make(chan struct{})
	runs := 0
	runner := &fakeRunner{onRun: func() {
		runs++
		if runs != 1 {
			return
		}
		go func() {
			time.Sleep(150 * time.Millisecond)
			if err := os.MkdirAll(filepath.Dir(storage), 0o755); err != nil {
				return
			}
			if err := os.WriteFile(storage, []byte(`{"userDataProfiles":[{"name":"work","location":"abc"}]}`), 0o644); err != nil {
				return
			}
			close(persisted)
		}()
	}}
	adapter := Adapter{
		Command: "code", UserDataDir: filepath.Join(root, "data"), ExtensionsDir: filepath.Join(root, "ext"), Runner: runner,
		StopForDatabaseWrite: func(context.Context) error {
			select {
			case <-persisted:
				return nil
			default:
				return errors.New("profile process stopped before persistence")
			}
		},
	}
	if err := adapter.EnsureProfile(context.Background(), "work"); err != nil {
		t.Fatal(err)
	}
}

func TestSettingsScopesAndInheritance(t *testing.T) {
	root := t.TempDir()
	storage := filepath.Join(root, "data", "User", "globalStorage", "storage.json")
	mustWrite(t, storage, `{"preserved":{"value":1},"userDataProfiles":[{"name":"work","location":"abc","icon":"rocket","useDefaultFlags":{"settings":true}}]}`)
	adapter := Adapter{Command: "code", UserDataDir: filepath.Join(root, "data"), ExtensionsDir: filepath.Join(root, "ext")}
	ctx := context.Background()
	if err := adapter.WriteSettings(ctx, runtimeio.DefaultScope(), settings.Document{"default": true}); err != nil {
		t.Fatal(err)
	}
	if err := adapter.WriteSettings(ctx, runtimeio.ProfileScope("work"), settings.Document{"profile": true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "data", "User", "settings.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "data", "User", "profiles", "abc", "settings.json")); err != nil {
		t.Fatal(err)
	}
	inheritance := cookbook.Inheritance{Keybindings: true, Snippets: true}
	if err := adapter.SetInheritance(ctx, runtimeio.ProfileScope("work"), inheritance); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(storage)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	if value["preserved"].(map[string]any)["value"] != float64(1) {
		t.Fatalf("storage = %s", data)
	}
	profile := value["userDataProfiles"].([]any)[0].(map[string]any)
	if profile["icon"] != "rocket" {
		t.Fatalf("profile = %#v", profile)
	}
	flags := profile["useDefaultFlags"].(map[string]any)
	if flags["keybindings"] != true || flags["snippets"] != true || flags["settings"] != nil {
		t.Fatalf("flags = %#v", flags)
	}
}

func TestExtensionOperationsUsePlatformCLI(t *testing.T) {
	runner := &fakeRunner{output: []byte("Zulu.Extension@2.0\r\nalpha.extension@1.0\r\n")}
	adapter := Adapter{Command: "code", UserDataDir: "/data", ExtensionsDir: "/ext", Runner: runner}
	ctx := context.Background()
	ids, err := adapter.Extensions(ctx, runtimeio.ProfileScope("work"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []runtimeio.Extension{{ID: "Zulu.Extension", Version: "2.0"}, {ID: "alpha.extension", Version: "1.0"}}) {
		t.Fatalf("ids = %v", ids)
	}
	if err := adapter.InstallExtension(ctx, runtimeio.ProfileScope("work"), "Mixed.Case"); err != nil {
		t.Fatal(err)
	}
	if err := adapter.UninstallExtension(ctx, runtimeio.DefaultScope(), "Mixed.Case"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runner.calls[1].args, []string{"--user-data-dir", "/data", "--extensions-dir", "/ext", "--profile", "work", "--install-extension", "Mixed.Case"}) {
		t.Fatalf("install args = %v", runner.calls[1].args)
	}
	if err := adapter.InstallExtension(ctx, runtimeio.ProfileScope("work"), "/archive/sample.vsix"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runner.calls[3].args, []string{"--user-data-dir", "/data", "--extensions-dir", "/ext", "--profile", "work", "--install-extension", "/archive/sample.vsix", "--force"}) {
		t.Fatalf("VSIX args = %v", runner.calls[3].args)
	}
}

func TestRuntimeArtifactPathsAndProfileTasksCapability(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	mustWrite(t, filepath.Join(data, "User", "globalStorage", "storage.json"), `{"userDataProfiles":[{"name":"work","location":"abc"}]}`)
	adapter := Adapter{UserDataDir: data}
	ctx := context.Background()
	profile := runtimeio.ProfileScope("work")
	if err := adapter.WriteKeybindings(ctx, profile, runtimeartifact.Array{map[string]any{"key": "ctrl+x"}}); err != nil {
		t.Fatal(err)
	}
	if err := adapter.WriteMCP(ctx, profile, runtimeartifact.Object{"servers": map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := adapter.WriteSnippets(ctx, profile, runtimeartifact.Snippets{"go.json": runtimeartifact.Object{"Probe": map[string]any{"body": []any{"x"}}}}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"keybindings.json", "mcp.json", "snippets/go.json"} {
		if _, err := os.Stat(filepath.Join(data, "User", "profiles", "abc", name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := adapter.WriteTasks(ctx, profile, runtimeartifact.Object{}); !errors.Is(err, runtimeio.ErrUnsupported) {
		t.Fatalf("profile Tasks error = %v", err)
	}
	if err := adapter.WriteTasks(ctx, runtimeio.DefaultScope(), runtimeartifact.Object{"version": "2.0.0", "tasks": []any{}, "inputs": []any{}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(data, "User", "tasks.json")); err != nil {
		t.Fatal(err)
	}
}

func TestMissingObjectArtifactIsEmpty(t *testing.T) {
	adapter := Adapter{UserDataDir: t.TempDir()}
	value, err := adapter.ReadMCP(context.Background(), runtimeio.DefaultScope())
	if err != nil || value == nil || len(value) != 0 {
		t.Fatalf("MCP = %#v, %v", value, err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
