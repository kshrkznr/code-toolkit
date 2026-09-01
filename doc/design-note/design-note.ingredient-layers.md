# Knowledge.design-note.ingredient-layers.md
============================================================

# Why Ingredient Layers Are Vocabulary, Not Hierarchy

## Background

CTK repeatedly encountered opportunities to introduce another grouping level,
including Runtime Categories, Groups, Environment Types, and Feature Sets.
Each proposal looked useful in isolation, but most added structure without
naming a genuinely different kind of responsibility.

That distinction clarified what an Ingredient Layer is for.

## Decision

An Ingredient Layer introduces vocabulary for a distinct conceptual dimension.
It does not exist merely to create a deeper hierarchy or organize more files.

```text
OS
Platform
Runtime
Profile
Extension
```

Each name helps a reader interpret an Ingredient's role before reading its
contents. The Layers are not simply directory levels.

## A generic Capability Layer was considered

Almost every Ingredient could technically be represented through one generic
Capability Layer:

```text
Capability
├── Windows
├── VSCode
├── Java
├── Git
└── Markdown
```

This would be flexible, but the role of an Ingredient would remain ambiguous
until its contents were inspected. Named Layers improve communication when
they distinguish kinds of responsibility that recur across Recipes.

The current Core therefore does not include Capability as an accepted Layer.
File movement, renaming, and Runtime/Profile composition remain sufficient for
most current Recipes.

One observed reuse problem did require a narrower answer: Runtime and Profile
Ingredients needed to share named lists of concrete editor Extensions. The Go
implementation adopted [Extension Set composition](../note/note.extension-set.md)
as an optional Kitchen Note. Its direction is fixed to Runtime/Profile → Set →
concrete Extension, so it solves that reuse without creating a generic Layer
or a second Recipe hierarchy.

There is no standing Capability Future after that implementation. Open a new
Future candidate only when a concrete requirement must compose non-Extension
Ingredients and can state a clear resolution and review model.

## Layers express intent that files cannot

Configuration formats often use comments as their only vocabulary for intent:

```jsonc
// Git
"git.autofetch": true,

// Java
"java.configuration.updateBuildConfiguration": "interactive"
```

The comments are not configuration. They name concepts that the format cannot
otherwise express.

CTK can give those concepts durable names through Ingredients and Layers. A
name such as `runtime.git.settings.json` is more than file organization: it
makes *Git* an explicit reusable responsibility inside the Cookbook.

Named concepts can be composed, searched, inspected, and varied. The useful
intent is preserved even though its representation changes from comments to
Cookbook vocabulary.

## Why Layers remain optional

A Cookbook may begin with one generated Ingredient:

```text
runtime.draft.settings.json
```

There is no need to invent finer Layers or Ingredients before recurring
responsibilities become visible. The goal is to express meaning, not to
maximize structure.

The same reasoning applies to Variants. A Variant does not add hierarchy; it
provides a context-specific representation of the same named responsibility.

## Layers reduce ambiguity

Without a meaningful Layer, a name such as `Java` leaves several readings
open:

- an operating system
- a language runtime
- an editor capability
- a purpose-specific environment

Placing the Ingredient under Runtime or another accepted Layer narrows that
reading before the file is opened.

Exact labels remain less important than the distinction they communicate.
Names support understanding; they do not create the responsibility by
themselves.

## Vocabulary also creates design pull

Layer names are not neutral. Treating an IDE family as Platform naturally
makes a default Runtime and named Profiles easier to imagine. That pull can
help readers form a model quickly, but it can also make one Recipe
interpretation appear mandatory.

Core therefore defines relative Layer boundaries while leaving concrete
contents to each Recipe. Design review should continue to ask whether a name
clarifies responsibility or unnecessarily narrows otherwise valid Recipes.

## Boundary

A new Layer should be introduced only when CTK needs vocabulary for a genuinely
different conceptual dimension. Better file organization alone is not enough.

Accepted Layer responsibilities belong to the Cookbook Concept API. This
Design Note preserves the rationale behind keeping the vocabulary small.

## Related documents

- [Cookbook Concept API](../core/core.cookbook.md)
- [Why CTK Uses Cookbook Vocabulary](design-note.cookbook.md)
- [Variant](../note/note.variant.md)
