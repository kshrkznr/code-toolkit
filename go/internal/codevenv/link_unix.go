//go:build !windows

package codevenv

import "os"

func createSelectionLink(target, link string) error {
	return os.Symlink(target, link)
}
