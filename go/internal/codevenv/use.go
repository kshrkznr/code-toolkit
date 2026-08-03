package codevenv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"code-toolkit/internal/distribution"
	"code-toolkit/internal/platform"
)

type UseResult struct {
	Platform string
	Current  string
	Changed  bool
}

func Use(ctx context.Context, distRoot string, target distribution.Distribution, stopper platform.ProcessStopper) (UseResult, error) {
	platformName := target.Recipe.Platform
	current := filepath.Join(distRoot, "current."+platformName)
	oldTarget, err := os.Readlink(current)
	if err != nil {
		if os.IsNotExist(err) {
			return UseResult{}, fmt.Errorf("platform is not active: %s (run: ctk activate %s)", platformName, platformName)
		}
		return UseResult{}, fmt.Errorf("read current selection for %s: %w", platformName, err)
	}
	oldTarget = absoluteLinkTarget(current, oldTarget)
	if samePath(oldTarget, target.Path) {
		return UseResult{Platform: platformName, Current: target.Name}, nil
	}

	lock := filepath.Join(distRoot, ".selection."+platformName+".lock")
	unlock, err := acquireProcessLock(lock, "selection already in progress for platform: "+platformName)
	if err != nil {
		return UseResult{}, err
	}
	defer unlock()

	oldTarget, err = os.Readlink(current)
	if err != nil {
		return UseResult{}, fmt.Errorf("re-read current selection for %s: %w", platformName, err)
	}
	oldTarget = absoluteLinkTarget(current, oldTarget)
	if samePath(oldTarget, target.Path) {
		return UseResult{Platform: platformName, Current: target.Name}, nil
	}

	paths := []string{filepath.Join(oldTarget, ".data"), filepath.Join(target.Path, ".data")}
	if err := stopper.StopForSelection(ctx, platformName, paths...); err != nil {
		return UseResult{}, err
	}
	if err := replaceSelection(current, target.Path); err != nil {
		return UseResult{}, err
	}
	return UseResult{Platform: platformName, Current: target.Name, Changed: true}, nil
}

func replaceSelection(current, target string) error {
	return replaceSelectionWith(current, target, os.Rename)
}

func replaceSelectionWith(current, target string, rename func(string, string) error) error {
	dir := filepath.Dir(current)
	name := filepath.Base(current)
	next, err := os.MkdirTemp(dir, "."+name+".next-")
	if err != nil {
		return fmt.Errorf("prepare selection: %w", err)
	}
	if err := os.Remove(next); err != nil {
		return fmt.Errorf("prepare selection link: %w", err)
	}
	if err := createSelectionLink(target, next); err != nil {
		return fmt.Errorf("create selection link: %w", err)
	}
	defer os.Remove(next)

	resolved, exists, err := linkTarget(next)
	if err != nil || !exists || !samePath(resolved, target) {
		return fmt.Errorf("validate selection link: target mismatch")
	}

	previous, err := os.MkdirTemp(dir, "."+name+".previous-")
	if err != nil {
		return fmt.Errorf("prepare previous selection: %w", err)
	}
	if err := os.Remove(previous); err != nil {
		return fmt.Errorf("prepare previous selection path: %w", err)
	}
	if err := rename(current, previous); err != nil {
		return fmt.Errorf("backup current selection: %w", err)
	}
	if err := rename(next, current); err != nil {
		rollbackErr := rename(previous, current)
		if rollbackErr != nil {
			return errors.Join(fmt.Errorf("replace current selection: %w", err), fmt.Errorf("restore previous selection: %w", rollbackErr))
		}
		return fmt.Errorf("replace current selection: %w", err)
	}
	_ = os.Remove(previous)
	return nil
}

func absoluteLinkTarget(link, target string) string {
	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}
	return filepath.Clean(filepath.Join(filepath.Dir(link), target))
}

func samePath(left, right string) bool {
	a, errA := canonicalPath(left)
	b, errB := canonicalPath(right)
	return errA == nil && errB == nil && filepath.Clean(a) == filepath.Clean(b)
}

func canonicalPath(path string) (string, error) {
	value, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(value); err == nil {
		value = resolved
	}
	return filepath.Clean(value), nil
}
