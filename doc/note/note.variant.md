# Knowledge.note.variant.md
============================================================

# Settings Variant Usage

Platform selection belongs to the Recipe. Ingredient responsibilities should
remain Platform-independent.

## Purpose

This Note records the current operational scope of Settings Variants and the
guidance that emerged from using them.

The Cookbook Concept API defines Variant responsibility. This Note does not
extend that responsibility to new Resources or dimensions.

## Background

While CTK Runtimes were being organized, some Settings needed to differ by
operating system:

```text
runtime/git/settings.json
runtime/git/macos.settings.json
runtime/git/windows.settings.json
```

Moving those Settings into the OS Layer would have moved responsibility merely
because one representation differed. Variant emerged as a way to preserve the
Runtime responsibility while expressing only the environmental difference.

## The Base remains the subject

A Variant is not a separate responsibility. It is a context-specific
representation of the same Resource owned by the Base Ingredient.

For example, `macos.settings.json` beside `settings.json` means *the macOS
Variant of these Settings*, not *Settings owned by macOS*.

```text
Base Ingredient responsibility
├── Base Settings
├── macOS Settings Variant
└── Windows Settings Variant
```

The Base Ingredient remains the subject in every case.

## Current scope and precedence

The current implementation applies Variants only to Settings. It resolves OS
and Platform independently in this order:

```text
Base → OS Variant → Platform Variant
```

A later Platform Variant may replace a value supplied by the Base or OS
Variant.

```text
runtime/common/settings.json
runtime/common/macos.settings.json
runtime/common/code.settings.json
```

Combined OS-and-Platform filenames such as `macos.code.settings.json` are not
part of the current form.

Extensions, Ingredients, and Recipes express composition rather than one
Resource representation. Applying the same Variant mechanism to them would
blur responsibility, so they remain outside the current scope.

## Introduce a Variant after observing a difference

Multiple Platforms alone do not justify multiple representations. Begin with
the shared responsibility whenever possible:

```text
runtime.common
```

Introduce a Variant when actual use shows that one environment needs different
behavior while the responsibility remains the same:

```text
runtime.common
        ├── runtime.common.macos
        └── runtime.common.windows
```

For example, Vim key handling was initially shared. On Windows, `<Ctrl-K>`
conflicted with common editor shortcuts; macOS used `<Cmd-K>` and did not need
the same customization. The Runtime responsibility stayed in place, and only
the Windows Settings Variant was introduced.

This suggests the following loose guidance:

- Start with the shared responsibility.
- Add a Variant only after observing a concrete environmental difference.
- Keep the Base Ingredient's responsibility unchanged.
- Use a Variant to express a different representation, not to relocate
  ownership.

## Boundary and future questions

Other dimensions, such as a Java version or tool choice, may eventually expose
the same need. CTK should not generalize Variant dimensions before repeated
usage establishes both the difference and the responsibility that remains
stable across it.

When that evidence appears, reconsider the Variant boundary in Core and the
relevant implementation documents. Until then, Settings with independent OS
and Platform dimensions remain the documented operational scope.

## Related documents

- [Cookbook Concept API](../core/core.cookbook.md)
- [Why Ingredient Layers Are Vocabulary, Not
  Hierarchy](../design-note/design-note.ingredient-layers.md)
