# Knowledge.design-note.vsix-acquisition.md
============================================================

# Why CTK Keeps VSIX Acquisition Explicit

A VS Code-family Platform can normally install an Extension by ID through its
own CLI and Repository. The same CLI can also install a local VSIX path. Both
are ordinary Platform capabilities; Visual Studio Code documents both forms in
its [command-line interface](https://code.visualstudio.com/docs/configure/command-line#_working-with-extensions).

CTK previously combined normal installation with a second behavior: after
observing an installed version, it automatically contacted a Repository,
downloaded the exact VSIX, and stored it in the shared Pool. That was useful for
offline reconstruction, but it made artifact acquisition an implicit side
effect of ordinary environment construction.

## Installation and acquisition carry different responsibility

Passing an Extension ID to the Platform preserves the Platform's own
Repository selection, authentication, compatibility checks, and installation
path:

```text
Recipe Extension ID
  → Platform-owned CLI operation
  → Platform-owned Repository behavior
```

Downloading a VSIX into CTK storage is a different act:

```text
observed Extension ID and version
  → CTK contacts a configured Repository
  → CTK stores a reusable local artifact
```

Once CTK performs the second flow, the artifact may later be supplied to a
different Platform or included in an Archive. Technical availability does not
establish permission or compatibility. Marketplace offers can have
publisher-specific terms; Microsoft's Marketplace guidance likewise directs
customers to review the applicable
[Publisher Terms](https://learn.microsoft.com/en-us/legal/marketplace/marketplace-terms#3-publisher-terms).

CTK should therefore avoid presenting "downloadable" as equivalent to "usable
with any VS Code-family IDE" or "redistributable".

## Current decision

- Keep Extension ID installation through the Platform CLI as the normal route.
- Keep local VSIX and existing Pool artifacts usable through Platform-owned
  install operations.
- Make CTK-owned Repository download and Pool update explicit through
  `extension-pool: refresh`.
- Default `extension-pool` to `reuse`, which performs no CTK-owned download.
- Do not turn a failed Platform Repository install into a new Repository
  download. A previously cached secondary Pool artifact may still be tried.
- Require an exact local artifact for Archive creation unless the Recipe
  explicitly permits `refresh`.
- Treat Repository terms, individual Extension licenses, target-Platform use,
  and redistribution permission as user decisions outside CTK's guarantees.

## Trade-off

The default no longer prepares an offline VSIX cache automatically. A user who
wants that cache must state the intent in the Recipe, and an Archive can fail
until the required exact artifacts are supplied locally or acquisition is
enabled.

This cost is intentional. Ordinary Builds remain convenient through the
Platform path, while the operation that turns CTK into an artifact acquisition
client is visible and reviewable.

## Boundary

This note explains why acquisition is explicit. The
[Cookbook Representation Contract](../contract/contract.cookbook.md) defines
the accepted strategy values, and [Extension Resolution](../note/note.extension-resolve.md)
defines their operational behavior. CTK validates VSIX structure, identity,
and version; it does not provide legal advice or certify an Extension for a
particular Platform.
