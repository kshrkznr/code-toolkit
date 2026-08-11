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
	process, err := windowsProcessName(platform)
	if err != nil {
		return err
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
	script := windowsStopScript(process, selection)
	if output, err := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", script).CombinedOutput(); err != nil {
		return fmt.Errorf("stop platform processes: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func windowsProcessName(platform string) (string, error) {
	switch platform {
	case "code":
		return "Code.exe", nil
	case "codium":
		return "VSCodium.exe", nil
	case "kiro":
		return "Kiro.exe", nil
	case "cursor":
		return "Cursor.exe", nil
	case "devin-desktop":
		return "Devin.exe", nil
	default:
		return "", fmt.Errorf("platform process management is not configured for: %s", platform)
	}
}

func windowsStopScript(process, selection string) string {
	return strings.Join([]string{
		"$all = @(Get-CimInstance Win32_Process | Where-Object { $_.Name -eq '" + process + "' })",
		"$items = $all | Where-Object {",
		// Cursor can use Cursor.exe itself for language servers and other workers
		// without a --type argument. Only a process without a same-name parent is
		// a desktop root; stopping that root lets its process tree exit normally.
		"  $all.ProcessId -notcontains $_.ParentProcessId -and",
		"  $_.Name -eq '" + process + "' -and $_.CommandLine -notmatch '--type=' -and",
		"  (" + selection + ")",
		"}",
		"$items | ForEach-Object { Stop-Process -Force -Id $_.ProcessId }",
		"$items | ForEach-Object { Wait-Process -Id $_.ProcessId -ErrorAction SilentlyContinue }",
	}, "\n")
}
