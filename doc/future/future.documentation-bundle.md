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
    show the Core Node README, not every Core document

ctk docs resolve <terms...>
    find candidate documents from maintained metadata without searching bodies

ctk docs show <canonical-identity-or-path>[#heading]
    print one detailed document

ctk docs export <directory>
    packaged documents in their relative repository structure
```

These forms illustrate the candidate responsibility. Names, arguments, output
format, and group boundaries remain open.

## Default navigation view

The default `ctk docs` output should be a generated Bootstrap document rather
than the complete root README or Documentation Resolver presented one after
the other.

The current source composition is:

```text
README.md
    # Concept Domains
        through the Concept Domain and Concept API catalog
        stopping before # Installation

doc/README.md
    complete Documentation Resolver
```

This keeps the Concept Domain and Concept API vocabulary together with the
question-based Resolver while leaving public onboarding content such as Why,
Explore with AI, Installation, and Getting Started in the repository and full
bundle rather than the default view.

The composition should be owned by a dedicated checked-in Bootstrap template,
not by copying those sections into another maintained Knowledge document. A
candidate template could use small build placeholders such as:

```text
{{ include-range "README.md" from="# Concept Domains" before="# Installation" }}
{{ include-document "doc/README.md" }}
```

The exact placeholder syntax remains an implementation representation, but a
selector should identify the source path, heading level and text, and range
boundary explicitly. Generation must fail when a selected heading is missing,
duplicated at the selected level, or appears after its boundary. A heading
rename therefore becomes a visible Bootstrap update rather than silently
dropping or widening context.

The generated Bootstrap is a derived view and should be listed and hashed as
such in the Packaged Documentation Manifest. Its source sections remain owned
by `README.md` and `doc/README.md`; the template owns only their composition,
small transition text, and provenance presentation.

The template and its validator should be introduced together. Release uses the
validated generated view, but ordinary verification should also detect stale
selectors before Release preparation.

## Document index, Resolve, and Show

The generated Manifest should index each bundled document by the metadata it
already exposes rather than introducing a second tagging scheme. Useful index
fields include:

- canonical identity, when the document declares one;
- repository-relative path;
- a concise Node alias such as `core` or `contract`, when one applies;
- document title and headings;
- document role or selected bundle group;
- source content hash.

Canonical identity is one stable lookup alias and a useful grep candidate. It
does not need special visual weight in generated output, and documents such as
an implementation README may remain addressable by repository path when they
do not declare a Knowledge identity.

`ctk docs resolve <terms...>` could perform deterministic search over this
index: canonical identity, repository-relative path, Node alias, document
title, and headings. It should not search document bodies. Resolve returns a
small ordered candidate list containing identity, path, and title rather than
concatenating matching documents. Exact identity, path, or Node alias matches
precede token matches. Token results prefer more matched terms, then the
metadata field, while ties remain path-ordered and visible. Resolve is document
discovery, not full-text search or semantic question answering.

This boundary makes documentation discoverability observable. When a likely
question cannot find the responsible document from its maintained metadata,
the preferred response is to improve its title, headings, or Node route. A
documentation-only Issue is useful even when no implementation defect exists.
Full-text investigation belongs to ordinary repository tools over an exported
Bundle rather than an increasingly semantic Resolve ranking.

`ctk docs show <reference>` could resolve an exact canonical identity or
repository-relative path and print that one document. An optional heading
fragment selects one section, allowing an indexed See also link to keep its
original precision. Missing and ambiguous references should report candidates
rather than selecting one silently.

Single-document structural navigation can keep that exact lookup useful
without introducing a cross-document hierarchy. Two complementary forms are:

```text
ctk docs toc <canonical-identity-or-path>
    print the document heading tree with references accepted by Show

ctk docs show <canonical-identity-or-path>#<heading> --depth <depth-or-range>
    project selected ancestor and descendant levels around one heading
```

`toc` operates on one exact document. Its output should preserve heading
nesting and expose the exact fragment for each entry, including deterministic
suffixes for duplicate Markdown anchors, so a person or AI assistant can pass
an entry directly to `show`. It is a local map of one file, not another Node or
Bundle-wide navigation layer.

`--depth` is valid only when `show` receives a heading fragment. Relative level
zero is the selected heading and its direct body. Negative levels add ancestors
on the unique path to that heading, while positive levels add descendants:

```text
--depth 0       selected heading and direct body
--depth -1      parent and selected heading, equivalent to -1..0
--depth 1       selected heading and direct children, equivalent to 0..1
--depth -1..2   parent through grandchildren, inclusively
```

A range must contain level zero. Ancestor projection includes each selected
ancestor heading and its direct body but omits sibling sections, so it is a
structural projection rather than one contiguous source slice. Descendant
projection includes only sections within the selected heading. Omitting
`--depth` preserves the ordinary Show behavior of printing the complete
selected subtree.

TOC and depth projection should share one Markdown heading model. At minimum it
must retain level, text, generated anchor, source range, parent, and children;
ignore apparent headings inside fenced code blocks; and apply the same
duplicate-anchor rules used by heading Show. The generated Manifest may retain
its concise heading index until a richer persisted representation proves
necessary.

The Go M3.5 implementation accepts this single-document boundary. Its exact
syntax and observable behavior are owned by the Go Documentation Bundle
Contract; cross-document hierarchy remains outside this direction.

A concise Node alias may be exposed as a direct shortcut. For example,
`ctk docs core` is equivalent to showing `Knowledge.core.md` or
`doc/core/README.md`; it prints the Core Node README rather than concatenating
every Core document. Workbench, Integration, Contract, Note, and other useful
bundled Nodes can follow the same rule without creating a separate category
dump behavior. Node aliases belong in the central index and must not collide
with command names such as `resolve`, `show`, `toc`, or `export`.

`ctk docs resolve core` remains distinct: it may discover the Node and several
detailed documents, then let the reader or AI assistant choose one result to
Show.

## Links in generated output

Source documents keep their ordinary relative Markdown links. Exported source
documents also keep those links because their repository-relative directory
structure is preserved.

Concatenated or terminal output no longer has one source directory from which a
relative target can be interpreted. For those views, generation should:

- keep the original Markdown link label;
- resolve the target from the source document before composition;
- render an included target as a repository-root-relative path plus any
  heading fragment;
- make that rendered path acceptable directly to `ctk docs show`;
- route an intentionally excluded target through the common exact-version
  GitHub repository reference;
- fail when the original target does not exist.

Generated source-boundary annotations can keep the originating repository path
visible around each composed section. Canonical identity lines remain when the
selected source contains them; the Bootstrap does not invent a new Knowledge
identity for extracted README sections.

## Candidate content boundary

The current direction is to carry nearly all Knowledge needed to understand,
use, and troubleshoot the current CTK version, but not the entire repository
documentation tree.

Inclusion should follow the document's primary responsibility:

| Bundle | Primary responsibility | Current examples |
| --- | --- | --- |
| Include | Public entry, navigation, accepted concepts, agreements, current operational guidance, and rationale needed to interpret current behavior | `README.md`, Documentation Resolver, Core, Workbench, Integration, Contract, Note, Design Note, current Go README and Contracts |
| Include as a concrete entry | A small maintained example that helps a reader apply the concepts without defining them | Author's Recipe Node |
| Include as an explicit Project Knowledge exception | A self-contained observation directly useful while applying bundled current Knowledge | Collaborative Review Surfaces |
| Repository only by default | How Knowledge was created, reviewed, or evolved | Project Knowledge, including its Experiments, Notes, and Design Notes |
| Repository only | Unsettled candidate directions | Future |
| Repository only | Personal or generated instance data rather than CTK responsibility | Author's Recipe Inspect snapshots |
| Repository only | Contribution process, tests, retained implementation evidence, and other development-only material | `.github`, test data, and the retained Bash documentation |

Note and Design Note remain included by default because operational guidance
and accepted design rationale can explain how or why the current product
behaves. A document retained mainly as historical evidence may become an
explicit repository-only exception without changing the default for its role.

Project Knowledge has the opposite default. Most of it is valuable repository
context but is not necessary beside every installed binary. A Project
Knowledge document may become an explicit bundle exception when it:

- is directly useful while applying or reviewing current bundled Knowledge;
- is self-contained enough to use without replaying its originating
  Experiment or conversation;
- states its non-authoritative boundary clearly;
- adds context not already supplied by a smaller accepted document.

The current explicit exception is [Collaborative Review
Surfaces](../project-knowledge/note/note.collaborative-review-surfaces.md).
Bundled [Workbench Review](../note/note.workbench-review.md) already routes to
it when review needs shared context about participants, AI assistants, and
mechanical views. Its originating Experiments and the Project Knowledge Node
remain repository-only and use the common exact-version GitHub route when
referenced.

Bundle inclusion does not promote this Note into accepted Knowledge or change
its canonical identity. Document role describes authority; Bundle selection
describes availability in one distribution. The Bundle Definition should keep
the exception visible as one exact path rather than including the surrounding
Project Knowledge subtree.

No additional repository-only exception is currently identified for the main
`doc/note/` and `doc/design-note/` roles. Their operational guidance and current
design rationale remain included by default even when a document also preserves
some historical context.

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

## Repository-only links

Bundled Knowledge may link to Project Knowledge, Future, generated Author's
Recipe Inspect snapshots, or other material intentionally kept in the
repository. The bundle does not need a packaged stub or an exact replacement
URL for every excluded document.

During bundle generation, a relative link whose target is valid repository
content but excluded by the Bundle Definition should become a repository-only
reference to the exact-version GitHub repository root. Its presentation can be
as small as:

```text
Repository-only material. See the CTK repository for this version:
https://github.com/kshrkznr/code-toolkit/tree/<tag-or-commit>
```

The source Knowledge remains unchanged. Link conversion applies only to the
generated bundle and export, keeping normal repository navigation precise
without carrying that precision into the smaller distribution view.

The generator must distinguish intentional exclusion from a broken source
link. A relative target that does not exist in the source repository is a
validation failure; it must not be hidden behind the general GitHub route.
Ordinary external links and links between included documents remain unchanged.

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

Export is required for the first Release that presents packaged documentation
as complete. Until Export exists, tests should evaluate Resolve as metadata
navigation and record full-text search scenarios as pending Export coverage,
not broaden Resolve to compensate for the missing command.

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

## Revisit when

Reconsider this candidate when CTK prepares a package-manager or other
binary-only distribution, or when repeated AI-assisted use shows that sharing
the repository solely to obtain version-matched Knowledge is unnecessary
friction.

Track the current exploration in
[Issue #20](https://github.com/kshrkznr/code-toolkit/issues/20).
