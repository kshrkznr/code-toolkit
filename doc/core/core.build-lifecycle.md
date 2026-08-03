# Knowledge.core.build-lifecycle.md
============================================================

# Build Lifecycle

## Definition

The Build Lifecycle defines how a Runtime is constructed from a Recipe.

It resolves the Recipe and generates a reproducible Runtime that can be used independently of the source Recipe.

---

## Responsibility

|Lifecycle|Responsibility|
|---|---|
|Build|Resolve a Recipe or Archive into a Runtime|
|Apply|Update an existing Runtime from a Recipe or Archive|
|CodeVenv|Select which Runtime is active|

---

## Analogy

The Build Lifecycle is conceptually similar to a software build pipeline.

Instead of producing application binaries,
CTK produces reusable IDE Runtime environments.

---

## Conceptual diagram

```text
Recipe / Archive
        │
        ▼
     Resolve
        │
        ▼
      Build
        │
        ▼
     Runtime
        │
   ┌────┴────┐
   ▼         ▼
 Apply   CodeVenv
```

Another view:

```text
Recipe
    │
    ▼
  Build
    │
    ▼
 Runtime
    │
    ▼
CodeVenv
    │
    ▼
 IDE
```

---

## Notes

The Build Lifecycle is responsible for creating and updating Runtime environments.

Persistence is handled by a separate lifecycle.

```text
Build Lifecycle

Recipe
    │
    ▼
 Runtime

Persistence Lifecycle

Runtime
    │
    ▼
 Lock
    │
 ┌──┴──┐
 ▼     ▼
Freeze Archive
```

The two lifecycles are complementary:

- Build creates Runtime environments.
- Persistence preserves Runtime environments.

After Build or Apply completes, CTK automatically runs Lock against the resulting Runtime. This records the observed state without making Lock part of Build or Apply's responsibility.

---

## See also

- [Core Concept Domain](README.md)
- [Cookbook](core.cookbook.md) — defines the Recipe resolved by Build
- [Persistence Lifecycle](core.persistence-lifecycle.md) — observes and
  preserves the resulting Runtime
- [Extension Resolution](../note/note.extension-resolve.md) — describes how
  extension artifacts are prepared when a Recipe-based Runtime is constructed
  or recovered

---

# Build

## Definition

Build resolves a Recipe or Archive into a Runtime.

The generated Runtime is self-contained and can be used independently of the source definition.

---

## Responsibility

Build is responsible for constructing Runtime environments.

It resolves declarative definitions into executable Runtime layouts.

It does not activate Runtimes (CodeVenv),
nor preserve them (Persistence Lifecycle).

---

## Analogy

Build is conceptually similar to a software build system.

Instead of producing application binaries,
CTK produces IDE Runtime environments.

---

## Notes

Build accepts multiple sources.

Typical inputs include:

- Recipe
- Archive

Both are resolved into the same Runtime structure.

---

## Conceptual diagram

```text
Recipe / Archive
        │
        ▼
      Resolve
        │
        ▼
       Build
        │
        ▼
      Runtime
```

---

# Apply

## Definition

Apply updates an existing Runtime from a Recipe or Archive.

Unlike Build, Apply preserves the existing Runtime and only synchronizes its contents.

---

## Responsibility

Apply is responsible for synchronizing Runtime contents.

It updates an existing Runtime without recreating it.

It does not resolve Runtime selection (CodeVenv),
nor preserve Runtime state (Persistence Lifecycle).

---

## Analogy

Apply is conceptually similar to incremental synchronization.

Instead of rebuilding a Runtime,
it updates an existing Runtime in place.

---

## Notes

Apply accepts the same inputs as Build.

Typical inputs include:

- Recipe
- Archive

Build and Apply share the same resolution process.

The difference is the target.

- Build creates a Runtime.
- Apply updates an existing Runtime.

---

## Conceptual diagram

```text
Recipe / Archive
        │
        ▼
      Resolve
        │
        ▼
      Apply
        │
        ▼
 Existing Runtime
```
