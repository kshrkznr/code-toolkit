package codevenv

import (
	"fmt"
	"os/exec"
	"strings"
)

func createSelectionLink(target, link string) error {
	output, err := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create junction: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
