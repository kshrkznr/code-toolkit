package platform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type darwinProcessStopper struct{}

func newProcessStopper() ProcessStopper { return darwinProcessStopper{} }

func (darwinProcessStopper) StopForSelection(ctx context.Context, platform string, runtimePaths ...string) error {
	return stopDarwinProcesses(ctx, platform, false, runtimePaths...)
}

func (darwinProcessStopper) StopRuntime(ctx context.Context, platform string, runtimePaths ...string) error {
	return stopDarwinProcesses(ctx, platform, true, runtimePaths...)
}

func stopDarwinProcesses(ctx context.Context, platform string, runtimeOnly bool, runtimePaths ...string) error {
	definition, err := Lookup(platform)
	if err != nil {
		return fmt.Errorf("platform process management is not configured for: %s", platform)
	}
	processDefinition, ok := definition.OS["darwin"]
	if !ok {
		return fmt.Errorf("platform process management is not configured for: %s", platform)
	}
	output, err := exec.CommandContext(ctx, "ps", "-axo", "pid=,args=").Output()
	if err != nil {
		return fmt.Errorf("inspect platform processes: %w", err)
	}

	for _, line := range strings.Split(string(output), "\n") {
		pid, args, ok := parseProcess(line)
		if !ok || !relevantProcessDefinition(processDefinition.Process, args, runtimePaths) || runtimeOnly && !matchesRuntimePath(args, runtimePaths) {
			continue
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("find platform process %d: %w", pid, err)
		}
		if err := process.Signal(syscall.SIGTERM); err != nil && !isProcessDone(err) {
			return fmt.Errorf("stop platform process %d: %w", pid, err)
		}
		if err := waitStopped(ctx, process, 10*time.Second); err != nil {
			return fmt.Errorf("confirm platform process %d stopped: %w", pid, err)
		}
	}
	return nil
}

func matchesRuntimePath(args string, runtimePaths []string) bool {
	for _, path := range runtimePaths {
		if path != "" && strings.Contains(args, path) {
			return true
		}
	}
	return false
}

func parseProcess(line string) (int, string, bool) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 {
		return 0, "", false
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, "", false
	}
	return pid, strings.Join(fields[1:], " "), true
}

func relevantProcess(platform, args string, runtimePaths []string) bool {
	definition, err := Lookup(platform)
	if err != nil {
		return false
	}
	osDefinition, ok := definition.OS["darwin"]
	if !ok {
		return false
	}
	return relevantProcessDefinition(osDefinition.Process, args, runtimePaths)
}

func relevantProcessDefinition(process ProcessDefinition, args string, runtimePaths []string) bool {
	if strings.Contains(args, " Helper") || strings.Contains(args, " --type=") {
		return false
	}
	for _, path := range runtimePaths {
		if path != "" && strings.Contains(args, path) {
			return true
		}
	}
	if strings.Contains(args, "--user-data-dir") {
		return false
	}
	for _, identity := range process.Identities {
		if strings.Contains(args, identity) {
			return true
		}
	}
	return false
}

func waitStopped(ctx context.Context, process *os.Process, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := process.Signal(syscall.Signal(0)); isProcessDone(err) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("timeout after %s", timeout)
		case <-ticker.C:
		}
	}
}

func isProcessDone(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "process already finished") || strings.Contains(err.Error(), "no such process"))
}
