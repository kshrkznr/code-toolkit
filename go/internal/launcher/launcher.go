package launcher

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"code-toolkit/internal/distribution"
)

type Launcher struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	GOOS   string
}

func New() *Launcher {
	return &Launcher{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr, GOOS: runtime.GOOS}
}

func OverrideName(goos string) string {
	if goos == "windows" {
		return "run.cmd"
	}
	return "run.sh"
}

func (l *Launcher) Launch(ctx context.Context, dist distribution.Distribution, args []string) error {
	command, commandArgs, override, err := l.command(dist, args)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, command, commandArgs...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = l.Stdin, l.Stdout, l.Stderr
	if err := cmd.Run(); err != nil {
		kind := "platform"
		if override {
			kind = "launch override"
		}
		return fmt.Errorf("%s failed: %w", kind, err)
	}
	return nil
}

func (l *Launcher) command(dist distribution.Distribution, args []string) (string, []string, bool, error) {
	if override := l.override(dist.Path); override != "" {
		if l.GOOS == "windows" {
			return "cmd.exe", append([]string{"/c", override}, args...), true, nil
		}
		return override, args, true, nil
	}

	platform, err := exec.LookPath(dist.Recipe.Platform)
	if err != nil {
		return "", nil, false, fmt.Errorf("platform command not found: %s", dist.Recipe.Platform)
	}
	platformArgs := []string{
		"--user-data-dir", filepath.Join(dist.Path, ".data"),
		"--extensions-dir", filepath.Join(dist.Path, ".ext"),
	}
	return platform, append(platformArgs, args...), false, nil
}

func (l *Launcher) override(distPath string) string {
	path := filepath.Join(distPath, OverrideName(l.GOOS))
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ""
	}
	return path
}
