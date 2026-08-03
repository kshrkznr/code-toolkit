# project-knowledge.experiment.position-and-lens.md
============================================================

# Experiment: Position and Lens in Design Review

## Starting question

Does separating a participant's Position from the Lens used for observation
make design and documentation discussions easier to align?

The hypothesis emerged during CTK development. It has not been adopted as a CTK
concept or required review method.

## Hypothesis

Position describes who a participant is reasoning as:

- Author
- Maintainer
- User

Lens describes what the participant is trying to observe or evaluate:

- Maintainability
- Performance
- Simplicity
- Learning

Separating them may make it easier to distinguish a change of responsibility
from a change of evaluation axis.

```text
Position
    who is reasoning?
        +
Lens
    what are they observing?
        ↓
Explicit review context
```

## Why the distinction may matter

Discussions can become misaligned when Position and Lens are mixed together.

A disagreement between an Author and a Maintainer differs from a tradeoff
between Maintainability and Simplicity. Stating both dimensions may help
participants identify which kind of movement occurred.

## Possible relationship to CTK

The distinction appears compatible with CTK's responsibility-first design:

- Responsibility resembles Position because it identifies an accountable
  role or boundary.
- Inspect and Review can switch observation targets in a way that resembles
  changing Lens.

This is only a candidate correspondence. It does not redefine Responsibility,
Inspect, or Review.

## What to observe

- whether design reviews become easier to align
- whether documentation reviews distinguish audience from quality criteria
- whether AI-assisted conversations can state their current Position and Lens
  without adding distracting ceremony
- whether the distinction produces decisions that were difficult to reach
  without it
