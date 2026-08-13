# Knowledge.note.platform-registry.md
============================================================

# Built-in Platform Registry

The Go implementation resolves every incorporated VS Code ecosystem Platform
through one Built-in Registry. This Note explains the operational boundary of
that Registry after the first centralization refactoring.

It does not define an external file format or make user-defined Platforms
available. Those remain in the
[Workspace Platform Definitions Future](../future/future.platform-registry.md).

## What the Registry owns

The Registry centralizes declarative differences for `code`, `codium`, `kiro`,
`cursor`, and `devin-desktop`:

- Platform identity and command
- macOS and Windows Host User data, managed User, and Extension paths
- process identities and registered additional filter IDs by OS
- ordered Extension Pool Repository candidates and download capability

Every consumer resolves the same definition rather than maintaining another
Platform switch. Launcher, Direct Launcher, Runtime construction, CodeVenv,
process management, and convergence therefore share one Platform identity.

```text
Built-in definitions
        ↓ validate
Platform Registry
        ├── Host path resolution
        ├── process identity + named filters
        └── ordered Pool Repository policy
                    ↓
              CTK services
```

## Declaration and behavior remain separate

Registry data selects values or a named strategy already implemented by CTK.
It does not contain executable behavior.

```text
Platform declaration                  CTK implementation
--------------------                  ------------------
Host path base + relative path   ───► Host path resolver
process identity                 ───► candidate discovery
additional filter ID             ───► registered process filter
Repository ID + download flag    ───► Pool lookup and Repository connector
```

The shared VS Code Runtime Adapter continues to own Settings, Profiles,
Extensions, Runtime Artifacts, Profile persistence, and CLI semantics. CodeVenv
continues to own lifecycle transactions, recovery, links or Junctions, process
ownership, stopping, and waiting. Repository implementations continue to own
endpoints, exact VSIX acquisition, response parsing, and artifact validation.

## Host path resolution

Built-in Host roots use a registered base and a relative path. Current bases
are `home`, `application-support`, and `roaming-application-data`. The managed
User path is relative to the resolved User data root.

Paths are interpreted using the target OS represented by the definition. A Go
binary running on Windows must not reinterpret a stored macOS declaration with
Windows separators, and the reverse is also true.

The resolver also understands an OS-native absolute path without a base. That
shape is covered as a Registry invariant but is not currently loaded from
Workspace configuration.

## Process selection

Process selection composes Platform data with CTK behavior:

```text
Platform process identity
        ↓
CTK default candidate filters
        ↓
registered additional filter IDs
        ↓
CTK Runtime ownership, stop, and wait
```

The Registry accepts only registered additional filter IDs. It does not execute
regular expressions, scripts, interpreters, or external providers. The current
Windows definitions all declare `same-name-root`, preserving the behavior that
excludes a same-named child process from desktop-root selection.

## Pool and Repository resolution

Each Platform owns an ordered array of Repository candidates. Every candidate
has a Repository ID and a flag indicating whether CTK may acquire an exact VSIX
through that Repository's registered connector.

Resolution remains a two-stage CTK operation:

1. inspect every Repository-scoped local Pool candidate in declared order;
2. after a complete cache miss, and only for a Recipe request of
   `extension-pool: refresh`, try download-enabled candidates in the same order.

The Platform flag declares capability. `reuse | refresh` remains the Recipe's
request for one operation. Repository protocol and validation do not become
Platform behavior.

## Registration invariants

Built-in registration fails before Host mutation when a definition is
incomplete or contains an invalid reference. Current validation and parity
tests establish:

- unique Platform identity and complete macOS and Windows declarations
- resolvable Host paths for every claimed OS
- registered process filter and Repository IDs
- no duplicate Repository candidate within one Platform
- an exact acquisition connector for every download-enabled Repository
- parity of all five built-in commands, paths, process declarations, and Pool
  order

Unknown Platform identities are rejected rather than inferred from a command.
Identity and command remain distinct, and CTK does not infer that two identities
with the same Host paths describe the same installed application.

## Adding another Built-in Platform

Centralization removes the earlier need to update several Platform switches,
but it does not turn incorporation into a data-only support claim.

1. Follow the
   [VS Code Ecosystem Platform Intake](../project-knowledge/note/note.vscode-ecosystem-platform-intake.md).
2. Add the complete definition to the Registry.
3. Add a named filter or Repository connector only when observed behavior
   requires one.
4. Update completeness and focused behavior tests.
5. Validate every claimed OS and record the result in the
   [Go Platform Support Inventory](../../go/doc/platform-support.md).

`Supported` remains reserved for Platforms incorporated and validated by CTK.

## Related documents

- [Platform Differences as Boundary Evidence](note.platform-boundary-evidence.md)
- [Extension Resolution](note.extension-resolve.md)
- [Why CTK Keeps Platform Inside the VS Code Ecosystem](../design-note/design-note.vscode-ecosystem-scope.md)
- [Why Platform Definitions Belong to Workspace Integration](../design-note/design-note.platform-definition-scope.md)
- [Platform Runtime I/O Contract](../contract/contract.platform-runtime-io.md)
- [Go Platform Support Inventory](../../go/doc/platform-support.md)
