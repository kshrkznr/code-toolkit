package codevenv

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"code-toolkit/internal/converge"
	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/recovery"
	"code-toolkit/internal/runtimeio"
	"code-toolkit/internal/runtimelock"
	"code-toolkit/internal/settings"
)

type lifecycleStopper struct{ selectionCalls [][]string }

func (s *lifecycleStopper) StopForSelection(_ context.Context, _ string, paths ...string) error {
	s.selectionCalls = append(s.selectionCalls, append([]string(nil), paths...))
	return nil
}
func (*lifecycleStopper) StopRuntime(context.Context, string, ...string) error { return nil }

type filesystemRuntime struct{ userData, extensions string }

func (r *filesystemRuntime) Scopes(context.Context) ([]runtimeio.Scope, error) {
	return []runtimeio.Scope{runtimeio.DefaultScope()}, nil
}
func (*filesystemRuntime) EnsureProfile(context.Context, string) error { return nil }
func (*filesystemRuntime) SetInheritance(context.Context, runtimeio.Scope, cookbook.Inheritance) error {
	return nil
}
func (*filesystemRuntime) ReadInheritance(context.Context, runtimeio.Scope) (cookbook.Inheritance, error) {
	return cookbook.Inheritance{}, nil
}
func (r *filesystemRuntime) ReadSettings(context.Context, runtimeio.Scope) (settings.Document, error) {
	data, err := os.ReadFile(filepath.Join(r.userData, "User", "settings.json"))
	if os.IsNotExist(err) {
		return settings.Document{}, nil
	}
	if err != nil {
		return nil, err
	}
	return settings.Parse(data)
}
func (r *filesystemRuntime) WriteSettings(_ context.Context, _ runtimeio.Scope, document settings.Document) error {
	data, err := settings.Marshal(document)
	if err != nil {
		return err
	}
	path := filepath.Join(r.userData, "User", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
func (r *filesystemRuntime) Extensions(context.Context, runtimeio.Scope) ([]runtimeio.Extension, error) {
	data, err := os.ReadFile(filepath.Join(r.extensions, "ids.txt"))
	if os.IsNotExist(err) {
		return []runtimeio.Extension{}, nil
	}
	if err != nil {
		return nil, err
	}
	var result []runtimeio.Extension
	for _, value := range strings.Fields(string(data)) {
		result = append(result, runtimeio.Extension{ID: value, Version: "1.0.0"})
	}
	return result, nil
}
func (r *filesystemRuntime) InstallExtension(_ context.Context, _ runtimeio.Scope, id string) error {
	values, err := r.Extensions(context.Background(), runtimeio.DefaultScope())
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(values)+1)
	for _, value := range values {
		ids = append(ids, value.ID)
	}
	ids = append(ids, id)
	sort.Strings(ids)
	if err := os.MkdirAll(r.extensions, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.extensions, "ids.txt"), []byte(strings.Join(ids, "\n")+"\n"), 0o644)
}
func (r *filesystemRuntime) UninstallExtension(_ context.Context, _ runtimeio.Scope, id string) error {
	values, err := r.Extensions(context.Background(), runtimeio.DefaultScope())
	if err != nil {
		return err
	}
	var ids []string
	for _, value := range values {
		if value.ID != id {
			ids = append(ids, value.ID)
		}
	}
	return os.WriteFile(filepath.Join(r.extensions, "ids.txt"), []byte(strings.Join(ids, "\n")), 0o644)
}

func TestActivateHealthCheckAndDeactivate(t *testing.T) {
	root := t.TempDir()
	distRoot := filepath.Join(root, "dist")
	hostData := filepath.Join(root, "host-data")
	hostUser := filepath.Join(hostData, "User")
	hostExtensions := filepath.Join(root, "host-extensions")
	if err := os.MkdirAll(hostUser, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(hostExtensions, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hostUser, "settings.json"), []byte("{\"editor.fontSize\": 15}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hostExtensions, "ids.txt"), []byte("Publisher.One\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stopper := &lifecycleStopper{}
	locks := runtimelock.Store{Attempts: 1}
	service := Service{
		DistRoot: distRoot,
		HostPaths: func(string) (HostPaths, error) {
			return HostPaths{UserData: hostData, User: hostUser, Extensions: hostExtensions}, nil
		},
		Runtime: func(_, userData, extensions string) (runtimeio.Runtime, error) {
			return &filesystemRuntime{userData: userData, extensions: extensions}, nil
		},
		Stopper: stopper,
		Locks:   locks,
		Recovery: recovery.Service{
			Pool: converge.Pool{Root: filepath.Join(root, "pool")}, Locks: locks,
		},
	}

	activated, err := service.Activate(context.Background(), "code", ActivateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if activated.NoOp || activated.Current != "origin.code" {
		t.Fatalf("Activate() = %+v", activated)
	}
	if info, err := os.Stat(distRoot); err != nil || !info.IsDir() {
		t.Fatalf("Activate did not create Distribution root: info=%v err=%v", info, err)
	}
	health, err := service.CheckHealth("code", HostPaths{UserData: hostData, User: hostUser, Extensions: hostExtensions})
	if err != nil || health.State != "healthy" {
		t.Fatalf("health=%+v err=%v", health, err)
	}
	second, err := service.Activate(context.Background(), "code", ActivateOptions{})
	if err != nil || !second.NoOp {
		t.Fatalf("second Activate()=%+v err=%v", second, err)
	}
	if len(stopper.selectionCalls) != 1 || len(stopper.selectionCalls[0]) != 0 {
		t.Fatalf("activation stop calls = %#v", stopper.selectionCalls)
	}

	deactivated, err := service.Deactivate(context.Background(), "code", DeactivateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if deactivated.Forced {
		t.Fatalf("Deactivate() = %+v", deactivated)
	}
	if info, err := os.Lstat(hostUser); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("Host User was not restored physically: info=%v err=%v", info, err)
	}
	if info, err := os.Lstat(hostExtensions); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("Host Extensions were not restored physically: info=%v err=%v", info, err)
	}
	if _, err := os.Lstat(filepath.Join(distRoot, "current.code")); !os.IsNotExist(err) {
		t.Fatalf("current.code still exists: %v", err)
	}
	extensions, err := (&filesystemRuntime{userData: hostData, extensions: hostExtensions}).Extensions(context.Background(), runtimeio.DefaultScope())
	if err != nil || len(extensions) != 1 || extensions[0].ID != "Publisher.One" {
		t.Fatalf("restored extensions=%+v err=%v", extensions, err)
	}
	if len(stopper.selectionCalls) != 2 || len(stopper.selectionCalls[1]) != 1 {
		t.Fatalf("deactivation stop calls = %#v", stopper.selectionCalls)
	}
}

func TestForceEmptyDeactivateDoesNotRequireOrigin(t *testing.T) {
	root := t.TempDir()
	distRoot := filepath.Join(root, "dist")
	runtimeRoot := filepath.Join(distRoot, "selected")
	current := filepath.Join(distRoot, "current.code")
	hostData := filepath.Join(root, "host-data")
	hostUser := filepath.Join(hostData, "User")
	hostExtensions := filepath.Join(root, "host-extensions")
	for _, directory := range []string{filepath.Join(runtimeRoot, ".data", "User"), filepath.Join(runtimeRoot, ".ext"), hostData} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(runtimeRoot, ".data", "User", "settings.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(runtimeRoot, current); err != nil {
		t.Fatal(err)
	}
	if err := createManagedLink(filepath.Join(current, ".data", "User"), hostUser); err != nil {
		t.Fatal(err)
	}
	if err := createManagedLink(filepath.Join(current, ".ext"), hostExtensions); err != nil {
		t.Fatal(err)
	}
	stopper := &lifecycleStopper{}
	service := Service{
		DistRoot: distRoot,
		HostPaths: func(string) (HostPaths, error) {
			return HostPaths{UserData: hostData, User: hostUser, Extensions: hostExtensions}, nil
		},
		Stopper: stopper,
	}
	result, err := service.Deactivate(context.Background(), "code", DeactivateOptions{ForceEmpty: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Forced || !result.Empty {
		t.Fatalf("result=%+v", result)
	}
	for _, path := range []string{hostUser, hostExtensions} {
		info, err := os.Lstat(path)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("empty Host path %s: info=%v err=%v", path, info, err)
		}
		entries, err := os.ReadDir(path)
		if err != nil || len(entries) != 0 {
			t.Fatalf("Host path is not empty %s: entries=%v err=%v", path, entries, err)
		}
	}
	if _, err := os.Lstat(current); !os.IsNotExist(err) {
		t.Fatalf("current Selection remains: %v", err)
	}
	if _, err := os.Stat(runtimeRoot); err != nil {
		t.Fatalf("selected Distribution was removed: %v", err)
	}
	if len(stopper.selectionCalls) != 1 || len(stopper.selectionCalls[0]) != 1 {
		t.Fatalf("stop calls=%#v", stopper.selectionCalls)
	}
}

func TestUnhealthyActivationDirectsForcedDeactivate(t *testing.T) {
	root := t.TempDir()
	origin := filepath.Join(root, "origin.code")
	if err := os.MkdirAll(filepath.Join(origin, ".data", "User"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(origin, ".ext"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(origin, filepath.Join(root, "current.code")); err != nil {
		t.Fatal(err)
	}
	host := HostPaths{UserData: filepath.Join(root, "host"), User: filepath.Join(root, "host", "User"), Extensions: filepath.Join(root, "extensions")}
	if err := os.Symlink(filepath.Join(root, "current.code", ".ext"), host.Extensions); err != nil {
		t.Fatal(err)
	}
	service := Service{DistRoot: root, HostPaths: func(string) (HostPaths, error) { return host, nil }}
	health, err := service.CheckHealth("code", host)
	if err != nil || health.State != "unhealthy" || health.UserState != "missing" {
		t.Fatalf("health=%+v err=%v", health, err)
	}
	_, err = service.Activate(context.Background(), "code", ActivateOptions{})
	if err == nil || !strings.Contains(err.Error(), "deactivate code --force") {
		t.Fatalf("Activate() error = %v", err)
	}
}

func TestRollbackActivationRestoresBackupsAndRemovesNewLinks(t *testing.T) {
	root := t.TempDir()
	origin := filepath.Join(root, "origin.code")
	originBackup := origin + ".ctk-backup"
	if err := os.MkdirAll(originBackup, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(origin, 0o755); err != nil {
		t.Fatal(err)
	}
	current := filepath.Join(root, "current.code")
	if err := os.Symlink(origin, current); err != nil {
		t.Fatal(err)
	}
	hostUser := filepath.Join(root, "host", "User")
	hostExtensions := filepath.Join(root, "extensions")
	if err := os.MkdirAll(filepath.Dir(hostUser), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(current, ".data", "User"), hostUser); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(current, ".ext"), hostExtensions); err != nil {
		t.Fatal(err)
	}
	if err := rollbackJournal(journal{Operation: "activate", Phase: "host-linked", Current: current, Origin: origin, OriginBackup: originBackup, HostUser: hostUser, HostExtensions: hostExtensions}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(current); !os.IsNotExist(err) {
		t.Fatalf("current link remains: %v", err)
	}
	if _, err := os.Lstat(hostUser); !os.IsNotExist(err) {
		t.Fatalf("Host User link remains: %v", err)
	}
	if _, err := os.Lstat(hostExtensions); !os.IsNotExist(err) {
		t.Fatalf("Host Extensions link remains: %v", err)
	}
	if _, err := os.Stat(origin); err != nil {
		t.Fatalf("origin backup was not restored: %v", err)
	}
}

func TestValidateJournalRejectsUnmanagedBackup(t *testing.T) {
	root := t.TempDir()
	paths := HostPaths{User: filepath.Join(root, "host", "User"), Extensions: filepath.Join(root, "extensions")}
	j := journal{Platform: "code", Current: filepath.Join(root, "current.code"), Origin: filepath.Join(root, "origin.code"), HostUser: paths.User, HostExtensions: paths.Extensions, HostUserBackup: filepath.Join(root, "unrelated")}
	if err := validateJournal(root, paths, j); err == nil {
		t.Fatal("validateJournal() error = nil")
	}
}
