package docbundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"unicode"
)

type Bundle struct {
	manifest   Manifest
	bootstrap  []byte
	documents  map[string][]byte
	byIdentity map[string][]string
	byAlias    map[string]string
}

type Candidate struct {
	Identity string
	Path     string
	Title    string
	Aliases  []string
	Score    int
}

func Open(archive []byte) (*Bundle, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open Documentation Bundle: %w", err)
	}
	entries := map[string][]byte{}
	casePaths := map[string]string{}
	for _, file := range reader.File {
		if file.Name != path.Clean(file.Name) || !validArchivePath(file.Name) {
			return nil, fmt.Errorf("invalid Documentation Bundle path: %s", file.Name)
		}
		if !file.Mode().IsRegular() {
			return nil, fmt.Errorf("Documentation Bundle entry is not a regular file: %s", file.Name)
		}
		if _, exists := entries[file.Name]; exists {
			return nil, fmt.Errorf("duplicate Documentation Bundle path: %s", file.Name)
		}
		folded := strings.ToLower(file.Name)
		if previous, exists := casePaths[folded]; exists {
			return nil, fmt.Errorf("case-insensitive Documentation Bundle collision: %s and %s", previous, file.Name)
		}
		stream, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open Documentation Bundle entry %s: %w", file.Name, err)
		}
		content, readErr := io.ReadAll(stream)
		closeErr := stream.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read Documentation Bundle entry %s: %w", file.Name, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close Documentation Bundle entry %s: %w", file.Name, closeErr)
		}
		entries[file.Name] = content
		casePaths[folded] = file.Name
	}
	manifestBytes, ok := entries[ManifestPath]
	if !ok {
		return nil, fmt.Errorf("Documentation Bundle Manifest is missing")
	}
	manifest, err := decodeManifest(manifestBytes)
	if err != nil {
		return nil, err
	}
	bootstrap, ok := entries[manifest.Bootstrap.Path]
	if !ok || digest(bootstrap) != manifest.Bootstrap.SHA256 {
		return nil, fmt.Errorf("Documentation Bootstrap is missing or does not match its digest")
	}
	documents := make(map[string][]byte, len(manifest.Documents))
	byIdentity := map[string][]string{}
	byAlias := map[string]string{}
	contentEntries := map[string][]byte{manifest.Bootstrap.Path: bootstrap}
	for _, document := range manifest.Documents {
		content, ok := entries[document.Path]
		if !ok || digest(content) != document.SHA256 {
			return nil, fmt.Errorf("Documentation Bundle document is missing or does not match its digest: %s", document.Path)
		}
		if _, exists := documents[document.Path]; exists {
			return nil, fmt.Errorf("duplicate document in Documentation Manifest: %s", document.Path)
		}
		documents[document.Path] = content
		contentEntries[document.Path] = content
		if document.Identity != "" {
			byIdentity[document.Identity] = append(byIdentity[document.Identity], document.Path)
		}
		for _, alias := range document.Aliases {
			if previous, exists := byAlias[alias]; exists {
				return nil, fmt.Errorf("duplicate documentation Node alias %s: %s and %s", alias, previous, document.Path)
			}
			byAlias[alias] = document.Path
		}
	}
	if aggregateDigest(contentEntries) != manifest.ContentSHA256 {
		return nil, fmt.Errorf("Documentation Bundle aggregate digest mismatch")
	}
	for name := range entries {
		if name == ManifestPath || name == manifest.Bootstrap.Path {
			continue
		}
		if _, ok := documents[name]; !ok {
			return nil, fmt.Errorf("Documentation Bundle contains unmanifested entry: %s", name)
		}
	}
	return &Bundle{manifest: manifest, bootstrap: bootstrap, documents: documents, byIdentity: byIdentity, byAlias: byAlias}, nil
}

func (bundle *Bundle) Manifest() Manifest {
	return bundle.manifest
}

func (bundle *Bundle) Bootstrap() []byte {
	return bytes.Clone(bundle.bootstrap)
}

func (bundle *Bundle) Show(reference string) ([]byte, error) {
	value, fragment := splitReference(reference)
	paths := []string{}
	if _, ok := bundle.documents[value]; ok {
		paths = append(paths, value)
	}
	paths = append(paths, bundle.byIdentity[value]...)
	paths = compactSorted(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("documentation reference not found: %s", value)
	}
	if len(paths) > 1 {
		return nil, fmt.Errorf("ambiguous documentation reference %s: %s", value, strings.Join(paths, ", "))
	}
	content := rewriteLinks(string(bundle.documents[paths[0]]), paths[0], bundle.documents, repositoryReference(bundle.manifest.Repository, Metadata{Revision: bundle.manifest.Revision, Tag: bundle.manifest.Tag}))
	if fragment != "" {
		section, err := selectHeading(content, fragment)
		if err != nil {
			return nil, fmt.Errorf("show %s: %w", reference, err)
		}
		content = section
	}
	return []byte(content), nil
}

func (bundle *Bundle) ShowNode(alias string) ([]byte, error) {
	target, ok := bundle.byAlias[alias]
	if !ok {
		return nil, fmt.Errorf("documentation Node alias not found: %s", alias)
	}
	return bundle.Show(target)
}

func (bundle *Bundle) Resolve(terms []string) []Candidate {
	normalized := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term != "" {
			normalized = append(normalized, term)
		}
	}
	if len(normalized) == 0 {
		return nil
	}
	result := []Candidate{}
	for _, document := range bundle.manifest.Documents {
		identity := strings.ToLower(document.Identity)
		documentPath := strings.ToLower(document.Path)
		aliases := strings.ToLower(strings.Join(document.Aliases, "\n"))
		metadata := strings.ToLower(document.Title + "\n" + strings.Join(document.Headings, "\n"))
		content := strings.ToLower(string(bundle.documents[document.Path]))
		score := -1
		if len(normalized) == 1 && (normalized[0] == identity || normalized[0] == documentPath) {
			score = 0
		} else if len(normalized) == 1 && containsExactFold(document.Aliases, normalized[0]) {
			score = 1
		} else if containsAll(identity+"\n"+documentPath+"\n"+aliases+"\n"+metadata, normalized) {
			score = 2
		} else if containsAll(content, normalized) {
			score = 3
		}
		if score >= 0 {
			result = append(result, Candidate{Identity: document.Identity, Path: document.Path, Title: document.Title, Aliases: append([]string(nil), document.Aliases...), Score: score})
		}
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Score != result[right].Score {
			return result[left].Score < result[right].Score
		}
		return result[left].Path < result[right].Path
	})
	return result
}

func decodeManifest(content []byte) (Manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse Documentation Manifest: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			err = fmt.Errorf("multiple JSON values are not supported")
		}
		return Manifest{}, fmt.Errorf("parse Documentation Manifest: %w", err)
	}
	if manifest.FormatVersion != formatVersion {
		return Manifest{}, fmt.Errorf("unsupported Documentation Manifest format-version %d", manifest.FormatVersion)
	}
	if manifest.Bootstrap.Path != BootstrapPath {
		return Manifest{}, fmt.Errorf("unsupported Documentation Bootstrap path: %s", manifest.Bootstrap.Path)
	}
	return manifest, nil
}

func splitReference(reference string) (string, string) {
	value, fragment, found := strings.Cut(reference, "#")
	if !found {
		return reference, ""
	}
	return value, fragment
}

func selectHeading(content, fragment string) (string, error) {
	lines := strings.SplitAfter(content, "\n")
	start := -1
	level := 0
	for index, line := range lines {
		match := headingPattern.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
		if match == nil {
			continue
		}
		if start < 0 {
			if markdownAnchor(match[2]) == fragment {
				start = index
				level = len(match[1])
			}
			continue
		}
		if len(match[1]) <= level {
			return strings.Join(lines[start:index], ""), nil
		}
	}
	if start < 0 {
		return "", fmt.Errorf("heading not found: #%s", fragment)
	}
	return strings.Join(lines[start:], ""), nil
}

func markdownAnchor(value string) string {
	var output strings.Builder
	previousHyphen := false
	for _, character := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(character), unicode.IsDigit(character), character == '_', character == '-':
			output.WriteRune(character)
			previousHyphen = character == '-'
		case unicode.IsSpace(character):
			if !previousHyphen && output.Len() > 0 {
				output.WriteByte('-')
				previousHyphen = true
			}
		}
	}
	return strings.Trim(output.String(), "-")
}

func containsAll(value string, terms []string) bool {
	for _, term := range terms {
		if !containsTerm(value, term) {
			return false
		}
	}
	return true
}

func containsTerm(value, term string) bool {
	return strings.Contains(value, term)
}

func containsExactFold(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(value, want) {
			return true
		}
	}
	return false
}

func compactSorted(values []string) []string {
	sort.Strings(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func validArchivePath(value string) bool {
	return value != "" && !strings.HasPrefix(value, "/") && !strings.Contains(value, `\`) && value != "." && value != ".." && !strings.HasPrefix(value, "../")
}
