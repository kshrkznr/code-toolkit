package converge

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"code-toolkit/internal/runtimeio"
	"code-toolkit/internal/runtimelock"
)

type Downloader interface {
	Download(context.Context, string, runtimeio.Extension, string) error
}

type HTTPDownloader struct{ Client *http.Client }

func (d HTTPDownloader) Download(ctx context.Context, repository string, extension runtimeio.Extension, destination string) error {
	parts := strings.SplitN(extension.ID, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid Extension ID %q", extension.ID)
	}
	var url string
	switch repository {
	case "visual-studio-marketplace":
		url = fmt.Sprintf("https://marketplace.visualstudio.com/_apis/public/gallery/publishers/%s/vsextensions/%s/%s/vspackage", parts[0], parts[1], extension.Version)
	case "open-vsx":
		url = fmt.Sprintf("https://open-vsx.org/api/%s/%s/%s/file/%s-%s.vsix", parts[0], parts[1], extension.Version, extension.ID, extension.Version)
	default:
		return fmt.Errorf("unsupported Extension repository %q", repository)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "code-toolkit-go")
	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("download returned %s", response.Status)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, response.Body)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

type PoolUpdater struct {
	Root       string
	Downloader Downloader
}

func (u PoolUpdater) Update(ctx context.Context, platform string, snapshot runtimelock.Snapshot, report *Report) {
	seen := map[string]runtimeio.Extension{}
	for _, extension := range snapshot.Default.Extensions {
		seen[extension.ID+"@"+extension.Version] = extension
	}
	for _, profile := range snapshot.Profiles {
		for _, extension := range profile.Extensions {
			seen[extension.ID+"@"+extension.Version] = extension
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		u.updateOne(ctx, platform, seen[key], report)
	}
}

func (u PoolUpdater) updateOne(ctx context.Context, platform string, extension runtimeio.Extension, report *Report) {
	subject := extension.ID + "@" + extension.Version
	if extension.ID == "" || extension.Version == "" || strings.ContainsAny(extension.ID+extension.Version, `/\\`) {
		report.Add(Operation{Action: "update Extension Pool", Subject: subject, Status: Unresolved, Err: fmt.Errorf("versioned Extension observation required")})
		return
	}
	repositories := []string{platformRepository(platform)}
	if repositories[0] != "visual-studio-marketplace" {
		repositories = append(repositories, "visual-studio-marketplace")
	}
	artifact := poolArtifactName(extension)
	for _, repository := range repositories {
		matches, _ := poolArtifacts(filepath.Join(u.Root, repository), extension.ID)
		for _, match := range matches {
			if strings.EqualFold(filepath.Base(match), artifact) {
				report.Add(Operation{Action: "retain Extension Pool artifact", Subject: subject, Status: Completed})
				return
			}
		}
	}
	downloader := u.Downloader
	if downloader == nil {
		downloader = HTTPDownloader{}
	}
	for _, repository := range repositories {
		directory := filepath.Join(u.Root, repository)
		if err := os.MkdirAll(directory, 0o755); err != nil {
			continue
		}
		staging, err := os.MkdirTemp(directory, ".download-")
		if err != nil {
			continue
		}
		path := filepath.Join(staging, artifact)
		err = downloader.Download(ctx, repository, extension, path)
		if err == nil {
			err = ValidateVSIX(path)
		}
		if err == nil {
			err = os.Rename(path, filepath.Join(directory, artifact))
		}
		_ = os.RemoveAll(staging)
		if err != nil {
			continue
		}
		matches, _ := poolArtifacts(directory, extension.ID)
		for _, old := range matches {
			if filepath.Base(old) != artifact {
				_ = os.Remove(old)
			}
		}
		report.Add(Operation{Action: "store Extension Pool artifact", Subject: subject, Status: Completed})
		return
	}
	report.Add(Operation{Action: "update Extension Pool", Subject: subject, Status: Unresolved, Err: fmt.Errorf("artifact unavailable from configured repositories")})
}

// EnsureExact returns a validated exact-version Pool artifact, downloading it
// through the Platform repository order when necessary. Unlike ordinary Pool
// update, failure is returned to callers such as Archive creation.
func (u PoolUpdater) EnsureExact(ctx context.Context, platform string, extension runtimeio.Extension) (string, error) {
	if extension.ID == "" || extension.Version == "" || strings.ContainsAny(extension.ID+extension.Version, `/\\`) {
		return "", fmt.Errorf("versioned Extension observation required: %s@%s", extension.ID, extension.Version)
	}
	repositories := []string{platformRepository(platform)}
	if repositories[0] != "visual-studio-marketplace" {
		repositories = append(repositories, "visual-studio-marketplace")
	}
	artifact := poolArtifactName(extension)
	for _, repository := range repositories {
		matches, _ := poolArtifacts(filepath.Join(u.Root, repository), extension.ID)
		for _, path := range matches {
			if strings.EqualFold(filepath.Base(path), artifact) {
				if err := ValidateVSIXExact(path, extension); err == nil {
					return path, nil
				}
			}
		}
	}
	downloader := u.Downloader
	if downloader == nil {
		downloader = HTTPDownloader{}
	}
	for _, repository := range repositories {
		directory := filepath.Join(u.Root, repository)
		if err := os.MkdirAll(directory, 0755); err != nil {
			continue
		}
		staging, err := os.MkdirTemp(directory, ".download-")
		if err != nil {
			continue
		}
		path := filepath.Join(staging, artifact)
		err = downloader.Download(ctx, repository, extension, path)
		if err == nil {
			err = ValidateVSIXExact(path, extension)
		}
		final := filepath.Join(directory, artifact)
		if err == nil {
			_ = os.Remove(final)
			err = os.Rename(path, final)
		}
		_ = os.RemoveAll(staging)
		if err == nil {
			return final, nil
		}
	}
	return "", fmt.Errorf("artifact unavailable from configured repositories: %s@%s", extension.ID, extension.Version)
}

func ValidateVSIX(path string) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("invalid VSIX archive: %w", err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name == "extension/package.json" || file.Name == "extension.vsixmanifest" {
			return nil
		}
	}
	return fmt.Errorf("invalid VSIX archive: manifest not found")
}

func ValidateVSIXExact(path string, expected runtimeio.Extension) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("invalid VSIX archive: %w", err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != "extension/package.json" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		var metadata struct {
			Publisher string `json:"publisher"`
			Name      string `json:"name"`
			Version   string `json:"version"`
		}
		decodeErr := json.NewDecoder(reader).Decode(&metadata)
		closeErr := reader.Close()
		if decodeErr != nil {
			return fmt.Errorf("parse VSIX package metadata: %w", decodeErr)
		}
		if closeErr != nil {
			return closeErr
		}
		identity := metadata.Publisher + "." + metadata.Name
		// VS Code reports installed Extension IDs in lowercase, while a VSIX
		// package may retain publisher/name casing (for example,
		// alefragnani.Bookmarks). Treat only this external metadata boundary as
		// case-insensitive; Lock and Ingredient comparisons remain exact.
		if !strings.EqualFold(identity, expected.ID) || metadata.Version != expected.Version {
			return fmt.Errorf("VSIX metadata mismatch: got %s@%s, expected %s@%s", identity, metadata.Version, expected.ID, expected.Version)
		}
		return nil
	}
	return fmt.Errorf("VSIX package metadata not found")
}
