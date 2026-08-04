package archive

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/converge"
	"github.com/kshrkznr/code-toolkit/go/internal/cookbook"
	"github.com/kshrkznr/code-toolkit/go/internal/distribution"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeartifact"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

func TestCreateValidatesAndPreservesExactProfileVersions(t *testing.T) {
	root := t.TempDir()
	cookbookRoot := filepath.Join(root, "cookbook")
	distPath := filepath.Join(root, "dist", "demo")
	for _, directory := range []string{filepath.Join(distPath, ".data"), filepath.Join(distPath, ".ext"), filepath.Join(distPath, ".meta"), filepath.Join(cookbookRoot, "ingredient")} {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	recipePath := filepath.Join(distPath, ".meta", "recipe.yaml")
	recipeData := []byte("name: demo\nos: macos\nplatform: code\nprofile: [work]\nconfig:\n  dist-strategy:\n    lock-mode: reuse\n  profile-strategy:\n    work:\n      settings: profile\n")
	if err := os.WriteFile(recipePath, recipeData, 0644); err != nil {
		t.Fatal(err)
	}
	plan, err := (cookbook.Repository{Root: filepath.Join(cookbookRoot, "ingredient")}).Resolve(recipePath)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := runtimelock.Snapshot{FormatVersion: runtimelock.FormatVersion, RecipeName: "demo", Platform: "code", ObservedAt: time.Now(), Default: runtimelock.ScopeSnapshot{Settings: settings.Document{}, Extensions: []runtimeio.Extension{{ID: "sample.ext", Version: "1.0"}}}, Profiles: []runtimelock.ScopeSnapshot{{Name: "work", Settings: settings.Document{}, Extensions: []runtimeio.Extension{{ID: "sample.ext", Version: "2.0"}}, Inheritance: plan.Profiles[0].Inheritance}}}
	if err := (runtimelock.Store{}).Seal(distPath, recipePath, snapshot, plan); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distPath, "run.sh"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	poolRoot := filepath.Join(root, ".vsix", "visual-studio-marketplace")
	writeVSIX(t, filepath.Join(poolRoot, "sample.ext-1.0.vsix"))
	writeVSIX(t, filepath.Join(poolRoot, "sample.ext-2.0.vsix"))
	dist, err := distribution.Load(filepath.Join(root, "dist"), "demo")
	if err != nil {
		t.Fatal(err)
	}
	service := Service{Cookbook: cookbook.Repository{Root: filepath.Join(cookbookRoot, "ingredient")}, Locks: runtimelock.Store{}, Pool: converge.PoolUpdater{Root: filepath.Join(root, ".vsix")}, Now: func() time.Time { return time.Unix(1, 0).UTC() }}
	result, err := service.Create(context.Background(), filepath.Join(root, "archive"), dist, Options{OnConflict: "abort"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Bundle.Manifest.Extensions) != 2 || len(result.Warnings) != 1 {
		t.Fatalf("result = %#v", result)
	}
	if _, err := os.Stat(filepath.Join(result.Bundle.Path, "launch-override", "run.sh")); err != nil {
		t.Fatal(err)
	}
	originalManifest := result.Bundle.Manifest
	result.Bundle.Manifest.Extensions = result.Bundle.Manifest.Extensions[:1]
	if err := writeManifest(result.Bundle.Path, result.Bundle.Manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(result.Bundle.Path); err == nil {
		t.Fatal("expected Lock and Archive Extension mismatch")
	}
	if err := writeManifest(result.Bundle.Path, originalManifest); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(result.Bundle.Path, "vsix", "sample.ext-1.0.vsix"))
	if err != nil {
		t.Fatal(err)
	}
	data[0] ^= 1
	if err := os.WriteFile(filepath.Join(result.Bundle.Path, "vsix", "sample.ext-1.0.vsix"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(result.Bundle.Path); err == nil {
		t.Fatal("expected integrity failure")
	}
}

func TestVerifyTreatsTasksEnvelopeAsSemanticEmpty(t *testing.T) {
	expected := runtimelock.Snapshot{RecipeName: "demo", Platform: "code", Default: runtimelock.ScopeSnapshot{Tasks: runtimeartifact.Object{
		"version": "2.0.0", "tasks": []any{}, "inputs": []any{},
	}}}
	actual := runtimelock.Snapshot{RecipeName: "demo", Platform: "code", Default: runtimelock.ScopeSnapshot{Tasks: runtimeartifact.Object{}}}
	if err := Verify(expected, actual); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRejectsMissingExtensionVersion(t *testing.T) {
	snapshot := runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.ext"}}}}
	if _, err := uniqueExtensions(snapshot); err == nil {
		t.Fatal("expected version error")
	}
}

func writeVSIX(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("extension/package.json")
	version := "1.0"
	if strings.Contains(filepath.Base(path), "-2.0.") {
		version = "2.0"
	}
	if err == nil {
		_, err = entry.Write([]byte(fmt.Sprintf(`{"publisher":"sample","name":"ext","version":"%s"}`, version)))
	}
	if closeErr := writer.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
}
