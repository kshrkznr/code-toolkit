# Knowledge.authors-recipe.md
============================================================

# Author's Recipe

Author's Recipe shows how one participant maps a development style onto CTK.

This Node preserves the author's mental model, design decisions, and current
Recipe examples. It is not a best practice or a recommended configuration.

CTK defines the responsibilities. Author's Recipe explains one choice of how to
use them.

It is similar to talking with one participant in the project about how they use
CTK and why they made particular choices. It is one perspective, not a lead or
authoritative interpretation.

Other useful Recipes should be consulted when they fit the current question.
Better interpretations and proposals are welcome, including choices that differ
from this Recipe while respecting CTK's responsibilities.

## Responsibility

Author's Recipe is responsible for making one concrete interpretation and its
trade-offs inspectable.

It may show how the author maps Layer responsibilities, composes Recipes, and
uses the resulting environments. It does not define Concept API
responsibilities, a canonical Cookbook layout, or a configuration that other
users should copy.

When an example and an accepted Concept differ, resolve the current boundary
through [Core](../core/README.md), [Workbench](../workbench/README.md), or the
[Documentation Resolver](../README.md).

## Navigate by question

Start with the question closest to the current work:

- How does the author interpret Runtime and Profile?
  → [My Mental Model](#my-mental-model)
- Which trade-offs shape the author's Recipe?
  → [Design Decisions](#design-decisions)
- How are the author's current Recipes composed and resolved?
  → [Inspect Views](#inspect-views)
- How should generated inventories be reviewed without treating them as
  canonical Cookbook structure?
  → [Workbench Review](../note/note.workbench-review.md)

The routes expose one participant's context. They are examples to inspect and
adapt, not a preferred reading of CTK.

## Relationship

```text
              CTK
               │
      Responsibilities
               │
      ┌────────┴────────┐
      │                 │
  My Recipe      Your Recipe
      │
 ┌────┴────┐
 │         │
VSCode   Kiro
```

Related Recipes may realize the same workflow on different Platforms.

VS Code, Kiro, or future Platforms may compose that workflow from different
Platform-specific capabilities.

---

## My Mental Model

When organizing my development environment, I use the following interpretation.

- Runtime = my default development experience
- Profile = purpose-specific differences

For me, Runtime represents the environment I expect every time I start
developing.

Theme, Terminal, Language Support, common Extensions, and base Settings all
belong to Runtime because they define my everyday experience rather than a
specific task.

Profiles are intentionally lightweight.

Most Profiles start empty and only gain Settings or Extensions when a task
truly requires different behavior.

In other words, I try to keep Runtime stable and let Profiles describe only the differences.

---

## Design Decisions

Some examples of decisions in my Recipe are:

- Runtime represents common development experience rather than programming languages.
- Profiles represent workflows rather than technologies.
- Review-related Profiles are shared across different Runtimes.
- Language Servers are isolated when it improves performance or maintainability.

These are not CTK requirements.

They are simply trade-offs that fit the way I work.

---

## Building My Recipe

My Recipe is built by composing Ingredients.

For example,

- Platform selects the IDE family.
- Runtime provides the default development experience.
- Profiles add task-specific differences.

The resulting Recipe can then be built into a concrete environment for a
specific Platform.

---

## Example Workflow

```text
Connect
    ↓
Review

Ops
    ↓
Inspect
```

The important part is not reproducing this exact Recipe.

The goal is to build your own Recipe by composing Ingredients in a way that
matches your own workflow.

---

## Inspect Views

The files under `inspect/` are generated Recipe Views of selected current
Recipes. Each directory keeps the Recipe composition, resolution summary, and
typed Artifact inventories together as one review context.

These Views are snapshots, not maintained copies of the Cookbook. They may be
regenerated or replaced as the author's Recipes evolve. Their generated paths,
groupings, and Artifact contents do not define a canonical Cookbook or
repository layout.

Start with each `summary.md` to understand provenance and resolved contents,
then open only the typed Artifact relevant to the current question.

| Recipe View | Current composition shown |
| --- | --- |
| [kiro-default](inspect/recipe.kiro-default/summary.md) | Baseline Kiro environment with shared purpose-specific Profiles |
| [vscode-default](inspect/recipe.vscode-default/summary.md) | Baseline VS Code environment with the ChatGPT Runtime responsibility |
| [vscode-development](inspect/recipe.vscode-development/summary.md) | VS Code with Vim, Remote Development, and Dev Container responsibilities |
| [vscode-golang](inspect/recipe.vscode-golang/summary.md) | VS Code with Go and ChatGPT Runtime responsibilities |

The useful result is not the exact list of Settings or Extensions. The Views
make the author's current responsibility mapping visible enough to discuss,
compare, and reconsider.
