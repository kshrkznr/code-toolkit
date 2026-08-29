# Go.contract.workspace.md
============================================================

# Go Workspace Contract

This Contract realizes the shared
[Workspace Concept API](../../../doc/integration/integration.workspace.md) for
the Go implementation.

## Discovery

Go resolves a Workspace in this order:

1. `CTK_HOME`, when explicitly configured;
2. the current directory or its ancestors;
3. the repository-local location relative to the executable.

A Workspace is discoverable when it contains either:

- both `cookbook/recipe` and `cookbook/ingredient`; or
- `.config/workspace.yaml`.

After selecting the nearest marker, Go validates its configuration and
Cookbook Source. Invalid configuration is reported rather than skipped in
search of a different Workspace.

## Configuration

`.config/workspace.yaml` is optional. Its current representation is:

```yaml
paths:
  cookbook-source: /path/to/cookbook
  dist: /path/to/dist
```

Relative values resolve from `CTK_HOME`; absolute values remain absolute.
Unknown fields and multiple YAML documents are rejected. An empty or absent
value uses its default.

Resolved Cookbook Source must contain `recipe/` and `ingredient/` directories.
The Dist root may be absent until an operation creates content there.

## Resolved paths

| Responsibility | Default | Configurable |
| --- | --- | --- |
| Cookbook Source | `CTK_HOME/cookbook` | `paths.cookbook-source` |
| Dist | `CTK_HOME/dist` | `paths.dist` |
| Workbench Draft and Inspect | `CTK_HOME/cookbook` | No |
| Archive | `CTK_HOME/archive` | No |
| Extension Pool | `CTK_HOME/.vsix` | No |

Recipe selection, Ingredient resolution, Kitchen Notes, and Freeze Commit
targets use Cookbook Source. Freeze Draft, View, Sync, and Workbench opening use
the Workspace-local Workbench root.

Go validates Workspace configuration before dispatching a command, including a
command whose later operation would otherwise be read-only. This keeps one
invocation on one inspected path model and prevents Host mutation under an
invalid Workspace.

## Non-migration boundary

Go does not create a configuration file, persist `CTK_HOME`, remember previous
locations, or move content when a value changes. Archive and Extension Pool
overrides are not accepted by the current schema.
