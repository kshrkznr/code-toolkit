# project-knowledge.design-note.concept-api.md
============================================================

# Design Idea: Concept API

## Background

As CTK evolved, discussions gradually shifted away from individual commands toward higher-level concepts.

Recent design discussions naturally focused on concepts such as:

- Recipe
- Ingredient
- Build
- Freeze
- Inspect
- Variant

rather than their CLI commands.

This suggests that CTK is increasingly being understood through its conceptual model rather than its implementation.

---

## Concept API

One way to view CTK is as a collection of public concepts.

These concepts form the stable interface exposed to users, AI agents, and documentation.

Examples include:

```text
Recipe
Ingredient
Build
Freeze
Inspect
Variant
Capability
```

The implementation (CLI, directory layout, internal scripts, etc.) exists to realize these concepts.

In this view, CTK exposes a **Concept API** rather than merely a CLI.

---

## Relationship with README

README may act as a **Concept API Catalog**.

Its purpose is not to explain implementation details.

Instead, it introduces the concepts available in CTK and directs readers toward the appropriate Knowledge documents.

Conceptually:

```text
README
    │
    ▼
Concept API Catalog
    │
    ▼
Knowledge
    │
    ▼
Implementation
```

---

## Relationship with Knowledge

Knowledge documents become specifications of Concept APIs.

For example:

```text
Concept API
│
├── Cookbook
├── Build Lifecycle
└── Persistence Lifecycle
```

Each Concept Group introduces related concepts and their responsibilities.

This naturally aligns with the existing Core Knowledge structure.

---

## Relationship with Documentation Resolver

Documentation Resolver operates on Concept APIs.

Rather than searching for commands or source files, it identifies the relevant concept and routes readers to the corresponding Knowledge.

Example:

```text
"I want to update Ingredients."

        │

        ▼

Persistence Lifecycle

        │

        ▼

Freeze Draft
Freeze Commit
```

---

## Relationship with CLI

CLI is one implementation interface for the Concept API.

```text
Concept API
        │
        ├── CLI
        ├── Directory Structure
        ├── Build Pipeline
        └── Future Interfaces
```

The conceptual model remains stable even if implementations evolve.

---

## Benefits

Thinking in terms of Concept APIs helps separate stable concepts from implementation details.

This provides:

- a clearer role for README
- a natural entry point for AI agents
- better alignment with Knowledge
- a stable public model independent of CLI commands
- easier design discussions centered on responsibilities rather than implementations

The idea does not introduce new concepts.

It gives a name to an interface that has gradually emerged through the project's evolution.
