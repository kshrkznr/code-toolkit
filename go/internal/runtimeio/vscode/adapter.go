package vscode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

type Runner interface {
	Run(context.Context, string, []string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, command string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w: %s", command, err, strings.TrimSpace(stderr.String()))
	}
	return output, nil
}

type Adapter struct {
	Command              string
	UserDataDir          string
	ExtensionsDir        string
	Runner               Runner
	StopForDatabaseWrite func(context.Context) error
}

type profileRecord struct {
	Name            string          `json:"name"`
	Location        string          `json:"location"`
	UseDefaultFlags map[string]bool `json:"useDefaultFlags,omitempty"`
}

func (a Adapter) Scopes(context.Context) ([]runtimeio.Scope, error) {
	profiles, _, err := a.profiles()
	if err != nil {
		return nil, err
	}
	result := []runtimeio.Scope{runtimeio.DefaultScope()}
	for _, profile := range profiles {
		result = append(result, runtimeio.ProfileScope(profile.Name))
	}
	sort.Slice(result[1:], func(i, j int) bool { return result[i+1].Name < result[j+1].Name })
	return result, nil
}

func (a Adapter) EnsureProfile(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("profile name is empty")
	}
	if _, err := a.profile(name); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_, err := a.runner().Run(ctx, a.Command, a.baseArgs(runtimeio.ProfileScope(name)))
	if err != nil {
		return fmt.Errorf("create profile %q: %w", name, err)
	}
	if err := a.stopForDatabaseWrite(ctx); err != nil {
		return err
	}
	for attempt := 0; attempt < 20; attempt++ {
		if _, err := a.profile(name); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("profile %q was not created", name)
}

func (a Adapter) ReadSettings(_ context.Context, scope runtimeio.Scope) (settings.Document, error) {
	path, err := a.settingsPath(scope)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return settings.Document{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read settings %s: %w", path, err)
	}
	document, err := settings.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse settings %s: %w", path, err)
	}
	return document, nil
}

func (a Adapter) WriteSettings(_ context.Context, scope runtimeio.Scope, document settings.Document) error {
	path, err := a.settingsPath(scope)
	if err != nil {
		return err
	}
	data, err := settings.Marshal(document)
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}

func (a Adapter) ReadKeybindings(_ context.Context, scope runtimeio.Scope) (runtimeartifact.Array, error) {
	path, err := a.profileFilePath(scope, "keybindings.json")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return runtimeartifact.Array{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Keybindings %s: %w", path, err)
	}
	value, err := runtimeartifact.ParseArray(data)
	if err != nil {
		return nil, fmt.Errorf("parse Keybindings %s: %w", path, err)
	}
	return value, nil
}

func (a Adapter) WriteKeybindings(_ context.Context, scope runtimeio.Scope, value runtimeartifact.Array) error {
	path, err := a.profileFilePath(scope, "keybindings.json")
	if err != nil {
		return err
	}
	data, err := runtimeartifact.Marshal(value)
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}

func (a Adapter) ReadTasks(_ context.Context, scope runtimeio.Scope) (runtimeartifact.Object, error) {
	if !scope.IsDefault() {
		return nil, fmt.Errorf("Profile-local Tasks: %w", runtimeio.ErrUnsupported)
	}
	return a.readObject(filepath.Join(a.UserDataDir, "User", "tasks.json"), "Tasks")
}

func (a Adapter) WriteTasks(_ context.Context, scope runtimeio.Scope, value runtimeartifact.Object) error {
	if !scope.IsDefault() {
		return fmt.Errorf("Profile-local Tasks: %w", runtimeio.ErrUnsupported)
	}
	return a.writeObject(filepath.Join(a.UserDataDir, "User", "tasks.json"), value)
}

func (a Adapter) ReadMCP(_ context.Context, scope runtimeio.Scope) (runtimeartifact.Object, error) {
	path, err := a.profileFilePath(scope, "mcp.json")
	if err != nil {
		return nil, err
	}
	return a.readObject(path, "MCP")
}

func (a Adapter) WriteMCP(_ context.Context, scope runtimeio.Scope, value runtimeartifact.Object) error {
	path, err := a.profileFilePath(scope, "mcp.json")
	if err != nil {
		return err
	}
	return a.writeObject(path, value)
}

func (a Adapter) ReadSnippets(_ context.Context, scope runtimeio.Scope) (runtimeartifact.Snippets, error) {
	directory, err := a.profileDirectory(scope, "snippets")
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return runtimeartifact.Snippets{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Snippets %s: %w", directory, err)
	}
	result := runtimeartifact.Snippets{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		value, err := runtimeartifact.ParseObject(data)
		if err != nil {
			return nil, fmt.Errorf("parse Snippet %s: %w", entry.Name(), err)
		}
		result[entry.Name()] = value
	}
	return result, nil
}

func (a Adapter) WriteSnippets(_ context.Context, scope runtimeio.Scope, value runtimeartifact.Snippets) error {
	directory, err := a.profileDirectory(scope, "snippets")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if _, keep := value[entry.Name()]; !keep {
			if err := os.Remove(filepath.Join(directory, entry.Name())); err != nil {
				return err
			}
		}
	}
	for name, document := range value {
		if filepath.Base(name) != name {
			return fmt.Errorf("unsafe Snippet filename %q", name)
		}
		data, err := runtimeartifact.Marshal(document)
		if err != nil {
			return err
		}
		if err := atomicWrite(filepath.Join(directory, name), data); err != nil {
			return err
		}
	}
	return nil
}

func (a Adapter) readObject(path, label string) (runtimeartifact.Object, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return runtimeartifact.Object{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s %s: %w", label, path, err)
	}
	value, err := runtimeartifact.ParseObject(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s %s: %w", label, path, err)
	}
	return value, nil
}

func (a Adapter) writeObject(path string, value runtimeartifact.Object) error {
	data, err := runtimeartifact.Marshal(value)
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}

func (a Adapter) Extensions(ctx context.Context, scope runtimeio.Scope) ([]runtimeio.Extension, error) {
	output, err := a.runner().Run(ctx, a.Command, append(a.baseArgs(scope), "--list-extensions", "--show-versions"))
	if err != nil {
		return nil, err
	}
	var result []runtimeio.Extension
	for _, line := range strings.Split(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n") {
		if value := strings.TrimSpace(line); value != "" {
			id, version := value, ""
			if index := strings.LastIndex(value, "@"); index > 0 {
				id, version = value[:index], value[index+1:]
			}
			result = append(result, runtimeio.Extension{ID: id, Version: version})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (a Adapter) InstallExtension(ctx context.Context, scope runtimeio.Scope, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("extension ID is empty")
	}
	args := append(a.baseArgs(scope), "--install-extension", id)
	if strings.HasSuffix(strings.ToLower(id), ".vsix") {
		args = append(args, "--force")
	}
	_, err := a.runner().Run(ctx, a.Command, args)
	return err
}

func (a Adapter) UninstallExtension(ctx context.Context, scope runtimeio.Scope, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("extension ID is empty")
	}
	_, err := a.runner().Run(ctx, a.Command, append(a.baseArgs(scope), "--uninstall-extension", id))
	return err
}

func (a Adapter) SetInheritance(ctx context.Context, scope runtimeio.Scope, inheritance cookbook.Inheritance) error {
	if scope.IsDefault() {
		return fmt.Errorf("default scope has no profile inheritance")
	}
	if err := a.stopForDatabaseWrite(ctx); err != nil {
		return err
	}
	data, err := os.ReadFile(a.storagePath())
	if err != nil {
		return err
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse profile storage: %w", err)
	}
	var profiles []map[string]json.RawMessage
	if err := json.Unmarshal(root["userDataProfiles"], &profiles); err != nil {
		return fmt.Errorf("parse profiles: %w", err)
	}
	found := false
	for i := range profiles {
		var name string
		if err := json.Unmarshal(profiles[i]["name"], &name); err != nil {
			return fmt.Errorf("parse profile name: %w", err)
		}
		if name != scope.Name {
			continue
		}
		found = true
		flags := map[string]bool{}
		existingFlags := map[string]bool{}
		if raw := profiles[i]["useDefaultFlags"]; len(raw) != 0 {
			if err := json.Unmarshal(raw, &existingFlags); err != nil {
				return fmt.Errorf("parse profile inheritance: %w", err)
			}
		}
		for key, value := range map[string]bool{"settings": inheritance.Settings, "keybindings": inheritance.Keybindings, "tasks": inheritance.Tasks, "mcp": inheritance.MCP, "snippets": inheritance.Snippets} {
			if inheritance.Unmanaged[key] {
				if existingFlags[key] {
					flags[key] = true
				}
				continue
			}
			if value {
				flags[key] = true
			}
		}
		if len(flags) == 0 {
			delete(profiles[i], "useDefaultFlags")
		} else {
			encodedFlags, err := json.Marshal(flags)
			if err != nil {
				return fmt.Errorf("encode inheritance: %w", err)
			}
			profiles[i]["useDefaultFlags"] = encodedFlags
		}
	}
	if !found {
		return fmt.Errorf("profile %q: %w", scope.Name, os.ErrNotExist)
	}
	encoded, err := json.Marshal(profiles)
	if err != nil {
		return fmt.Errorf("encode profiles: %w", err)
	}
	root["userDataProfiles"] = encoded
	data, err = json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile storage: %w", err)
	}
	return atomicWrite(a.storagePath(), append(data, '\n'))
}

func (a Adapter) ReadInheritance(_ context.Context, scope runtimeio.Scope) (cookbook.Inheritance, error) {
	if scope.IsDefault() {
		return cookbook.Inheritance{}, nil
	}
	profile, err := a.profile(scope.Name)
	if err != nil {
		return cookbook.Inheritance{}, err
	}
	flags := profile.UseDefaultFlags
	return cookbook.Inheritance{
		Settings: flags["settings"], Keybindings: flags["keybindings"], Tasks: flags["tasks"], MCP: flags["mcp"], Snippets: flags["snippets"],
	}, nil
}

func (a Adapter) settingsPath(scope runtimeio.Scope) (string, error) {
	return a.profileFilePath(scope, "settings.json")
}

func (a Adapter) profileFilePath(scope runtimeio.Scope, name string) (string, error) {
	directory, err := a.profileDirectory(scope, "")
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, name), nil
}

func (a Adapter) profileDirectory(scope runtimeio.Scope, child string) (string, error) {
	directory := filepath.Join(a.UserDataDir, "User")
	if !scope.IsDefault() {
		profile, err := a.profile(scope.Name)
		if err != nil {
			return "", err
		}
		directory = filepath.Join(directory, "profiles", profile.Location)
	}
	if child != "" {
		directory = filepath.Join(directory, child)
	}
	return directory, nil
}

func (a Adapter) profile(name string) (profileRecord, error) {
	profiles, _, err := a.profiles()
	if err != nil {
		return profileRecord{}, err
	}
	for _, profile := range profiles {
		if profile.Name == name {
			return profile, nil
		}
	}
	return profileRecord{}, fmt.Errorf("profile %q: %w", name, os.ErrNotExist)
}

func (a Adapter) profiles() ([]profileRecord, map[string]json.RawMessage, error) {
	data, err := os.ReadFile(a.storagePath())
	if os.IsNotExist(err) {
		return nil, map[string]json.RawMessage{}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read profile storage: %w", err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, nil, fmt.Errorf("parse profile storage: %w", err)
	}
	var profiles []profileRecord
	if raw := root["userDataProfiles"]; len(raw) != 0 {
		if err := json.Unmarshal(raw, &profiles); err != nil {
			return nil, nil, fmt.Errorf("parse profiles: %w", err)
		}
	}
	return profiles, root, nil
}

func (a Adapter) storagePath() string {
	return filepath.Join(a.UserDataDir, "User", "globalStorage", "storage.json")
}

func (a Adapter) baseArgs(scope runtimeio.Scope) []string {
	args := []string{"--user-data-dir", a.UserDataDir, "--extensions-dir", a.ExtensionsDir}
	if !scope.IsDefault() {
		args = append(args, "--profile", scope.Name)
	}
	return args
}

func (a Adapter) runner() Runner {
	if a.Runner != nil {
		return a.Runner
	}
	return ExecRunner{}
}

func (a Adapter) stopForDatabaseWrite(ctx context.Context) error {
	if a.StopForDatabaseWrite == nil {
		return nil
	}
	if err := a.StopForDatabaseWrite(ctx); err != nil {
		return fmt.Errorf("stop Platform for Profile database update: %w", err)
	}
	return nil
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".ctk-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return fmt.Errorf("chmod temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
