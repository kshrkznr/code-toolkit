# Knowledge.contract.md
============================================================

# Contracts

Contracts define explicit behavioral and representation agreements shared by
CTK implementations.

They preserve the boundaries that an implementation must satisfy without
turning one language's types, packages, files, commands, or migration strategy
into universal CTK requirements.

## Responsibility

Contracts are responsible for distinguishing accepted agreements from current
implementation choices and unresolved questions.

A shared Contract may define:

- **Required Capability** — behavior or information every implementation must
  provide, independent of representation.
- **Required Source Compatibility** — source formats, such as Cookbook files,
  that implementations reuse directly.
- **Safety Invariant** — a property that must hold without prescribing the
  mechanism used to achieve it.
- **Recommended Strategy** — non-binding guidance that remains useful across
  implementations.
- **Current Resolution** — an accepted shared interpretation of a Contract.
- **Open Question** — behavior that has not yet been accepted as a stable
  agreement.

Documents may retain the shorter `Required Contract` heading where all items in
the section belong to one clearly stated category.

## Navigate by question

Start with the boundary closest to the current work:

- How are Recipe and Ingredient sources represented?
  → [Cookbook Representation](contract.cookbook.md)
- How are Cookbook sources resolved into Runtime content?
  → [Cookbook Resolution](contract.cookbook-resolution.md)
- What must a generated Runtime Distribution provide?
  → [Distribution](contract.distribution.md)
- How does CTK integrate a Distribution with an IDE Platform?
  → [Code Environment Integration](contract.code-environment.md)
- What capabilities read and apply Platform Runtime content?
  → [Platform Runtime I/O](contract.platform-runtime-io.md)
- How are Keybindings, Snippets, Tasks, and MCP composed?
  → [Runtime Artifacts](contract.runtime-artifacts.md)
- What safety boundaries govern Build, Apply, Lock, and convergence?
  → [Runtime Convergence](contract.runtime-convergence.md)
- What agreements govern Freeze, Inspect, View, and Sync?
  → [Freeze and Inspect Workbench](contract.workbench.md)
- How are completed Runtimes preserved and reconstructed?
  → [Archive](contract.archive.md)
- What interaction behavior belongs to the CTK CLI boundary?
  → [CLI Interaction](contract.cli.md)

These routes are entry points rather than a second exhaustive Repository Map.
Use the [Documentation Resolver](../README.md) when the current question
belongs to a Concept, Design Note, operational Note, or implementation.

## Implementation-specific Contracts

An implementation may define additional agreements for its own observable
behavior and representation. Those Contracts belong with that implementation.

The primary Go implementation declares its implementation-specific Contracts
under [`go/doc/contract`](../../go/doc/contract/README.md). A Go Contract may
choose a concrete strategy for satisfying a shared Contract, but it does not
redefine the shared CTK boundary.

Reference-implementation details remain in the retained implementation or its
historical documentation unless they are necessary to explain an active
cross-implementation compatibility boundary.

## Boundary

Contracts do not define:

- stable conceptual responsibilities that belong to Core or Workbench
- one language's internal package or type design
- operational advice that does not establish an agreement
- historical migration steps that no longer constrain an implementation
- a universal repository or Cookbook organization

CTK defines transformation and compatibility Contracts rather than one
mandatory implementation architecture.
