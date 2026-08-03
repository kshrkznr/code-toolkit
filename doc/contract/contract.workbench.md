# Knowledge.contract.workbench.md
============================================================

# Freeze and Inspect Workbench Contract

A Workbench is review material for observing, comparing, and evolving Cookbook
Source. It is not canonical Cookbook Source, a Runtime Lock, or an Archive.

## Shared capability

Freeze Draft and Inspect operations use compatible observable content and may
share Artifact renderers. Their intent differs:

- Freeze Draft contains proposed changes for later editing and Freeze Commit.
- View is a disposable Inventory of one completed source.
- Sync compares two completed sources without assigning ownership.

Freeze Draft may omit a typed Artifact when its comparison result is `SAME`;
the Summary remains evidence that the Artifact was compared. View renders the
selected source without treating current Cookbook state as its comparison
baseline. Sync renders the transition from its left source to its right source.

Settings, Extensions, Keybindings, Snippets, Tasks, and MCP participate only
when they are supported and managed for the applicable Runtime scope. Profile
ownership follows the Runtime Artifact Contract.

## Source and observation boundary

Distribution observation uses a trusted complete Lock. The applicable Recipe
Lock mode determines whether observation is refreshed, reused, or selected
interactively. A failed refresh does not silently fall back to stale trusted
state.

Recipe and Ingredient Views resolve Cookbook Source without requiring a
Distribution or creating Platform state. Distribution View observes Runtime
state without assigning it to current Ingredients.

Archive is not an Inspect source. Reconstruct an Archive as a Distribution when
Runtime observation is needed.

## View and Sync

Viewpoints expose the identity of their source kind and source document or
Distribution. Untyped input must be classified by content without ambiguous
precedence guessing.

Sync accepts completed Distribution or Recipe sources in either position.
Ingredient is not a Sync source. Omitted sides may use interactive selection,
but the resulting operation receives explicit left and right values.

Adopting View or Sync material does not require a CTK operation. A reviewer may
copy selected typed Artifacts into a Draft, and an implementation may provide a
Commit operation for explicit partial adoption.

## Classification boundary

Freeze Draft reports observed differences. It does not infer which Ingredient
should own an observed value.

The Workbench may list Ingredients selected by a Recipe and other discoverable
Ingredients as review context. That Inventory is not an ownership assignment.
Recipe- and Ingredient-aware mapping belongs to their corresponding View
sources. Responsibility assignment remains with human or automated review and
Freeze Edit and Commit.

## Workbench replacement

When a target Workbench exists, interactive execution asks the user to Abort or
Replace. Explicit conflict behavior remains available for automation.

- interactive execution defaults to asking
- non-interactive execution defaults to aborting
- replacement constructs and validates the next Workbench in staging
- the current Workbench remains until the staged result is ready
- failure retains or restores the previous usable Workbench

Freeze Draft may retain one previous generation for local recovery. Disposable
Inspect viewpoints require atomic publication but need not retain history.

## Summary and history boundary

A Workbench provides a human-readable Summary containing enough provenance and
result information to review the generated Artifacts. It may include:

- source and Recipe identity
- OS and Platform
- observation and generation time
- generated Artifact inventory and comparison status
- observed Extension versions
- diagnostic Recipe differences

The Summary is informational. Freeze Commit must not parse it, depend on edits
to it, or treat it as authorization. Typed Draft Artifacts are the only Runtime
Artifact Commit input.

Workbench replacement is responsible for safe publication, not collaborative
version control. Git or another external text-oriented tool may own longer
history and rollback.

## Workbench placement

A Workbench is logically associated with Cookbook review, but its physical path
is not a Required Capability or Required Source Compatibility rule.

Recommended repository-visible locations are:

```text
cookbook/draft
cookbook/inspect/<viewpoint>
```

Recipe and Ingredient resolution must not consume generated Workbench content
as canonical Source. A Cookbook remains valid without either directory.

## Representation requirements

An editable Workbench representation must:

- support line-oriented comparison and editing
- produce deterministic ordering and stable repeated output
- preserve JSON object hierarchy and JSON value types
- distinguish dotted property names from nested object paths
- support meaning-preserving reconstruction
- allow observed content to be divided among or combined into Ingredient
  candidates without unnecessary editing hazards

No one Settings assignment syntax is Required Source Compatibility. Gron, CTK
JSON Flat, or another reversible representation may satisfy these requirements.

A repository-visible Workbench should group observations by Artifact kind while
retaining the structure required by that Artifact's Runtime I/O boundary.
Exact filenames and Markdown styling are implementation choices.

## Difference direction

Freeze Draft differences describe the transition from the current Cookbook
result to the observed Distribution:

- `-` is current Cookbook content
- `+` is observed Distribution content proposed for Commit

Sync uses the same representation from its left source to its right source.
Missing comparison source content may be rendered as an all-added candidate.
Unavailable content must remain distinguishable and does not prevent unrelated
Artifacts from being reviewed.

## Freeze Commit boundary

Freeze Commit applies only typed Draft inputs that are present. An absent typed
Artifact means no operation, not deletion or convergence to empty state.

Commit targets are relative to the Cookbook Ingredient root. Absolute paths,
parent traversal, and resolution outside that root are invalid. Each Artifact
validates targets against its own Cookbook Source boundary.

Lightweight Patch semantics may treat removals as conditional and additions as
proposed values. Unresolved removals are reported rather than silently deleting
a different current value.

Commit constructs and validates all affected files before publication and
rolls the set back on publication failure. It does not claim trusted-Lock or
resolved-Runtime verification; ownership and configuration meaning remain with
the reviewer.

Comment-bearing source must not lose comments without an explicit acceptance
boundary. Commit does not silently delete empty Resource files unless a
separate cleanup operation defines that behavior.

## Implementation-specific resolution

The primary implementation's viewpoint names, typed Artifact filenames, CTK
JSON Flat presentation, Summary shape, command flags, Freeze Commit parsing,
and Merge Rules integration are defined by the
[Go Workbench Contract](../../go/doc/contract/contract.workbench.md).
