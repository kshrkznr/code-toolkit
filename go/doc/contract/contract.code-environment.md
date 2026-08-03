# Go.contract.code-environment.md
============================================================

# Go Code Environment Contract

This Contract specializes the shared
[Code Environment Contract](../../../doc/contract/contract.code-environment.md).

## Managed representation

- Selected Runtimes are connections named `dist/current.<platform>`.
- Imported defaults are stored as `dist/origin.<platform>`.
- macOS host connections use symbolic links.
- Windows host connections use directory junctions.
- Distribution Platform identity comes from `.meta/recipe.yaml`.

These paths retain interoperability with the Bash reference representation.
They are Go persisted state, not a universal CodeVenv layout.

## Activation and switching

Go stops only relevant Platform processes and confirms they stopped before host
path replacement. Repeated `activate` is an integration health check: it is a
no-op only when Selection and both host connections agree. An incomplete or
inconsistent integration directs the caller to explicit deactivation recovery
rather than silently repairing activation.

Selection changes are serialized per Platform. Go prepares and validates the
new Selection, backs up the previous Selection, replaces it, and rolls back
when publication fails.

## Process lock

The per-Platform operation lock records its owner process. An interrupted lock
is reclaimed only after its owner is no longer alive. A legacy ownerless lock
directory is treated as stale only after a short publication grace period so a
concurrent operation is not mistaken for abandoned state.

## Transaction and recovery

Activation and deactivation persist an atomic transaction journal before host
mutation. Recovery resumes or rolls back only when the recorded phase and
observed paths identify one safe action. Ambiguity preserves the journal and
returns a hard failure.

Forced deactivation may reconstruct a missing host path from the trusted
origin. A physical directory or unexpected connection already at that path is
first preserved as diagnostic and rollback state. `--force` does not permit Go
to guess through missing trusted evidence.

The explicit `--force-empty` escape hatch is destructive empty deactivation,
not trusted-origin restoration. It preserves recoverable host content and the
selected Distribution, initializes empty physical Host User and Extensions
paths, and removes Selection only after successful host publication.

## Windows output boundary

Go normalizes CRLF from Platform command output at the adapter boundary.
Profile identities and path-bearing output must not retain line-ending bytes.
