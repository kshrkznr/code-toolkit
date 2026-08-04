package converge

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
)

type fakeDownloader struct{ fail map[string]bool }

func (f fakeDownloader) Download(_ context.Context, repository string, _ runtimeio.Extension, destination string) error {
	if f.fail[repository] {
		return errors.New("unavailable")
	}
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(file)
	entry, err := archive.Create("extension/package.json")
	if err == nil {
		_, err = entry.Write([]byte(`{}`))
	}
	if closeErr := archive.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}

func TestPoolUpdaterFallsBackAndReplacesOldVersion(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, "visual-studio-marketplace", "sample.id-1.0.vsix")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	updater := PoolUpdater{Root: root, Downloader: fakeDownloader{fail: map[string]bool{"open-vsx": true}}}
	report := Report{}
	updater.Update(context.Background(), "kiro", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.id", Version: "2.0"}}}}, &report)
	newPath := filepath.Join(root, "visual-studio-marketplace", "sample.id-2.0.vsix")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old artifact remains: %v", err)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func TestPoolUpdaterFailureIsUnresolved(t *testing.T) {
	report := Report{}
	updater := PoolUpdater{Root: t.TempDir(), Downloader: fakeDownloader{fail: map[string]bool{"visual-studio-marketplace": true}}}
	updater.Update(context.Background(), "code", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.id", Version: "1.0"}}}}, &report)
	if report.HasFailures() || len(report.Operations) != 1 || report.Operations[0].Status != Unresolved {
		t.Fatalf("report = %#v", report)
	}
}

func TestPoolUpdaterNormalizesStorageIDToLowerCase(t *testing.T) {
	root := t.TempDir()
	updater := PoolUpdater{Root: root, Downloader: fakeDownloader{fail: map[string]bool{}}}
	report := Report{}
	updater.Update(context.Background(), "kiro", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "emilast.LogFileHighlighter", Version: "2.8.0"}}}}, &report)
	path := filepath.Join(root, "open-vsx", "emilast.logfilehighlighter-2.8.0.vsix")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func TestValidateVSIXExactAcceptsPackageIdentityCaseDifference(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alefragnani.bookmarks-14.1.1.vsix")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entry, err := archive.Create("extension/package.json")
	if err == nil {
		_, err = entry.Write([]byte(`{"publisher":"alefragnani","name":"Bookmarks","version":"14.1.1"}`))
	}
	if closeErr := archive.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateVSIXExact(path, runtimeio.Extension{ID: "alefragnani.bookmarks", Version: "14.1.1"})
	if err != nil {
		t.Fatalf("ValidateVSIXExact() error = %v", err)
	}
}
