package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	stableVersionPattern = regexp.MustCompile(`^v([0-9]+\.[0-9]+\.[0-9]+)$`)
	checksumLinePattern  = regexp.MustCompile(`^([0-9a-f]{64})[[:space:]]+\*?([^[:space:]]+)$`)
)

type release struct {
	Tag          string
	Version      string
	DarwinARM64  string
	DarwinAMD64  string
	WindowsAMD64 string
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("ctk-package-update", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	version := flags.String("version", "", "published stable CTK tag")
	checksumsPath := flags.String("checksums", "", "published checksums.txt path")
	homebrewFormula := flags.String("homebrew-formula", "", "Homebrew ctk.rb path")
	scoopManifest := flags.String("scoop-manifest", "", "Scoop ctk.json path")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return usageError()
	}
	for name, value := range map[string]string{
		"version":          *version,
		"checksums":        *checksumsPath,
		"homebrew-formula": *homebrewFormula,
		"scoop-manifest":   *scoopManifest,
	} {
		if value == "" {
			return fmt.Errorf("%s is required: %w", name, usageError())
		}
	}

	release, err := loadRelease(*version, *checksumsPath)
	if err != nil {
		return err
	}
	if err := requireExistingFile(*homebrewFormula, "ctk.rb"); err != nil {
		return err
	}
	if err := requireExistingFile(*scoopManifest, "ctk.json"); err != nil {
		return err
	}

	formula := renderHomebrewFormula(release)
	manifest, err := renderScoopManifest(release)
	if err != nil {
		return err
	}
	for _, target := range []struct {
		path    string
		content []byte
	}{
		{path: *homebrewFormula, content: []byte(formula)},
		{path: *scoopManifest, content: manifest},
	} {
		changed, err := writeIfChanged(target.path, target.content)
		if err != nil {
			return err
		}
		state := "current"
		if changed {
			state = "updated"
		}
		fmt.Fprintf(output, "%s: %s\n", state, target.path)
	}
	return nil
}

func loadRelease(tag, checksumsPath string) (release, error) {
	match := stableVersionPattern.FindStringSubmatch(tag)
	if match == nil {
		return release{}, fmt.Errorf("package delivery requires a stable v-prefixed semantic version: %s", tag)
	}
	file, err := os.Open(checksumsPath)
	if err != nil {
		return release{}, fmt.Errorf("open checksums: %w", err)
	}
	defer file.Close()

	expected := []string{
		fmt.Sprintf("ctk_%s_darwin_arm64.tar.gz", tag),
		fmt.Sprintf("ctk_%s_darwin_amd64.tar.gz", tag),
		fmt.Sprintf("ctk_%s_windows_amd64.zip", tag),
	}
	wanted := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		wanted[name] = struct{}{}
	}
	found := make(map[string]string, len(expected))
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		match := checksumLinePattern.FindStringSubmatch(line)
		if match == nil {
			return release{}, fmt.Errorf("invalid checksum line: %s", line)
		}
		name := match[2]
		if filepath.Base(name) != name {
			return release{}, fmt.Errorf("unexpected Release checksum entry: %s", name)
		}
		if _, ok := wanted[name]; !ok {
			return release{}, fmt.Errorf("unexpected Release checksum entry: %s", name)
		}
		if _, duplicate := found[name]; duplicate {
			return release{}, fmt.Errorf("duplicate Release checksum entry: %s", name)
		}
		found[name] = match[1]
	}
	if err := scanner.Err(); err != nil {
		return release{}, fmt.Errorf("read checksums: %w", err)
	}
	var missing []string
	for _, name := range expected {
		if _, ok := found[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return release{}, fmt.Errorf("missing Release checksum entries: %s", strings.Join(missing, ", "))
	}
	return release{
		Tag:          tag,
		Version:      match[1],
		DarwinARM64:  found[expected[0]],
		DarwinAMD64:  found[expected[1]],
		WindowsAMD64: found[expected[2]],
	}, nil
}

func renderHomebrewFormula(value release) string {
	return fmt.Sprintf(`class Ctk < Formula
  ARM64_SHA256 = %q.freeze
  AMD64_SHA256 = %q.freeze

  desc "Compose and reproduce VS Code-family environments"
  homepage "https://github.com/kshrkznr/code-toolkit"
  url "https://github.com/kshrkznr/code-toolkit/releases/download/%s/ctk_%s_darwin_#{Hardware::CPU.arm? ? "arm64" : "amd64"}.tar.gz"
  sha256 Hardware::CPU.arm? ? ARM64_SHA256 : AMD64_SHA256
  license "MIT"

  depends_on :macos

  def install
    bin.install "ctk"
    doc.install "LICENSE", "THIRD_PARTY_NOTICES"
    generate_completions_from_executable bin/"ctk", shell_parameter_format: :cobra, shells: [:bash, :zsh, :fish]
  end

  test do
    assert_match "ctk v#{version}", shell_output("#{bin}/ctk version")
    assert_match "source: packaged", shell_output("#{bin}/ctk docs status")
    assert_match "Usage:", shell_output("#{bin}/ctk --help")
  end
end
`, value.DarwinARM64, value.DarwinAMD64, value.Tag, value.Tag)
}

type scoopManifest struct {
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Homepage     string            `json:"homepage"`
	License      string            `json:"license"`
	Architecture scoopArchitecture `json:"architecture"`
	Bin          string            `json:"bin"`
	Notes        []string          `json:"notes,omitempty"`
	Checkver     scoopCheckver     `json:"checkver"`
	Autoupdate   scoopAutoupdate   `json:"autoupdate"`
}

type scoopArchitecture struct {
	AMD64 scoopDownload `json:"64bit"`
}

type scoopDownload struct {
	URL  string `json:"url"`
	Hash string `json:"hash,omitempty"`
}

type scoopCheckver struct {
	GitHub string `json:"github"`
}

type scoopAutoupdate struct {
	Architecture scoopArchitecture `json:"architecture"`
}

func renderScoopManifest(value release) ([]byte, error) {
	manifest := scoopManifest{
		Version:     value.Version,
		Description: "Compose and reproduce VS Code-family environments",
		Homepage:    "https://github.com/kshrkznr/code-toolkit",
		License:     "MIT",
		Architecture: scoopArchitecture{AMD64: scoopDownload{
			URL:  fmt.Sprintf("https://github.com/kshrkznr/code-toolkit/releases/download/%s/ctk_%s_windows_amd64.zip", value.Tag, value.Tag),
			Hash: value.WindowsAMD64,
		}},
		Bin: "ctk.exe",
		Notes: []string{
			"PowerShell completion is available but is not added to $PROFILE automatically.",
			"To enable it for future sessions, add this line to $PROFILE:",
			"ctk completion powershell | Out-String | Invoke-Expression",
		},
		Checkver: scoopCheckver{GitHub: "https://github.com/kshrkznr/code-toolkit"},
		Autoupdate: scoopAutoupdate{Architecture: scoopArchitecture{AMD64: scoopDownload{
			URL: "https://github.com/kshrkznr/code-toolkit/releases/download/v$version/ctk_v$version_windows_amd64.zip",
		}}},
	}
	content, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("render Scoop manifest: %w", err)
	}
	content = append(content, '\n')
	return bytes.ReplaceAll(content, []byte("\n"), []byte("\r\n")), nil
}

func requireExistingFile(path, name string) error {
	if filepath.Base(path) != name {
		return fmt.Errorf("package target must be named %s: %s", name, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("inspect package target %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("package target is not a regular file: %s", path)
	}
	return nil
}

func writeIfChanged(path string, content []byte) (bool, error) {
	existing, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read package target %s: %w", path, err)
	}
	if bytes.Equal(existing, content) {
		return false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".ctk-package-update-*")
	if err != nil {
		return false, fmt.Errorf("create temporary package target: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return false, fmt.Errorf("write temporary package target: %w", err)
	}
	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		temporary.Close()
		return false, fmt.Errorf("set temporary package target mode: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return false, fmt.Errorf("close temporary package target: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return false, fmt.Errorf("replace package target %s: %w", path, err)
	}
	return true, nil
}

func usageError() error {
	return errors.New("usage: ctk-package-update -version <vX.Y.Z> -checksums <checksums.txt> -homebrew-formula <ctk.rb> -scoop-manifest <ctk.json>")
}
