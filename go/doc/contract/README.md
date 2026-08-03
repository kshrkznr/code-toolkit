# Go.contract.md
============================================================

# Go Implementation Contracts

These Contracts define observable behavior, representations, and safety
strategies selected by the primary Go implementation of CTK.

Read the corresponding shared Contract first. Shared Contracts define what CTK
implementations agree on; Go Contracts define how this implementation satisfies
that agreement where callers, persisted state, or recovery behavior depend on a
concrete choice.

## Navigate by question

- How does Go interpret Cookbook Source?
  → [Cookbook](contract.cookbook.md) and
  [Cookbook Resolution](contract.cookbook-resolution.md)
- How does Go represent and launch a Distribution?
  → [Distribution](contract.distribution.md)
- How does Go preserve and reconstruct an exact Runtime?
  → [Archive](contract.archive.md)
- How does the native CLI select values and handle cancellation?
  → [CLI](contract.cli.md)
- How does Go activate, switch, and recover CodeVenv state?
  → [Code Environment](contract.code-environment.md)
- How does Go publish Locks and recover Runtime state?
  → [Runtime Convergence](contract.runtime-convergence.md)
- Which additional Runtime Artifacts does Go support?
  → [Runtime Artifacts](contract.runtime-artifacts.md)
- How does Go render and commit Workbench Artifacts?
  → [Workbench](contract.workbench.md)

## Boundary

These documents do not redefine Core or shared Contract responsibilities. An
internal package detail belongs here only when changing it would alter an
observable command, persisted representation, interoperability boundary, or
safety guarantee.

Historical milestone labels and completed migration plans are not Contracts.
They remain in Project Knowledge when their development context is useful.
