# Knowledge.note.workbench-review.md
============================================================

# Workbench Review

This Note provides practical guidance for reviewing Draft and Inspect
Workbenches.

The accepted responsibilities and lifecycle belong to the [Workbench Concept
Domain](../workbench/README.md). This Note does not define another Workbench
type or Commit contract.

## Choose the Workbench from the question

```text
Known Runtime change to return to Cookbook
    └── freeze draft → Draft → freeze commit

What is in one completed source?
    └── view → Inspect inventory

What differs between two completed states?
    └── sync → Inspect comparison
```

Draft is commit-oriented. Inspect is disposable understanding material. They
share review representations so selected Inspect material can be moved into
Draft, but that does not make Inspect itself a Commit source.

## Read the Workbench as one review context

Open the Workbench directory rather than treating one generated Artifact as the
whole result.

- `summary.md` explains result and provenance.
- `recipe.draft.yaml` preserves or diagnoses Recipe context.
- typed Draft Artifacts contain the detailed Inventory or Difference.

Current implementations may provide different Artifact kinds. Read only the
ones relevant to the current question.

In the current Markdown representation:

- `summary.md` is informational.
- Inventory sections are copyable or reviewable context.
- only operations under `## Difference` in Draft typed Artifacts are Freeze
  Commit input.

The exact Commit boundary belongs to
[`contract.workbench.md`](../contract/contract.workbench.md).

## Reduce structural noise before interpretation

Large generated Artifacts do not always need to be read from beginning to end.
Use a smaller, reversible view when it makes the current question easier to
answer.

Examples include:

- selecting one typed Artifact
- using `view` for one Recipe, Dist, Layer, or Ingredient scope
- using `sync` to isolate a completed-state difference
- extracting one relevant section for discussion
- generating a disposable grouping or report

Mechanical transformation should preserve the observed meaning. Its purpose is
to make review easier, not to decide the final Ingredient responsibility.

## Views are not responsibilities

A review view can group material by prefix, feature, source, or another useful
lens. That grouping is temporary.

One view may later be divided among several Ingredients, and one Ingredient may
receive material found through several views. Assign responsibility during the
review; do not infer it from the generated grouping alone.

Review Artifacts are expected to be replaced or discarded. Preserve only the
agreed Cookbook change or another intentionally durable outcome.

## CTK JSON Flat Format

An implementation may use CTK JSON Flat Format for line-oriented review of
JSON-like Artifacts. CTK does not require every implementation or Workbench to
use that representation.

When the format is present, see
[`note.ctk-json-flat-format.md`](note.ctk-json-flat-format.md) for its path and
value notation. Markdown headings, Inventory sections, Difference markers, and
Commit targets belong to the Workbench representation rather than to the Flat
Format itself.

## When review needs interpretation

The operational boundary above is usually sufficient. When a review also needs
help establishing shared context or deciding how people, AI, and mechanical
tools should participate, see the loose Project Knowledge observation
[`note.collaborative-review-surfaces.md`](../project-knowledge/note/note.collaborative-review-surfaces.md).

That observation is optional guidance, not Workbench behavior.
