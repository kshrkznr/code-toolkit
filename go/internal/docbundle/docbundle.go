package docbundle

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

const (
	DefinitionPath = "doc/documentation-bundle.yaml"
	BootstrapPath  = ".ctk-docs/bootstrap.md"
	ManifestPath   = ".ctk-docs/manifest.json"
	formatVersion  = 1
)

var (
	aliasPattern           = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	headingPattern         = regexp.MustCompile(`^(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	linkPattern            = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
	includeDocumentPattern = regexp.MustCompile(`^\{\{ include-document "([^"]+)" \}\}$`)
	includeRangePattern    = regexp.MustCompile(`^\{\{ include-range "([^"]+)" from="([^"]+)" before="([^"]+)" \}\}$`)
	reservedAliases        = map[string]bool{"resolve": true, "show": true, "export": true}
)

type Definition struct {
	FormatVersion     int               `yaml:"format-version"`
	Repository        string            `yaml:"repository"`
	Documents         DocumentSelection `yaml:"documents"`
	Nodes             map[string]string `yaml:"nodes"`
	BootstrapTemplate string            `yaml:"bootstrap-template"`
}

type DocumentSelection struct {
	Files   []string `yaml:"files"`
	Trees   []string `yaml:"trees"`
	Exclude []string `yaml:"exclude"`
}

type Metadata struct {
	Version  string
	Revision string
	Tag      string
}

type Manifest struct {
	FormatVersion    int                `json:"format-version"`
	Version          string             `json:"ctk-version"`
	Revision         string             `json:"source-revision"`
	Tag              string             `json:"release-tag,omitempty"`
	Repository       string             `json:"repository"`
	DefinitionSHA256 string             `json:"definition-sha256"`
	ContentSHA256    string             `json:"content-sha256"`
	Bootstrap        GeneratedDocument  `json:"bootstrap"`
	Documents        []ManifestDocument `json:"documents"`
}

type GeneratedDocument struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ManifestDocument struct {
	Path     string   `json:"path"`
	Identity string   `json:"identity,omitempty"`
	Title    string   `json:"title"`
	Headings []string `json:"headings"`
	Aliases  []string `json:"aliases,omitempty"`
	SHA256   string   `json:"sha256"`
}

type Result struct {
	Archive   []byte
	Manifest  Manifest
	Bootstrap []byte
	Documents map[string][]byte
}

func Generate(repositoryRoot string, metadata Metadata) (Result, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Result{}, fmt.Errorf("resolve repository root: %w", err)
	}
	definitionBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(DefinitionPath)))
	if err != nil {
		return Result{}, fmt.Errorf("read Bundle Definition: %w", err)
	}
	definition, err := decodeDefinition(definitionBytes)
	if err != nil {
		return Result{}, err
	}
	documents, err := selectDocuments(root, definition)
	if err != nil {
		return Result{}, err
	}
	aliases, err := validateNodes(definition.Nodes, documents)
	if err != nil {
		return Result{}, err
	}
	if err := validateLinks(root, documents); err != nil {
		return Result{}, err
	}

	templatePath, err := cleanRelative(definition.BootstrapTemplate)
	if err != nil {
		return Result{}, fmt.Errorf("invalid Bootstrap template path: %w", err)
	}
	templateBytes, err := readRegular(root, templatePath, false)
	if err != nil {
		return Result{}, fmt.Errorf("read Bootstrap template: %w", err)
	}
	bootstrap, err := renderBootstrap(templateBytes, documents, definition.Repository, metadata)
	if err != nil {
		return Result{}, err
	}

	manifestDocuments := make([]ManifestDocument, 0, len(documents))
	identities := map[string]string{}
	for _, documentPath := range sortedKeys(documents) {
		content := documents[documentPath]
		identity, title, headings := inspectMarkdown(content)
		if identity != "" {
			if previous, ok := identities[identity]; ok {
				return Result{}, fmt.Errorf("duplicate canonical identity %s: %s and %s", identity, previous, documentPath)
			}
			identities[identity] = documentPath
		}
		manifestDocuments = append(manifestDocuments, ManifestDocument{
			Path: documentPath, Identity: identity, Title: title, Headings: headings,
			Aliases: aliases[documentPath], SHA256: digest(content),
		})
	}

	contentEntries := make(map[string][]byte, len(documents)+1)
	for name, content := range documents {
		contentEntries[name] = content
	}
	contentEntries[BootstrapPath] = bootstrap
	manifest := Manifest{
		FormatVersion: formatVersion,
		Version:       metadata.Version, Revision: metadata.Revision, Tag: metadata.Tag,
		Repository: definition.Repository, DefinitionSHA256: digest(definitionBytes),
		ContentSHA256: aggregateDigest(contentEntries),
		Bootstrap:     GeneratedDocument{Path: BootstrapPath, SHA256: digest(bootstrap)},
		Documents:     manifestDocuments,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("encode Documentation Manifest: %w", err)
	}
	manifestBytes = append(manifestBytes, '\n')

	archiveEntries := make(map[string][]byte, len(contentEntries)+1)
	for name, content := range contentEntries {
		archiveEntries[name] = content
	}
	archiveEntries[ManifestPath] = manifestBytes
	archive, err := writeZIP(archiveEntries)
	if err != nil {
		return Result{}, err
	}
	return Result{Archive: archive, Manifest: manifest, Bootstrap: bootstrap, Documents: documents}, nil
}

func decodeDefinition(content []byte) (Definition, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	var definition Definition
	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, fmt.Errorf("parse Bundle Definition: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			err = fmt.Errorf("multiple YAML documents are not supported")
		}
		return Definition{}, fmt.Errorf("parse Bundle Definition: %w", err)
	}
	if definition.FormatVersion != formatVersion {
		return Definition{}, fmt.Errorf("unsupported Bundle Definition format-version %d", definition.FormatVersion)
	}
	parsed, err := url.Parse(definition.Repository)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return Definition{}, fmt.Errorf("Bundle Definition repository must be an absolute HTTPS URL")
	}
	if definition.BootstrapTemplate == "" {
		return Definition{}, fmt.Errorf("Bundle Definition bootstrap-template is required")
	}
	return definition, nil
}

func selectDocuments(root string, definition Definition) (map[string][]byte, error) {
	selected := map[string][]byte{}
	casePaths := map[string]string{}
	add := func(candidate string) error {
		cleaned, err := cleanRelative(candidate)
		if err != nil {
			return err
		}
		if path.Ext(cleaned) != ".md" {
			return fmt.Errorf("selected document is not Markdown: %s", cleaned)
		}
		if _, exists := selected[cleaned]; exists {
			return fmt.Errorf("duplicate selected document: %s", cleaned)
		}
		folded := strings.ToLower(cleaned)
		if previous, exists := casePaths[folded]; exists {
			return fmt.Errorf("case-insensitive document collision: %s and %s", previous, cleaned)
		}
		content, err := readRegular(root, cleaned, true)
		if err != nil {
			return err
		}
		selected[cleaned] = content
		casePaths[folded] = cleaned
		return nil
	}
	for _, file := range definition.Documents.Files {
		if err := add(file); err != nil {
			return nil, fmt.Errorf("select Bundle file %s: %w", file, err)
		}
	}
	for _, tree := range definition.Documents.Trees {
		cleaned, err := cleanRelative(tree)
		if err != nil {
			return nil, fmt.Errorf("invalid Bundle tree %s: %w", tree, err)
		}
		rootPath := filepath.Join(root, filepath.FromSlash(cleaned))
		info, err := os.Lstat(rootPath)
		if err != nil {
			return nil, fmt.Errorf("inspect Bundle tree %s: %w", cleaned, err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("Bundle tree must be a directory without symlink substitution: %s", cleaned)
		}
		err = filepath.WalkDir(rootPath, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("Bundle tree contains symlink: %s", filePath)
			}
			if entry.IsDir() || filepath.Ext(filePath) != ".md" {
				return nil
			}
			relative, err := filepath.Rel(root, filePath)
			if err != nil {
				return err
			}
			return add(filepath.ToSlash(relative))
		})
		if err != nil {
			return nil, fmt.Errorf("select Bundle tree %s: %w", cleaned, err)
		}
	}
	for _, excluded := range definition.Documents.Exclude {
		cleaned, err := cleanRelative(excluded)
		if err != nil {
			return nil, fmt.Errorf("invalid Bundle exclusion %s: %w", excluded, err)
		}
		if _, ok := selected[cleaned]; !ok {
			return nil, fmt.Errorf("Bundle exclusion is not selected: %s", cleaned)
		}
		delete(selected, cleaned)
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("Bundle Definition selects no documents")
	}
	return selected, nil
}

func validateNodes(nodes map[string]string, documents map[string][]byte) (map[string][]string, error) {
	aliases := map[string][]string{}
	for alias, target := range nodes {
		if !aliasPattern.MatchString(alias) {
			return nil, fmt.Errorf("invalid documentation Node alias: %s", alias)
		}
		if reservedAliases[alias] {
			return nil, fmt.Errorf("documentation Node alias collides with subcommand: %s", alias)
		}
		cleaned, err := cleanRelative(target)
		if err != nil {
			return nil, fmt.Errorf("invalid documentation Node target %s: %w", target, err)
		}
		if _, ok := documents[cleaned]; !ok {
			return nil, fmt.Errorf("documentation Node target is not selected: %s", cleaned)
		}
		aliases[cleaned] = append(aliases[cleaned], alias)
	}
	for target := range aliases {
		sort.Strings(aliases[target])
	}
	return aliases, nil
}

func renderBootstrap(template []byte, documents map[string][]byte, repository string, metadata Metadata) ([]byte, error) {
	var output strings.Builder
	lines := strings.SplitAfter(string(template), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if matches := includeDocumentPattern.FindStringSubmatch(trimmed); matches != nil {
			documentPath, err := cleanRelative(matches[1])
			if err != nil {
				return nil, fmt.Errorf("invalid Bootstrap document path: %w", err)
			}
			content, ok := documents[documentPath]
			if !ok {
				return nil, fmt.Errorf("Bootstrap document is not selected: %s", documentPath)
			}
			output.WriteString(rewriteLinks(string(content), documentPath, documents, repositoryReference(repository, metadata)))
			continue
		}
		if matches := includeRangePattern.FindStringSubmatch(trimmed); matches != nil {
			documentPath, err := cleanRelative(matches[1])
			if err != nil {
				return nil, fmt.Errorf("invalid Bootstrap range path: %w", err)
			}
			content, ok := documents[documentPath]
			if !ok {
				return nil, fmt.Errorf("Bootstrap range document is not selected: %s", documentPath)
			}
			section, err := headingRange(string(content), matches[2], matches[3])
			if err != nil {
				return nil, fmt.Errorf("Bootstrap range %s: %w", documentPath, err)
			}
			output.WriteString(rewriteLinks(section, documentPath, documents, repositoryReference(repository, metadata)))
			continue
		}
		if strings.Contains(trimmed, "{{") || strings.Contains(trimmed, "}}") {
			return nil, fmt.Errorf("unknown or malformed Bootstrap placeholder: %s", trimmed)
		}
		output.WriteString(line)
	}
	return []byte(output.String()), nil
}

func headingRange(content, from, before string) (string, error) {
	if !headingPattern.MatchString(from) || !headingPattern.MatchString(before) {
		return "", fmt.Errorf("range selectors must be exact ATX headings")
	}
	lines := strings.SplitAfter(content, "\n")
	fromIndexes := []int{}
	beforeIndexes := []int{}
	for index, line := range lines {
		value := strings.TrimRight(line, "\r\n")
		if value == from {
			fromIndexes = append(fromIndexes, index)
		}
		if value == before {
			beforeIndexes = append(beforeIndexes, index)
		}
	}
	if len(fromIndexes) != 1 || len(beforeIndexes) != 1 {
		return "", fmt.Errorf("range headings must each occur exactly once: %q=%d %q=%d", from, len(fromIndexes), before, len(beforeIndexes))
	}
	if fromIndexes[0] >= beforeIndexes[0] {
		return "", fmt.Errorf("range start must precede boundary")
	}
	return strings.Join(lines[fromIndexes[0]:beforeIndexes[0]], ""), nil
}

func validateLinks(root string, documents map[string][]byte) error {
	for _, documentPath := range sortedKeys(documents) {
		for _, match := range linkPattern.FindAllStringSubmatch(string(documents[documentPath]), -1) {
			target := strings.TrimSpace(match[2])
			if _, _, ok := resolveRelativeLink(documentPath, target); !ok {
				continue
			}
			resolved, _, _ := resolveRelativeLink(documentPath, target)
			info, err := os.Stat(filepath.Join(root, filepath.FromSlash(resolved)))
			if err != nil || info.IsDir() {
				return fmt.Errorf("broken Markdown link in %s: %s", documentPath, target)
			}
		}
	}
	return nil
}

func rewriteLinks(content, source string, documents map[string][]byte, repositoryReference string) string {
	return linkPattern.ReplaceAllStringFunc(content, func(value string) string {
		match := linkPattern.FindStringSubmatch(value)
		resolved, fragment, ok := resolveRelativeLink(source, strings.TrimSpace(match[2]))
		if !ok {
			return value
		}
		target := repositoryReference
		if _, included := documents[resolved]; included {
			target = resolved + fragment
		}
		return "[" + match[1] + "](" + target + ")"
	})
}

func resolveRelativeLink(source, target string) (string, string, bool) {
	if target == "" || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "<") {
		return "", "", false
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Path == "" {
		return "", "", false
	}
	resolved := path.Clean(path.Join(path.Dir(source), parsed.Path))
	if !fs.ValidPath(resolved) {
		return "", "", false
	}
	fragment := ""
	if parsed.Fragment != "" {
		fragment = "#" + parsed.Fragment
	}
	return resolved, fragment, true
}

func inspectMarkdown(content []byte) (string, string, []string) {
	lines := strings.Split(string(content), "\n")
	identity := ""
	title := ""
	headings := []string{}
	for index, line := range lines {
		match := headingPattern.FindStringSubmatch(strings.TrimRight(line, "\r"))
		if match == nil {
			continue
		}
		text := strings.TrimSpace(match[2])
		headings = append(headings, text)
		if index == 0 && match[1] == "#" && isCanonicalIdentity(text) {
			identity = text
			continue
		}
		if title == "" && match[1] == "#" {
			title = text
		}
	}
	if title == "" {
		title = identity
	}
	return identity, title, headings
}

func isCanonicalIdentity(value string) bool {
	return strings.HasSuffix(value, ".md") && !strings.ContainsAny(value, " \t/\\")
}

func cleanRelative(value string) (string, error) {
	if value == "" || strings.Contains(value, `\`) || path.IsAbs(value) {
		return "", fmt.Errorf("path must be a non-empty repository-relative slash path: %s", value)
	}
	cleaned := path.Clean(value)
	if cleaned != value || !fs.ValidPath(cleaned) {
		return "", fmt.Errorf("path is not clean or escapes the repository: %s", value)
	}
	return cleaned, nil
}

func readRegular(root, relative string, markdown bool) ([]byte, error) {
	filePath := filepath.Join(root, filepath.FromSlash(relative))
	info, err := os.Lstat(filePath)
	if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", relative, err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("path must be a regular file without symlink substitution: %s", relative)
	}
	if markdown && path.Ext(relative) != ".md" {
		return nil, fmt.Errorf("path must be Markdown: %s", relative)
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relative, err)
	}
	return content, nil
}

func repositoryReference(repository string, metadata Metadata) string {
	reference := metadata.Tag
	if reference == "" {
		reference = metadata.Revision
	}
	if reference == "" {
		return strings.TrimSuffix(repository, "/")
	}
	return strings.TrimSuffix(repository, "/") + "/tree/" + url.PathEscape(reference)
}

func aggregateDigest(entries map[string][]byte) string {
	hash := sha256.New()
	for _, name := range sortedKeys(entries) {
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write([]byte(digest(entries[name])))
		hash.Write([]byte{'\n'})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func digest(content []byte) string {
	value := sha256.Sum256(content)
	return hex.EncodeToString(value[:])
}

func writeZIP(entries map[string][]byte) ([]byte, error) {
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, name := range sortedKeys(entries) {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(0o644)
		header.SetModTime(time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC))
		entry, err := writer.CreateHeader(header)
		if err != nil {
			return nil, fmt.Errorf("create Bundle ZIP entry %s: %w", name, err)
		}
		if _, err := entry.Write(entries[name]); err != nil {
			return nil, fmt.Errorf("write Bundle ZIP entry %s: %w", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close Bundle ZIP: %w", err)
	}
	return output.Bytes(), nil
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
