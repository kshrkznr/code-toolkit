# project-knowledge.note.collaborative-review-surfaces.md
============================================================

# Collaborative Review Surfaces

During CTK development, temporary review surfaces repeatedly helped
collaborators discuss a repository without first agreeing on its final
organization.

This Note records a loose observation from that work. It is not a required
collaboration workflow and does not define the CTK Workbench contract.

## Observation

A useful division emerged:

```text
Mechanical tool
    organizes observable material
        ↓
Participants
    interpret it in the current context
        ↓
Owner
    accepts, rejects, or redirects the change
```

These are roles rather than fixed actors, and the boundaries are intentionally
soft. The same participant may perform several roles. A person or AI assistant
may help generate or interpret a view, while acceptance remains with the owner.
The durable point is that a mechanical grouping should not silently become a
design decision.

## Reduce noise before assigning responsibility

Generated state can be mechanically correct while still being difficult to
review. A reversible inventory, difference, or temporary grouping can make the
question smaller before semantic discussion begins.

This changed the order of some CTK reviews:

```text
Observed state
    ↓
Temporary review view
    ↓
Discuss what is visible
    ↓
Responsibility becomes clearer
    ↓
Record only the agreed result
```

The temporary view does not need to know the final architecture. It only needs
to preserve enough meaning for the participants to reason about it.

## Shared context has several parts

The first collaborative-design experiment found that different documents and
Artifacts supplied different parts of the discussion context:

- Knowledge supplied established concepts and ways of judging them.
- Recipe supplied the current composition and a bounded context for questions.
- Draft supplied concrete differences being considered.
- Inspect later supplied disposable inventories and comparisons without
  implying a pending Cookbook change.

Together they made it possible to discuss responsibility while looking at the
same evidence. No single document replaced the conversation.

## Use material both participants can inspect

Collaboration was easier when all participants could inspect the same ordinary,
repository-native material. Familiar text formats made review and diffs
accessible without requiring a participant-specific representation.

This is an observation about shared visibility, not a requirement to use
Markdown, YAML, or any other specific format. A different representation may
serve equally well when all participants can inspect the relevant evidence.

## Recipe as a conversation anchor

A generic question such as:

> What extensions should I install?

can become more contextual when a Recipe is available:

> Given this Recipe and its purpose, what would you recommend?

The Recipe does not authorize the recommendation. It narrows the environment,
composition, and intent that the participants can inspect together. Other
Knowledge, current state, and user context remain relevant.

## Boundary

- A review surface is an aid, not a responsibility model.
- Mechanical organization does not replace semantic review.
- Interpretation does not replace acceptance by the owner.
- A Recipe provides context, not a universal best practice.
- Temporary review material should not become durable merely because it was
  useful during one discussion.

The observation is useful when a review feels too broad or participants appear
to be reasoning from different evidence. It does not need to be applied to
every change.

## Related documents

- `../experiment/experiment.role-oriented-ai-collaboration.md` — observation
  that collaboration is structured around roles rather than fixed actors
- `../../note/note.workbench-review.md` — current operational review guidance
- `../../workbench/README.md` — accepted Draft and Inspect responsibilities
- `../experiment/experiment.freeze-draft-collaborative-design.md` — primary
  experiment record
