# Go.contract.cli.md
============================================================

# Go CLI Contract

This Contract specializes the shared
[CLI Interaction Contract](../../../doc/contract/contract.cli.md).

## Native Selector

- Interactive selection is implemented inside the Go binary and does not
  require an external selector executable.
- Single selection is primary; filtering is an optional presentation aid.
- Zero candidates report no selectable target.
- One candidate is selected without initializing the terminal UI.
- Multiple candidates open the Native Selector only in interactive execution.
- Explicit command arguments always bypass selection.

The CLI depends on a focused Selector abstraction. Application services produce
deterministically ordered candidate values, domain operations receive explicit
values, and concrete terminal UI types and errors do not cross the boundary.
Selector behavior and cancellation mapping are independently testable.

The current terminal presentation uses `huh` behind that abstraction. Replacing
the library must not change command or domain behavior.

## Cancellation

Cancellation performs no operation, changes no state, and exits with status
`130`. It is distinct from both application failure and successful selection.

## Default invocation

Invoking `ctk` without a command opens Native command selection. `ctk help` and
`ctk select` remain explicit operations, and all domain commands remain usable
without interactive selection.

This interactive-first default is a Go product preference rather than a shared
Core capability.

## Workspace-independent self-description

The standalone Go executable dispatches commands that describe content carried
by that executable before Workspace discovery and validation.

The current self-description boundary contains:

- `ctk help`, `ctk -h`, and `ctk --help`;
- `ctk version` and `ctk --version`;
- `ctk docs` and its documentation lookup subcommands.

These commands do not read Cookbook Source, Dist, Archive, Extension Pool,
Workbench, or Host integration paths. Their argument validation and usage
errors also occur without requiring a Workspace.

### Help context observation

`ctk help` prints static command help first, then a concise best-effort summary
of the context the current invocation would use. It reports:

- the selected Workspace path and whether selection came from `CTK_HOME`, the
  current-directory ancestry, or the executable-relative fallback;
- a Workspace discovery or configuration diagnostic when selection or
  validation fails;
- the version and revision of the packaged Documentation Bundle, or a
  diagnostic when the executable does not carry a valid Bundle.

Context observation does not load Cookbook content, Dist, Archive, Extension
Pool, Workbench, or Host state. A missing or invalid Workspace and a missing
Documentation Bundle do not suppress static help or change its successful exit.
Paths below the current user's home use `~` in this concise display; paths
outside home remain visible.

Help does not discover or select a local documentation clone. Packaged
Knowledge remains the default. Detailed packaged/local provenance and mismatch
diagnostics remain the responsibility of `ctk docs status` and an explicit
per-invocation `--source` selection.

All lifecycle, Runtime, Cookbook, Workbench, and Workspace inspection commands
retain the existing Workspace discovery and validation boundary. Classifying a
new command as read-only is not sufficient to move it into self-description;
the command must describe only content carried by the executable.

The reusable distribution observation is recorded by [Binary
Self-Description Outside a
Workspace](../../../doc/note/note.binary-self-description.md). Documentation
lookup and packaged representation are defined by the [Go Documentation Bundle
Contract](contract.documentation-bundle.md).

## Workspace-independent bootstrap and completion

Two additional commands run before Workspace discovery without becoming
lifecycle operations or extending self-description:

- `ctk init <path> [--exclude-sample]` creates the optional footing defined by
  the [Go Workspace Contract](contract.workspace.md);
- `ctk completion <bash|zsh|fish|powershell>` writes a shell-completion script
  to standard output.

Completion is static. It completes CTK commands, subcommands, options, and
closed option values without reading Recipe, Distribution, Cookbook,
Workbench, Archive, Extension Pool, Host, or Workspace state. Its hidden shell
protocol follows the same Workspace-independent boundary. The generated
scripts are suitable for direct installation by a binary package definition;
the command does not modify shell configuration itself.

## Explicit collision behavior

Go exposes conflict choices through command-specific `--on-conflict` values.
Build supports `suffix|abort` and diagnostic `--keep-staging`; Archive also
supports `replace`; Workbench operations support `abort|replace`.

When Build conflict behavior is omitted interactively, the Selector offers the
next numeric suffix or Abort. Non-interactive Build defaults to `suffix` and
may request `abort` explicitly.

## Package boundary

Selector code belongs to the CLI boundary rather than Cookbook, Distribution,
CodeVenv, or lifecycle domain packages. Commands may share the Selector but
must not move selection into domain services.

## Open Questions

- Whether filtering should be immediately active or entered through a key.
- Whether future candidates need descriptions or previews.
- Whether the focused Selector should remain a prompt or grow into a broader
  terminal application.
