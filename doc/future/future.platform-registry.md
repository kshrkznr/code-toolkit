# Knowledge.future.platform-registry.md
============================================================

# Future: Workspace Platform Definitions

CTK now resolves its five incorporated VS Code ecosystem Platforms through one
validated Built-in Registry. A later slice may let a selected Workspace add
complete Platform definitions without rebuilding CTK.

This Future preserves that unimplemented extension point. The operational
behavior of the current Registry has moved to the
[Built-in Platform Registry Note](../note/note.platform-registry.md).

## Current foundation

The Go implementation already centralizes:

- identity and command
- macOS and Windows Host paths
- process identities and registered additional filter IDs
- ordered Pool Repository candidates and download capability

Named filters, Repository connectors, the shared VS Code Runtime Adapter, and
CodeVenv lifecycle behavior remain implemented services. Completeness tests and
macOS and Windows validation cover all five Built-in definitions.

The remaining candidate is loading an equivalent validated definition from:

```text
CTK_HOME/.config/platform/*.yaml
```

## Candidate definition

```yaml
name: cursor-compatible
command: cursor-compatible
host:
  macos:
    user-data:
      base: application-support
      path: CursorCompatible
    user: User
    extensions:
      base: home
      path: .cursor-compatible/extensions
  windows:
    user-data:
      base: roaming-application-data
      path: CursorCompatible
    user: User
    extensions:
      base: home
      path: .cursor-compatible/extensions
process:
  macos:
    identity:
      executable: CursorCompatible.app/Contents/MacOS/CursorCompatible
  windows:
    identity:
      executable: CursorCompatible.exe
    additional-filters:
      - same-name-root
pool-policy:
  - repository: cursor-marketplace
    download: true
  - repository: visual-studio-marketplace
    download: false
```

This representation remains a candidate until file loading, diagnostics, and
real Workspace use validate the exact schema.

## Host path forms

A Host root uses one of two forms:

- known `base` plus relative `path`;
- omitted `base` plus an OS-native absolute `path`.

Known bases initially include `home`, `application-support`, and
`roaming-application-data`. Unknown bases, expansion expressions, `~`, and
environment-variable interpolation are rejected. A path with `base` must be
relative; a path without `base` must be absolute.

The managed `user` path remains relative to the User data root. Path validation
uses the target OS declared in the definition rather than the OS on which a file
happens to be inspected.

## Process selection

An external definition may declare process identity and select registered
additional filters:

```text
Platform process identity
        ↓
CTK default filters
        ↓
registered additional filter IDs
        ↓
CTK Runtime ownership, stop, and wait
```

The first loading slice should accept Built-in Filter Registry IDs only. It
should not discover or execute regular expressions, scripts, interpreters, or
external providers.

If an external filter provider is later observed as necessary, its location,
manifest, runtime, trust, timeout, versioned I/O, diagnostics, and failure
behavior need one independent design. A shell script and an interpreter-backed
provider must identify their execution requirements explicitly rather than
being inferred from a filename.

## Pool and Repository separation

Pool policy remains an ordered array inside the Platform definition. It is not
a separately named policy object.

```text
Platform pool-policy[]
    Repository ID + download capability
              ↓
Repository Registry
    exact acquisition and validation connector
```

Resolution first inspects every Repository-scoped local Pool candidate in
order. Only after a complete miss, and only when the Recipe requests
`extension-pool: refresh`, may CTK try candidates with `download: true`.

At registration:

- every Repository ID must resolve through the Repository Registry;
- duplicate IDs in one Platform are rejected;
- `download: true` requires an exact VSIX acquisition connector;
- `download: false` may use the ID only as a local Pool identity.

The first loading slice should accept Built-in Repository IDs and connectors
only. External Repository connector providers have the same design burden as
external process filters and remain independently deferred.

## Identity and collision boundary

Platform identity and command are distinct. Command duplication is permitted,
but CTK does not infer a Platform from a command.

An external definition cannot replace a Built-in identity. A different command
or integration uses a new complete identity such as `cursor-code`; there is no
partial command binding or Recipe command override.

If two identities resolve the same Host paths, current CodeVenv filesystem and
Selection checks still reject conflicting managed links or Junctions. CTK does
not attempt to infer whether two independently authored definitions represent
the same application beyond observable managed evidence. Compatibility of the
custom definition remains its author's responsibility.

## Workspace ownership

External definitions are Workspace integration configuration, not Recipe
Source, Kitchen Notes, or OS user-global configuration.

```text
CTK_HOME/
├── .config/
│   ├── workspace.yaml       optional location overrides
│   └── platform/            external Platform definitions
├── cookbook/
│   ├── draft/               Workspace-local Workbench output
│   └── inspect/             Workspace-local Workbench output
├── dist/
├── archive/
└── .vsix/
```

An optional Workspace configuration may later relocate only static Cookbook
Source and the Dist root:

```yaml
paths:
  cookbook-source: /path/to/cookbook
  dist: /path/to/dist
```

`cookbook-source` contains `recipe/` and `ingredient/`. Workbench `draft/` and
`inspect/` remain below `CTK_HOME/cookbook` so generated review state does not
enter a versioned Source repository accidentally.

Archive and Pool remain under `CTK_HOME` until a concrete sharing requirement
is observed. No `ctk init` is required: absent configuration resolves current
default directories. CTK does not persist `CTK_HOME`, track an old Dist root,
or migrate managed state after a location change.

Package managers distribute the binary and compiled Built-in definitions. They
do not own or update Workspace configuration, Cookbook, Dist, Archive, or Pool
state.

## Support boundary

`Supported` is reserved for Built-in Platforms incorporated and validated by
CTK. A Workspace definition is a user extension point, not a self-declared CTK
support claim.

The definition author owns compatibility with the selected application,
command, Host paths, process declarations, Repository policy, and intended OS.
CTK remains responsible for schema validation and its common lifecycle,
recovery, and artifact safety invariants.

Promoting a Workspace definition to Built-in requires normal Intake,
implementation review, automated tests, and real-machine validation on every
claimed OS.

## Initial implementation boundary

The first useful slice would include:

1. load complete definitions from `CTK_HOME/.config/platform/*.yaml`;
2. validate identity, per-OS paths, process declarations, filter references,
   ordered Repository candidates, and download capability before Host mutation;
3. merge valid external identities with Built-ins without permitting Built-in
   replacement;
4. report file- and field-specific diagnostics;
5. prove that existing Built-in completeness and lifecycle behavior is
   unchanged;
6. validate at least one disposable Workspace definition on each claimed OS.

After implementation, definition locations, fields, diagnostics, and examples
should move to a Platform Definition Note. Provider mechanisms that remain
unimplemented should stay in Future.

## Deferred independently

- external process filter providers
- external Repository connector providers
- automatic Dist relocation or migration
- multiple active Workspace selection
- Archive or Pool location overrides
- non-VS Code ecosystem Runtime adapters

## Related knowledge

- [Built-in Platform Registry](../note/note.platform-registry.md)
- [Why Platform Definitions Belong to Workspace Integration](../design-note/design-note.platform-definition-scope.md)
- [Why CTK Keeps Platform Inside the VS Code Ecosystem](../design-note/design-note.vscode-ecosystem-scope.md)
- [VS Code Ecosystem Platform Intake](../project-knowledge/note/note.vscode-ecosystem-platform-intake.md)
- [Platform Differences as Boundary Evidence](../note/note.platform-boundary-evidence.md)
- [Platform Runtime I/O Contract](../contract/contract.platform-runtime-io.md)
- [Go Platform Support Inventory](../../go/doc/platform-support.md)
