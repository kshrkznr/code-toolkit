# Knowledge.note.codevenv.md
============================================================

# CodeVenv Operations

This Note collects practical observations for using CodeVenv. Stable lifecycle
and responsibility belong to
[`integration.code-venv.md`](../integration/integration.code-venv.md), and host
safety requirements belong to the Code Environment Contract.

## Choose persistent selection or temporary launch

Use CodeVenv when the Platform's ordinary command should use a selected
Runtime:

```text
ctk activate code
ctk use vscode-golang
code .
```

Use `ctk launch <dist>` when one Distribution should start temporarily without
activating its Platform or changing the selected Runtime.

```text
Persistent selection             Temporary invocation
activate → use → code .           launch <dist>
host-integrated                   activation-free
remains until changed             one invocation
```

Launch is an adjacent Code Environment path, not a CodeVenv selection
operation.

## Platform command is the integration key

A Platform command is the normal CLI entry point for a compatible VS Code-family
application, such as `code` or `kiro`.

The command is a better integration key than a product label or Distribution
name because it identifies the host entry point CodeVenv manages. Recipe
provenance declares the command, so `use` does not infer it from filenames.

Selection is independent per activated command:

```text
current.code
current.kiro
```

Changing the selected Runtime for `code` does not change `kiro`.

## Activation is an explicit host decision

`activate <platform>` authorizes CodeVenv to manage the normal invocation of
that Platform command.

Activation observes the current default environment, constructs a trusted
origin Runtime, and only then redirects the Platform's User Data and Extensions
paths. This is intentionally different from merely selecting or launching a
Distribution.

Re-running activation for a managed Platform is a health check. A healthy
Selection and host connection produce a no-op. An incomplete or inconsistent
integration is reported rather than silently repaired.

## User Data and Extensions move together

For the current VS Code-family integration, User Data and installed Extensions
form one Runtime-switching boundary:

```text
Runtime identity
  = user-data-dir (.data)
  + extensions-dir (.ext)
```

Sharing or moving only the installed Extension directory is unsafe. Platform
metadata and fingerprints may depend on the complete Runtime state or its
physical location.

The complete `.data` directory is also broader than `.data/User`. Window,
session, cache, and other Platform-owned state may follow a selected Runtime
even when CTK does not interpret that state as a Cookbook Resource.

This is an observed VS Code-family storage boundary, not a promise that every
Platform keeps all authentication, agent history, indexes, or cloud state under
those paths.

## Recovery is explicit

Normal deactivation restores the imported default Runtime and removes CodeVenv
management for one Platform.

An unhealthy integration may offer `deactivate <platform> --force` only when
trusted origin and transaction evidence make recovery deterministic. CTK stops
when recovery would require guessing.

Do not manually remove or rewrite `current.<platform>`, `origin.<platform>`,
host connections, transaction journals, or trusted Lock files while a Platform
is managed. Such external mutation is outside the recovery guarantee.

`--force-empty` is a destructive last resort. It establishes empty physical
host paths; it does not restore the imported Runtime.

For the specialized current-Go procedure that adopts an existing completed
Distribution as recovery input, see
[`note.go-codevenv-origin-recovery.md`](note.go-codevenv-origin-recovery.md).

## Launch customization

A Distribution may provide a Launch Override, and current implementations may
also generate direct launchers. These are launch representation choices rather
than CodeVenv selection behavior.

See
[`design-note.direct-launcher-packaging.md`](../design-note/design-note.direct-launcher-packaging.md)
for the current packaging decision. Exact implementation boundaries remain in
the corresponding Contracts and language documentation.

## Related documents

- `../design-note/design-note.codevenv.md` — why CodeVenv remained
- `../integration/integration.code-venv.md` — current lifecycle and behavior
- `../contract/contract.code-environment.md` — host safety boundary
