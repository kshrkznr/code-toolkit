# project-knowledge.design-note.md
============================================================

# Project Knowledge Design Notes

Project Knowledge Design Notes preserve why CTK's Knowledge and documentation
structure took its current shape.

They record design rationale that has become clear enough to explain as a
decision, while leaving the accepted responsibility in the document that owns
it today.

## Responsibility

A Project Knowledge Design Note is responsible for preserving the reasoning
behind a consequential documentation or knowledge-structure decision.

Use this area when the useful question is why a structure emerged, which
problem it addressed, or where the resulting decision stops applying.

## Navigate by question

Start with the question closest to the current work:

- Why are related Concept APIs grouped into Concept Domains?
  → [Why Concept Domains Emerged](design-note.concept-domain.md)
- Why does CTK describe its concepts through Concept APIs?
  → [Design Idea: Concept API](design-note.concept-api.md)
- Why did the root README become a navigator rather than a complete manual?
  → [Why README Became a Navigator](design-note.documentation-onboarding.md)

These routes cover the current shelf, but the README is an entry point rather
than a second explanation of the decisions.

## Boundary

Project Knowledge Design Notes do not define:

- accepted Concept API or Concept Domain responsibilities
- required documentation structure or heading templates
- behavioral or representation Contracts
- unresolved questions whose exploration is still active

Use [Experiments](../experiment/README.md) while the question, alternatives, or
observation path remains the important material. Use
[Project Knowledge Notes](../note/README.md) for reusable observations that do
not need to preserve a consequential design decision.

The main [CTK Design Notes](../../design-note/README.md) explain consequential
choices in CTK's product and concept design. This directory focuses on the
evolution of CTK's Knowledge and documentation model.

## Local guidance

A Design Note should make the decision and its applicability understandable.
Background, alternatives, result, and boundary are useful when they fit the
subject, but no fixed heading template is required.
