# Knowledge.future.platform-registry.md
============================================================

# Future: Platform Registry and Workspace Definitions

CTK has five observed VS Code ecosystem Platforms. Their declarative
differences now resolve through one Built-in Registry; Workspace definition
loading remains a Future.

This Future preserves the agreed boundary for centralizing those differences
before implementation. It is structured so that, after implementation, the
operational parts can move to a Note without carrying candidate language into
user guidance.

## Intended promotion

After the Registry and external loading are implemented and validated:

- move definition locations, fields, validation, examples, and diagnostics to
  a Platform Definition Note;
- keep implementation coverage and real-machine evidence in the
  [Go Platform Support Inventory](../../go/doc/platform-support.md);
- keep the rationale for Workspace ownership in
  [Why Platform Definitions Belong to Workspace Integration](../design-note/design-note.platform-definition-scope.md);
- leave deferred provider mechanisms in Future rather than presenting them as
  available extension points.

## Current evidence

The Go implementation provides one VS Code ecosystem Runtime model. Before the
Registry refactoring, Platform differences were distributed across:

- built-in identity and command lists;
- macOS and Windows Host paths;
- macOS executable and Windows process identities;
- Windows same-name root filtering;
- Repository-scoped Pool lookup order;
- Repository-specific exact VSIX acquisition.

The Built-in Registry now centralizes those declarations while the service
implementations continue to own behavior. Other observations remain shared CTK
behavior rather than Platform choices:

- Profile creation waits for persistent identity before convergence continues;
- Host default and Dist processes are distinguished through isolated User data;
- process enumeration, ownership, stopping, and waiting remain CTK operations;
- lifecycle transactions, recovery, and artifact validation are invariant.

## Registry boundary

Built-in and Workspace definitions resolve to one validated internal shape.

```text
compiled Built-in definitions ─┐
                               ├─► Platform Registry ─► CTK services
Workspace definitions ─────────┘          │
                                          ├─► Filter Registry
                                          └─► Repository Registry
```

```text
PlatformDefinition
    Identity
    Command
    SupportedOS
    HostPathsByOS
    ProcessDefinitionByOS
        Identity
        AdditionalFilterIDs
    PoolRepositories
        RepositoryID
        DownloadSupported
```

Platform definitions contain declarations and registered IDs. They do not
contain arbitrary execution logic.

## Implemented foundation

The first refactoring centralizes the five Built-in definitions without loading
external files and without intentionally changing behavior.

It establishes that:

1. `code`, `codium`, `kiro`, `cursor`, and `devin-desktop` resolve through one
   Registry;
2. every claimed OS resolves complete Host paths, process identity and filters,
   and Pool Repository candidates;
3. launcher, Direct Launcher, Runtime construction, CodeVenv, and convergence
   consume the same resolved definition;
4. unknown Platform, filter, and Repository IDs fail before Host mutation;
5. Repository connectors and lifecycle behavior do not move into the Platform
   Registry.

Completeness and parity tests fix the five definitions, their OS-specific Host
and process data, their ordered Pool candidates, registered filter and
Repository references, and unknown-ID rejection. Launcher, Direct Launcher,
Runtime construction, CodeVenv, process management, and convergence resolve the
same Registry definition.

The current Windows implementation applies same-name root filtering to every
Built-in Platform. The first slice should therefore declare `same-name-root`
for all five Windows definitions. Restricting it to Cursor requires separate
process observation and is a later behavior change.

The remaining Future begins with Workspace definition loading and schema-level
diagnostics. It does not require another Built-in data migration.

## Workspace definition shape

A later slice may load definitions from:

```text
CTK_HOME/.config/platform/*.yaml
```

For example:

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

This example is a candidate representation until the implementation validates
the exact schema.

## Host path forms

A Host root uses one of two forms:

- known `base` plus relative `path`;
- omitted `base` plus an OS-native absolute `path`.

Known bases initially include `home`, `application-support`, and
`roaming-application-data`. Unknown bases, expansion expressions, `~`, and
environment-variable interpolation are rejected. A path with `base` must be
relative; a path without `base` must be absolute.

The managed `user` path remains relative to the User data root.

## Process selection

CTK composes Platform process selection from:

```text
Platform process identity
        ↓
CTK default filters
        ↓
registered additional filter IDs
        ↓
CTK Runtime ownership, stop, and wait
```

Default filters remain OS and VS Code ecosystem behavior. Platform definitions
may reference registered additional filters such as `same-name-root`.

The first implementation accepts Built-in Filter Registry IDs only. It does not
discover or execute scripts, interpreters, regular expressions, or external
providers. If that need is later observed, provider location, manifest,
runtime, trust, timeout, versioned I/O, diagnostics, and failure behavior must
be designed together as an independent Future.

## Pool and Repository separation

Pool policy is an ordered array inside the Platform definition. It is not a
separately named `PoolPolicyDefinition`.

```text
Platform pool-policy[]
    Repository ID + download capability
              ↓
Repository Registry
    exact acquisition and validation connector
```

Resolution proceeds in two stages:

1. inspect every Repository-scoped local Pool candidate in declared order;
2. only after a complete cache miss, and only when the Recipe requests
   `extension-pool: refresh`, try candidates with `download: true` in the same
   order.

The Platform flag is capability; `refresh` is the request for the current
operation. `reuse | refresh` therefore remains Recipe strategy rather than
Platform behavior.

At definition registration:

- every Repository ID must resolve through the Repository Registry;
- duplicate IDs in one Platform are rejected;
- `download: true` requires an exact VSIX acquisition connector;
- `download: false` may use the ID only as a local Pool identity;
- connector endpoints, protocol, response parsing, and asset validation remain
  Repository implementation.

The initial Repository Registry contains Built-in connectors only. External
connector providers, if later needed, require their own provider design.

## Identity and command collision

Platform identity and command are distinct. Command duplication is permitted,
but CTK does not infer a Platform from a command.

An external definition cannot replace a Built-in identity. A different command
or integration uses a new complete identity such as `cursor-code`; there is no
partial command binding or Recipe command override.

When two identities resolve the same Host paths, CTK does not infer that they
represent the same application. Existing CodeVenv health checks still reject
an Activate when the resolved Host path is already a managed link or Junction
without the matching Selection evidence. Compatibility beyond observable
filesystem and managed evidence belongs to the definition author.

## Workspace ownership

External Platform definitions are Workspace-local integration configuration,
not Recipe Source, Kitchen Notes, or OS user-global configuration.

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

An optional Workspace config may relocate only the static Cookbook Source and
Dist root initially:

```yaml
paths:
  cookbook-source: /path/to/cookbook
  dist: /path/to/dist
```

`cookbook-source` contains `recipe/` and `ingredient/`. Workbench `draft/` and
`inspect/` remain under `CTK_HOME/cookbook` so generated and review state does
not enter the versioned Source repository accidentally.

Archive and Pool remain under `CTK_HOME` until a concrete sharing requirement
is observed. There is no `ctk init` requirement: absent configuration resolves
the current default directories. CTK does not persist `CTK_HOME`, track an old
Dist root, or migrate managed state after a location change.

Package managers distribute the binary and compiled Built-in definitions.
They do not own or update Workspace configuration, Cookbook, Dist, Archive, or
Pool state.

## Support boundary

`Supported` is reserved for Built-in Platforms incorporated and validated by
CTK. A Workspace definition is available through a user extension point but is
not automatically or self-declared as CTK Supported.

The definition author is responsible for application compatibility, command,
Host paths, process declarations, Repository policy, and validation on the
intended OS and product versions. CTK remains responsible for schema
validation and for preserving its common lifecycle, recovery, and artifact
safety invariants.

Promoting an external definition to Built-in requires normal Intake,
implementation review, automated tests, and real-machine validation on every
claimed OS.

## Acceptance after refactoring

Registry implementation is followed by:

```text
automated parity and completeness tests
        ↓
targeted real-machine Intake on claimed OSes
        ↓
Go Platform Support Inventory update
```

The automated pass covers complete Built-in resolution, existing process and
Pool behavior, early rejection, lifecycle and recovery tests, Archive and
Runtime I/O tests, and Windows cross-build.

The real-machine pass re-observes command and application identity, redirected
Runtime paths, default and Dist process selection, Profiles, Extension and Pool
behavior, Build, Apply, Archive, CodeVenv lifecycle, Host restoration, and the
retained interruption scenarios. Unexecuted cells remain `Partial` or
`Not recorded`; implementation similarity does not fill evidence gaps.

The reduced
[Platform Validation Cookbook](../../test/platform-validation/README.md) is
available as public scenario input. It remains separate from Resolver unit-test
fixtures and from a user's daily Cookbook.

The macOS real-machine pass is complete for all five Built-in Platforms:
Build, Apply, Archive, Profile persistence, primary Repository Pool acquisition,
Activate, Deactivate, and Host restoration were observed. Windows validation is
the remaining claimed-OS continuation. Its executable handoff belongs to the
[Go Platform Support Inventory](../../go/doc/platform-support.md), not to a
local working memo. External definition loading remains after that parity pass;
it is not part of the Registry refactoring branch acceptance.

## Deferred independently

The following are not prerequisites for the Registry refactoring or initial
external definition loading:

- external process filter providers;
- external Repository connector providers;
- automatic Dist relocation or migration;
- multiple active Workspace selection;
- Archive or Pool location overrides;
- non-VS Code ecosystem Runtime adapters.

## Related knowledge

- [Why Platform Definitions Belong to Workspace Integration](../design-note/design-note.platform-definition-scope.md)
- [Why CTK Keeps Platform Inside the VS Code Ecosystem](../design-note/design-note.vscode-ecosystem-scope.md)
- [VS Code Ecosystem Platform Intake](../project-knowledge/note/note.vscode-ecosystem-platform-intake.md)
- [Platform Differences as Boundary Evidence](../note/note.platform-boundary-evidence.md)
- [Platform Runtime I/O Contract](../contract/contract.platform-runtime-io.md)
- [Go Platform Support Inventory](../../go/doc/platform-support.md)
