# Knowledge.note.extension-set.md
============================================================

# Extension Set as a Kitchen Note

## Purpose

This Note describes Extension Set composition as an optional Cookbook Kitchen
Note. It explains the reusable idea and its responsibility boundary; it does
not add Extension Set to Cookbook Core or require another implementation to
adopt it.

Each language README owns adoption. The adopting implementation's Contract
owns its exact Source representation, validation, resolution, review, and
compatibility behavior.

## Observed need

Runtime and Profile Ingredients sometimes need the same group of editor
Extensions. Copying those IDs into every `.extensions` Resource obscures that
the lists express one reusable responsibility and makes membership changes
easy to apply inconsistently.

An Extension Set gives that list a Cookbook-local name. An adopting resolver
may then let Runtime and Profile Extension Resources include the named list in
addition to declaring concrete Extension IDs directly.

```text
Runtime or Profile
        ↓ includes
Extension Set
        ↓ contains
concrete editor Extensions
```

The Set is deliberately one level above concrete Extensions. It does not
select Runtime or Profile Ingredients, and one Set does not include another.
This fixed direction prevents cycles and keeps the abstraction from growing
into a second Recipe composition system.

## Empty membership is meaningful

An Extension Set may contain zero or more concrete Extension IDs. A Set with
no `.extensions` Resource and a Set with a present empty Resource both
contribute an empty list.

The reference establishes the Set's identity; a separate existence manifest
is unnecessary for this narrow composition role. An absent member list is
therefore not the same as an unknown declaration syntax or malformed Source.

## Current Go adoption

The Go implementation adopts the exact `set:<name>` declaration form in
Runtime and Profile `.extensions` Resources. It resolves Set members through
the existing concrete Extension path, including Extension-owned Settings and
other Resources. Direct and Set-derived IDs are deduplicated into the same
deterministic concrete result.

The [Go Cookbook Contract](../../go/doc/contract/contract.cookbook.md) owns the
exact name grammar, compatible layouts, error behavior, source provenance, and
downgrade boundary. The [Go implementation
README](../../go/README.md#applied-kitchen-notes) is the adoption declaration.

## Review and reverse composition

Extension Set identity is Cookbook Source provenance, not Runtime state.
Recipe review can show which direct declaration or Set Resource contributed a
concrete Extension. Distribution, Lock, Archive, Apply, and Recovery continue
to operate on concrete observed Extensions.

Freeze can therefore show a used Set as context but cannot infer whether a
Runtime difference should update that Set, another Set, or a direct
declaration. That choice remains explicit user or AI review. Freeze Commit
must not create or silently rewrite Set membership.

## Compatibility boundary

A Cookbook Source using an optional Kitchen Note is portable only to
implementations that declare adoption. Go v0.6.2 reserves the `set:` form and
fails safely before mutation; Go v0.7.0 adds the adopted interpretation. Go
v0.6.1 and earlier are unsupported for Sources using Extension Sets.

The retained Bash implementation does not adopt this Kitchen Note and makes no
Extension Set behavior guarantee. Optional non-adoption does not require Bash
to emulate Go's reserved-prefix guard.

## Boundary and later Resources

The initial Extension Set owns only its concrete `.extensions` membership.
Membership does not use Variants. This keeps the implemented behavior aligned
with the observed reuse problem rather than speculating about OS or Platform
differences that have not appeared.

If a concrete need later gives a Set another Resource such as Settings, that
Resource may use the existing Variant model when its behavior is specified.
That possibility does not broaden the current Kitchen Note and does not imply
a generic Capability Layer.

## Related documents

- [Cookbook Kitchen Notes](../core/core.cookbook.kitchen-notes.md)
- [Why Ingredient Layers Are Vocabulary, Not Hierarchy](../design-note/design-note.ingredient-layers.md)
- [Go Cookbook Contract](../../go/doc/contract/contract.cookbook.md)
- [Go Workbench Contract](../../go/doc/contract/contract.workbench.md)
