package docbundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
	byPath     map[string]string
	byAlias    map[string]string
}

type Candidate struct {
	Identity string
	Path     string
	Title    string
	Aliases  []string
	Score    int
	Matched  []string
}

func OpenExecutable(executable string) (*Bundle, error) {
	content, err := os.ReadFile(executable)
	if err != nil {
		return nil, fmt.Errorf("read executable Documentation Bundle: %w", err)
	}
	return Open(content)
}

func AppendExecutable(executable string, archive []byte) error {
	if _, err := Open(archive); err != nil {
		return fmt.Errorf("validate Documentation Bundle before append: %w", err)
	}
	if existing, err := OpenExecutable(executable); err == nil && existing != nil {
		return fmt.Errorf("executable already contains a Documentation Bundle: %s", executable)
	}
	file, err := os.OpenFile(executable, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return fmt.Errorf("open executable for Documentation Bundle append: %w", err)
	}
	if _, err := file.Write(archive); err != nil {
		file.Close()
		return fmt.Errorf("append Documentation Bundle: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync executable Documentation Bundle: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close executable Documentation Bundle: %w", err)
	}
	return nil
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
	byPath := map[string]string{}
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
		byPath[strings.ToLower(document.Path)] = document.Path
		contentEntries[document.Path] = content
		if document.Identity != "" {
			byIdentity[strings.ToLower(document.Identity)] = append(byIdentity[strings.ToLower(document.Identity)], document.Path)
		}
		for _, alias := range document.Aliases {
			foldedAlias := strings.ToLower(alias)
			if previous, exists := byAlias[foldedAlias]; exists {
				return nil, fmt.Errorf("duplicate documentation Node alias %s: %s and %s", alias, previous, document.Path)
			}
			byAlias[foldedAlias] = document.Path
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
	return &Bundle{manifest: manifest, bootstrap: bootstrap, documents: documents, byIdentity: byIdentity, byPath: byPath, byAlias: byAlias}, nil
}

func (bundle *Bundle) Manifest() Manifest {
	return bundle.manifest
}

func (bundle *Bundle) Bootstrap() []byte {
	return bytes.Clone(bundle.bootstrap)
}

func (bundle *Bundle) RepositoryReference() string {
	return repositoryReference(bundle.manifest.Repository, Metadata{Revision: bundle.manifest.Revision, Tag: bundle.manifest.Tag})
}

func (bundle *Bundle) Show(reference string) ([]byte, error) {
	value, fragment := splitReference(reference)
	paths := []string{}
	if documentPath, ok := bundle.byPath[strings.ToLower(value)]; ok {
		paths = append(paths, documentPath)
	}
	paths = append(paths, bundle.byIdentity[strings.ToLower(value)]...)
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
	target, ok := bundle.byAlias[strings.ToLower(alias)]
	if !ok {
		return nil, fmt.Errorf("documentation Node alias not found: %s", alias)
	}
	return bundle.Show(target)
}

func (bundle *Bundle) Resolve(terms []string) []Candidate {
	normalized := searchTokens(strings.Join(terms, " "))
	if len(normalized) == 0 {
		return nil
	}
	result := []Candidate{}
	for _, document := range bundle.manifest.Documents {
		exactQuery := strings.ToLower(strings.TrimSpace(strings.Join(terms, " ")))
		if exactQuery == strings.ToLower(document.Identity) || exactQuery == strings.ToLower(document.Path) {
			result = append(result, Candidate{Identity: document.Identity, Path: document.Path, Title: document.Title, Aliases: append([]string(nil), document.Aliases...), Score: 1_000_000, Matched: normalized})
			continue
		}
		if containsExactFold(document.Aliases, exactQuery) {
			result = append(result, Candidate{Identity: document.Identity, Path: document.Path, Title: document.Title, Aliases: append([]string(nil), document.Aliases...), Score: 900_000, Matched: normalized})
			continue
		}

		zones := []struct {
			weight int
			tokens []string
		}{
			{weight: 1_200, tokens: searchTokens(document.Title)},
			{weight: 800, tokens: searchTokens(document.Identity + " " + document.Path + " " + strings.Join(document.Aliases, " "))},
			{weight: 400, tokens: searchTokens(strings.Join(document.Headings, " "))},
			{weight: 10, tokens: searchTokens(string(bundle.documents[document.Path]))},
		}
		matched := []string{}
		zoneScore := 0
		for _, query := range normalized {
			termScore := 0
			for _, zone := range zones {
				if tokenMatchesAny(query, zone.tokens) {
					termScore += zone.weight
				}
			}
			if termScore > 0 {
				matched = append(matched, query)
				zoneScore += termScore
			}
		}
		if len(matched) >= minimumTokenMatches(len(normalized)) {
			score := len(matched)*1_000 + zoneScore
			result = append(result, Candidate{Identity: document.Identity, Path: document.Path, Title: document.Title, Aliases: append([]string(nil), document.Aliases...), Score: score, Matched: matched})
		}
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Score != result[right].Score {
			return result[left].Score > result[right].Score
		}
		return result[left].Path < result[right].Path
	})
	return result
}

var searchStopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
	"be": true, "by": true, "for": true, "from": true, "how": true, "in": true,
	"into": true, "is": true, "it": true, "of": true, "on": true, "or": true,
	"the": true, "to": true, "with": true,
}

func searchTokens(value string) []string {
	tokens := []string{}
	seen := map[string]bool{}
	for _, token := range strings.FieldsFunc(strings.ToLower(value), func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsDigit(character)
	}) {
		if searchStopWords[token] || seen[token] {
			continue
		}
		seen[token] = true
		tokens = append(tokens, token)
	}
	return tokens
}

func tokenMatchesAny(query string, candidates []string) bool {
	for _, candidate := range candidates {
		if query == candidate {
			return true
		}
		if len([]rune(query)) >= 4 && len([]rune(candidate)) >= 4 && commonRunePrefix(query, candidate) >= 4 {
			return true
		}
	}
	return false
}

func commonRunePrefix(left, right string) int {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	limit := len(leftRunes)
	if len(rightRunes) < limit {
		limit = len(rightRunes)
	}
	for index := 0; index < limit; index++ {
		if leftRunes[index] != rightRunes[index] {
			return index
		}
	}
	return limit
}

func minimumTokenMatches(count int) int {
	if count <= 2 {
		return count
	}
	return (count*3 + 4) / 5
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
	anchors := map[string]int{}
	for index, line := range lines {
		match := headingPattern.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
		if match == nil {
			continue
		}
		base := markdownAnchor(match[2])
		anchor := base
		if duplicate := anchors[base]; duplicate > 0 {
			anchor = fmt.Sprintf("%s-%d", base, duplicate)
		}
		anchors[base]++
		if start < 0 {
			if anchor == fragment {
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
