# project-knowledge.experiment.ai-assisted-knowledge-lifecycle.md
============================================================

# Experiment: AI-Assisted Knowledge Lifecycle

## Starting question

Can AI help maintain the knowledge lifecycle without requiring contributors to
adopt a large set of new documentation and operating rules?

Additional process tends to become harder to sustain as its maintenance cost
grows. This experiment asks whether the human-facing part can remain small
while AI assists with organization, classification, and candidate destinations.

## Hypothesis

A lightweight practice may be enough:

- discuss the work with AI or record an ordinary memo
- leave a summary at useful transition points
- occasionally inventory the accumulated material

AI may then assist with organizing the material and proposing classifications
or promotion candidates. The owner still decides responsibility boundaries and
the final destination.

## Multiple entry paths

An AI chat should not be required. A contributor might begin with a simple
memo such as:

```text
tmp/
    2026-07-25.md
```

The memo could later be given to an AI assistant with a request to organize it
as Raw material. Both entry paths can join the same lifecycle:

```text
Conversation / Memo
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
    ├── Curated Knowledge
    └── Drop
```

Entry paths are flexible. Destinations branch according to information
responsibility rather than forming fixed maturity stages. See
[`Project Knowledge`](../README.md) for the current lifecycle and document
roles.

Unorganized or unsettled information can remain in Raw or another source
record. During inventory, related material may be arranged in a temporary View
or Workbench that people and AI can inspect together. Only the decisions need
to move into durable documents or a Reviewed source; the temporary view may be
discarded.

Earlier discussions used `Inbox` for a similar role. The nearest current role
is this disposable review context, not a fixed document type or storage path.

## Team Knowledge hypothesis

Personal Knowledge does not need to be shared unchanged. Only sufficiently
reusable and reproducible observations would be proposed as team Knowledge.

A review could ask:

- Is this appropriate as shared team knowledge?
- Is it reusable outside the original conversation?
- Is it still an unsettled Future direction?

The review would judge responsibility and reuse, not merely prose quality.

## CTK boundary

CTK does not manage AI systems. It may support the lifecycle of artifacts that
emerge from collaboration involving AI.

The approach should not depend on a specific assistant such as ChatGPT, Kiro,
Copilot, or Cursor.

## What remains open

- Is the three-part human-facing practice small enough to sustain?
- Which classifications can AI propose reliably without obscuring uncertainty?
- What evidence is sufficient before personal material becomes team Knowledge?
- Does the practice remain useful when the initial material did not come from
  an AI conversation?
