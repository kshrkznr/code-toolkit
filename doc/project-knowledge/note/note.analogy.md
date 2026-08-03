# project-knowledge.note.analogy.md
============================================================

# Analogy as a Bridge

Analogy helped CTK establish shared understanding before its own vocabulary
became natural.

This Note records a loose Project Knowledge guideline. It is not a required
documentation structure, and every CTK concept does not need an Analogy.

## Translation anchor

An Analogy is more than a figurative explanation. It can act as a translation
anchor between a CTK concept and a familiar reference pattern.

```text
CTK Concept
    ↓
Familiar reference pattern
    ↓
Similarity
    ↓
Difference and boundary
```

The familiar pattern provides an entry point. The similarity helps recover the
concept quickly. The difference prevents the reference pattern from replacing
CTK's actual responsibility model.

For example, Python virtual environments help introduce CodeVenv, but they do
not define CodeVenv's responsibilities. Git helps explain Freeze and Draft, but
CTK is not a source-control system.

## When to use an Analogy

If a concept is unfamiliar, begin with the nearest familiar Analogy and then
return to the CTK concept.

If the discussion already uses CTK terminology naturally, continue with the CTK
concept directly. An Analogy is a bridge, not required vocabulary.

## Three lifetimes

Analogies observed during CTK development have different useful lifetimes.

### Public Analogy

A Public Analogy continues to help readers understand a current concept.

Keep it near the concept it explains, such as in Knowledge, Core, or
Integration. State both the similarity and where the comparison stops.

### Design-review Analogy

A Design-review Analogy helps maintainers inspect responsibilities and
boundaries. It does not need to become part of the public explanation.

See `note.analogy-design-review.md` for one observed use.

### Temporary thought experiment

A temporary mapping may be useful during one discussion and have no durable
value afterward.

Preserve the responsibility or observation discovered through it. The mapping
itself may be discarded.

## Preservation guideline

Preserve an Analogy when it still does at least one of the following:

- provides a useful entrance to a current concept
- records why a concept became understandable
- helps reconstruct a responsibility or boundary
- captures a reusable observation about collaboration or review

An Analogy may be removed when it duplicates a clearer current explanation,
describes an obsolete model, or requires more explanation than the CTK concept
it was meant to clarify.

The goal is not to maintain an Analogy catalog. The goal is to retain useful
bridges into the current concepts.

