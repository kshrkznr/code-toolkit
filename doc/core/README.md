# Knowledge.core.md
============================================================

# Concept Domain: Core

Core defines the stable responsibilities that form CTK's conceptual
foundation.

Recipes, platforms, commands, and repository layouts may evolve.

The responsibilities described by Core are intended to remain stable as those
representations change.

## Responsibility

CTK organizes its concepts around **responsibilities** rather than
implementations.

A responsibility defines **what a concept is responsible for**, independent of
how it is implemented.

This allows the same conceptual model to be applied across different platforms,
tools, and repository layouts.

Responsibilities provide the common vocabulary shared throughout CTK. Other
Knowledge documents build upon these concepts or explain how they are applied,
integrated, observed, and evolved.

Responsibilities answer questions such as:

- What is this concept responsible for?
- Where are its boundaries?
- How does it relate to neighboring concepts?

## Navigate by question

This README is the entry node for the Core Concept Domain. Start with the
question closest to the current work:

- How are reusable environments organized and composed?
  → [Cookbook](core.cookbook.md)
- How are Recipes or Archives transformed into usable environments?
  → [Build Lifecycle](core.build-lifecycle.md)
- How are environments observed, reviewed, and preserved?
  → [Persistence Lifecycle](core.persistence-lifecycle.md)

These routes are entry points, not an exhaustive Repository Map. Once inside a
Core document, use its local relationships, such as `See also`, when they help
the current question cross into a neighboring concept.

## Boundary

Core intentionally avoids describing:

- Implementation details
- Repository structure
- Platform-specific behavior
- Personal workflow preferences

Those concerns belong to Integration, language implementations, Author's
Recipe, Design Notes, or operational Notes as appropriate. Return to the
[Documentation Resolver](../README.md) when the current question no
longer belongs to Core.

## Shared documentation guidance

Core documents should gradually converge toward a common structure that makes
accepted responsibilities easy to find.

Typical sections:

- Why
- Definition
- Responsibility
- Analogy
- Notes
- See also
- Future (when applicable)

Not every section is required. In particular, `See also` is useful only when a
related document provides a natural next step.

Choose the structure that best explains the concept with minimal context.

---

### Section reference

The following descriptions are guidelines rather than strict rules.

#### Why

Why this responsibility exists.

#### Definition

Defines the concept itself.

#### Responsibility

Explains the responsibility and its boundaries.

#### Analogy

Provides a mental model that helps explain the concept.

#### Notes

Supplemental information, design considerations, and operational notes.

#### See also

Related concepts that are useful to read next.

#### Future

Potential future directions that naturally extend the responsibility.
