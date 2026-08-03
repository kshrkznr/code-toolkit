package codevenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProcessLockRejectsLiveOwnerAndReleases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selection.lock")
	unlock, err := acquireProcessLock(path, "busy")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := acquireProcessLock(path, "busy"); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("second acquire error = %v", err)
	}
	unlock()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("lock remains after release: %v", err)
	}
}

func TestProcessLockReclaimsDeadOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selection.lock")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	owner := `{"pid":2147483647,"token":"dead","created":"2026-01-01T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(path, processLockOwner), []byte(owner), 0o600); err != nil {
		t.Fatal(err)
	}
	unlock, err := acquireProcessLock(path, "busy")
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}

func TestProcessLockReclaimsOldLegacyDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selection.lock")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Minute)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	unlock, err := acquireProcessLock(path, "busy")
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}

func TestProcessLockPreservesRecentIncompleteDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selection.lock")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := acquireProcessLock(path, "busy"); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("acquire error = %v", err)
	}
}
