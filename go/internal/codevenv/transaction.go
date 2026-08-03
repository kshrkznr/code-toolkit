package codevenv

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type journal struct {
	Operation      string `json:"operation"`
	Platform       string `json:"platform"`
	Phase          string `json:"phase"`
	Current        string `json:"current"`
	CurrentTarget  string `json:"currentTarget,omitempty"`
	Origin         string `json:"origin"`
	OriginBackup   string `json:"originBackup,omitempty"`
	HostUser       string `json:"hostUser"`
	HostUserBackup string `json:"hostUserBackup,omitempty"`
	HostExtensions string `json:"hostExtensions"`
	HostExtBackup  string `json:"hostExtensionsBackup,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

func journalPath(distRoot, platform string) string {
	return filepath.Join(distRoot, ".codevenv."+platform+".journal.json")
}

func writeJournal(path string, value journal) error {
	if value.CreatedAt == "" {
		value.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode CodeVenv journal: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create CodeVenv journal directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".journal-*")
	if err != nil {
		return fmt.Errorf("create CodeVenv journal staging: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(append(data, '\n')); err != nil {
		temporary.Close()
		return fmt.Errorf("write CodeVenv journal: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close CodeVenv journal: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("publish CodeVenv journal: %w", err)
	}
	return nil
}

func readJournal(path string) (journal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return journal{}, err
	}
	var value journal
	if err := json.Unmarshal(data, &value); err != nil {
		return journal{}, fmt.Errorf("parse CodeVenv journal: %w", err)
	}
	return value, nil
}

func copyTree(source, target string) error {
	info, err := os.Lstat(source)
	if os.IsNotExist(err) {
		return os.MkdirAll(target, 0o755)
	}
	if err != nil {
		return fmt.Errorf("inspect %s: %w", source, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("copy tree source is not a directory: %s", source)
	}
	return filepath.Walk(source, func(path string, value os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if value.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
				return err
			}
			return os.Symlink(linkTarget, destination)
		}
		if value.IsDir() {
			return os.MkdirAll(destination, value.Mode().Perm())
		}
		if !value.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type while copying %s", path)
		}
		return copyFile(path, destination, value.Mode().Perm())
	})
}

func copyFile(source, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func plannedBackup(path, suffix string) (string, error) {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	backup := path + suffix
	if _, err := os.Lstat(backup); err == nil {
		return "", fmt.Errorf("backup path already exists: %s", backup)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return backup, nil
}

func executeBackup(path, backup string) error {
	if backup == "" {
		return nil
	}
	if err := os.Rename(path, backup); err != nil {
		return fmt.Errorf("preserve %s: %w", path, err)
	}
	return nil
}
