# Knowledge.future.documentation-bundle.md
============================================================

# Future: Packaged Documentation Bundle

CTK's documentation is part of its inspectable operating surface, but a
standalone binary or package-manager installation does not carry the repository
that currently provides that surface.

A future `docs` command could make version-matched CTK Knowledge available
without requiring a Workspace or repository checkout. The command and bundle
remain candidates; this document does not define their CLI, configuration, or
packaging Contract.

## Candidate responsibility

A Packaged Documentation Bundle could be responsible for:

- carrying published Knowledge that matches the binary version;
- providing a small default context from which a person or AI assistant can
  navigate CTK's documentation roles;
- exposing additional document groups without requiring the entire repository;
- identifying the exact GitHub tag or commit that owns the complete source;
- exporting the packaged documents with their relative structure and
  provenance intact.

The existing [Documentation Resolver](../README.md) would continue to own
question-based navigation and the authoritative Repository Map. Packaging would
make that Resolver and the documents it routes to available in another
distribution context; it would not create a second Resolver or documentation
authority.

## Candidate experience

The smallest useful command could return an AI-readable Markdown bundle rather
than an interactive terminal view:

```text
ctk docs
    Documentation Resolver
    Concept Domain entry points
    Concept API navigation
    exact-version source link

ctk docs core
ctk docs workbench
ctk docs integration
ctk docs contract
    scoped Knowledge views

ctk docs export <directory>
    packaged documents in their relative repository structure
```

These forms illustrate the candidate responsibility. Names, arguments, output
format, and group boundaries remain open.

## Candidate content boundary

The current direction is to carry nearly all Knowledge needed to understand,
use, and troubleshoot the current CTK version, but not the entire repository
documentation tree.

Inclusion should follow the document's primary responsibility:

| Bundle | Primary responsibility | Current examples |
| --- | --- | --- |
| Include | Public entry, navigation, accepted concepts, agreements, current operational guidance, and rationale needed to interpret current behavior | `README.md`, Documentation Resolver, Core, Workbench, Integration, Contract, Note, Design Note, current Go README and Contracts |
| Include as a concrete entry | A small maintained example that helps a reader apply the concepts without defining them | Author's Recipe Node |
| Repository only | How Knowledge was created, reviewed, or evolved | Project Knowledge, including its Experiments, Notes, and Design Notes |
| Repository only | Unsettled candidate directions | Future |
| Repository only | Personal or generated instance data rather than CTK responsibility | Author's Recipe Inspect snapshots |
| Repository only | Contribution process, tests, retained implementation evidence, and other development-only material | `.github`, test data, and the retained Bash documentation |

Note and Design Note remain included by default because operational guidance
and accepted design rationale can explain how or why the current product
behaves. A document retained mainly as historical evidence may become an
explicit exception without changing the default for its role.

This boundary should not be expressed by adding a required bundle tag to every
Knowledge document. Canonical identity and document role already communicate
what a document is responsible for; repeating distribution metadata in each
file would introduce another classification that can drift.

A central Bundle Definition could select role-owned paths and record the small
number of exceptions. Validation could then report:

- a selected path that no longer exists;
- a new documentation role that has no bundle decision;
- an included document whose local link targets repository-only content;
- an exception that no longer differs from its role default.

The binary could carry this selected set as one versioned bundle while
presenting only a navigation view by default. Selection controls availability;
`ctk docs`, scoped views, and export control how much of the available set is
shown for one task.

## Bundle Definition and generated Manifest

The current direction separates two related artifacts:

```text
doc/documentation-bundle.yaml
    checked-in Bundle Definition
    owns selection policy and explicit exceptions
            ↓ generate and verify
Packaged Documentation Manifest
    generated inventory and provenance
    travels inside the binary and exported directory
```

The checked-in Definition should live under `doc/` because documentation
distribution is part of the documentation boundary. Its selected paths should
remain repository-root-relative because the bundle also includes `README.md`
and current Go implementation documents outside `doc/`.

The Definition could contain a format version, included role-owned paths, and
explicit exceptions. It should not contain source revisions, content hashes,
or generation timestamps that would require a checked-in update after every
ordinary documentation edit.

The generated Manifest could contain:

- Manifest and Bundle Definition format versions;
- CTK version, exact source revision, and Release tag when one applies;
- a digest of the Bundle Definition;
- the sorted repository-relative document inventory and a hash for each file;
- one aggregate content digest derived deterministically from that inventory.

Wall-clock generation time should not participate in reproducible content. If
an informational timestamp becomes useful, it should remain outside the
content identity or derive from stable source provenance.

The Definition and generator should be introduced together. Checking in an
unvalidated YAML shape before any consumer exists would turn a Future example
into a de facto configuration format without proving that it selects and
verifies the intended documents.

## Version and source provenance

Packaged documentation should default to the source built with the binary. A
release binary should route to its exact GitHub tag; a development binary could
route to its recorded commit. `main` or the latest Release may describe
behavior different from the installed executable.

Output and exported content could identify at least:

- CTK version and source revision;
- documentation source, such as embedded or local;
- included document identities and repository-relative paths;
- a content hash or equivalent provenance when useful.

An explicitly selected local clone could support documentation development and
AI work over an editable tree. It should not silently replace matching packaged
Knowledge merely because a clone is nearby. A local revision that differs from
the binary should remain visible rather than being presented as version-matched
documentation.

Local-source status can use the same Definition and generated Manifest rather
than relying on one ambiguous `dirty` label. It could report independently:

- whether the local Git revision matches the binary revision;
- whether the local Bundle Definition matches the embedded Definition digest;
- whether selected local document bytes match the embedded aggregate digest;
- which selected paths differ from their local revision.

Changes outside the selected document set may make the repository dirty, but
should not make the documentation bundle dirty. That repository state may be
shown as an additional diagnostic. A non-Git source can report an unknown
revision while still computing its Definition and content digests.

Whether local selection belongs to an explicit command option, Workspace
configuration, or another integration surface remains open. The default
`docs` path should still work when no Workspace can be discovered.

## Export candidate

Export could preserve paths such as `README.md`, `doc/core/`, and
`doc/contract/` so that relative links and document roles remain inspectable.
The generated Manifest should travel with those files so an exported directory
can be verified without the original binary or repository.

The first export behavior should be deliberately narrow:

- accept an absent or empty target directory;
- refuse to merge with or replace a non-empty target;
- write to a sibling staging directory and publish the completed tree
  atomically when the OS permits it;
- preserve repository-relative paths while rejecting traversal, symlinks,
  special files, duplicate paths, and cross-platform case collisions;
- write files in sorted order with fixed file and directory modes;
- reserve the generated Manifest path so source content cannot replace it;
- verify the exported aggregate digest before publication.

An explicit replace or merge mode can remain a later candidate with its own
conflict and recovery Contract. It is not necessary for the first reproducible
export.

## Release verification

Release generation can use the same artifacts as a closed verification chain:

```text
clean checkout at the requested tag
        ↓
load and validate Bundle Definition
        ↓
generate one ordered bundle and Manifest
        ↓
embed the same generated input in every target binary
        ↓
compare a native binary's exposed Manifest with the generated Manifest
        ↓
regenerate from a fresh tag checkout and compare the aggregate digest
```

Release should fail when the requested tag does not identify the source
revision, the checkout is dirty, a selected or excepted path is invalid, or
regenerated Definition and content digests differ. Cross-compiled binaries do
not each need a separate documentation selection pass; they should consume the
same already verified generated bundle input.

## Go CLI boundary

Workspace-independent dispatch is not a shared Packaged Documentation Bundle
responsibility. It arises from the current Go distribution unit: one standalone
binary must be able to explain itself even when no CTK Workspace or repository
checkout is present.

The reusable observation is recorded by [Binary Self-Description Outside a
Workspace](../note/note.binary-self-description.md). If the Go CLI implements
this direction, exact dispatch order and behavior for `help`, `version`, and
`docs` belong in the
[Go CLI Contract](../../go/doc/contract/contract.cli.md) and Go source. Other
implementations are not required to adopt the same boundary when their
distribution context differs.

## Boundary

This candidate does not define:

- a new Concept API or a second Documentation Resolver;
- generated answers to arbitrary natural-language questions;
- a promise to include every repository file in a binary;
- silent network access or silent selection of a local clone;
- accepted CLI syntax, configuration fields, output schemas, or overwrite
  behavior;
- package-manager publication itself.

Cookbooks, Distributions, Workbench output, and temporary review material are
not published Knowledge merely because they exist beside documentation in a
development checkout.

## Open questions

- Which documents form the default navigation view within the selected set?
- Which retained historical Notes or Design Notes need repository-only
  exceptions?
- How should canonical identities and relative links appear in concatenated
  Markdown output?
- Should a link from bundled Knowledge to repository-only material become an
  exact-version GitHub link, an unavailable route, or a small packaged stub?

## Revisit when

Reconsider this candidate when CTK prepares a package-manager or other
binary-only distribution, or when repeated AI-assisted use shows that sharing
the repository solely to obtain version-matched Knowledge is unnecessary
friction.

Track the current exploration in
[Issue #20](https://github.com/kshrkznr/code-toolkit/issues/20).
