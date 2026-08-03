# project-knowledge.experiment.raw-inventory-review.md
============================================================

# Experiment: Reviewing Raw by Theme and Responsibility

## Review status

**Reviewed after completion.**

The ordinary Raw shelves were inventoried and retired. The reusable editing
observation is recorded separately in
[`note.raw-inventory.md`](../note/note.raw-inventory.md).

This Experiment remains as a record of what was actually tried, which judgment
lenses mattered, and how the process felt after one complete pass.

## Starting question

Could CTK reduce its accumulated Raw material without treating deletion as the
goal, flattening useful design history, or forcing immature ideas into
specification-like documents?

The Raw shelves had originally served as a place to avoid losing early ideas.
As Core, Note, Design Note, Project Knowledge, Future, and Experiment acquired
clearer responsibilities, the same files increasingly mixed:

- accepted concepts
- obsolete vocabulary
- design rationale
- practical observations
- analogies that had helped understanding
- thought experiments
- implementation-era assumptions
- placeholders with no conclusion

A file-level Keep / Delete decision would have been too coarse.

## Working hypothesis

A temporary staging layer and a theme-first review could make each decision
small enough to discuss.

```text
Mixed Raw
    ↓ collect one theme
Temporary inventory view
    ↓ compare with current Knowledge
Adopt / Preserve / Discuss / Drop
    ↓ author review
Durable destination or deletion
```

The staging view was intentionally disposable. It held comparisons and
proposed decisions, while tracked Knowledge and reviewed thought records held
the durable result.

## Procedure that emerged

The inventory was not executed as one fixed plan. The following cycle emerged
through repeated batches:

1. List cross-cutting themes before editing source files.
2. Select one bounded theme or heading.
3. Read the current Knowledge responsible for that subject.
4. Collect only the related Raw fragments and implementation evidence.
5. Propose `Adopt`, `Preserve`, `Discuss`, or `Drop` for each function.
6. Discuss the boundary and destination with the author.
7. Apply only the accepted change.
8. Inspect the diff and leave uncertain surrounding text for later.
9. Retire a mixed source only after all of its themes have reached decisions.

Later in the work, source retirement became a separate pass. This was useful:
promotion decisions were made while the source still existed, and deletion
became a consequence of completed review rather than an objective imposed at
the beginning.

## Judgment lenses used in practice

### Current responsibility

The current Core, Workbench, Integration, and other curated documents were the
baseline. Historical Raw wording did not silently regain authority merely
because it was older or more detailed.

### Function of the fragment

The review asked what a passage was doing before asking where it should live.
It might be:

- explaining a concept
- preserving why a decision was made
- recording an operational observation
- proposing a future direction
- testing a hypothesis
- providing an Analogy
- repeating current material
- preserving an obsolete assumption

This often split one apparent topic into several document responsibilities.

### Destination and authority

The destination followed the information's role, not a linear maturity ladder.
Note, Experiment, Design Note, Future, and accepted Knowledge were sibling
outcomes of review.

An attractive or well-written passage did not automatically deserve a stronger
document role.

### Understanding value

The inventory did not ask only:

> Is this part of the accepted responsibility?

It also asked:

> Did this explanation help us reach the current understanding?

This lens rescued useful Analogy material without allowing Analogy to replace
Core Concepts.

### Thought-path value

Experiments were not deleted merely because their observations had been
processed. When the path, rejected options, or original question remained
useful, the Experiment stayed as a Reviewed record and linked to its durable
outcomes.

Mixed operational Raw did not receive the same protection automatically. It
was retired when its useful fragments had destinations and no bounded thought
record remained.

### Duplication and drift

Text could be removed confidently when it only repeated current material, used
obsolete names, described an implementation that no longer existed, or implied
a stronger rule than the current design accepted.

The useful intention was sometimes rewritten in a smaller form before the old
wording was removed.

### Uncertainty

When a boundary remained unclear, the default was to keep the fragment or hold
the theme rather than invent a destination.

This mattered most for Raw Contract. The inventory used current Contracts as
evidence where necessary, but did not decide their long-term document role.
Contract review remained separate and incremental.

## What was observed

### Theme-first review reduced accidental coupling

One Raw file frequently contained README copy, Core candidates, AI
collaboration observations, old commands, and design-history fragments.
Reviewing a theme across files made comparison easier than promoting or deleting
a whole file.

### A destination made deletion easier

The difficult question was rarely whether a paragraph could be deleted. It was
whether its useful function survived elsewhere.

Once that question had an answer, later source retirement was mostly
mechanical. The last large deletions did not feel destructive because the
important decisions had already been made in smaller conversations.

### The staging layer acted as a discussion cushion

The temporary inventory created a place where a proposal could be visible
without immediately changing Raw or declaring a final Knowledge role.

This made `Discuss` a real state rather than a delay hidden inside a document.
It also allowed the author to correct interpretation before tracked material
was changed.

### The inventory clarified document roles

The work did more than sort existing text. Repeated cases sharpened several
boundaries:

- Note preserves reusable observations and loose guidance.
- Experiment preserves the path of thought.
- Design Note preserves product design rationale and evolution.
- Future preserves unsettled candidate directions.
- Project Knowledge describes how CTK itself was developed.
- Inbox was closest to a temporary review context, not a required path or
  permanent document type.

### Ordinary Raw and `raw`-named documents were not one category

Completing the ordinary Raw shelves did not authorize deleting
`experiment.raw.*`, reclassifying `contract.raw.*`, or folding Raw-labelled
Notes and Design Notes into the same cleanup.

The filename recorded history; the containing document role determined the
review lifecycle.

## Friction and limitations

The process was effective, but not lightweight.

- The temporary inventory became large enough that its own next-step pointers
  occasionally became stale and needed housekeeping.
- Some themes depended on earlier decisions, so the order could not be fully
  parallelized.
- Repeatedly comparing Raw, current documents, implementation evidence, and
  prior inventory decisions took time.
- The four statuses helped discussion but did not decide responsibility on
  their own.

The process would be excessive for a small, coherent note with an obvious
destination. It was useful because CTK's Raw had accumulated across several
stages of conceptual change.

## Reflection after one pass

The work felt less like cleaning a folder and more like giving each piece of
information an appropriate home.

The temporary staging layer was especially helpful. It allowed decisions to be
made slowly without turning uncertainty into permanent structure. The author
could inspect each proposed boundary, and the source could remain intact until
the surrounding themes were understood.

It was also important that deletion happened late. By the time the final mixed
Raw files were removed, their disappearance was unsurprising: useful Analogy,
design rationale, operational observations, Future candidates, and thought
records had already found explicit destinations.

The overall procedure felt worthwhile for this repository. That is an
observation from one CTK inventory, not evidence that every project should use
the same statuses, staging directory, theme order, or amount of discussion.

## Durable outcomes

- [`note.raw-inventory.md`](../note/note.raw-inventory.md) — loose reusable
  guidance for theme- and responsibility-based inventory
- [`note.collaborative-review-surfaces.md`](../note/note.collaborative-review-surfaces.md)
  — temporary shared views and inspectable repository-native material
- [`README.md`](../README.md) — current Project Knowledge purpose and the Note /
  Experiment lifecycle

The discarded staging inventory is not required to interpret this record.
