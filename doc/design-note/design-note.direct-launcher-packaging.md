# Knowledge.design-note.direct-launcher-packaging.md
============================================================

# Why CTK Keeps Distribution-Local Launchers

CTK considered macOS `.app` bundles and Windows `.exe` launchers as future Go
Build Strategy choices. The current decision is not to keep them on the Go
implementation roadmap.

The readable command launcher remains the preferred representation:

```text
dist/<name>/<name>       macOS and Unix-like hosts
dist/<name>/<name>.cmd   Windows hosts
```

This is not a technical limitation. It is a decision about the interface that
best represents Distribution ownership and movement safety.

## Runtime integrity is more important than application appearance

A VS Code-family Runtime owns `.data` and `.ext` as a pair. Installed
Extensions may bind metadata or fingerprints to their Runtime state and
physical layout. Moving or sharing only part of that pair can produce
Extension integrity failures.

An `.app` or standalone `.exe` visually suggests a self-contained application
that may be copied or moved independently. That expectation is unsafe when the
real launch target is a managed Distribution:

```text
Distribution identity
  = metadata
  + .data
  + .ext
  + launch behavior
```

Packaging only the entry point would hide this relationship without making the
Runtime portable. Packaging the complete Runtime would increase size and still
need to preserve Platform-owned storage invariants.

The current command launcher is intentionally located inside the Distribution.
Its relative paths make the ownership boundary visible, keep the generated
projection readable, and allow the complete Distribution to move as one unit.

## Existing public launch interfaces are sufficient

When `ctk` is on `PATH`, a Runtime can be launched from any directory:

```text
ctk launch <distribution> [target...]
```

When `ctk` is not on `PATH`, the generated command launcher can be invoked by
path. A user who wants a GUI entry point can place a small shell script, batch
file, or Windows shortcut in any convenient location. That movable interface
calls CTK or the Distribution-local launcher while the Runtime remains under
CTK management.

```text
movable shell / batch / shortcut
  → ctk launch <identity>
  → managed Distribution
```

This provides most practical GUI-launch benefits without introducing another
packaging, signing, update, or identity layer.

## When packaged launchers may become appropriate

Packaged launchers would become worth reconsidering only if CTK deliberately
adopts a product model where Distributions are internal and users interact
only with a published interface:

```text
CTK-managed storage
  └── hidden or non-user-facing Distribution

Applications / Start Menu / Desktop
  └── published launcher interface
```

In that model, the launcher should remain a reference rather than pretend to
contain a movable Runtime.

- A macOS `.app` may contain a launcher and an explicit symlink or identity
  reference to CTK-managed storage.
- Windows should normally use a `.lnk` or a small generic launcher that resolves
  `CTK_HOME` and Distribution identity. A per-Distribution native executable is
  not automatically required.
- Hard links are not a general Runtime representation because they do not model
  directories, cannot cross volumes, and obscure copy and ownership semantics.
- A missing or moved Distribution must produce an explicit diagnostic rather
  than silently creating a second Runtime.

This approach also requires concrete decisions for `CTK_HOME`, multiple
workspaces, launcher publication and removal, rename behavior, upgrades,
signing, notarization, and Windows trust presentation.

## Current decision

- Keep readable Distribution-local command launchers as the Go default.
- Keep `ctk launch <identity>` as the stable movable public interface.
- Use user-authored shell, batch, or OS shortcut files for optional GUI entry
  points.
- Do not implement `.app` or `.exe` only for appearance.
- Keep packaged launchers off the Go roadmap. A future concrete
  non-user-facing Distribution model may create a new proposal, rather than
  implicitly reopening this one.

This is a Go implementation strategy, not a Core prohibition. Another language
or product layer may choose packaged launchers while preserving the same
Distribution and Runtime integrity contracts.

See also:

- [CodeVenv Note](../note/note.codevenv.md)
- [Distribution Contract](../contract/contract.distribution.md)
- [Binary distribution Future](../future/future.candidates.md)
