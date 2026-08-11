# CTK Go

This directory contains the primary native Go implementation of CTK.

The Windows release is a native executable. Using the primary Go
implementation on Windows does not require Git Bash.

This README is the current entry point for Go behavior and implementation
decisions. Historical observations from the Bash-to-Go migration are secondary
context in the Project Knowledge
[Responsibility-First Implementation Migration](../doc/project-knowledge/experiment/experiment.responsibility-first-implementation-migration.md)
Experiment. The preserved Bash implementation remains available under `bash/`
as a historical and behavioral reference.

## Implementation principles

- Write idiomatic Go rather than translating Bash functions and pipelines.
- Model domain responsibilities explicitly with Go types and packages.
- Keep filesystem and Platform integration at clear boundaries.
- Prefer deterministic behavior and explicit errors.
- Introduce interfaces only for a concrete second implementation or useful
  test boundary.
- Add tests against Contracts and observable behavior.
- Preserve unknown source data when a read-modify-write workflow requires it.
- Avoid host mutation until its recovery boundary is documented and tested.

## Applied Kitchen Notes

The Go implementation currently adopts the Merge Rules Kitchen Note for
Settings composition:

- Settings Resources enter the merge stream in Cookbook Core resolution order.
- Objects merge recursively.
- Arrays, scalars, and `null` use the later value.
- Array paths absent from Merge Rules use later-value `replace`.
- Cookbook-wide exact paths may use deterministic, canonical-value-deduplicated
  `union` through `cookbook/kitchen-notes/go.merge-rules.yaml`.
- Merge Rules are not selected per Recipe.

Notes not declared here are not part of Go Cookbook interpretation.

## Implementation Contracts

Observable behavior and persisted representations selected by the Go
implementation are indexed under
[`doc/contract`](doc/contract/README.md). Read them together with the shared
Contracts under [`../doc/contract`](../doc/contract/README.md).

The shared Contract defines the implementation-independent agreement. The Go
Contract records concrete choices such as trusted Manifest formats, Native CLI
selection, CodeVenv recovery, and Workbench Commit syntax.

Recipe Extension installation uses the Platform-owned CLI by default. Exact
VSIX acquisition into `.vsix` is separate: omitted
`config.dist-strategy.extension-pool` means `reuse`, so Go reads existing Pool
artifacts without downloading new ones. `refresh` explicitly enables Go-owned
Repository download after Runtime observation and when an Archive requires a
missing exact artifact.

## Package layout

```text
go/
├── doc/contract/          Go implementation Contracts
├── cmd/ctk/              CLI entry point
├── release.sh            Cross-platform Release artifact builder
└── internal/
    ├── archive/          Hashed snapshot packaging and validation
    ├── buildinfo/        Binary version and build provenance
    ├── cli/selector/     Native single-selection boundary
    ├── codevenv/         Runtime selection and transactions
    ├── converge/         Settings and Extension convergence reports
    ├── cookbook/         Pure Cookbook-to-Runtime-Plan resolution
    ├── directlauncher/   Generated host-facing command launchers
    ├── distribution/     Distribution discovery and metadata
    ├── flatformat/       Editable Workbench Settings representation
    ├── launcher/         Override and native launch resolution
    ├── lifecycle/        Recipe/Archive Build, Apply, and Lock orchestration
    ├── mergerules/       Go Kitchen Note merge-rule loading
    ├── platform/         Host-specific process integration
    ├── recipe/           Recipe representation and loading
    ├── recovery/         Internal trusted-Lock Runtime reconstruction
    ├── runtimeio/        Platform Runtime capability boundary
    ├── runtimelock/      Runtime observation and trusted Lock publication
    ├── settings/         JSONC parsing and VS Code merge semantics
    └── workbench/        Freeze and Inspect Artifact lifecycle
```

## Build and test

Build the repository binary:

```bash
go -C go build -o ../bin/ctk ./cmd/ctk
```

Run tests and static analysis:

```bash
go -C go test ./...
go -C go vet ./...
```

The repository entry points are:

```text
bin/ctk        Primary Go implementation
bash/bin/ctk   Bash reference implementation
bash/scripts/  Bash reference source
```

## Release artifacts

Create versioned macOS and Windows artifacts with checksums:

```bash
go/release.sh v0.2.0
```

The builder produces:

```text
release/v0.2.0/
├── ctk_v0.2.0_darwin_arm64.tar.gz
├── ctk_v0.2.0_darwin_amd64.tar.gz
├── ctk_v0.2.0_windows_amd64.zip
└── checksums.txt
```

Each platform archive contains the executable, the CTK `LICENSE`, and generated
`THIRD_PARTY_NOTICES`. Regenerate the checked-in notice inventory after
dependency changes:

```bash
go/third-party-notices.sh
```

Repository builds remain the development path. Versioned Releases are the
intended source for future Homebrew and Scoop distribution.

The module path is `github.com/kshrkznr/code-toolkit/go`, matching the public
repository and the `go/` module root. Binary Releases and the `ctk` command name
remain independent of the module path.

## Workspace discovery

CTK resolves its workspace in this order:

1. `CTK_HOME`, when explicitly configured.
2. The current directory or its ancestors.
3. The repository-local location relative to the executable.

A valid workspace contains both `cookbook/recipe` and
`cookbook/ingredient`. This allows a Homebrew or Scoop binary to operate on
Cookbook state without installing that state alongside the executable.

## Commands

```text
ctk activate [platform] [--force]
ctk build [recipe-or-archive] [--on-conflict suffix|abort] [--keep-staging] [--force]
ctk apply [recipe-or-archive] [dist] [--force]
ctk archive [dist] [--on-conflict suffix|replace|abort]
ctk lock [dist]
ctk freeze draft [dist] [--on-conflict abort|replace]
ctk freeze commit [--force]
ctk view [source] [--on-conflict abort|replace]
ctk view dist [dist] [--on-conflict abort|replace]
ctk view recipe [recipe] [--on-conflict abort|replace]
ctk view ingredient [all|layer|layer.name] [--on-conflict abort|replace]
ctk sync [left] [right] [--on-conflict abort|replace]
ctk list
ctk current [platform]
ctk deactivate [platform] [--force|--force-empty]
ctk use [dist]
ctk launch [dist] [--] [args...]
ctk workbench [draft|inspect] [viewpoint] [--editor command]
ctk select
ctk version
ctk help
```

`launch` forwards trailing file, directory, and Platform arguments without
changing them. Use an empty Distribution slot or `--` to select a Distribution
interactively while still providing launch targets:

```text
ctk launch vscode-default .
ctk launch "" .
ctk launch -- .
ctk launch vscode-default -- file.go another-directory
```

Relative launch targets remain relative to the caller's working directory.

## Open a Workbench

Open an existing Draft or Inspect Workbench in an editor:

```text
ctk workbench
ctk workbench draft
ctk workbench inspect
ctk workbench inspect dist.vscode-default --editor code
```

Omitting the kind selects between the available Draft and Inspect areas.
Omitting an Inspect viewpoint selects one generated viewpoint. The editor is
resolved from `--editor`, then `$EDITOR`, then `code` when available, with
`vim` as the final fallback. CTK opens the Workbench directory so its Summary
and typed Draft Artifacts can be reviewed together.

Invoking `ctk` without a command opens the Native command Selector. Explicit
commands remain available for automation and shell history.

## Current implementation

The Go implementation provides the complete M0-M8 lifecycle:

- Native CLI selection, launch, and Runtime switching.
- Cookbook resolution with Settings variants and Merge Rules.
- Recipe and Archive Build/Apply with trusted Locks and Extension Pool support.
- CodeVenv activation and deactivation with recovery and safety gates.
- Freeze Draft/Commit, View, and Sync Workbench operations.
- Exact-version, hashed Archive creation and offline reconstruction.
- Direct Launcher generation for macOS command files and Windows `.cmd` files.
- Primary `ctk` binary and versioned Release artifact generation.

Windows builds and generated `.cmd` files have automated coverage and
cross-build validation. Host-specific behavior continues to be refined from
real-machine feedback.

## Observed Platform behavior

VSCodium support uses the `codium` command, the `VSCodium` Host user-data
directory, and `~/.vscode-oss/extensions`. Its Extension Pool resolution
matches Kiro: Open VSX first, then Visual Studio Marketplace. The adapter and
host-specific process management have automated coverage. End-to-end
observations on Windows x64 and macOS Apple Silicon completed Build,
activation, `use`, isolated launch, and normal deactivation. Windows preserved
`naterkane.gremlins@0.26.1` across that lifecycle. macOS separately confirmed
isolated Open VSX installation and uninstallation, named Profile persistence,
symlink-based Host redirection, Runtime stopping, and physical Host
restoration. The observed Windows network required process-local
`NODE_OPTIONS=--use-system-ca` for Open VSX access.

Windsurf was renamed to Devin Desktop in June 2026. Devin Desktop 3.7.16 for
Apple Silicon exposes `devin-desktop` as its Platform command, stores Host user
data under `~/Library/Application Support/Devin/User`, and stores Extensions
under `~/.devin/extensions`. Its application root is
`/Applications/Devin.app/Contents/MacOS/Devin`; helpers retain the VS
Code-family `Devin Helper` and `--type=...` representation.

The application still carries explicit migration identity for Windsurf:
`Windsurf` and `.windsurf` are named as the old data locations, the macOS
bundle identifier remains `com.exafunction.windsurf`, and its Extension
Gallery is served from `marketplace.windsurf.com`. CTK therefore uses the
current `devin-desktop` identity for new Distributions while treating the
Windsurf names as migration evidence, not as a second supported Platform
command.

An isolated macOS observation confirmed Settings and Extension path
redirection, Extension list/install/uninstall, named Profile persistence, and
the CTK lifecycle from Recipe View and Build through Archive, activation,
launch, selection, Freeze Draft, and deactivation. Exact Pool acquisition uses
`windsurf-marketplace` first; CTK follows the Gallery's selected Open VSX asset
without treating direct Open VSX as another repository. An exact Visual Studio
Marketplace artifact may be used as a secondary local Pool candidate when the
normal Devin install reports that the Extension is unavailable.

Devin Desktop 3.7.16 for Windows x64 uses `Devin.exe`,
`%APPDATA%\Devin\User`, and `%USERPROFILE%\.devin\extensions`. The desktop
root has no `--type` argument; same-name helpers use `--type=...`, and the
bundled Devin agent also has a distinct lower-case `devin.exe` path. An
observed Windows lifecycle completed activation, Freeze Draft, Build, Archive,
Apply from Archive, `use`, isolated launch, and normal deactivation. Host paths
were redirected with junctions, the named `core` Profile persisted, and
deactivation restored physical Host directories without a remaining Devin
process, Selection, transaction journal, or backup. The restored Setting was
semantically unchanged but its JSON formatting was normalized.

Runtime-only stopping during a separate Build left the running
`origin.devin-desktop` root and its descendants intact. Forced activation
interruption at both `host-backups-planned` and `host-backed-up` was recovered
by the next lifecycle invocation, restoring the physical Host paths, original
Setting hash, and Extension version without a remaining journal, backup, or
partial Selection. After the deeper interruption, recovery completed but the
same invocation's new Runtime convergence failed once; a subsequent activation
and deactivation completed normally.

In the observed network environment, Marketplace installation required
`NODE_OPTIONS=--use-system-ca`; without it the Devin CLI reported
`unable to verify the first certificate`. This is retained as an environment
and CLI observation rather than a CTK Platform requirement.

Cursor 3.15.6 for macOS exposes a Cursor-owned Extension Gallery at
`marketplace.cursorapi.com`. Although the gallery is based on Open VSX, it is
the Platform repository boundary: it also carries Anysphere-specific builds
and may apply its own synchronization and selection behavior. CTK therefore
uses `cursor-marketplace` as Cursor's primary VSIX Pool repository and
`visual-studio-marketplace` as its fallback. It does not bypass the Cursor
Gallery through a direct Open VSX fallback.

Cursor Gallery acquisition resolves a full Extension ID through the Gallery
query API, selects the exact observed version, and downloads its
`Microsoft.VisualStudio.Services.VSIXPackage` asset. CTK validates the resulting
VSIX identity and version before treating it as an exact Pool artifact.

Cursor 3.15.6 for Windows x64 uses `%APPDATA%\Cursor\User` and
`%USERPROFILE%\.cursor\extensions` as its managed Host paths. Cursor can run
language servers and other workers through `Cursor.exe` without a `--type`
argument, so process stopping identifies the desktop root by same-name process
ancestry rather than treating every such process as a root. Stopping that root
allowed its descendants to exit, while Runtime-only stopping during Build left
the separately running default Host intact.

The observed Windows lifecycle completed activation, Freeze Draft, Inspect,
Archive, a four-Profile Build, Apply from Archive, `use`, `launch`, normal use,
and deactivation with origin recovery. Host User and Extension redirection used
junctions. Forced interruption after both backup planning and completed Host
backup was recovered by the next lifecycle invocation without a remaining
journal, partial Selection, transaction backup, or observable Host content
drift.

On Kiro for Windows, an Open VSX installation of
`emilast.logfilehighlighter` initially failed and CTK continued to its
secondary Visual Studio Marketplace Pool candidate. Kiro CLI reported
`unable to verify the first certificate`; this was a TLS certificate
validation failure rather than an Extension ID or registry identity problem.

In the observed environment, applying the VS Code-family setting
`"http.proxyStrictSSL": false` allowed Kiro CLI to install the extension from
Open VSX, but disabled strict TLS validation and did not affect the observed
Cursor Extension CLI. The preferred observed resolution is to launch CTK with
`NODE_OPTIONS=--use-system-ca`, allowing both Platform CLIs to use the Windows
trust store while retaining certificate and Extension signature verification.
CTK and its Platform subprocesses inherit this external environment setting;
the Go binary does not inject it. This observation does not currently require
an OS or Platform Extension Resource Variant.

Marketplace-facing spelling remains Cookbook input. Independently, the Go
implementation normalizes only its internal VSIX Pool filename key to lower
case.

## Future

- Optional log and Operation Report presentation modes: normal, verbose,
  quiet, and JSON.
- Release signing, notarization, and concrete Homebrew Tap/Scoop Bucket setup.
- The relationship among the CLI, `CTK_HOME`, and multiple Workspaces for
  binary distribution, including placement boundaries for Cookbook, Dist,
  Archive, and Pool. The open questions remain in
  `doc/future/future.candidates.md`.
- Configurable no-argument behavior if a concrete non-interactive use case
  requires it.
