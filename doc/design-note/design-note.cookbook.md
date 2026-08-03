# Knowledge.design-note.cookbook.md
============================================================

# Why CTK Uses Cookbook Vocabulary

## Background

CTK first adopted *Recipe* as a familiar way to describe reusable environment
composition. As more reusable Settings, Extensions, Snippets, and Tasks
appeared, the model needed a name for the material selected by a Recipe and a
name for the collection that held both.

The vocabulary had to clarify responsibility before a reader inspected a file
or implementation detail.

## Alternatives considered

### Module

*Module* had begun to describe two different ideas: automatically resolved
Extension details and reusable capabilities selected explicitly by a Recipe.
The name mixed resolution behavior with conceptual responsibility.

### Component

*Component* accurately suggested a reusable part, but it was too general to
explain the part's relationship to a Recipe. It also overlapped with *Module*
and made the model switch from a cooking metaphor to generic system
terminology:

```text
Recipe
  selects
Component
```

The missing concept was not simply a reusable component. It was material made
available to, and selected by, a Recipe.

## Decision

CTK adopted *Ingredient* for a reusable responsibility selected by a Recipe,
and *Cookbook* for the conceptual container that holds Ingredients, Recipes,
and their supporting lifecycle.

```text
Cookbook
├── Ingredients
└── Recipes
      └── select Ingredients
```

The relationship is recognizable without requiring *Ingredient* to behave
like an object from a particular software framework.

Looking back at Chef confirmed that Recipe, Ingredient, and Cookbook already
formed familiar vocabulary. CTK uses that vocabulary as an homage and adapts
its responsibilities to development-environment composition; it does not
reproduce Chef's model.

## Why Platform, Runtime, and Profile appear as Layers

Once Ingredient became the reusable unit, CTK needed vocabulary for distinct
kinds of Ingredient responsibility.

### Platform

CTK needed to treat VS Code-family applications such as VS Code, Kiro, and
Cursor through a common responsibility. *Platform* names the development
application that supplies the capabilities used by an environment.

Platform does not replace the operating system and does not mean every kind of
software platform. It is CTK's name for this specific Ingredient
responsibility.

### Runtime

Runtime names Ingredients that contribute to a Recipe's default experience on
the selected Platform. Theme, Terminal behavior, language support, common
Extensions, and baseline Settings can participate in that experience.

The name borrows the idea of a common operating foundation. It does not mean a
language runtime or virtual machine, and CTK does not require every Recipe to
interpret Runtime contents in the same way.

### Profile

Profile names Ingredients that contribute purpose-specific differences on top
of a Recipe's default Runtime. A Profile may contain only Extensions, only
Settings, or another small difference; its responsibility comes from the named
composition it supports rather than its size.

This produces a useful reading aid:

```text
OS
  ↓
Platform
  ↓
Runtime
  ↓
Profile
```

The diagram is vocabulary for distinguishing responsibilities. It is not a
required application architecture or a directory hierarchy.

## A Recipe maps one development style onto the vocabulary

A Recipe does not explain the Layer model. It records one composition made
with that model.

For example, the author's current interpretation maps IDE families to
Platform, a shared everyday editor experience to Runtime, and purposes such as
development or inspection to Profile. Another Recipe may interpret Runtime and
Profile differently while preserving their accepted boundaries.

That distinction is why Author's Recipe exists. It shows one concrete mapping
and its trade-offs without presenting that mapping as a CTK best practice.

## Ingredient selection and assembly strategy remain separate

The top level of a Recipe declares which Ingredients participate in an
environment. Its optional `config` section declares how the selected
Ingredients are assembled into a Distribution.

```text
Recipe top level  → Ingredient selection
Recipe config     → Distribution assembly strategy
```

Keeping these responsibilities separate allows the Recipe to remain readable
as a composition while still expressing choices such as Lock handling or
Profile Artifact ownership. Strategy is not another Ingredient Layer.

## Named Ingredients preserve intent

Traditional configuration files often rely on comments to explain why values
belong together:

```jsonc
// Git
"git.autofetch": true
```

CTK can promote that intent into a named Ingredient. The representation changes
from an informal comment to vocabulary that can be reused, composed, searched,
inspected, and varied.

This does not require a Cookbook to introduce many Ingredients immediately. A
single generated Draft Ingredient is a valid starting point. New names become
useful when repeated work reveals a responsibility worth preserving.

## Boundary

This Design Note explains how the vocabulary emerged. Current Cookbook,
Ingredient, Recipe, Variant, and Layer responsibilities are defined by the Core
Cookbook Concept API.

It does not prescribe:

- one directory layout for a Cookbook
- one interpretation of Runtime or Profile contents
- immediate decomposition into many Ingredients
- adoption of the author's personal Recipe

## Related documents

- [Cookbook Concept API](../core/core.cookbook.md)
- [Why Ingredient Layers Are Vocabulary, Not Hierarchy](design-note.ingredient-layers.md)
- [Author's Recipe](../authors-recipe/README.md)
- [Recipe Build Strategy](../note/note.recipe-build-strategy.md)
