# Knowledge.contract.cli.md
============================================================

# CLI Interaction Contract

This document records the boundary between explicit CTK commands and optional
interactive selection.

Interactive selection helps a person provide a command argument. It does not
change the responsibility of the operation that receives the selected value.

## Explicit Commands

### Required Contract

- Every operation has an explicit command form suitable for shell history,
  documentation, tests, and automation.
- Supplying all required arguments bypasses interactive selection.
- Domain operations receive explicit values and do not invoke or depend on an
  interactive UI implementation.
- Interactive selection must resolve to the same input accepted by the explicit
  command form.

Examples:

```text
ctk use vscode-golang
ctk launch vscode-golang
ctk deactivate code
ctk build vscode-golang.macos.yaml --on-conflict suffix
ctk apply vscode-golang.macos.yaml vscode-golang.1
ctk lock vscode-golang.1
```

### Open Questions

- Whether every future interactive action must expose an equivalent explicit
  subcommand or flag.
- Whether commands should provide a machine-readable mode in addition to their
  explicit human-readable form.

---

## Selector

### Required Contract

When an operation allows its target argument to be omitted, selection follows
one general rule:

| Candidate count | Result |
| ---: | --- |
| 0 | Report that no selectable target exists. |
| 1 | Select the only candidate without opening an interactive UI. |
| 2 or more | Ask the user to select one candidate through an interactive Selector. |

- Candidate generation belongs to the relevant application service.
- Candidate presentation and selection belong to the CLI Selector.
- The selected value must be one of the supplied candidates.
- Candidate ordering must be deterministic before the Selector is opened.
- A Selector must not perform the operation associated with the selected value.
- The Selector may provide filtering without changing candidate values or the
  single-selection contract.

Conceptual boundary:

```text
Application service
    │
    └── candidates
            │
            ▼
       CLI Selector
            │
    explicit selected value
            │
            ▼
     Domain operation
```

### Open Questions

- Whether candidates need display labels distinct from the value passed to the
  operation.
- Whether Selector preview content becomes useful for Recipes, Distributions,
  or Archives.

---

## Cancellation

### Required Contract

- Cancelling interactive selection performs no operation and changes no state.
- Cancellation is distinct from both successful selection and application
  failure.
- The CLI exits with status `130` after an interactive selection is cancelled.
- CTK must not present cancellation as an application failure.

Cancellation represents an intentional incomplete interaction: it is neither a
successful operation nor an application error.

### Open Questions

- Whether a concise cancellation message is useful after the Native Selector
  closes.

---

## Default Invocation

### Required Contract

- CTK provides Help as an explicit operation.
- CTK may provide command selection as an explicit interactive operation.
- Changing the default no-argument behavior must not remove either explicit
  operation.

Conceptual command forms:

```text
ctk help
ctk select
```

### Open Questions

- Whether no-argument behavior becomes configurable in the future.
- If configurable, whether the preference belongs in a CTK configuration file
  or another explicit mechanism.

Changing this preference later does not change the underlying explicit command
Contracts.

## Implementation-specific resolution

The primary implementation's Native Selector, command defaults, collision
flags, and internal CLI boundary are defined by the
[Go CLI Contract](../../go/doc/contract/contract.cli.md).
