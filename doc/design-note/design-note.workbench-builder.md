# Knowledge.design-note.workbench-builder.md
============================================================

# Workbench Builder as a Historical Lens

## Background

CTK was once interpreted as a **Workbench Builder for the VS Code ecosystem**.

VS Code-family applications already provide Settings, Extensions, Profiles,
Workspaces, CLI integration, and other environment capabilities. From this
view, CTK does not replace those capabilities. It adds an explicit way to
compose, review, reproduce, and evolve an environment built from them.

## The lens

```text
VS Code-family Platform capabilities
        ↓ composed through
Recipe and Cookbook lifecycle
        ↓ produces
Purpose-specific development environment
```

This interpretation helped explain why CTK could remain relatively thin. The
Platform owns the application capabilities; CTK owns the composition and
lifecycle expressed through its Concepts.

## Why it was not adopted as public positioning

The lens used *Workbench* to mean the complete purpose-specific development
environment.

CTK later gave Workbench a narrower accepted responsibility:

- Draft is the review area for a known Runtime change.
- Inspect is the disposable review area for inventories and comparisons.

Using Workbench for both the complete environment and its temporary review
areas would make one term identify different objects and lifecycles.

CTK therefore continues to describe its public purpose through development
environments, Cookbooks, Recipes, and Ingredients. **Workbench Builder** remains
a historical interpretation rather than a product category or Concept API.

## Durable observation

The useful conclusion does not depend on the retired label:

> CTK builds on the capabilities of an existing Platform. It adds composition
> and lifecycle rather than replacing the Platform itself.

This is a design lens, not an additional Core responsibility. Current Platform
and Workbench responsibilities remain defined by their respective Knowledge
documents.

## Related documents

- `../core/core.cookbook.md` — Platform, Recipe, and Ingredient responsibilities
- `../workbench/README.md` — current Draft and Inspect Workbench meaning
- `../integration/integration.code-venv.md` — integration with Platform commands
