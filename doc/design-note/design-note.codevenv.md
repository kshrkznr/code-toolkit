# Knowledge.design-note.codevenv.md
============================================================

# Why CodeVenv Remained

## Background

CodeVenv began as an experiment in applying Python virtual-environment ideas to
IDE Runtimes.

CTK later gained `launch`, which can start an isolated Distribution without
changing host state. That raised a reasonable question: why retain the heavier
CodeVenv host integration at all?

## The retained value

CodeVenv preserves the Platform's ordinary entry point.

```text
ctk activate code
ctk use vscode-golang
code .
```

After activation, users can select a Runtime and continue using the familiar
Platform command, shortcuts, file associations, and other native entry paths.
They do not need to replace every invocation with `ctk launch`.

This made CodeVenv more than an isolation experiment. It connected CTK's
Runtime lifecycle to an existing everyday workflow.

## Why Launch is not equivalent

`ctk launch <dist>` is intentionally lighter:

- it does not require Platform activation
- it does not change the selected Runtime
- it affects only that invocation

This is useful for temporary or parallel use. CodeVenv instead establishes a
persistent selection for one Platform command until another Runtime is selected
or the Platform is deactivated.

Both paths remain useful because they make different host-integration choices.

## The corresponding cost

Preserving ordinary Platform usage requires CodeVenv to redirect Platform-owned
User Data and Extensions paths. That is a host mutation, not merely a launcher
choice.

Activation is therefore explicit. CTK observes and reconstructs an origin
Runtime before taking responsibility for the Platform command, and deactivation
restores that imported state.

The safety guarantees depend on CTK's managed Selection, origin, host
connections, and transaction evidence remaining intact. Manual or concurrent
external mutation of that state is outside the recovery guarantee; CTK should
stop rather than infer a missing history.

The exact safety boundary belongs to the Code Environment Contract, not this
Design Note.

## Decision

CodeVenv remains the persistent, host-integrated Runtime-selection path.

Launch remains the activation-free path for one temporary invocation. The
availability of Launch does not remove the native-workflow value that justified
keeping CodeVenv.

## Related documents

- `../integration/integration.code-venv.md` — current Code Environment behavior
- `../contract/contract.code-environment.md` — host-integration safety
  boundary
- `../note/note.codevenv.md` — practical CodeVenv observations
