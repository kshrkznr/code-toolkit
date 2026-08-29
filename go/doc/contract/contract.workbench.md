# Go.contract.workbench.md
============================================================

# Go Workbench Contract

This Contract specializes the shared
[Freeze and Inspect Workbench Contract](../../../doc/contract/contract.workbench.md).

## Commands and sources

```text
ctk freeze draft [dist] [--on-conflict abort|replace]
ctk freeze commit [--force]
ctk view [source] [--on-conflict abort|replace]
ctk view dist [dist] [--on-conflict abort|replace]
ctk view recipe [recipe] [--on-conflict abort|replace]
ctk view ingredient [all|layer|layer.name] [--on-conflict abort|replace]
ctk sync [left] [right] [--on-conflict abort|replace]
```

Omitted interactive sources use the Native Selector. Distribution, Recipe, and
Ingredient View use content-based source classification when the source kind is
not explicit. Sync accepts Dist-to-Dist, Dist-to-Recipe, Recipe-to-Dist, and
Recipe-to-Recipe; Ingredient is not a Sync source.

## Repository placement

```text
CTK_HOME/cookbook/draft
CTK_HOME/cookbook/inspect/dist.<name>
CTK_HOME/cookbook/inspect/recipe.<name>
CTK_HOME/cookbook/inspect/ingredient.<name>
CTK_HOME/cookbook/inspect/sync.<left>.<right>
```

Generation uses staging and atomic replacement. Freeze Draft retains at most
one `.old` generation. Inspect viewpoints are disposable and do not retain
`.old` after successful replacement.

These paths remain Workspace-local when `.config/workspace.yaml` selects an
external Cookbook Source. Recipe and Ingredient reads and Freeze Commit writes
use that resolved Source; generated Draft and Inspect state does not move with
it. The [Go Workspace Contract](contract.workspace.md) owns path resolution.

## Typed Artifacts

The Go Workbench uses one review file per Artifact kind:

```text
summary.md
recipe.draft.yaml
settings.draft.md
extensions.draft.md
keybindings.draft.md
snippets.draft.md
tasks.draft.md
mcp.draft.md
```

Freeze Draft omits a typed file whose managed content is `SAME`. View retains
available Inventory Artifacts. `summary.md` presents result, provenance,
Profile topology, exact observed Extension versions, Recipe difference, and
observation timing; Freeze Commit never reads it.

Summary Artifact status uses `SAME`, `DIFFERENT`, `MISSING`, and `UNAVAILABLE`.
Settings counts use CTK JSON Flat paths, Extensions count exact IDs, and Recipe
comparison reports status plus a diagnostic unified diff rather than
field-level counts. When scopes observe multiple versions for one exact
Extension ID, the Summary preserves every distinct version.

Summary provenance renders paths inside the active CTK Workspace with the
literal `$CTK_HOME/<workspace-relative path>` identity. This keeps generated
review material portable and avoids publishing a host-local absolute path.
Sources outside the Workspace retain their supplied path because CTK cannot
assign them a Workspace identity.

Settings assignments use the optional CTK JSON Flat Format. Extensions remain
exact line-oriented IDs. Other Runtime Artifacts retain the JSON or named-file
structure defined by their shared Artifact Contract.

## Draft heading and difference format

Typed Markdown uses:

```text
# <Artifact> Draft
## Inventory: Used by Recipe / Inventory: Available but Unused
### <copyable Resource name>
## Difference
### <Ingredient-relative Commit target>
#### <optional review group>
```

Level-three headings under `## Difference` are plain paths relative to
`cookbook/ingredient`. `Target:` prefixes, code spans, and alternate target
spellings are not accepted Go Commit syntax.

Unified-diff markers describe the transition into the proposal: `-` is source
content and `+` is proposed content. Freeze Draft compares current Cookbook to
observed Distribution; Sync compares left to right.

## Freeze Commit inputs

`recipe.draft.yaml` and each supported `*.draft.md` are independent optional
inputs. Presence opts that Artifact into Commit. Absence means no operation.
The Workbench remains after success.

Settings targets end in supported `settings.json` or `settings.jsonc` forms;
Extension targets are `extensions` or end in `.extensions`; other targets must
match their Keybindings, Tasks, MCP, or Snippet Resource forms. Absolute paths,
`..`, and paths escaping the Ingredient root are errors.

Repeated target sections are aggregated. Distinct logical paths may share a
target. Incompatible additions for the same logical path are invalid.

Patch behavior is:

- `-` removes only a matching current value
- missing or mismatched `-` is unresolved and does not stop other Commit work
- `+` adds or sets the proposed value
- removals apply deepest-first and additions container-first

Go constructs every affected file in staging, validates syntax and targets,
and publishes the set with rollback on failure. Existing empty Resources remain
empty files or objects; Commit does not delete them.

## Recipe and JSONC publication

Recipe Commit resolves identity by document `name`, `os`, and `platform`. One
match is updated; multiple matches are an ambiguity error; no match creates the
convenience filename `<name>.<os>.<platform>.yaml` without making filename the
identity.

JSON and comment-free JSONC may be rewritten semantically. Comment-bearing
JSONC aborts unless `--force` explicitly accepts comment loss.

## Merge Rules

Go Merge Rules are read from `cookbook/kitchen-notes/go.merge-rules.yaml`.
Exact logical Settings paths apply Cookbook-wide and are not selected per
Recipe. Unsupported versions, malformed paths, array indices, Workbench
selectors, and Union rules resolving to non-arrays are errors. Semantic rewrite
need not preserve YAML comments or formatting.

## Open Questions

- Additional Platform Artifact kinds beyond the supported Go Adapter matrix.
- Review-specific generated Views and alternative grouping adapters.
