package converge

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
)

type Downloader interface {
	Download(context.Context, string, runtimeio.Extension, string) error
}

const (
	cursorGalleryURL           = "https://marketplace.cursorapi.com/_apis/public/gallery"
	windsurfGalleryURL         = "https://marketplace.windsurf.com/vscode/gallery"
	galleryExtensionNameFilter = 7
	galleryIncludeExactAssets  = 1 | 2 | 16 | 128 // versions, files, version properties, and asset URI
)

type HTTPDownloader struct {
	Client             *http.Client
	CursorGalleryURL   string
	WindsurfGalleryURL string
}

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
	case "cursor-marketplace":
		var err error
		url, err = d.cursorVSIXURL(ctx, extension)
		if err != nil {
			return err
		}
	case "windsurf-marketplace":
		var err error
		url, err = d.windsurfVSIXURL(ctx, extension)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported Extension repository %q", repository)
	}
	return d.downloadURL(ctx, url, destination)
}

func (d HTTPDownloader) downloadURL(ctx context.Context, source, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
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

type galleryQuery struct {
	Filters    []galleryFilter `json:"filters"`
	AssetTypes []string        `json:"assetTypes"`
	Flags      int             `json:"flags"`
}

type galleryFilter struct {
	Criteria  []galleryCriterion `json:"criteria"`
	Page      int                `json:"pageNumber"`
	PageSize  int                `json:"pageSize"`
	SortBy    int                `json:"sortBy"`
	SortOrder int                `json:"sortOrder"`
}

type galleryCriterion struct {
	FilterType int    `json:"filterType"`
	Value      string `json:"value"`
}

type galleryResponse struct {
	Results []struct {
		Extensions []struct {
			Publisher struct {
				Name string `json:"publisherName"`
			} `json:"publisher"`
			Name     string `json:"extensionName"`
			Versions []struct {
				Version string `json:"version"`
				Files   []struct {
					AssetType string `json:"assetType"`
					Source    string `json:"source"`
				} `json:"files"`
			} `json:"versions"`
		} `json:"extensions"`
	} `json:"results"`
}

func (d HTTPDownloader) cursorVSIXURL(ctx context.Context, extension runtimeio.Extension) (string, error) {
	gallery := d.CursorGalleryURL
	if gallery == "" {
		gallery = cursorGalleryURL
	}
	return d.galleryVSIXURL(ctx, "Cursor Marketplace", gallery, extension, func(source string) error {
		return validateCursorAssetURL(gallery, source)
	})
}

func (d HTTPDownloader) windsurfVSIXURL(ctx context.Context, extension runtimeio.Extension) (string, error) {
	gallery := d.WindsurfGalleryURL
	if gallery == "" {
		gallery = windsurfGalleryURL
	}
	return d.galleryVSIXURL(ctx, "Windsurf Marketplace", gallery, extension, func(source string) error {
		return validateWindsurfAssetURL(gallery, source)
	})
}

func (d HTTPDownloader) galleryVSIXURL(ctx context.Context, marketplace, gallery string, extension runtimeio.Extension, validateAsset func(string) error) (string, error) {
	payload, err := json.Marshal(galleryQuery{
		Filters: []galleryFilter{{
			Criteria: []galleryCriterion{{FilterType: galleryExtensionNameFilter, Value: extension.ID}},
			Page:     1, PageSize: 10,
		}},
		AssetTypes: []string{},
		Flags:      galleryIncludeExactAssets,
	})
	if err != nil {
		return "", fmt.Errorf("encode %s query: %w", marketplace, err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(gallery, "/")+"/extensionquery", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "code-toolkit-go")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json;api-version=3.0-preview.1")
	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("%s query returned %s", marketplace, response.Status)
	}
	var result galleryResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode %s query: %w", marketplace, err)
	}
	for _, group := range result.Results {
		for _, candidate := range group.Extensions {
			identity := candidate.Publisher.Name + "." + candidate.Name
			if !strings.EqualFold(identity, extension.ID) {
				continue
			}
			for _, version := range candidate.Versions {
				if version.Version != extension.Version {
					continue
				}
				for _, file := range version.Files {
					if file.AssetType != "Microsoft.VisualStudio.Services.VSIXPackage" {
						continue
					}
					if err := validateAsset(file.Source); err != nil {
						return "", err
					}
					return file.Source, nil
				}
			}
		}
	}
	return "", fmt.Errorf("%s artifact unavailable: %s@%s", marketplace, extension.ID, extension.Version)
}

func validateCursorAssetURL(gallery, source string) error {
	galleryURL, err := url.Parse(gallery)
	if err != nil {
		return fmt.Errorf("parse Cursor Marketplace URL: %w", err)
	}
	assetURL, err := url.Parse(source)
	if err != nil {
		return fmt.Errorf("parse Cursor Marketplace asset URL: %w", err)
	}
	if (galleryURL.Scheme != "http" && galleryURL.Scheme != "https") ||
		(assetURL.Scheme != "http" && assetURL.Scheme != "https") ||
		(galleryURL.Scheme == "https" && assetURL.Scheme != "https") ||
		!strings.EqualFold(assetURL.Host, galleryURL.Host) {
		return fmt.Errorf("reject Cursor Marketplace asset URL: %s", source)
	}
	return nil
}

func validateWindsurfAssetURL(gallery, source string) error {
	galleryURL, err := url.Parse(gallery)
	if err != nil {
		return fmt.Errorf("parse Windsurf Marketplace URL: %w", err)
	}
	assetURL, err := url.Parse(source)
	if err != nil {
		return fmt.Errorf("parse Windsurf Marketplace asset URL: %w", err)
	}
	allowedHost := strings.EqualFold(assetURL.Host, galleryURL.Host) || strings.EqualFold(assetURL.Host, "open-vsx.org")
	if galleryURL.Scheme != "https" || assetURL.Scheme != "https" || !allowedHost {
		return fmt.Errorf("reject Windsurf Marketplace asset URL: %s", source)
	}
	return nil
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
	repositories := platformRepositories(platform)
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
	if path, err := u.ResolveExact(platform, extension); err == nil {
		return path, nil
	}
	repositories := platformRepositories(platform)
	artifact := poolArtifactName(extension)
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

// ResolveExact returns an already cached, validated exact-version Pool
// artifact without contacting any Repository.
func (u PoolUpdater) ResolveExact(platform string, extension runtimeio.Extension) (string, error) {
	if extension.ID == "" || extension.Version == "" || strings.ContainsAny(extension.ID+extension.Version, `/\\`) {
		return "", fmt.Errorf("versioned Extension observation required: %s@%s", extension.ID, extension.Version)
	}
	artifact := poolArtifactName(extension)
	var validationErr error
	for _, repository := range platformRepositories(platform) {
		matches, _ := poolArtifacts(filepath.Join(u.Root, repository), extension.ID)
		for _, path := range matches {
			if !strings.EqualFold(filepath.Base(path), artifact) {
				continue
			}
			if err := ValidateVSIXExact(path, extension); err != nil {
				validationErr = fmt.Errorf("validate Extension Pool artifact %s: %w", path, err)
				continue
			}
			return path, nil
		}
	}
	if validationErr != nil {
		return "", validationErr
	}
	return "", fmt.Errorf("exact Extension Pool artifact unavailable: %s@%s", extension.ID, extension.Version)
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
