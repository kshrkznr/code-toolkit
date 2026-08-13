# project-knowledge.note.vscode-ecosystem-platform-intake.md
============================================================

# VS Code Ecosystem Platform Intake

A VS Code-family application may expose familiar CLI flags without giving CTK
enough evidence to manage its Host Runtime safely.

This Note preserves a reusable sequence for observing a new Platform from its
actual boundaries rather than adding it from its name or apparent compatibility.
It is neither a support certification specification nor a mandatory test
standard.

## Observation sequence

Observe a new Platform in layers:

1. Record the installed application, Platform command, version, architecture,
   application or executable identity, default User data path, and Extension
   path.
2. Use temporary `--user-data-dir` and `--extensions-dir` paths to verify CLI
   isolation before changing the real Host environment.
3. Observe default Settings and Extensions, then exercise Extension install,
   versioned listing, and uninstall behavior against the Platform's own
   Marketplace.
4. Create a named Profile and inspect its storage identity, Profile-local
   Artifacts, inheritance representation, and Extension scope.
5. Observe root and helper process command lines for both the default Runtime
   and an isolated Runtime. Verify that CTK can stop only the intended root
   without inferring process ownership from the application name alone.
6. Reflect those observations in the implementation before exercising
   `activate`, `freeze`, `build`, `apply`, `archive`, `use`, `launch`, and
   `deactivate` with a disposable Cookbook and a recoverable Host environment.
7. On every OS for which support is claimed, observe OS-specific paths, process
   behavior, link representation, line endings, and transaction recovery.

This order separates four questions that are easy to conflate:

- Does the command exist?
- Can Runtime paths be isolated?
- Can the shared Adapter preserve the meaning of Settings, Extensions, and
  Profiles?
- Can CTK manage the Host lifecycle safely?

## When to repeat observation

The full sequence is primarily for the initial intake of a Platform or for
adding support on another OS. It is not a recurring certification checklist for
every upstream application release.

Repeat only the relevant parts when new evidence suggests that an integration
boundary may have changed, for example:

- the Platform command, packaging identity, Host paths, or process tree changes
- Profile storage or CLI behavior changes
- the Platform changes its Gallery, Repository, or Extension policy
- an update introduces a lifecycle regression or an unexplained Runtime
  difference
- CTK begins relying on another Platform capability

A version-number change alone does not require a complete intake pass.

## Evidence to retain

During initial intake or a targeted re-observation, retain evidence that allows
a later review to reconsider the affected boundary, not only the final success
or failure:

- observation date, application version, OS, architecture, and installer
  identity
- Platform command and actual root and helper process identities
- Host User data and Extension paths
- files and database entries created in the redirected Runtime
- Settings and Extension results for each default or Profile scope
- Platform Gallery or Repository identity and acquisition results
- lifecycle operations and remaining processes, links, journals, or backups
- interruption point, recovery result on the next invocation, and semantic or
  byte-level differences
- execution-environment conditions such as TLS proxies and certificate stores

Retain transient product versions and network conditions as observations. Do
not automatically promote them into general Platform requirements.

## Boundary-first interpretation

When unexpected behavior appears, inspect the actual file, command output,
process, link, or Gallery response before changing shared design.

```text
Unexpected behavior
    ↓
Inspect concrete evidence
    ↓
Locate the representation boundary
    ↓
Normalize or select a named strategy at that boundary
    ↓
Change shared responsibility only when repeated evidence requires it
```

For example, CRLF appeared at an ingestion boundary, Junctions at managed-link
resolution, Cursor worker processes at process ancestry, and Profile persistence
at Adapter operation ordering. Failures that look similar should not
automatically be assigned to the same responsibility.

## From observation to implementation status

After adding a Platform to the Go implementation, continue to record these as
separate forms of evidence:

- differences incorporated as declarations or implementation branches
- capabilities covered by shared Adapter tests
- differences covered by Platform-name-specific tests
- OS, product version, and lifecycle range observed on real machines
- unobserved areas and candidates for renewed observation

The
[Go Platform Support Inventory](../../../go/doc/platform-support.md) owns the
current relationship between the Go implementation and real-machine evidence.
This Note owns the observation practice; it does not define the Supported
Platform list or current validation status.

## Related knowledge

- [Platform Differences as Boundary Evidence](../../note/note.platform-boundary-evidence.md)
- [VS Code Ecosystem Platform Intake experiment](../experiment/experiment.vscode-ecosystem-platform-intake.md)
- [Code Environment Integration Contract](../../contract/contract.code-environment.md)
