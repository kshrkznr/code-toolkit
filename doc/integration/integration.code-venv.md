# Knowledge.integration.code-venv.md
============================================================

# Concept API: CodeVenv

## Definition

CodeVenv provides isolated Runtime selection for IDE environments.

Like Python's `venv`, it lets users select independent IDE Runtimes while preserving the host IDE installation.

A Runtime declares its Platform command in its Recipe metadata. A Platform command is the CLI integration boundary for a VS Code-family application. `code` and `kiro` are examples; other compatible applications may provide their own commands.

---

## Responsibility

CodeVenv is responsible for selecting which Runtime is active for each activated Platform command.

It does not build Runtimes (Build),
nor does it preserve them (Persistence Lifecycle).

Its responsibility is the activation, selection, and restoration of IDE Runtimes.

---

## Analogy

Inspired by Python virtual environments.

Instead of switching Python packages,
CodeVenv switches IDE Runtime environments.

---

## Platform activation

Before CodeVenv can redirect a Platform command, the user explicitly activates that command.

```text
# Examples
ctk activate code
ctk activate kiro
```

When the Platform argument is omitted in an interactive invocation, CTK offers
the intersection of its host-integration adapters and Platform commands found
on `PATH`. An explicitly named Platform remains available for automation and
direct invocation.

Activation observes the Platform into a temporary Recipe and trusted Lock,
then reconstructs an `origin` Runtime before placing the Platform under
CodeVenv management. The origin Extension area starts empty and is rebuilt from
exact Extension IDs rather than moved installation files or metadata.

The default, unprofiled extension state is treated as Runtime state. Named Profile extensions remain purpose-specific differences.

Platform commands are independent. For example, activating `code` does not activate or change `kiro`.

Calling `activate` again for an already-managed Platform performs an activation
health check. When the selected Runtime and both redirected host paths agree,
the command succeeds without changing state. If only part of the integration
remains, `activate` reports the unhealthy state and directs the user to forced
deactivation; it does not silently repair host paths.

## Runtime selection

A Distribution exposes the Platform information from its Recipe provenance.
The current directory representation stores that provenance in
`.meta/recipe.yaml`.

```yaml
# Example: a VS Code Runtime
platform: code
```

`ctk use <dist>` resolves this metadata and changes the selected Runtime for that command.

```text
# Examples
ctk use vscode-default  → selected Runtime for code
ctk use kiro-default    → selected Runtime for kiro
```

The Platform command must be activated before a Runtime can be selected. This keeps host integration an explicit user decision.

CodeVenv exposes the selected Runtime through `ctk current`.

```text
# Example
ctk current
code: vscode-default
kiro: kiro-default

ctk current code
vscode-default
```

Without an argument, `current` shows all activated Platforms. With a Platform command, it shows only that Platform's selected Runtime.

## Launch

Code Environment provides two adjacent usage paths.

- CodeVenv persistently changes the Runtime selected for an activated Platform
  command through `ctk use`.
- `ctk launch <dist> [targets...]` temporarily starts a Distribution without
  changing any selected Runtime.
- `ctk launch "" <targets...>` selects the Distribution interactively while preserving explicit file or directory targets.

This separation allows permanent environment switching and temporary experimentation.

`launch` does not require Platform activation. It resolves the Distribution's
Launch capability without changing the selected Runtime. A Launch Override,
native Platform adapter, application bundle, executable, or another
representation may provide that capability.

Launch is therefore part of Code Environment Integration, but it is not a
CodeVenv selection operation.

## Deactivation

`ctk deactivate <platform>` restores the host state for one Platform command and removes it from CodeVenv management.

`ctk deactivate <platform> --force` also provides the explicit recovery path
for an incomplete activation. CTK preserves unexpected physical host content
or link state before reconstructing the imported default environment. It does
not continue when the previous managed state is not trustworthy enough for a
safe recovery.

When only one Platform is active, `ctk deactivate` selects it automatically.
When multiple Platforms are active, the user selects one through CTK's
interactive Selector.

### Observation

CodeVenv redirects the Platform's User directory and extensions directory while leaving the host IDE installation intact.

Some IDE state (for example window/session/cache) may still be stored outside the redirected User directory.

This does not affect Runtime reproducibility because CTK treats the Runtime itself as self-contained.

---

## Conceptual diagram

```text
Recipe metadata
    │
    ▼
Platform command
    │
    ├── ctk activate <command>
    │       │
    │       ▼
    │   Default state → temporary Lock → origin Recovery
    │
    └── ctk use <dist>
            │
            ▼
      Selected Runtime
            │
            ▼
      Normal command usage

Distribution
    │
    └── ctk launch <dist>
            │
            ▼
      Temporary Runtime
```
