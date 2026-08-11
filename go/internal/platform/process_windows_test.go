//go:build windows

package platform

import (
	"strings"
	"testing"
)

func TestWindowsStopScriptSelectsOnlyDesktopRoots(t *testing.T) {
	script := windowsStopScript("Cursor.exe", "$_.CommandLine.Contains('C:\\runtime\\.data')")

	for _, expected := range []string{
		"$all = @(Get-CimInstance Win32_Process",
		"$all.ProcessId -notcontains $_.ParentProcessId",
		"$_.Name -eq 'Cursor.exe'",
		"$_.CommandLine -notmatch '--type='",
		"$_.CommandLine.Contains('C:\\runtime\\.data')",
		"Stop-Process -Force -Id $_.ProcessId",
		"Wait-Process -Id $_.ProcessId",
	} {
		if !strings.Contains(script, expected) {
			t.Fatalf("windowsStopScript() missing %q:\n%s", expected, script)
		}
	}
}

func TestWindowsProcessNameIncludesDevinDesktop(t *testing.T) {
	got, err := windowsProcessName("devin-desktop")
	if err != nil || got != "Devin.exe" {
		t.Fatalf("windowsProcessName() = %q, %v", got, err)
	}
}
