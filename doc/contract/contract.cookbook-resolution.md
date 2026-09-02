# Knowledge.contract.cookbook-resolution.md
============================================================

# Cookbook Resolution Contract

Cookbook Resolution transforms reusable Cookbook Source into a resolved Runtime
description that can be consumed by Platform Runtime I/O and later lifecycle
orchestration.

The resolved result is referred to conceptually as a Runtime Plan. This name
does not prescribe a language type, file format, or persisted artifact.

```text
Cookbook Source
    │
    ▼
Cookbook Resolver
    │
    ▼
Runtime Plan
    │
    ▼
Platform Runtime I/O
```

## Required Capabilities

- Read the Recipe and Ingredient files defined by the Cookbook Representation
  Contract.
- Resolve Recipe identity, OS, Platform, Runtime Ingredients, Profile
  Ingredients, and applicable strategy.
- Produce the desired default Runtime and named Profile content.
- Preserve Recipe provenance required by later lifecycles.
- Preserve deterministic Resource resolution order.
- Preserve Extension ID spelling and letter case.
- Represent a Recipe whose selected Ingredients resolve to zero Resources.

The Runtime Plan may contain no Settings and no Extensions. A Recipe and its
Ingredient composition remain meaningful before concrete Resources exist.

## Purity Boundary

### Required Contract

Cookbook Resolution:

- does not launch or modify a Platform
- does not create or mutate a Distribution
- does not install or uninstall Extensions
- does not contact an Extension marketplace
- does not modify Recipe or Ingredient Source
- does not infer malformed Source through fallback interpretation

Existing but malformed Resources fail resolution. Missing Resources are valid
and contribute no content.

```text
Resource absent       → valid empty contribution
Resource present      → parse and validate
Resource malformed    → resolution error
All Resources absent  → valid empty Runtime Plan
```

## Runtime Plan Capabilities

A resolved Runtime description exposes, when declared:

- Recipe provenance
- OS identity
- Platform identity
- ordered default Runtime Settings
- default Runtime Extension requirements
- named Profiles
- ordered Profile Settings
- Profile Extension requirements
- Profile inheritance strategy
- Extension source policy used by later construction

The Runtime Plan expresses desired content. It does not prescribe where a
Distribution stores that content.

## Resource Resolution Order

### Required Contract

Ingredient arrays are ordered declarations. Resources selected by a later
Ingredient are resolved after Resources selected by an earlier Ingredient.

Within a variant-capable Ingredient, resolution proceeds:

```text
base
  ↓
OS variant
  ↓
Platform variant
```

The current default Settings stream is:

```text
OS Settings
  ↓
Platform Settings
  ↓
Runtime Extension Settings
  ↓
Runtime Ingredient Settings
  ↓
Profile content assigned to the default Runtime
```

An implementation that adopts Extension Set companion Artifacts inserts the
Set Resources between the concrete Extension Resources and their declaring
Runtime or Profile Ingredient:

```text
Runtime Extension
  ↓
Runtime Extension Set
  ↓
Runtime Ingredient
  ↓
Profile Extension
  ↓
Profile Extension Set
  ↓
Profile Ingredient
```

Repeated Extension or Set references do not repeat their Resource contribution
within one effective scope. This ordering extends the source stream without
changing an Artifact's composition semantics.

A Profile-local Settings stream begins with resolved default Settings and then
applies that Profile's Extension Settings and Profile Settings.

Resolution order is required because VS Code-family Settings use later-value
precedence. Implementations must not reorder semantically ordered Settings
Resources.

## Settings Composition

### Required Contract

- The resolved Settings presented to a Platform are readable by that Platform.
- The required Resource resolution order is preserved.

### Recommended composition strategy

The current VS Code-family strategy materializes one Settings document:

- objects merge recursively
- arrays use the later value
- scalars use the later value
- `null` uses the later value

Formatting and comments need not survive Runtime Plan materialization. Source
files remain unchanged. Comment-preserving read-modify-write behavior belongs to
later persistence workflow design.

## Extension Resolution

### Required Contract

- Default Runtime Extension strategy `runtime` resolves Runtime Extensions.
- Default Runtime Extension strategy `clean` resolves an empty default set.
- A named Profile resolves Runtime Extensions plus its Profile Extensions.
- Extension ID letter case is preserved.
- Extension Settings participate only when their Extension is resolved.
- Extension artifact acquisition is not part of Cookbook Resolution.

The resolved output may carry Extension ID and source policy. Marketplace,
Pool, Archive, and other artifact acquisition belong to Runtime construction.
When acquisition uses multiple Repository candidates, Runtime construction
must preserve enough source identity to apply the Platform's main Repository
policy. A secondary Pool artifact must not become indistinguishable from a
main-Repository artifact before installation policy is evaluated.

The Concept API does not prohibit a future Extension Resource format with
comments.

## Multiple Layout Matches

All documented Ingredient layouts are Required Source Compatibility forms.

Behavior when the same logical Resource exists in more than one layout is not a
cross-implementation Contract. An implementation must document whether it
selects one match or rejects the ambiguity.

Cookbooks should not depend on duplicate definitions across layouts.

## Strategy Scope

Cookbook Resolution interprets strategy that changes desired Runtime content,
including:

- default Runtime Extension mode
- Profile content inheritance
- Extension source policy

Lifecycle strategy such as Lock reuse is carried forward for later lifecycle
orchestration; it does not cause side effects during Cookbook Resolution.

## Open Questions

- Whether Recipe variant identity eventually includes Platform as well as name
  and OS.
- How unknown Recipe fields are retained for future read-modify-write workflows.
- Whether a Runtime Plan benefits from a persisted diagnostic form for Inspect.

## Implementation-specific resolution

The primary implementation's duplicate handling, ambiguity errors, and applied
Settings composition rules are defined by the
[Go Cookbook Resolution Contract](../../go/doc/contract/contract.cookbook-resolution.md).
