package directlauncher

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/recipe"
)

func TestGenerateUnixDirectLauncher(t *testing.T) {
	dist := distribution.Distribution{Name: "vscode-golang.1", Path: t.TempDir(), Recipe: recipe.Recipe{OS: "macos", Platform: "code"}}
	result, err := Generate(dist)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || filepath.Base(result.Path) != dist.Name {
		t.Fatalf("result = %#v", result)
	}
	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{marker, `exec "$BASE_DIR/run.sh" "$@"`, `--user-data-dir "$BASE_DIR/.data"`, `--extensions-dir "$BASE_DIR/.ext"`} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("launcher missing %q:\n%s", want, content)
		}
	}
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(result.Path)
		if info.Mode().Perm() != 0o755 {
			t.Fatalf("mode = %o", info.Mode().Perm())
		}
	}
}

func TestGenerateWindowsDirectLauncher(t *testing.T) {
	dist := distribution.Distribution{Name: "vscode-java", Path: t.TempDir(), Recipe: recipe.Recipe{OS: "windows", Platform: "code"}}
	result, err := Generate(dist)
	if err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(result.Path)
	if filepath.Base(result.Path) != "vscode-java.cmd" || !strings.Contains(string(content), `if exist "%BASE_DIR%run.cmd"`) || !strings.Contains(string(content), `--extensions-dir "%BASE_DIR%.ext"`) {
		t.Fatalf("result=%#v content=%s", result, content)
	}
}

func TestGenerateRejectsUnknownPlatform(t *testing.T) {
	dist := distribution.Distribution{Name: "sample", Path: t.TempDir(), Recipe: recipe.Recipe{OS: "windows", Platform: "code & echo unsafe"}}
	if _, err := Generate(dist); err == nil {
		t.Fatal("expected unknown Platform error")
	}
}

func TestGenerateSkipsLaunchOverrideNameCollision(t *testing.T) {
	dist := distribution.Distribution{Name: "run", Path: t.TempDir(), Recipe: recipe.Recipe{OS: "windows", Platform: "code"}}
	result, err := Generate(dist)
	if err != nil {
		t.Fatal(err)
	}
	if result.Warning == "" {
		t.Fatalf("result = %#v", result)
	}
	if _, err := os.Stat(filepath.Join(dist.Path, "run.cmd")); !os.IsNotExist(err) {
		t.Fatalf("conflicting launcher created: %v", err)
	}
}

func TestGeneratePreservesUnknownExistingFile(t *testing.T) {
	dist := distribution.Distribution{Name: "sample", Path: t.TempDir(), Recipe: recipe.Recipe{OS: "macos", Platform: "code"}}
	target := filepath.Join(dist.Path, dist.Name)
	if err := os.WriteFile(target, []byte("custom\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Generate(dist)
	if err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(target)
	if result.Warning == "" || string(content) != "custom\n" {
		t.Fatalf("result=%#v content=%q", result, content)
	}
}

func TestGenerateReplacesHistoricalCTKLauncher(t *testing.T) {
	dist := distribution.Distribution{Name: "sample", Path: t.TempDir(), Recipe: recipe.Recipe{OS: "macos", Platform: "code"}}
	target := filepath.Join(dist.Path, dist.Name)
	old := "#!/usr/bin/env bash\nBASE_DIR=\"$(cd \"$(dirname \"$0\")\" && pwd)\"\nexec \"$BASE_DIR/run.sh\" \"$@\"\n"
	if err := os.WriteFile(target, []byte(old), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Generate(dist)
	if err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(target)
	if !result.Changed || !strings.Contains(string(content), marker) {
		t.Fatalf("result=%#v content=%q", result, content)
	}
}

func TestUnixDirectLauncherStartsPlatformWithoutCTK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix launcher execution requires a Unix host")
	}
	root := t.TempDir()
	distPath := filepath.Join(root, "dist", "sample")
	binPath := filepath.Join(root, "platform-bin")
	if err := os.MkdirAll(distPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(binPath, 0755); err != nil {
		t.Fatal(err)
	}
	platform := filepath.Join(binPath, "code")
	if err := os.WriteFile(platform, []byte("#!/usr/bin/env bash\nprintf '%s\\n' \"$@\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	dist := distribution.Distribution{Name: "sample", Path: distPath, Recipe: recipe.Recipe{OS: "macos", Platform: "code"}}
	result, err := Generate(dist)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(result.Path, "workspace")
	command.Env = append(os.Environ(), "PATH="+binPath+":/usr/bin:/bin")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("launcher error = %v, output = %s", err, output)
	}
	for _, want := range []string{"--user-data-dir\n" + distPath + "/.data", "--extensions-dir\n" + distPath + "/.ext", "workspace"} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestUnixDirectLauncherUsesLaunchOverride(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix launcher execution requires a Unix host")
	}
	dist := distribution.Distribution{Name: "sample", Path: t.TempDir(), Recipe: recipe.Recipe{OS: "macos", Platform: "code"}}
	result, err := Generate(dist)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dist.Path, "run.sh"), []byte("#!/usr/bin/env bash\nprintf 'override:%s\\n' \"$1\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(result.Path, "argument").CombinedOutput()
	if err != nil {
		t.Fatalf("launcher error = %v, output = %s", err, output)
	}
	if string(output) != "override:argument\n" {
		t.Fatalf("output = %q", output)
	}
}
