# Knowledge.note.packaged-documentation-navigation.md
============================================================

# Packaged Documentation Navigation

## Context

A standalone CTK binary may carry enough version-matched Knowledge to explain
itself without a repository checkout. People can read that Knowledge directly,
but AI assistants are likely to use it more frequently as a navigation surface
while answering a narrower implementation or usage question.

The useful experience is therefore not one large document dump. It is a path
from a broad vocabulary to the smallest responsible document section, with an
escape route to ordinary repository tools when structural navigation is not
enough.

## Progressive navigation

Packaged documentation works well as a progressive sequence:

```text
Bootstrap or Node
        ↓
metadata Resolve
        ↓
one-document TOC
        ↓
exact heading Show
        ↓
bounded ancestor or descendant context
```

Each step should reduce uncertainty without forcing the next reader to ingest
the complete Bundle. A Node gives local vocabulary, Resolve identifies likely
documents, TOC exposes one file's structure, and heading-relative output lets a
caller choose the context needed for the current question.

This sequence is useful to people as well as AI assistants. Its main benefit is
that navigation decisions remain visible and repeatable instead of being
hidden inside one semantic search result.

## Metadata navigation and full-text investigation

Canonical identity, repository path, Node alias, title, and headings form a
small maintained navigation interface. They are suitable for deterministic
document discovery; document bodies are not.

When expected Knowledge cannot be found from this metadata, treat the miss as
documentation feedback. A clearer title, heading, or Node route can improve
future navigation without changing search ranking or implementation behavior.

Full-text search remains an investigation tool. Exporting the Bundle into its
repository-relative structure makes ordinary tools such as `rg` available
without turning metadata Resolve into body search or semantic question
answering.

## Source provenance has independent dimensions

An explicitly selected local source should not be described by one ambiguous
`dirty` value. Useful questions are independent:

- Does its revision match the packaged source revision?
- Does it use the same Bundle Definition?
- Does its selected content match after provenance is normalized?
- Which selected documents differ from the local revision?
- Is the wider repository dirty only because of unrelated files?

A revision mismatch does not necessarily mean the documentation differs, and
an unrelated implementation edit does not make selected Knowledge dirty.
Keeping these dimensions separate lets a reader decide whether local content is
appropriate without presenting it as binary-matched by accident.

Packaged Knowledge should remain the default. Local content is safest when it
is selected explicitly for one task and its comparison state stays visible.

## Current Go relevance

The Go executable implements this observation through its packaged
Documentation Bundle and Workspace-independent `docs` commands. Exact command
forms, output streams, Bundle representations, source validation, and Release
verification belong to the
[Go Documentation Bundle Contract](../../go/doc/contract/contract.documentation-bundle.md).

The more general reason those commands can run before Workspace discovery is
described by [Binary Self-Description Outside a
Workspace](note.binary-self-description.md).

## Boundary

This Note does not define:

- one CLI syntax for every CTK implementation;
- a semantic search or generated-answer capability;
- a requirement to package every repository document;
- silent local clone discovery or persisted local source selection;
- exact Bundle, Manifest, export, or Release representations.

Those choices belong to an implementation Contract or remain Future candidates.
