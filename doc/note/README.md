# Knowledge.note.md
============================================================

# Notes

Notes preserve reusable operational observations and loose guidance for working
with CTK.

They help a reader apply, review, or troubleshoot accepted concepts in a
specific context. A Note may describe a useful practice without making that
practice mandatory or redefining the concept it supports.

## Responsibility

Notes are responsible for keeping practical knowledge reusable after it has
become clearer than a working fragment but does not belong to an accepted
Concept API or Contract.

A Note should make its context visible enough that a reader can judge whether
the observation applies to the current work.

## Navigate by question

Start with the question closest to the current work:

- How can I put my normal IDE back, stop using CTK, or remove generated
  environments safely?
  → [Leaving CTK](note.leaving-ctk.md)
- How should a CodeVenv be selected, activated, or recovered?
  → [CodeVenv Operations](note.codevenv.md)
- How should an origin Distribution be recovered cautiously?
  → [Go CodeVenv Origin Recovery](note.go-codevenv-origin-recovery.md)
- How should a Draft or Inspect result be reviewed?
  → [Workbench Review](note.workbench-review.md)
- How are JSON settings represented for inspection and composition?
  → [CTK JSON Flat Format](note.ctk-json-flat-format.md)
- How are Recipe build, Profile, and activation choices used in practice?
  → [Recipe Build Strategy](note.recipe-build-strategy.md)
- When and how may an implementation extend Resource combination semantics?
  → [Merge Rules as a Kitchen Note](note.merge-rules.md)
- How are extension artifacts resolved and diagnosed?
  → [Extension Resolution](note.extension-resolve.md)
- What do Platform-specific failures reveal about CTK integration boundaries?
  → [Platform Differences as Boundary Evidence](note.platform-boundary-evidence.md)
- How does the current Built-in Platform Registry separate declarations from
  CTK behavior?
  → [Built-in Platform Registry](note.platform-registry.md)
- What has been observed about Variant responsibility and usage?
  → [Variant](note.variant.md)
- Why might a standalone CTK binary expose its own help, version, or packaged
  documentation before Workspace discovery?
  → [Binary Self-Description Outside a Workspace](note.binary-self-description.md)
- How can packaged Knowledge be navigated progressively, searched by metadata,
  and compared with an explicit local source?
  → [Packaged Documentation Navigation](note.packaged-documentation-navigation.md)
- How does a published CTK Release become reviewable Homebrew and Scoop
  updates, and where does target-device verification begin?
  → [Package-manager Delivery](note.package-manager-delivery.md)

These are useful entry points rather than an exhaustive document catalog.
Search by the current concept when none of these questions fit.

## Boundary

Notes do not define:

- accepted Concept API responsibilities
- behavioral or representation Contracts
- required repository or Cookbook layouts
- one mandatory development practice

Use the [Documentation Resolver](../README.md) when the question belongs
to Core, Workbench, Integration, a Contract, or another document role.

## Local guidance

Note structure follows the observation being preserved. Context, observation,
guidance, status, and boundaries are useful when they clarify applicability,
but no common heading set is required.

Links to related documents, including a `See also` section, are optional. Add
them when they provide a natural next step rather than to satisfy a template.
