# Go.contract.runtime-artifacts.md
============================================================

# Go Runtime Artifact Contract

This Contract specializes the shared
[Runtime Artifact Contract](../../../doc/contract/contract.runtime-artifacts.md).

## Supported Artifacts

Go resolves, converges, observes, Locks, recovers, archives, views, and freezes
managed Keybindings, Snippets, Tasks, and MCP alongside Settings and Extensions.

The VS Code-family Adapter currently maps:

- Keybindings to scope-appropriate `keybindings.json`
- Snippets to the scope-appropriate `snippets/` file collection
- MCP to scope-appropriate `mcp.json`
- default User Tasks to `<User>/tasks.json`

Profile-local Tasks is explicitly unsupported until repeatable native Platform
I/O is confirmed. It is reported as unsupported rather than valid empty content.

## Cookbook ambiguity

Go supports the compatible shared layout families and rejects multiple physical
definitions of the same logical Resource.

## Extension Set companion Artifacts

A referenced Extension Set may provide managed Settings, Keybindings,
Snippets, Tasks, and MCP Resources. Go applies their existing Artifact
composition rules without Set-specific merge semantics.

Within an effective scope, resolution proceeds:

```text
Runtime Extension Resources
  → Runtime Extension Set Resources
  → Runtime Ingredient Resources
  → Profile Extension Resources
  → Profile Extension Set Resources
  → Profile Ingredient Resources
```

Repeated concrete Extension and Set references contribute Resources only at
their first applicable position. Set companion Resources follow the Artifact's
existing Default and named-Profile ownership strategy. An unmanaged Artifact
is not resolved, including a malformed companion Resource that would otherwise
participate.

## Snippet Commit behavior

Runtime observation represents a rename as old-name absence plus new-name
addition. Freeze Commit does not infer rename intent or delete Resource files.
Accepting only the removal side converges the old Snippet Resource to a valid
empty object; accepting the addition may create the new named Resource.

## Tasks normalization

Go compares absent Tasks, an empty root object, and the default `2.0.0` envelope
with empty `tasks` and `inputs` as the same empty semantic state. Empty arrays
are ignored in an otherwise non-empty comparison.

This normalization applies to Workbench, Recovery, and Archive verification.
Lock retains the Platform's actual document, and Build or Apply may write the
explicit VS Code-compatible envelope.

## Lock completeness

The Go Lock Manifest distinguishes absent observations, valid empty content,
unmanaged scopes, and unsupported content. When the Recipe requires a managed
Profile-local Artifact, a trusted Lock must contain that observation.

New supported Artifacts participate in semantic verification, Workbench, and
Archive completeness through the same Manifest boundary.
