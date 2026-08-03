# Project Knowledge

Project Knowledge preserves **how CTK evolved**, not only **what CTK became**.

Accepted Knowledge describes the project's current concepts, responsibilities,
contracts, and behavior. Project Knowledge preserves the observations,
exploration, and rationale through which that understanding emerged.

## Responsibility

Project Knowledge is responsible for keeping CTK's development history useful
to future work without turning that history into specification.

It may preserve:

- discussions and observations that changed the design
- patterns discovered through implementation, review, or collaboration
- questions and rejected paths that explain how an idea developed
- rationale behind consequential Knowledge or documentation decisions

The material may remain exploratory, become reusable guidance, or explain a
decision. These are different document roles rather than levels of authority.

## Navigate by current need

Choose the Node that matches what should be preserved now:

- Is the question, exploration, uncertainty, or path of thought still the
  important material?
  → [Experiments](experiment/README.md)
- Has an observation become reusable without replaying the full exploration?
  → [Project Knowledge Notes](note/README.md)
- Has a consequential Knowledge or documentation decision become clear enough
  to explain through its rationale and boundary?
  → [Project Knowledge Design Notes](design-note/README.md)

```text
Project Knowledge
    ├── Experiment
    │     preserves exploration and the path of thought
    ├── Note
    │     preserves a reusable observation or loose guideline
    └── Design Note
          preserves rationale behind a consequential decision
```

These Nodes provide local entry points and guidance. They are not a required
promotion sequence or a complete inventory of every document.

## Boundary

Project Knowledge is not CTK's specification.

It does not define:

- accepted Concept API or Concept Domain responsibilities
- architecture or lifecycle that current CTK behavior must follow
- behavioral or representation Contracts
- mandatory repository, Cookbook, or documentation structures

Use the main [Documentation Resolver](../README.md) when the current question
is about accepted CTK Knowledge. If a historical observation conflicts with an
accepted document, the accepted document owns the current responsibility.

Project Knowledge records how an idea emerged; it does not gain authority by
preserving that history.

## Knowledge lifecycle

Project Knowledge participates in a branching knowledge lifecycle:

```text
Conversation
    ↓
Summary
    ↓
Raw
    ↓
Inventory / Review
    ├── Note
    ├── Experiment
    ├── Design Note
    ├── Future
    ├── Core or other curated Knowledge
    └── Drop
```

Inventory and Review determine the responsibility of useful material. The
destinations are roles, not mandatory stages in one maturity ladder.

In particular:

- Note and Future are separate destinations. A Note is reusable guidance;
  Future preserves an unsettled candidate direction.
- An Experiment may remain after a result moves to a Note, Design Note, or
  accepted Knowledge. The path that produced the result may still be useful.
- Promotion does not require erasing uncertainty, rejected paths, or historical
  context that remains valuable.
- Drop means not promoting the material into another durable role. A Reviewed
  source may still remain when it preserves useful context, and Drop should not
  be inferred merely because an observation is unfinished.

## Why preserve this

Once a design becomes familiar, it is easy to forget why it became convincing
or which alternatives shaped its boundary.

Project Knowledge helps future collaborators:

- continue a design conversation without reconstructing all of its context
- revisit earlier reasoning without mistaking it for current authority
- recognize recurring patterns across implementation and review
- understand not only a decision, but the discovery that made it possible

---

> *This directory is less about preserving decisions, and more about preserving discoveries.*
