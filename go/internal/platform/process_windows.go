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
	definition, err := Lookup(platform)
	if err != nil {
		return fmt.Errorf("platform process management is not configured for: %s", platform)
	}
	osDefinition, ok := definition.OS["windows"]
	if !ok || len(osDefinition.Process.Identities) != 1 {
		return fmt.Errorf("platform process management is not configured for: %s", platform)
	}
	process := osDefinition.Process.Identities[0]

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
	script := windowsStopScript(process, selection, osDefinition.Process.AdditionalFilters...)
	if output, err := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", script).CombinedOutput(); err != nil {
		return fmt.Errorf("stop platform processes: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func windowsProcessName(platform string) (string, error) {
	definition, err := Lookup(platform)
	if err != nil {
		return "", fmt.Errorf("platform process management is not configured for: %s", platform)
	}
	osDefinition, ok := definition.OS["windows"]
	if !ok || len(osDefinition.Process.Identities) != 1 {
		return "", fmt.Errorf("platform process management is not configured for: %s", platform)
	}
	return osDefinition.Process.Identities[0], nil
}

func windowsStopScript(process, selection string, additionalFilters ...string) string {
	lines := []string{
		"$all = @(Get-CimInstance Win32_Process | Where-Object { $_.Name -eq '" + process + "' })",
		"$items = $all | Where-Object {",
	}
	for _, filter := range additionalFilters {
		if filter == FilterSameNameRoot {
			lines = append(lines, "  $all.ProcessId -notcontains $_.ParentProcessId -and")
		}
	}
	lines = append(lines,
		"  $_.Name -eq '"+process+"' -and $_.CommandLine -notmatch '--type=' -and",
		"  ("+selection+")",
		"}",
		"$items | ForEach-Object { Stop-Process -Force -Id $_.ProcessId }",
		"$items | ForEach-Object { Wait-Process -Id $_.ProcessId -ErrorAction SilentlyContinue }",
	)
	return strings.Join(lines, "\n")
}
