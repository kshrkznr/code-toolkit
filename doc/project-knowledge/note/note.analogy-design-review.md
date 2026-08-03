# project-knowledge.note.analogy-design-review.md
============================================================

# Analogy as a Design Review Tool

This Note records one secondary effect observed while using Analogy during CTK
design.

Analogy was introduced to make unfamiliar concepts easier to understand. During
discussion, it also helped expose awkward responsibilities and unnecessary
abstractions.

This is an example of Analogy use, not a required CTK design method.

## Observation

When a responsibility boundary felt unnatural but was difficult to explain, the
design was temporarily mapped to a familiar model.

```text
Current design
    ↓
Map it to a familiar model
    ↓
Observe unnatural responsibilities or missing concepts
    ↓
Return to CTK
    ↓
Refine the actual responsibility model
```

The familiar model did not decide whether the design was correct. It provided a
different view from which inconsistencies were easier to notice.

## Cookbook observation

CTK originally explored generic concepts such as `Component` and `Module` for
reusable Recipe material.

Mapping the design to the cooking vocabulary made a different boundary visible:

```text
Cookbook
├── Recipes
└── Ingredients
```

The useful result was not that cooking terminology sounded better. The mapping
made the responsibility "material used by a Recipe" easier to see, while the
generic intermediate abstractions became less necessary.

The resulting `Cookbook`, `Recipe`, and `Ingredient` concepts were then defined
by CTK responsibilities rather than by the metaphor.

## Boundary

Analogy does not provide design authority.

- A natural mapping does not prove that a responsibility is correct.
- CTK does not need to reproduce the referenced system.
- Temporary details of the thought experiment do not need to survive.
- The result must still be explained using CTK's own concepts and boundaries.

The durable output of an Analogy review is the responsibility discovered or
clarified. The metaphor itself is retained only when it continues to help.

## Related documents

- `note.analogy.md` — why Analogies are preserved and their different lifetimes
- `../../design-note/design-note.cookbook.md` — the resulting Cookbook design rationale

