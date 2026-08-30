# Go.contract.workspace.md
============================================================

# Go Workspace Contract

This Contract realizes the shared
[Workspace Concept API](../../../doc/integration/integration.workspace.md) for
the Go implementation.

## Discovery

Go resolves a Workspace in this order:

1. `CTK_HOME`, when explicitly configured;
2. the current directory or its ancestors.

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

## Activation bootstrap

When interactive `ctk activate` does not discover a Workspace and `CTK_HOME`
is not configured, Go prompts for a Workspace path before activation. The
editable suggestion is `~/ctk`: Enter accepts it, another path replaces it,
and Escape cancels without writing.

After path confirmation, Go creates only the minimum footing and continues the
requested activation in the same invocation. Activation bootstrap does not
write the executable sample. The imported `origin.<platform>` Distribution and
subsequent Freeze workflow provide the starting point for that Workspace.

An explicitly configured but invalid `CTK_HOME` is reported rather than
replaced or initialized. Non-interactive activation and other
Workspace-dependent commands do not create a Workspace implicitly; they report
how to run within a Workspace, set `CTK_HOME`, or use `ctk init <path>`.

The suggested path is a visible, user-owned Workspace location. It is not a
user-global active-Workspace selector, and CTK does not select it automatically
on later invocations outside that Workspace. The user may work below the new
Workspace or configure `CTK_HOME` for later discovery. Bootstrap output states
that boundary so a successful first activation does not imply persisted
selection.

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
