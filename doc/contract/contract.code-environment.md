# Knowledge.contract.code-environment.md
============================================================

# Code Environment Integration Contract

This document records the capability and safety boundary between CTK and a
VS Code ecosystem Platform.

A Platform is a VS Code-family application integration such as `code` or
`kiro`; those examples are not a fixed supported-app list. JetBrains, Eclipse,
and other IDE ecosystems are outside the current CTK product boundary rather
than alternative Platform adapter kinds.

## Platform Integration

### Required Capabilities

- A Recipe identifies the Platform required by its Runtime.
- CTK can launch that Platform with the isolated Runtime represented by a
  Distribution.
- CTK can observe the Runtime state required by Lock.
- CTK can apply Platform extensions and configuration required by Build and
  Apply.
- Platform extension identifiers cross the integration boundary without CTK
  changing their letter case.

### Platform representation

The current directory adapter launches a Platform command with:

```text
<platform> \
  --user-data-dir <dist>/.data \
  --extensions-dir <dist>/.ext \
  <additional arguments>
```

It also uses Platform CLI operations to list, install, and uninstall extensions
and to open named Profiles.

These flags and paths are implementation details within the VS Code ecosystem,
not user-selectable Runtime Adapter kinds.

Platform definitions may declare observed application differences such as the
command, Host paths, process identity, and Extension Pool behavior. They do not
select another IDE ecosystem or replace Runtime I/O semantics.

A Distribution Launch Override may replace launch execution where supported.
It may therefore start Eclipse, a JetBrains IDE, or another application while
CTK acts only as its launcher. CTK does not infer Build, Apply, Lock, Archive,
Profile, Extension, or CodeVenv capabilities for that application's Runtime.

---

## CodeVenv Lifecycle

### Required Capabilities

#### Activate

- Import the Platform's current default environment as an origin Distribution.
- Change the Platform's default behavior so CodeVenv can select its Runtime.
- Activation is an explicit user decision for one Platform.
- Activating an already-managed Platform checks the health of its host
  integration.
- A healthy integration succeeds without changing state.
- An incomplete or inconsistent integration is reported without silently
  repairing or replacing host paths.

#### Use

- Change the activated Platform's default behavior to use the selected
  Distribution.
- Resolve the Platform from Distribution provenance rather than its artifact
  name.
- Selecting one Platform does not change another Platform.
- Selecting the already-current Distribution succeeds without changing state.

#### Deactivate

- Stop CodeVenv management for one Platform.
- Restore the Platform's imported default environment.
- Forced deactivation can recover an incomplete activation when the previous
  managed state is trustworthy.
- Host content encountered during forced recovery is preserved before CTK
  replaces or reconstructs it.

### Safety Invariants

- User data and installed extensions are treated as one Runtime state.
- Activation does not destroy the imported default environment.
- Runtime switching prevents concurrent Platform access that could mix or
  corrupt old and new Runtime state.
- A completed switch exposes the new Runtime.
- A failed switch retains or restores the previous usable Runtime.
- A lifecycle operation does not leave a broken intermediate state as the
  Platform default.
- Re-running activation does not silently mutate an unhealthy integration.
- Recovery from an unhealthy integration is an explicit deactivation decision.
- Operations for one Platform do not change another Platform's state.

### Managed-state integrity boundary

CodeVenv safety guarantees assume that CTK's managed Selection, trusted origin,
host connections, transaction journal, and recovery backups are not manually or
concurrently changed outside the lifecycle operation that owns them.

For example, removing or replacing `current.<platform>` while its Platform is
activated can destroy the evidence that connects the host paths to the selected
Runtime. Deleting an origin, journal, or recovery backup can similarly remove
the state required for deterministic restoration.

CTK must report an incomplete or ambiguous integration and stop when observed
state no longer agrees with the managed evidence. It is not required to infer
which external change occurred, recreate missing history, or guess a recovery
target. `--force` does not relax this boundary.

An explicitly destructive empty-deactivation operation may still be available
when its own preconditions are satisfied, but it is not restoration of the
previous managed Runtime.

### Activation observation and publication

Activation separates read-only observation from origin construction and host
integration:

1. Validate host paths, current Selection, and any persisted transaction.
2. Stop the default host Runtime and observe its User and Extension state
   without changing host paths.
3. Derive a temporary Recipe and seal a trusted source Lock from that
   observation.
4. Construct a candidate origin in staging. Copy the complete Host User tree,
   but start the candidate Extension area empty and recover Extensions from the
   trusted Lock by exact ID.
5. Obtain a fresh Lock and semantically compare it with the source observation.
6. Continue only when verification matches or the user explicitly accepts a
   forceable semantic difference.
7. Publish the candidate origin while retaining the previous origin for
   rollback.
8. Persist a transaction journal before changing the current Selection or host
   paths, then redirect the Selection, Host User, and Host Extensions as one
   logical operation.
9. Validate the completed integration before discarding backups and the
   journal.

The source observation and trusted Lock are complete before origin generation
begins. Copying Host User into staging is part of origin construction, not the
source observation. The full User tree is retained because Runtime state may
include keybindings, snippets, tasks, MCP configuration, storage, and other
Platform-owned data not yet interpreted by CTK.

Installed Extension files and Platform-owned Extension inventory metadata are
not copied into origin. They are observations used to reconstruct a clean
Extension area through the Platform adapter.

An existing origin is not replaced until the candidate origin has passed
recovery and verification. It remains recoverable until host redirection has
also completed successfully.

### Deactivation restoration

Deactivation validates the active Selection and trusted origin before changing
host paths. It then:

1. Stops the default host Runtime and the Runtime referenced by the current
   Selection.
2. Persists a transaction journal.
3. Removes the managed Host User and Host Extensions connections.
4. Restores a physical Host User tree from origin.
5. Creates an empty physical Host Extensions area at its final host path and
   recovers Extensions there from the trusted origin Lock.
6. Obtains a fresh observation and performs semantic verification.
7. Continues on a match or an explicitly accepted forceable difference, then
   removes the current Selection and transaction backups.

Extension recovery occurs at the final physical Host Extensions path. CTK does
not reconstruct extensions elsewhere and move them into place, because a
Platform may bind installed Extension metadata or fingerprints to the actual
installation location.

If restoration fails, the persisted transaction state is used to reconnect the
previous current Runtime when that rollback is safely determinable.

### Process coordination

Host lifecycle operations stop only Runtime processes that can access paths
being replaced:

- Activation stops the default Platform Runtime launched without isolated
  user-data and extensions arguments.
- Deactivation stops that default Runtime and the isolated Runtime referenced
  by the current Selection.
- A recovery adapter may additionally stop only the staging Runtime it started
  for Platform database operations.
- Other independently launched Distributions are not stopped solely because
  they use the same Platform command.

### Persistent transaction journal

Before its first host mutation, activation or deactivation persists a journal
containing the operation, Platform, resolved paths, backups, prior Selection,
and current phase. Each externally visible transition advances that journal
atomically.

On a later invocation, CTK detects an unfinished journal before beginning a new
lifecycle operation. It rolls back or resumes only when the recorded state and
observed filesystem make that action unambiguous. Otherwise it preserves the
journal and diagnostics and returns a hard failure rather than guessing.

The journal representation and phase names are implementation details. Its
persistence before host mutation and its ability to identify the previous
usable state are safety requirements.

### Safety Gate boundary

Semantic differences between the source Lock, temporary Recipe, and recovered
Lock may be accepted explicitly:

- interactive execution offers Abort or Force
- non-interactive execution requires an explicit `--force`
- forced continuation retains the temporary Recipe, source Lock, recovered
  Lock, and Verification Report as diagnostics

Malformed provenance, unreadable trusted state, failed Platform operations,
ambiguous transaction recovery, and an origin that cannot be constructed are
hard failures and cannot be bypassed by Force. Forced deactivation may repair
an incomplete host integration only when its trusted origin and prior managed
state make the restoration deterministic.

An implementation may provide a separate explicit empty-deactivation escape
hatch. This is not trusted-origin restoration and must not be represented as
such. It discards the Host Runtime connection, initializes empty physical Host
User and Extensions paths, preserves the selected Distribution, and removes
the active Selection only after Platform stop and recoverable host-path backup.
It must be rejected for an already inactive Platform and must require an
unambiguous destructive option in non-interactive use.

## VS Code-family Profile Integration

### Required Capabilities

- CTK can represent named IDE Profiles as purpose-specific Runtime differences.
- The Platform integration preserves the distinction between default Runtime
  content and Profile-local content.
- Profile identity crosses operating-system boundaries without encoding or line
  ending corruption.

### Current representation

- The current adapter reads VS Code-family Profile metadata from Platform user
  data storage.
- `useDefaultFlags` identifies default-inherited and Profile-local content.
- Windows command output is normalized at the adapter boundary.

The VS Code-family storage schema is an adapter dependency, not a general CTK
contract.

---

## Required Source Compatibility

Code Environment implementations must preserve Cookbook meaning and Concept API
behavior. They do not need to preserve another implementation's process
commands, link layouts, temporary files, or Distribution internals.

---

## Open Questions

- How a VS Code ecosystem Platform advertises optional capabilities that differ
  from the currently observed CLI behavior.
- How CodeVenv state is represented without a retained implementation bridge.
- How packaged Distributions expose Lock observation.

## Implementation-specific resolution

The primary implementation's Selection links, origin representation,
transaction and process-lock strategy, force recovery, and retained Bash
interoperability are defined by the
[Go Code Environment Contract](../../go/doc/contract/contract.code-environment.md).

The reusable observation sequence is preserved in the
[VS Code Ecosystem Platform Intake Note](../project-knowledge/note/note.vscode-ecosystem-platform-intake.md).
The implemented declaration and service boundary is described in the
[Built-in Platform Registry Note](../note/note.platform-registry.md).
Current Go implementation coverage and Platform-specific observations are
inventoried in the
[Go Platform Support Inventory](../../go/doc/platform-support.md).
