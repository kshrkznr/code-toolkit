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

One possible packaging strategy is to carry all published Knowledge in a
versioned bundle while presenting only a navigation set by default. This could
keep relative document relationships complete without making every invocation
return the entire collection. A curated partial bundle remains possible if a
closed and maintainable document boundary becomes clearer.

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

Whether local selection belongs to an explicit command option, Workspace
configuration, or another integration surface remains open. The default
`docs` path should still work when no Workspace can be discovered.

## Export candidate

Export could preserve paths such as `README.md`, `doc/core/`, and
`doc/contract/` so that relative links and document roles remain inspectable.
An accompanying manifest could record bundle version, source revision,
document inventory, and hashes.

Export should not make replacement or merge behavior implicit. Empty-target,
conflict, atomic publication, and explicit-overwrite rules need observation
before they become a Contract.

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

- Which documents form the default navigation set?
- Should the binary carry all published Knowledge or a selected closed set?
- Which document roles belong in scoped views and a full export?
- How should canonical identities and relative links appear in concatenated
  Markdown output?
- How should a local source expose dirty state and revision mismatch?
- Which manifest and conflict rules make export safe and reproducible?
- Should `help`, `version`, and `docs` share one Workspace-independent command
  boundary in the Go CLI?
- How should Release generation verify that the packaged bundle matches the
  source tag or commit?

## Revisit when

Reconsider this candidate when CTK prepares a package-manager or other
binary-only distribution, or when repeated AI-assisted use shows that sharing
the repository solely to obtain version-matched Knowledge is unnecessary
friction.

Track the current exploration in
[Issue #20](https://github.com/kshrkznr/code-toolkit/issues/20).
