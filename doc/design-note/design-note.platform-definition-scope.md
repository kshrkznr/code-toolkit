# Knowledge.design-note.platform-definition-scope.md
============================================================

# Why Platform Definitions Belong to Workspace Integration

External Platform definitions need a home that makes their ownership visible.
They describe machine integration for a VS Code ecosystem application:
commands, Host paths, process identity, process filters, and Extension Pool
policy.

CTK considered three nearby homes: Recipe Source, Kitchen Notes, and user-global
configuration. None expresses that responsibility as clearly as Workspace
integration configuration.

## Platform definitions are not Cookbook Source

Recipes and Ingredients describe reusable desired Runtime content. A Platform
definition describes how CTK reaches and manages an installed application on a
particular machine.

Putting command aliases or absolute Host paths in Recipe Source would make a
portable Cookbook depend on one machine's installation. The Recipe should keep
selecting a stable Platform identity while the selected Workspace resolves the
corresponding integration definition.

## Platform definitions are not Kitchen Notes

Kitchen Notes belong to a Cookbook. They optionally supplement how one
implementation interprets that Cookbook during Build.

A Platform definition has a wider operational effect:

- it resolves Host User data and Extension paths;
- it participates in process selection and stopping;
- it controls CodeVenv Activate, Use, and Deactivate boundaries;
- it selects local Pool identities and permitted Repository acquisition;
- it applies to Launch, Lock, recovery, and other operations beyond Cookbook
  interpretation.

Calling this a Kitchen Note would broaden Kitchen Notes from Cookbook
interpretation into machine integration and Host lifecycle policy. CTK keeps
that boundary narrow instead.

Workbench state illustrates the same separation. An independently versioned
Cookbook Source may contain `recipe/` and `ingredient/`, while `draft/` and
`inspect/` remain Workspace-local output under `CTK_HOME/cookbook`.

## Platform definitions are not user-global defaults

A user-global definition would change Platform resolution outside the selected
CTK Workspace. That makes it harder to determine which integration rules govern
Host mutation and which configuration must accompany the managed state.

Workspace-local configuration keeps the relationship explicit:

```text
selected CTK_HOME
    ├── Platform integration definitions
    ├── Dist and CodeVenv evidence
    ├── Archive and Pool
    └── Workspace-local Workbench state
```

The definition author remains responsible for compatibility with the chosen
application. CTK remains responsible for validating the definition and
preserving common safety invariants.

## Package managers do not own the Workspace

Homebrew, Scoop, WinGet, or another package manager may install and upgrade the
CTK binary. The package contains compiled Built-in Platform definitions, but it
does not install user definitions into a package prefix or user-global
configuration location.

The Workspace remains user-owned and survives binary upgrades independently.

## Current decision

- Load external Platform definitions as Workspace integration configuration
  under `CTK_HOME/.config/platform`.
- Keep Recipe and Ingredient content in Cookbook Source.
- Keep Kitchen Notes limited to implementation-specific Cookbook
  interpretation.
- Keep Workbench Draft and Inspect output under `CTK_HOME/cookbook`, even when
  static Cookbook Source is located elsewhere.
- Do not add user-global Platform configuration or a multi-Workspace selector.
- Let package managers distribute the binary and Built-in definitions only.
- Reserve `Supported` for Platforms incorporated and validated by CTK; users
  guarantee compatibility for their Workspace definitions.

## Boundary

The [Built-in Platform Registry](../note/note.platform-registry.md) records the
implemented declaration and service boundary. The
[Workspace Platform Definitions Future](../future/future.platform-registry.md)
preserves the candidate external schema, loading sequence, and deferred
provider boundaries.
The [Kitchen Notes Concept](../core/core.cookbook.kitchen-notes.md) defines the
narrow Cookbook interpretation boundary. The
[Code Environment Integration Contract](../contract/contract.code-environment.md)
owns common Host lifecycle safety.
