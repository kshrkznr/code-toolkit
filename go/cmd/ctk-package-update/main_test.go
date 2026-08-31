package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUpdatesBothPackageDefinitions(t *testing.T) {
	directory := t.TempDir()
	checksums := filepath.Join(directory, "checksums.txt")
	formula := filepath.Join(directory, "ctk.rb")
	manifest := filepath.Join(directory, "ctk.json")
	writeTestFile(t, checksums, strings.Join([]string{
		strings.Repeat("a", 64) + "  ctk_v1.2.3_darwin_arm64.tar.gz",
		strings.Repeat("b", 64) + "  ctk_v1.2.3_darwin_amd64.tar.gz",
		strings.Repeat("c", 64) + "  ctk_v1.2.3_windows_amd64.zip",
	}, "\n")+"\n")
	writeTestFile(t, formula, "old formula\n")
	writeTestFile(t, manifest, "{}\n")

	var output bytes.Buffer
	err := run([]string{
		"-version", "v1.2.3",
		"-checksums", checksums,
		"-homebrew-formula", formula,
		"-scoop-manifest", manifest,
	}, &output)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"updated: " + formula, "updated: " + manifest} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output %q does not contain %q", output.String(), expected)
		}
	}
	formulaContent, err := os.ReadFile(formula)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"v1.2.3", strings.Repeat("a", 64), strings.Repeat("b", 64)} {
		if !strings.Contains(string(formulaContent), expected) {
			t.Fatalf("formula does not contain %q", expected)
		}
	}
	manifestContent, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"version": "1.2.3"`, "ctk_v1.2.3_windows_amd64.zip", strings.Repeat("c", 64)} {
		if !strings.Contains(string(manifestContent), expected) {
			t.Fatalf("manifest does not contain %q", expected)
		}
	}
	withoutCRLF := bytes.ReplaceAll(manifestContent, []byte("\r\n"), nil)
	if !bytes.Contains(manifestContent, []byte("\r\n")) || bytes.Contains(withoutCRLF, []byte("\n")) {
		t.Fatalf("Scoop manifest does not use CRLF consistently: %q", manifestContent)
	}

	output.Reset()
	if err := run([]string{
		"-version", "v1.2.3",
		"-checksums", checksums,
		"-homebrew-formula", formula,
		"-scoop-manifest", manifest,
	}, &output); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), "current:") != 2 {
		t.Fatalf("second run is not idempotent: %q", output.String())
	}
}

func TestLoadReleaseRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		content  string
		expected string
	}{
		{name: "prerelease", version: "v1.2.3-rc.1", expected: "requires a stable"},
		{name: "malformed", version: "v1.2.3", content: "not-a-checksum\n", expected: "invalid checksum line"},
		{name: "unexpected", version: "v1.2.3", content: strings.Repeat("a", 64) + "  other.zip\n", expected: "unexpected Release checksum"},
		{name: "path", version: "v1.2.3", content: strings.Repeat("a", 64) + "  nested/ctk_v1.2.3_darwin_arm64.tar.gz\n", expected: "unexpected Release checksum"},
		{name: "duplicate", version: "v1.2.3", content: strings.Repeat(strings.Repeat("a", 64)+"  ctk_v1.2.3_darwin_arm64.tar.gz\n", 2), expected: "duplicate Release checksum"},
		{name: "missing", version: "v1.2.3", content: strings.Repeat("a", 64) + "  ctk_v1.2.3_darwin_arm64.tar.gz\n", expected: "missing Release checksum"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checksums := filepath.Join(t.TempDir(), "checksums.txt")
			writeTestFile(t, checksums, test.content)
			_, err := loadRelease(test.version, checksums)
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("loadRelease error = %v, want containing %q", err, test.expected)
			}
		})
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
