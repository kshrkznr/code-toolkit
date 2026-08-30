# Knowledge.note.binary-self-description.md
============================================================

# Binary Self-Description Outside a Workspace

## Context

A package manager or standalone Release can install the CTK executable without
installing a CTK Workspace or repository checkout beside it.

Commands that explain the executable itself may still be useful in that
context. A person, script, or package workflow may need to discover available
commands, identify the installed version, or obtain packaged documentation
before choosing or creating a Workspace.

## Observation

Help, version provenance, and packaged documentation describe the installed
binary. They do not inherently need Cookbook Source, Dist, Archive, Extension
Pool, Workbench state, or Host integration paths.

Requiring Workspace discovery before answering those questions couples binary
self-description to state that the operation does not inspect. This is easy to
miss when the executable is normally run from a source checkout that is also a
valid Workspace, but becomes visible in binary-only distribution.

## Guidance

An implementation that distributes one standalone binary may place commands
about that binary before its Workspace-dependent dispatch boundary.

This is a responsibility test rather than a fixed command list:

```text
Does the command only describe content carried by this executable?
    yes -> it may run without selecting a Workspace
    no  -> resolve and validate the Workspace required by the operation
```

Moving a command across that boundary should not weaken validation for
Cookbook, lifecycle, Runtime, or other Workspace-dependent operations.

## Current Go relevance

The current Go CLI dispatches `help`, `version`, and packaged `docs` operations
before Workspace discovery. Exact commands, argument validation, output, exit
behavior, and initialization order belong in the
[Go CLI Contract](../../go/doc/contract/contract.cli.md).

The navigation and source-provenance observations learned from packaged
Knowledge are preserved by [Packaged Documentation
Navigation](note.packaged-documentation-navigation.md). Exact Bundle and CLI
behavior belongs in the [Go Documentation Bundle
Contract](../../go/doc/contract/contract.documentation-bundle.md), while
unsettled distribution candidates remain in [Packaged Documentation Bundle
Follow-ups](../future/future.documentation-bundle.md).

## Boundary

This Note does not define:

- a shared CLI Contract that every CTK implementation must follow;
- `help`, `version`, or `docs` syntax and output;
- permission for lifecycle or Host operations to bypass Workspace validation;
- a requirement that a non-Go implementation distribute one standalone
  binary;
- the contents or representation of a Packaged Documentation Bundle.

The observation becomes relevant when an implementation's distribution unit
is useful before a Workspace exists. Implementation-specific behavior remains
owned by that implementation's Contract and source.
