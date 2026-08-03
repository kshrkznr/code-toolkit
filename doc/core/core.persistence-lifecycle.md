# Knowledge.core.persistence-lifecycle.md
============================================================

# Persistence Lifecycle

## Definition

Persistence Lifecycle describes how knowledge evolves inside a Cookbook.

Rather than generating artifacts, it manages the lifecycle of reusable concepts.

The primary stages are:

Draft → Recipe → Archive

It does not build a Runtime.
It records, organizes, or packages an existing Runtime.

## Responsibility

|Lifecycle|Responsibility|
|---|---|
|Lock|Observe Runtime and generate a Manifest|
|Freeze|Convert a Manifest into Recipes|
|Archive|Convert a Manifest into a distributable bundle|

## Analogy

Persistence Lifecycle is conceptually similar to creating build artifacts.

Instead of preserving compiled binaries,
CTK preserves development environments.

## Conceptual diagram

```text
Current Runtime
        │
        ▼
      Lock
        │
     Manifest
   ┌────┴────┐
   ▼         ▼
Freeze    Archive
```

Another view:

```text
Current Runtime
        │
        ▼
      Lock
        │
   Snapshot / Manifest
   ┌────────┴─────────┐
   ▼                  ▼
Recipe             Offline Package
```

---

## See also

- [Core Concept Domain](README.md)
- [Cookbook](core.cookbook.md) — owns the reusable concepts updated by Freeze
- [Build Lifecycle](core.build-lifecycle.md) — constructs the Runtime later
  observed and preserved by this lifecycle

---

# Archive

## Definition

Archive preserves a Runtime as a self-contained distribution package.

It stores both the manifest and the required assets so that the Runtime can be reconstructed without accessing external repositories.

---

## Responsibility

Archive is responsible for packaging a Runtime for distribution.

It does not describe the desired configuration (Recipe),
nor does it observe the current Runtime (Lock).

Its responsibility is to preserve everything required for reconstruction.

---

## Analogy

Archive is conceptually closer to a fat JAR than a backup.

Both package everything required to run in another environment.

Unlike a fat JAR, Archive packages a development environment rather than an application.

---

## Notes

Archive stores both manifests and assets.

Typical assets include:

- VSIX packages
- Runtime metadata
- Other resources required for offline reconstruction

---

# Freeze
## Definition

Freeze is a review workflow for updating Recipes.

It compares the current Runtime with the resolved Recipe (`recipe_ref`) and generates a Draft for review.

Freeze never modifies Recipes directly.
The edited Draft becomes the source of truth for `freeze commit`.

---

## Responsibility

Freeze is responsible for organizing Runtime changes into Recipes.

It does not observe the Runtime (Lock),
nor does it package a Runtime for distribution (Archive).

Its responsibility is to generate and apply reviewable Recipe changes.

---

## Analogy

Freeze is inspired by the Git review workflow.

Rather than synchronizing files, Freeze generates a reviewable patch for Recipes.

```text
git add      → ctk freeze
diff         → draft
commit       → freeze commit
push         → build
CI           → Runtime
Artifact     → codevenv use
```

---

### Conceptual diagram

```text
Freeze
    │
    ▼
recipe_ref
    │
    ▼
Draft
    │
    ▼
Review
    │
    ▼
Commit
    │
    ▼
Recipe
```
