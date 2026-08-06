# Knowledge.core.cookbook.md
============================================================

# Concept API: Cookbook

## Concept Domain

Core

Cookbook is the conceptual container for reusable development environments.

It organizes reusable building blocks independently from platforms, implementations, or repository layouts.

A Cookbook consists of three primary concepts:

- Ingredients
- Recipes
- Drafts

These concepts describe how reusable environments are composed, refined, and eventually built into concrete artifacts.

---

## Why

CTK focuses on reusable responsibilities rather than reusable files.

A Cookbook provides a conceptual model for organizing those responsibilities without depending on a particular implementation.

This allows Recipes and Ingredients to evolve while preserving a stable mental model.

---

## Definition

A Cookbook is a collection of reusable Ingredients, Recipes, and their supporting lifecycle.

It represents the conceptual source of truth before concrete artifacts are generated.

---

## Responsibility

A Cookbook is responsible for organizing reusable concepts.

It does not define how artifacts are generated.

It does not describe repository layouts.

It provides the conceptual structure that Build and Persistence lifecycles operate on.

---
## Notes

Ingredients are designed to be composed like LEGO blocks.

Each Ingredient has a single responsibility.

Recipes compose multiple Ingredients to create complete development environments.

## Conceptual diagram

```text
             Cookbook

      ┌────────┴────────┐
      │                 │
Ingredients         Recipes
      │                 │
      └────────┬────────┘
               │
            Build
```

---

## See also

- [Core Concept Domain](README.md)
- [Build Lifecycle](core.build-lifecycle.md)
- [Persistence Lifecycle](core.persistence-lifecycle.md)
- [Cookbook Kitchen Notes](core.cookbook.kitchen-notes.md)

---

# Ingredients

## Why

Recipes should compose reusable responsibilities rather than duplicate configuration.

Ingredients provide those reusable building blocks.

---

## Definition

An Ingredient is a reusable unit that represents a single responsibility.

Examples include Platform, Runtime, Profile, and other reusable concepts.

An Ingredient may provide one or more reusable resources, such as Settings, Extensions, Snippets, or Tasks.

---

## Responsibility

Ingredients provide reusable building blocks for Recipes.

They intentionally avoid describing complete environments by themselves.

A complete environment emerges from combining multiple Ingredients inside a Recipe.

---

## Notes

Ingredients are organized by Layer and Name.

Not every Ingredient provides every resource.

Each Ingredient exposes only the resources required by its own responsibility.

## Conceptual diagram

```text
Ingredients
└── <layer>
    └── <ingredient>
        ├── settings
        ├── extensions
        ├── snippets
        ├── tasks
        └── ...
```

---

# Variant

## Why

An Ingredient responsibility may remain stable while a Resource needs a
different representation for a particular environment.

A Variant expresses that difference without creating a second responsibility.

## Definition

A Variant is a context-specific representation of an Ingredient Resource.

The Variant retains the identity and responsibility of its Base Ingredient.

## Responsibility

Variants express differences within an existing Ingredient responsibility.

They do not introduce a new Ingredient responsibility.

They do not relocate a responsibility to another Layer or change which
Ingredients a Recipe selects.

## Notes

Supported Variant dimensions, Resource types, and resolution order are
implementation interpretation. They are documented by the corresponding
Kitchen Notes, Contracts, and operational Notes.

---

# Recipes

## Why

Development environments are built by combining reusable responsibilities rather than defining everything from scratch.

Recipes capture those compositions in a reusable form.

## Definition

A Recipe defines how multiple Ingredients are combined to create a development environment.

Recipes describe composition rather than implementation.

## Responsibility

Recipes select Ingredients.

They do not define the internal responsibilities of those Ingredients.

They provide the composition that Build processes later transform into concrete artifacts.

## Analogy

A Recipe describes how Ingredients are combined.

Similar ideas can be found in technologies such as Chef Cookbooks or Dockerfiles, where reusable components are composed into a complete environment.

CTK uses the same mental model while remaining independent of any specific implementation.

Recipe changes may be prepared through the [Draft
Workbench](../workbench/README.md#draft). Workbench owns the review context;
Cookbook owns the durable Recipe and Ingredient resources accepted through the
Persistence Lifecycle.

====================================================================================================

# Ingredient Layers

## Definition

Ingredient Layers organize Ingredients by their primary responsibility.

Layers provide a consistent way to classify reusable Ingredients.

They are independent of repository layouts and implementation details.

---

## Responsibility

Layers classify Ingredients.

They do not prescribe architecture.

They do not define how Recipes should compose Ingredients.

Recipes remain free to interpret and combine Ingredients according to their own workflow.

---

## Conceptual diagram

```text
Ingredients
├── OS
├── Platform
├── Runtime
├── Profile
└── ...
```

---

## Notes

Layers exist to organize responsibilities rather than implementations.

Different Recipes may interpret the same Layers differently while still sharing the same conceptual model.


### Layer Design

CTK defines layer responsibilities.

It does not define how a Recipe should partition its architecture.

For example:

Recipe A

Runtime = Languages
Profile = Personas

Recipe B

Runtime = Editor Features
Profile = Languages

Both are valid CTK Recipes.

---

## OS

### Definition

OS represents the operating system on which a development environment is built.

Typical examples include Windows and macOS.

### Responsibility

OS provides the host environment for all higher Ingredient Layers.

It is responsible for operating system specific behavior.

It does not define development tools, editor configuration, or workflows.

---

# Platform

## Why

Reusable environment responsibilities should not need to be rebuilt merely
because the application hosting them changes.

Platform keeps application-specific capabilities separate so Runtime and
Profile responsibilities can be composed for another compatible application.

## Definition

Platform represents the development application that hosts a development environment.

Typical examples include VSCode, Kiro, and future compatible platforms.

## Responsibility

Platform provides the capabilities available to a development environment.

It defines the execution environment for Runtime and Profile Ingredients.

It does not define the user's workflow or personal preferences.

---

## Runtime

### Why

Named Profiles should be able to share a baseline without repeating each
common responsibility.

Runtime makes that baseline reusable across named compositions.

### Definition

Runtime identifies Ingredients that participate in the default Runtime composed
for a Recipe.

The Recipe decides what those Ingredients mean for its own workflow.

### Responsibility

Runtime contributes the baseline from which the Recipe's named Profiles are
constructed.

It does not prescribe whether that baseline is organized by language, editor
feature, shared experience, or another interpretation.

---

## Profile

### Why

A named composition should express only the differences required beyond the
Recipe's default Runtime.

Profile keeps those differences explicit without redefining the shared
baseline.

### Definition

Profile identifies a named composition constructed from a Recipe's default
Runtime and Profile Ingredients.

Each Recipe decides what its Profile names and contents represent.

### Responsibility

Profile contributes the differences required by one named composition without
redefining the Recipe's default Runtime.

It does not prescribe a particular workflow taxonomy or require Profiles to be
large or small.

---

## Extension

### Definition

Extension represents a reusable configuration unit for a single editor extension.

Each Extension Ingredient encapsulates the resources required to use that extension within a development environment.

Typical resources include extension-specific settings and related configuration.

### Responsibility

Extension owns configuration specific to a single editor extension.

It provides reusable extension configuration that may be resolved by other Ingredients.

It does not define development workflows or the overall development experience.

Those responsibilities belong to Runtime and Recipes.

### Notes

Extension Ingredients are resolved through other Ingredients rather than being referenced directly by Recipes.

Each Extension Ingredient should focus on a single editor extension.

### Conceptual diagram

```text
Ingredients
└── extension
    ├── gitlens
    ├── vim
    └── errorlens
```
