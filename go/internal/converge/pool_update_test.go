package converge

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/runtimeio"
	"github.com/kshrkznr/code-toolkit/go/internal/runtimelock"
)

type fakeDownloader struct {
	fail     map[string]bool
	attempts *[]string
}

func TestPoolUpdaterDownloadsVSCodiumArtifactOnlyFromOpenVSX(t *testing.T) {
	var attempts []string
	updater := PoolUpdater{Root: t.TempDir(), Downloader: fakeDownloader{attempts: &attempts}}
	report := Report{}
	updater.Update(context.Background(), "codium", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.id", Version: "1.0"}}}}, &report)
	if !slices.Equal(attempts, []string{"open-vsx"}) {
		t.Fatalf("attempts = %v", attempts)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func (f fakeDownloader) Download(_ context.Context, repository string, _ runtimeio.Extension, destination string) error {
	if f.attempts != nil {
		*f.attempts = append(*f.attempts, repository)
	}
	if f.fail[repository] {
		return errors.New("unavailable")
	}
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(file)
	entry, err := archive.Create("extension/package.json")
	if err == nil {
		_, err = entry.Write([]byte(`{}`))
	}
	if closeErr := archive.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}

func TestPoolUpdaterUsesCursorMarketplaceThenVisualStudioFallback(t *testing.T) {
	var attempts []string
	updater := PoolUpdater{Root: t.TempDir(), Downloader: fakeDownloader{
		fail: map[string]bool{"cursor-marketplace": true}, attempts: &attempts,
	}}
	report := Report{}
	updater.Update(context.Background(), "cursor", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.id", Version: "1.0"}}}}, &report)
	if !slices.Equal(attempts, []string{"cursor-marketplace", "visual-studio-marketplace"}) {
		t.Fatalf("attempts = %v", attempts)
	}
	if _, err := os.Stat(filepath.Join(updater.Root, "visual-studio-marketplace", "sample.id-1.0.vsix")); err != nil {
		t.Fatal(err)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func TestPoolUpdaterUsesWindsurfMarketplaceThenVisualStudioFallback(t *testing.T) {
	var attempts []string
	updater := PoolUpdater{Root: t.TempDir(), Downloader: fakeDownloader{
		fail: map[string]bool{"windsurf-marketplace": true}, attempts: &attempts,
	}}
	report := Report{}
	updater.Update(context.Background(), "devin-desktop", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.id", Version: "1.0"}}}}, &report)
	if !slices.Equal(attempts, []string{"windsurf-marketplace", "visual-studio-marketplace"}) {
		t.Fatalf("attempts = %v", attempts)
	}
	if _, err := os.Stat(filepath.Join(updater.Root, "visual-studio-marketplace", "sample.id-1.0.vsix")); err != nil {
		t.Fatal(err)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func TestHTTPDownloaderResolvesCursorMarketplaceVSIX(t *testing.T) {
	const galleryURL = "https://marketplace.cursorapi.test/gallery"
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		switch request.URL.Path {
		case "/gallery/extensionquery":
			if request.Method != http.MethodPost {
				t.Errorf("method = %s", request.Method)
			}
			var query galleryQuery
			if err := json.NewDecoder(request.Body).Decode(&query); err != nil {
				return nil, err
			}
			if len(query.Filters) != 1 || len(query.Filters[0].Criteria) != 1 || query.Filters[0].Criteria[0].FilterType != galleryExtensionNameFilter || query.Filters[0].Criteria[0].Value != "anysphere.remote-containers" {
				t.Errorf("query = %#v", query)
			}
			body = fmt.Sprintf(`{"results":[{"extensions":[{"publisher":{"publisherName":"Anysphere"},"extensionName":"remote-containers","versions":[{"version":"1.0.39","files":[]},{"version":"1.0.38","files":[{"assetType":"Microsoft.VisualStudio.Services.VSIXPackage","source":%q}]}]}]}]}`, "https://marketplace.cursorapi.test/vsix")
		case "/vsix":
			body = "cursor-vsix"
		default:
			return &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not Found", Body: io.NopCloser(strings.NewReader("not found")), Header: make(http.Header), Request: request}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: request}, nil
	})}

	destination := filepath.Join(t.TempDir(), "extension.vsix")
	downloader := HTTPDownloader{Client: client, CursorGalleryURL: galleryURL}
	if err := downloader.Download(context.Background(), "cursor-marketplace", runtimeio.Extension{ID: "anysphere.remote-containers", Version: "1.0.38"}, destination); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "cursor-vsix" {
		t.Fatalf("data=%q err=%v", data, err)
	}
}

func TestHTTPDownloaderResolvesWindsurfMarketplaceVSIX(t *testing.T) {
	const galleryURL = "https://marketplace.windsurf.test/vscode/gallery"
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		switch request.URL.Path {
		case "/vscode/gallery/extensionquery":
			var query galleryQuery
			if err := json.NewDecoder(request.Body).Decode(&query); err != nil {
				return nil, err
			}
			if len(query.Filters) != 1 || len(query.Filters[0].Criteria) != 1 || query.Filters[0].Criteria[0].Value != "editorconfig.editorconfig" {
				t.Errorf("query = %#v", query)
			}
			body = fmt.Sprintf(`{"results":[{"extensions":[{"publisher":{"publisherName":"EditorConfig"},"extensionName":"EditorConfig","versions":[{"version":"0.18.2","files":[{"assetType":"Microsoft.VisualStudio.Services.VSIXPackage","source":%q}]}]}]}]}`, "https://open-vsx.org/extension.vsix")
		case "/extension.vsix":
			body = "windsurf-vsix"
		default:
			return &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not Found", Body: io.NopCloser(strings.NewReader("not found")), Header: make(http.Header), Request: request}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: request}, nil
	})}

	destination := filepath.Join(t.TempDir(), "extension.vsix")
	downloader := HTTPDownloader{Client: client, WindsurfGalleryURL: galleryURL}
	if err := downloader.Download(context.Background(), "windsurf-marketplace", runtimeio.Extension{ID: "editorconfig.editorconfig", Version: "0.18.2"}, destination); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "windsurf-vsix" {
		t.Fatalf("data=%q err=%v", data, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestValidateCursorAssetURLRejectsAnotherHost(t *testing.T) {
	if err := validateCursorAssetURL(cursorGalleryURL, "https://example.com/extension.vsix"); err == nil {
		t.Fatal("expected foreign Cursor Marketplace asset URL to be rejected")
	}
}

func TestValidateWindsurfAssetURLAllowsSelectedOpenVSXAsset(t *testing.T) {
	if err := validateWindsurfAssetURL(windsurfGalleryURL, "https://open-vsx.org/api/sample/id/1.0/file/sample.id-1.0.vsix"); err != nil {
		t.Fatal(err)
	}
	if err := validateWindsurfAssetURL(windsurfGalleryURL, "https://example.com/extension.vsix"); err == nil {
		t.Fatal("expected foreign Windsurf Marketplace asset URL to be rejected")
	}
}

func TestPoolUpdaterFallsBackAndReplacesOldVersion(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, "visual-studio-marketplace", "sample.id-1.0.vsix")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	updater := PoolUpdater{Root: root, Downloader: fakeDownloader{fail: map[string]bool{"open-vsx": true}}}
	report := Report{}
	updater.Update(context.Background(), "kiro", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.id", Version: "2.0"}}}}, &report)
	newPath := filepath.Join(root, "visual-studio-marketplace", "sample.id-2.0.vsix")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old artifact remains: %v", err)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func TestPoolUpdaterFailureIsUnresolved(t *testing.T) {
	report := Report{}
	updater := PoolUpdater{Root: t.TempDir(), Downloader: fakeDownloader{fail: map[string]bool{"visual-studio-marketplace": true}}}
	updater.Update(context.Background(), "code", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "sample.id", Version: "1.0"}}}}, &report)
	if report.HasFailures() || len(report.Operations) != 1 || report.Operations[0].Status != Unresolved {
		t.Fatalf("report = %#v", report)
	}
}

func TestPoolUpdaterNormalizesStorageIDToLowerCase(t *testing.T) {
	root := t.TempDir()
	updater := PoolUpdater{Root: root, Downloader: fakeDownloader{fail: map[string]bool{}}}
	report := Report{}
	updater.Update(context.Background(), "kiro", runtimelock.Snapshot{Default: runtimelock.ScopeSnapshot{Extensions: []runtimeio.Extension{{ID: "emilast.LogFileHighlighter", Version: "2.8.0"}}}}, &report)
	path := filepath.Join(root, "open-vsx", "emilast.logfilehighlighter-2.8.0.vsix")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if report.HasFailures() {
		t.Fatalf("report = %#v", report)
	}
}

func TestValidateVSIXExactAcceptsPackageIdentityCaseDifference(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alefragnani.bookmarks-14.1.1.vsix")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entry, err := archive.Create("extension/package.json")
	if err == nil {
		_, err = entry.Write([]byte(`{"publisher":"alefragnani","name":"Bookmarks","version":"14.1.1"}`))
	}
	if closeErr := archive.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateVSIXExact(path, runtimeio.Extension{ID: "alefragnani.bookmarks", Version: "14.1.1"})
	if err != nil {
		t.Fatalf("ValidateVSIXExact() error = %v", err)
	}
}

func TestResolveExactUsesValidatedLocalArtifact(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "visual-studio-marketplace", "sample.ext-1.0.vsix")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entry, err := archive.Create("extension/package.json")
	if err == nil {
		_, err = entry.Write([]byte(`{"publisher":"sample","name":"ext","version":"1.0"}`))
	}
	if closeErr := archive.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}

	got, err := (PoolUpdater{Root: root}).ResolveExact("code", runtimeio.Extension{ID: "sample.ext", Version: "1.0"})
	if err != nil || got != path {
		t.Fatalf("ResolveExact() = %q, %v", got, err)
	}
}
