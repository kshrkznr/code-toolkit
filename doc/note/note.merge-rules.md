# Knowledge.note.merge-rules.md
============================================================

# Merge Rules as a Kitchen Note

## Purpose

This Note describes Merge Rules as one example of Kitchen Notes. It explains
the reusable idea and its responsibility boundary; the existence of this Note
does not declare that every language implementation applies the rules.

Each language README owns that adoption decision, while its implementation
documents own the concrete representation and Contracts.

## Why Merge Rules exist

Some Ingredient Resources cannot be combined correctly through one generic
value operation. Their target ecosystem may require additional rules for cases
such as conflict resolution, list replacement, or list Union.

Merge Rules describe how an implementation combines multiple representations
while preserving the effective behavior expected by that ecosystem.

For example, if several JSON Settings Resources resolve into one document, the
result should behave as if those Resources were applied in Cookbook resolution
order.

## Responsibility boundary

Recipe order, Resource resolution order, and Variant precedence are Cookbook
Core behavior. Merge Rules consume that ordered stream; they do not reorder it
or override those responsibilities.

```text
Cookbook Core resolution
        ↓ ordered Resources
adopted Merge Rule
        ↓ combination semantics
generated Artifact
```

Merge Rules belong to the Cookbook and affect its interpretation during Build
or another lifecycle operation that resolves Resources. They do not add a Core
concept or change the responsibility of the generated Artifact.

Different implementations may express the same Rule differently, adopt a
different subset, or adopt none. A Rule absent from a language README is
unimplemented for that language, not a cross-language compatibility error.

## Current Go adoption

The Go implementation currently adopts two array operations for Settings:

- Rule absence means later-value `replace`.
- An exact Settings property path declared by the Cookbook-wide Merge Rules
  uses `union`.

Go stores those Rules at:

```text
cookbook/kitchen-notes/go.merge-rules.yaml
```

Rules are Cookbook-wide and are not selected per Recipe. Go Build and Apply
consult the same Rules through shared Cookbook Resolution. Freeze Commit may
add a Union path when reviewed Workbench selectors request it.

The Go README and implementation Contracts own the schema, validation,
ordering, Workbench selector, and error requirements. Keeping those details
with the adopting implementation prevents this Note from becoming a second Go
specification.

## Boundary

This Note does not require another implementation to support Go's file format,
operations, or Workbench representation. It also does not make Merge Rules a
replacement for Core resolution or a per-Recipe strategy.

Add another Merge Rule only when the target ecosystem exposes a concrete
combination behavior that generic Build cannot preserve.

## Related documents

- [Cookbook Kitchen Notes](../core/core.cookbook.kitchen-notes.md)
- [Go implementation declarations](../../go/README.md#applied-kitchen-notes)
- [Go Cookbook Resolution Contract](../../go/doc/contract/contract.cookbook-resolution.md)
- [CTK JSON Flat Format](note.ctk-json-flat-format.md)
