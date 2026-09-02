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
Extensions. The group may also need shared Settings, Snippets, Keybindings,
Tasks, or MCP content to work as one capability. Copying those Resources into
every referring Ingredient obscures the reusable responsibility and makes its
configuration easy to evolve inconsistently.

An Extension Set gives that responsibility a Cookbook-local name. An adopting
resolver may let Runtime and Profile Extension Resources include the named Set
in addition to declaring concrete Extension IDs directly. The Set expands to
concrete IDs and may contribute companion Runtime Artifacts.

```text
Runtime or Profile
        ↓ includes
Extension Set
        ├── contains concrete editor Extensions
        └── contributes companion Runtime Artifacts
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

A referenced Set may also contribute Settings, Keybindings, Snippets, Tasks,
and MCP Resources. These companion Resources use the existing compatible
Ingredient layouts, Artifact composition semantics, ownership strategies, and
Base → OS → Platform Variant order. Set membership itself remains
non-variant. Within one effective scope, repeated Set and concrete Extension
references contribute their Resources only at the first applicable position.

The [Go Cookbook Contract](../../go/doc/contract/contract.cookbook.md) owns the
exact name grammar, compatible layouts, error behavior, source provenance, and
downgrade boundary. The [Go implementation
README](../../go/README.md#applied-kitchen-notes) is the adoption declaration.

## Review and reverse composition

Extension Set identity is Cookbook Source provenance, not Runtime state.
Recipe review can show the declaring Runtime or Profile Resource, the Set
Resources it selected, and which direct declaration or membership Resource
contributed a concrete Extension. Distribution, Lock, Archive, Apply, and
Recovery continue to operate on concrete resolved or observed Artifacts.

Freeze can therefore show used Set Resources as context but cannot infer
whether a Runtime difference should update that Set, another Set, or a direct
Runtime/Profile Resource. That choice remains explicit user or AI review.
Freeze Commit must not create or silently rewrite Set membership or companion
Resources.

## Compatibility boundary

A Cookbook Source using an optional Kitchen Note is portable only to
implementations that declare adoption. Go v0.6.2 reserves the `set:` form and
fails safely before mutation; Go v0.7.0 adds the adopted interpretation. Go
v0.6.1 and earlier are unsupported for Sources using Extension Sets.

Go v0.7.0 resolves only Set membership. A v0.7.x Cookbook that adds companion
Set Resources remains partially usable there: concrete Extensions resolve, but
the unrecognized companion Resources are ignored. The v0.7.x compatibility
policy accepts that partial downgrade instead of adding a new versioned Set
declaration.

The retained Bash implementation does not adopt this Kitchen Note and makes no
Extension Set behavior guarantee. Optional non-adoption does not require Bash
to emulate Go's reserved-prefix guard.

## Artifact and Layer boundary

The Set still represents an Extension group. A Resource specific to one member
belongs to `extension.<id>`, while purpose-wide content belongs to the owning
Runtime or Profile Ingredient. Set companion Resources own only configuration
that makes the Extension group work together.

The selection direction remains Runtime/Profile → Set → concrete Extension.
A Set does not select another Set, Runtime, or Profile, and a Recipe does not
select Sets directly. Companion Artifacts therefore do not imply a generic
Capability Layer or a second Recipe composition hierarchy.

## Related documents

- [Cookbook Kitchen Notes](../core/core.cookbook.kitchen-notes.md)
- [Why Ingredient Layers Are Vocabulary, Not Hierarchy](../design-note/design-note.ingredient-layers.md)
- [Go Cookbook Contract](../../go/doc/contract/contract.cookbook.md)
- [Go Workbench Contract](../../go/doc/contract/contract.workbench.md)
