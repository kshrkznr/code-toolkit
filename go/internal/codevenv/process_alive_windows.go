//go:build windows

package codevenv

import "golang.org/x/sys/windows"

const stillActive = 259

func processAlive(pid int) bool {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return err == windows.ERROR_ACCESS_DENIED
	}
	defer windows.CloseHandle(handle)
	var exitCode uint32
	if windows.GetExitCodeProcess(handle, &exitCode) != nil {
		return true
	}
	return exitCode == stillActive
}
