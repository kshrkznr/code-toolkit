# Go.contract.documentation-bundle.md
============================================================

# Go Documentation Bundle Contract

This Contract realizes the
[Packaged Documentation Bundle Future](../../../doc/future/future.documentation-bundle.md)
for the primary standalone Go executable.

## Source Definition

`doc/documentation-bundle.yaml` is the checked-in Bundle Definition. Its paths
are repository-root-relative even though the Definition is owned under `doc/`.

The initial strict representation is:

```yaml
format-version: 1
repository: https://github.com/kshrkznr/code-toolkit
documents:
  files: []
  trees: []
  exclude: []
nodes: {}
bootstrap-template: doc/documentation-bootstrap.md.tmpl
```

`files` selects exact regular Markdown files. `trees` recursively selects
regular Markdown files below exact directories. `exclude` removes exact files
from those selections and is not a glob mechanism. Node aliases map a concise
command value to one selected Node README.

Go rejects unknown fields, unsupported format versions, absolute paths, path
traversal, missing paths, symlinks, special files, duplicate selections,
case-insensitive path collisions, exclusions outside the selected set, Node
aliases colliding with documentation subcommands, and Node targets outside the
selected set.

## Bootstrap template

The checked-in Bootstrap template supports only:

```text
{{ bundle-provenance }}
{{ include-range "<path>" from="<heading>" before="<heading>" }}
{{ include-document "<path>" }}
```

The provenance placeholder emits the CTK version, source revision, and the
exact-tag-or-revision repository route recorded for the generated Bundle.

Paths must select bundled Markdown documents. A range heading is an exact
Markdown ATX heading including its level. It must occur exactly once, and the
`from` heading must precede the `before` heading. The starting heading is
included and the boundary heading is excluded.

Literal template text is preserved. Unknown or malformed placeholders fail
generation.

## Generated Bundle

The generator produces one deterministic ZIP with this logical layout:

```text
.ctk-docs/
├── bootstrap.md
└── manifest.json
README.md
doc/...
go/...
```

Selected source documents retain their repository-relative paths and bytes.
The Bootstrap is a generated view rather than another maintained Knowledge
source.

ZIP entries are path-sorted regular files with fixed modes and timestamps.
The generated Manifest records:

- Manifest format version;
- CTK version, source revision, and optional Release tag;
- repository URL and Bundle Definition digest;
- aggregate content digest;
- generated Bootstrap digest;
- each selected document's path, canonical identity when present, title,
  headings, aliases, and content digest.

The aggregate digest is derived from the sorted logical path and SHA-256 digest
of the Bootstrap and every selected source document. It excludes the generated
Manifest itself and wall-clock time.

## Link validation and generated views

Go validates ordinary relative Markdown links from each selected document.
Missing source targets fail generation. A valid target excluded by the Bundle
Definition is repository-only.

Selected source documents remain unchanged in the archive. In generated or
terminal views, Go resolves each relative target from its source document:

- included targets become repository-root-relative paths with fragments
  preserved;
- repository-only targets route to the exact tag or commit repository root;
- external URLs and same-document fragments remain unchanged.

## Document lookup

The Manifest index permits lookup by exact canonical identity,
repository-relative path, or Node alias. Canonical identity is read from the
first Markdown heading when it declares a CTK Knowledge or Go document
identity; it is not required for an implementation README addressable by path.

Resolve searches canonical identity, repository-relative path, Node alias,
title, and headings case-insensitively. It does not search document bodies.
Exact identity, path, and Node alias matches take priority. Other queries are
tokenized and common navigation words are ignored. Documents matching more
query terms precede documents matching fewer; matching metadata fields then
receive a small fixed priority. A result reports the matched query terms so a
caller can understand the ranking. Ties are repository-path ordered. Resolve
returns candidates rather than full documents; the CLI shows at most ten and
asks the caller to narrow a larger result.

Full-text search belongs to repository tools over `ctk docs export` output.
Until Export is implemented, its full-text scenarios remain pending Release
coverage and do not expand Resolve responsibility.

Show requires one identity or repository-relative path, matched
case-insensitively, with an optional heading fragment. Duplicate Markdown
heading anchors use the usual numeric suffix (`#responsibility-1`, then
`#responsibility-2`). Missing and ambiguous references fail rather than
selecting one document silently, and a miss routes the caller to Resolve and
the exact Bundle repository reference.

## Executable transport and CLI

Release assembly appends the deterministic Bundle ZIP to each Go executable.
ZIP readers locate the archive from its end record, so the native executable
prefix remains runnable while the Bundle can be verified with the same reader
used for a standalone archive. Appending refuses an invalid Bundle or an
executable that already contains one.

`ctk docs` reads and verifies the Bundle from its own resolved executable path
before rendering content. It does not discover or load a Workspace. The CLI
provides Bootstrap, Node listing and shortcuts, Resolve, and Show. Help remains
available even when a development binary has no Bundle.

## Boundary

The current implementation does not define local source configuration,
filesystem export, network fetching, semantic question answering, interactive
selection, or pager behavior. The generator and read model remain independently
testable from executable transport.
