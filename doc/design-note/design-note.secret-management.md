# Knowledge.design-note.secret-management.md
============================================================

# Why CTK Does Not Manage Secrets

Secret management was discussed during CTK design but intentionally left
outside the current scope.

The primary reason is responsibility.

CTK manages configuration as plain-text Artifacts. It does not attempt to
determine whether individual values are secrets.

This would otherwise require CTK to understand:

- VS Code core settings
- Extension-specific settings
- MCP configuration
- Future extension-defined schemas
- Secret identification rules

Maintaining those rules would introduce significant complexity while remaining
tightly coupled to external projects.

## Current approach

CTK provides two approaches instead.

- **Managed**
  - Configuration is managed by CTK as plain text.
  - Users are responsible for deciding whether sensitive values should be stored.

- **Unmanaged**
  - CTK intentionally ignores the artifact.
  - Users configure and maintain it manually, preserving existing workflows.

This allows projects to exclude credentials, tokens, or machine-specific configuration from CTK management when appropriate.

`Managed` and `Unmanaged` are responsibility categories in this Design Note,
not a one-to-one list of Recipe literals.

Managed content currently includes Recipe choices such as:

- `runtime`
- `profile`
- `clean`

`unmanaged` explicitly places the selected Artifact scope outside CTK
ownership. MCP uses `unmanaged` as its Default Profile compatibility default,
so MCP configuration enters the managed lifecycle only through an explicit
Recipe choice.

## Future direction

If secret management becomes necessary in the future, the preferred direction is **not** for CTK to become a secret manager.

Instead, CTK should remain responsible only for configuration structure while secret resolution is delegated to the target platform.

Examples include:

- SecretStorage provided by VS Code or an Extension
- Environment variable placeholders
- External secret providers
- Platform-specific credential systems

In this model, CTK manages placeholder references rather than secret values
themselves. Placeholder syntax is still ordinary opaque Artifact content to
CTK. CTK does not resolve a placeholder, validate that its provider exists, or
assume that a reference form supported by one Artifact or Platform is portable
to another.

If integration is required, CTK should expose extension points or adapters
rather than implementing a built-in secret management system. The target
Platform, Extension, execution environment, or external provider remains
responsible for resolving and protecting the referenced value.

See also:

- [Runtime Artifact Contract: MCP Secret Boundary](../contract/contract.runtime-artifacts.md#mcp-secret-boundary)
- [Recipe Build Strategy](../note/note.recipe-build-strategy.md)
