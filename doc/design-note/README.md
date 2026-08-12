# Knowledge.design-note.md
============================================================

# Design Notes

Design Notes preserve the rationale behind consequential CTK design choices.

They explain why a direction was adopted, why an alternative was left behind,
and which boundary keeps the decision applicable. The accepted responsibility
itself remains in Core, Workbench, Integration, or a Contract.

## Responsibility

A Design Note should make a past or current design choice understandable
without turning its history into a new specification.

Use a Design Note when the question is about reasoning, trade-offs, or design
evolution. Use the [Documentation Resolver](../README.md) when the
question is about required behavior or an accepted Concept API.

## Navigate by question

Start with the question closest to the current work:

- Why does CTK use Cookbook, Recipe, and Ingredient vocabulary?
  → [Why CTK Uses Cookbook Vocabulary](design-note.cookbook.md)
- Why are Ingredient Layers semantic vocabulary rather than a deeper
  hierarchy?
  → [Why Ingredient Layers Are Vocabulary, Not Hierarchy](design-note.ingredient-layers.md)
- Why was CodeVenv retained after activation-free Launch became available?
  → [Why CodeVenv Remained](design-note.codevenv.md)
- Why does CTK keep readable Distribution-local launchers instead of `.app`
  or `.exe` packaging?
  → [Why CTK Keeps Distribution-Local Launchers](design-note.direct-launcher-packaging.md)
- Why does CTK leave secret identification and resolution outside its current
  responsibility?
  → [Why CTK Does Not Manage Secrets](design-note.secret-management.md)
- Why is direct VSIX acquisition separate from normal Platform installation?
  → [Why CTK Keeps VSIX Acquisition Explicit](design-note.vsix-acquisition.md)
- Why does a Platform definition stay inside the VS Code ecosystem instead of
  selecting JetBrains, Eclipse, or another Runtime Adapter?
  → [Why CTK Keeps Platform Inside the VS Code Ecosystem](design-note.vscode-ecosystem-scope.md)
- Why are external Platform definitions Workspace integration configuration
  rather than Recipe Source, Kitchen Notes, or user-global settings?
  → [Why Platform Definitions Belong to Workspace Integration](design-note.platform-definition-scope.md)
- Why is *Workbench Builder* retained only as a historical lens?
  → [Workbench Builder as a Historical Lens](design-note.workbench-builder.md)

These routes cover the current Design Notes, but the README is an entry point
rather than a second specification. Search by the decision or concept when a
future note is not yet listed here.

## Boundary

Design Notes do not define:

- accepted Concept API responsibilities
- required behavior or representation Contracts
- implementation-specific operational procedures
- candidate directions that have not yet reached a decision

Operational observations belong in [Notes](../note/README.md). Unsettled
questions belong in Future or Experiment documents. When a Design Note and an
accepted document differ, resolve current responsibility through the accepted
document and use the Design Note to understand how that responsibility arose.

## Local guidance

A title should expose the decision or question being explained. Background,
alternatives, decision, trade-offs, and boundary are useful headings when they
fit the subject, but Design Notes do not require one fixed template.
