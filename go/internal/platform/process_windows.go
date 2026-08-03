package platform

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type windowsProcessStopper struct{}

func newProcessStopper() ProcessStopper { return windowsProcessStopper{} }

func (windowsProcessStopper) StopForSelection(ctx context.Context, platform string, runtimePaths ...string) error {
	return stopWindowsProcesses(ctx, platform, false, runtimePaths...)
}

func (windowsProcessStopper) StopRuntime(ctx context.Context, platform string, runtimePaths ...string) error {
	return stopWindowsProcesses(ctx, platform, true, runtimePaths...)
}

func stopWindowsProcesses(ctx context.Context, platform string, runtimeOnly bool, runtimePaths ...string) error {
	process := ""
	switch platform {
	case "code":
		process = "Code.exe"
	case "kiro":
		process = "Kiro.exe"
	default:
		return fmt.Errorf("platform process management is not configured for: %s", platform)
	}

	var pathChecks []string
	for _, path := range runtimePaths {
		if path == "" {
			continue
		}
		path = strings.ReplaceAll(path, "'", "''")
		pathChecks = append(pathChecks, "$_.CommandLine.Contains('"+path+"')")
	}
	selection := "($_.CommandLine -notmatch '--user-data-dir')"
	if len(pathChecks) > 0 {
		selection += " -or " + strings.Join(pathChecks, " -or ")
	}
	if runtimeOnly {
		if len(pathChecks) == 0 {
			return nil
		}
		selection = strings.Join(pathChecks, " -or ")
	}
	script := strings.Join([]string{
		"$items = Get-CimInstance Win32_Process | Where-Object {",
		"  $_.Name -eq '" + process + "' -and $_.CommandLine -notmatch '--type=' -and",
		"  (" + selection + ")",
		"}",
		"$items | ForEach-Object { Stop-Process -Force -Id $_.ProcessId }",
		"$items | ForEach-Object { Wait-Process -Id $_.ProcessId -ErrorAction SilentlyContinue }",
	}, "\n")
	if output, err := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", script).CombinedOutput(); err != nil {
		return fmt.Errorf("stop platform processes: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
