# Go.contract.documentation-bundle.md
============================================================

# Go Documentation Bundle Contract

This Contract defines the current Documentation Bundle behavior of the primary
standalone Go executable. Reusable navigation and provenance guidance belongs
to [Packaged Documentation
Navigation](../../../doc/note/note.packaged-documentation-navigation.md), while
unsettled distribution and navigation candidates remain in [Collected Future
Candidates](../../../doc/future/future.candidates.md#packaged-documentation-follow-ups).

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
It does not expand Resolve responsibility.

Show requires one identity or repository-relative path, matched
case-insensitively, with an optional heading fragment. Duplicate Markdown
heading anchors use the usual numeric suffix (`#responsibility-1`, then
`#responsibility-2`). Missing and ambiguous references fail rather than
selecting one document silently, and a miss routes the caller to Resolve and
the exact Bundle repository reference.

### Single-document structural navigation

`ctk docs toc <canonical-identity-or-path>` renders one exact document's ATX
heading tree as nested Markdown links. Each link uses the resolved
repository-relative path and exact heading fragment accepted by Show. The
leading canonical identity heading is document metadata and is omitted from
the TOC; the document title remains its first navigable entry.

TOC, heading Show, and depth projection share one parsed heading model.
Apparent headings inside fenced code blocks do not enter the model. Duplicate
anchors are assigned across the visible document headings before any subtree
is selected, so TOC and Show always agree.

`ctk docs show <canonical-identity-or-path>#<heading> --depth <N|A..B>` limits
output to structural levels relative to the selected heading. `--depth`
requires a heading fragment. Level `0` is the selected heading and its direct
body. Negative levels add ancestors on its unique heading path and their direct
bodies without siblings. Positive levels add descendants through the requested
tree depth. The projection is therefore not necessarily one contiguous source
range.

A single negative integer `N` means `N..0`; a single non-negative integer means
`0..N`. An explicit inclusive range must contain zero. Omitting `--depth`
preserves Show's complete selected-subtree behavior.

## Explicit local source

`ctk docs --source <repository> [<docs-operation>]` explicitly selects one
local documentation source for that invocation. `--source` precedes the docs
operation and accepts an absolute or current-directory-relative repository
root. CTK does not discover a nearby clone, read Workspace configuration, or
persist the selection. Omitting it always uses the packaged Bundle.

The local root must contain a valid Bundle Definition and all selected source
material. Local loading reuses the generator's traversal, regular-file,
symlink, duplicate, and case-collision validation. The generated local Bundle
drives Bootstrap, Node, Resolve, TOC, Show, and Export exactly as the packaged
read model does.

Local status compares independent dimensions rather than reducing them to one
dirty flag:

- local Git `HEAD` against the packaged source revision;
- local Bundle Definition digest against the packaged Definition digest;
- local generated content against the packaged content while using packaged
  metadata for a provenance-neutral comparison; status reports that comparison
  digest separately from the locally attributed Bundle digest;
- each locally selected document byte-for-byte against the blob at local
  `HEAD`;
- whole-repository dirty state as a separate diagnostic.

A non-Git source reports revision, selected-path dirty, and repository dirty as
unknown while Definition and content comparison remain available. Changes
outside the selected document set may make the repository dirty but do not
make selected-path status dirty.

`ctk docs status` reports packaged provenance. With `--source`, it also reports
the resolved local path, packaged comparisons, selected dirty paths, and the
independent repository state. A path below the current user's home replaces
that user-specific prefix with `<home>` before display. Every local navigation
operation emits a concise source and comparison diagnostic on stderr; document
or tabular content on stdout remains composable. Local content is therefore
never silently presented as version-matched packaged Knowledge.

## Filesystem Export

`ctk docs export <directory>` publishes the verified Bundle as its logical
directory tree, including the generated Manifest and Bootstrap. Selected source
documents retain their repository-relative paths and original bytes, making
ordinary repository full-text tools available without a clone.

The target must be absent or an empty directory. Its parent must already exist
as a directory and cannot itself be a symlink. A file, symlink, special file,
or non-empty target fails without merge or replacement.

Export writes fixed-mode files and directories to a sibling staging directory.
It then rereads the complete staged inventory, rejects missing, unexpected,
symlinked, special, case-colliding, or byte-mismatched entries, and publishes
the verified tree by replacing only the reserved empty target. A failure does
not publish a partial final tree or modify a pre-existing non-empty target.

Successful CLI output reports the absolute exported path and the aggregate
content digest recorded by the verified Manifest.

## Executable transport and CLI

Release assembly appends the deterministic Bundle ZIP to each Go executable.
ZIP readers locate the archive from its end record, so the native executable
prefix remains runnable while the Bundle can be verified with the same reader
used for a standalone archive. Appending refuses an invalid Bundle or an
executable that already contains one.

`ctk docs` reads and verifies the Bundle from its own resolved executable path
before rendering content. It does not discover or load a Workspace. The CLI
provides Status, Bootstrap, Node listing and shortcuts, Resolve, TOC, Show,
depth projection, and Export. Help remains available even when a development
binary has no Bundle.

## Release verification

Release assembly requires a clean checkout whose `HEAD` is the exact requested
Release tag. It generates the Bundle once with that version, full revision, and
tag, then independently regenerates it from a fresh `git archive` of the tag.
The byte streams must match.

Every target executable receives the same validated Bundle bytes. The Release
tool reopens each final executable without executing the target architecture
and verifies version, revision, tag, and aggregate content digest. When one
built target is native to the Release host, assembly also exercises version,
packaged Status, Bootstrap, Node, Resolve, TOC, Show, depth projection, local
source Status and Show, and Export from a directory without a Workspace.

Platform archives and their checksums are created only after the Bundle is
appended and verified, so the published checksum covers both executable and
documentation content.

## Boundary

The current implementation does not define persisted local source
configuration, network fetching, semantic question answering, interactive
selection, or pager behavior. The generator and read model remain
independently testable from executable transport.
