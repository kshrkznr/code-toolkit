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

## Optional initialization

`ctk init <path>` creates a discoverable Workspace footing at an explicit
path. It is a convenience for binary-only installation, not a required CTK
operation or a prerequisite for using an existing Workspace.

The path argument is required. Relative paths resolve from the caller's current
directory. Initialization does not perform Workspace discovery, select the new
Workspace for later invocations, persist or modify `CTK_HOME`, or create
`.config/workspace.yaml`.

The minimum footing contains empty `cookbook/recipe` and
`cookbook/ingredient` directories. By default initialization also writes the
small macOS and Windows `vscode-sample` Recipes and their Ingredients. The
`--exclude-sample` option creates only the minimum directories.

Existing directories and byte-identical sample files are retained. If any
sample target contains different content, initialization reports all detected
conflicts before writing sample content and does not overwrite them. There is
no force mode.

After initialization, current-directory discovery selects the new Workspace
from its root or descendants. An existing `CTK_HOME` continues to take
precedence. When it identifies another path, `ctk init` reports that the caller
must unset it to use current-directory discovery.

## Non-migration boundary

Go does not create a configuration file, persist `CTK_HOME`, remember previous
locations, or move content when a value changes. Archive and Extension Pool
overrides are not accepted by the current schema.
