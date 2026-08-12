# Go.platform-support.md
============================================================

# Go Platform Support Inventory

This document inventories the VS Code-family Platforms recognized by the current
Go implementation and the evidence behind their support status.

It complements the concise Supported Platforms table in the repository README.
`Implemented` means that the Go implementation contains the required Platform
differences. It does not mean that every capability was re-observed on every
product version and operating system.

Observed version numbers identify the application builds from which evidence
was collected. They are evidence snapshots, not a list of individually
supported version targets. CTK does not repeat the full intake solely because an
upstream version number changes; targeted re-observation is appropriate when a
CLI, path, process, Profile, Repository, or lifecycle boundary may have changed.

This is an implementation and validation inventory, not a shared CTK Contract or
a compatibility promise for unlisted applications.

## Snapshot

- Inventory date: 2026-08-12
- Implementation: Go primary implementation on `main`
- Host integration OSes: macOS and Windows
- Runtime Adapter: shared `internal/runtimeio/vscode` Adapter
- Built-in definitions: shared `internal/platform` Registry
- Built-in Platform commands: `code`, `codium`, `kiro`, `cursor`,
  `devin-desktop`

## Built-in Platform incorporation

Go currently incorporates all five observed Platform commands listed in
[Platform Differences as Boundary Evidence](../../doc/note/note.platform-boundary-evidence.md):
`code`, `codium`, `kiro`, `cursor`, and `devin-desktop`.

The application-owned command, Host data identity, Extension root, and primary
Repository are shared observations rather than Go-specific definitions. Go
incorporates those observations into Host integration and Runtime convergence.
Its secondary Pool candidates and exact acquisition strategies are CTK policy,
not application identity; they remain visible below and in
[Extension Resolution](../../doc/note/note.extension-resolve.md).

## Implementation incorporation matrix

| Difference | Shared or Platform-specific | Current implementation |
| --- | --- | --- |
| Runtime CLI, Settings, Profiles, Extensions, Runtime Artifacts | Shared VS Code-family behavior | `internal/runtimeio/vscode` |
| Identity, command, Host paths, process identities / filters, Pool candidates | Platform-specific declarations | `internal/platform/registry.go` |
| Host path resolution | Shared rules consuming Platform declarations | `internal/codevenv/host.go`, `internal/platform/registry.go` |
| macOS root/helper process selection | Platform identity with shared rules | `internal/platform/process_darwin.go` |
| Windows executable and desktop-root selection | Platform identity with declared ancestry filter | `internal/platform/process_windows.go` |
| Local Pool candidate order and download capability | Platform policy consumed by convergence | `internal/converge` |
| Exact Repository acquisition | Registered Repository-specific connector | `internal/repository`, `internal/converge/pool_update.go` |

The current CTK Pool order is:

```text
code          → visual-studio-marketplace
codium        → open-vsx → visual-studio-marketplace
kiro          → open-vsx → visual-studio-marketplace
cursor        → cursor-marketplace → visual-studio-marketplace
devin-desktop → windsurf-marketplace → visual-studio-marketplace
```

The first candidate follows the observed Platform-owned Repository. The second,
where present, is CTK's exact local Pool fallback and is not another online
Repository configured in the target application.

Built-in declarations now resolve through a single Registry while named Go
filters, Repository connectors, and lifecycle procedures retain behavior that
cannot be safely represented as data. Workspace definition loading remains a
separate Future and does not make a user definition `Supported`.

## Real-machine observation matrix

The table records what the retained observation reports establish for the
listed evidence snapshots. It is not a version-by-version support matrix.
`Partial` means that useful evidence exists but not at the same lifecycle
granularity as the more recent Platform passes. `Not recorded` does not mean
unsupported; it means this inventory has no equally specific retained
observation.

Some retained real-machine passes used Recipes from the author's private
Cookbook. Profile names such as `core`, `inspect`, `review`, and `struct`
identify those observed Runtime scopes; they do not refer to the similarly
named Go Resolver fixtures under `internal/cookbook/testdata`.

Future targeted passes can use the reduced, public
[Platform Validation Cookbook](../../test/platform-validation/README.md). It
contains one Runtime Ingredient, one named Profile, and one Extension available
from every built-in Platform's primary Repository rather than reproducing the
author's daily environment. Executed results are recorded below; the source's
presence alone is not real-machine evidence.

| Platform / observed version | OS | CLI and isolated paths | Named Profile | Build / Apply / Archive | CodeVenv lifecycle | Interrupted recovery |
| --- | --- | --- | --- | --- | --- | --- |
| Visual Studio Code 1.132.1–1.133.0 | macOS Apple Silicon | Observed through Host import and isolated Dist paths | Six named Profiles observed | Build, Apply, Archive, and Archive reconstruction observed | Activate, Use, Launch, Deactivate, and Host restoration observed | Not recorded |
| Visual Studio Code / baseline | Windows | Baseline implementation | Baseline implementation | Cross-build and historical use | Historical use | Automated lifecycle tests; no versioned real-machine row retained here |
| Kiro 1.0.242 | macOS Apple Silicon | Isolated Dist paths observed | One named Profile observed | Build, Apply, and Archive observed | Activate, Deactivate, and Host restoration observed | Not recorded |
| Kiro / earlier observations | Windows | CLI, CRLF, process, Repository behavior observed | Partial | Partial | Partial | Not recorded at equal granularity |
| VSCodium 1.126.04524 | macOS Apple Silicon | Observed | Observed | Build, Apply, and Archive observed | Activate, use, launch, deactivate observed | Not recorded |
| VSCodium 1.126.04524 | Windows x64 | Observed | Not recorded | Build observed; Apply and Archive not recorded | Activate, use, launch, deactivate observed | Not recorded |
| Cursor 3.15.6 | macOS Apple Silicon | Observed | Four-Profile private Build and reduced one-Profile Build observed | Build, Apply, and Archive observed | Activate, Freeze Draft, normal use observed | Not recorded |
| Cursor 3.15.6 | Windows x64 | Observed | Four Profiles observed | Build, Apply from Archive, Inspect, Archive observed | Activate, Freeze Draft, use, launch, deactivate observed | Two activation interruption phases observed |
| Devin Desktop 3.7.16 | macOS Apple Silicon | Observed | Observed separately | Build, Apply, Archive observed | Activate, Freeze Draft, use, launch, deactivate observed | Not recorded |
| Devin Desktop 3.7.16 | Windows x64 | Observed | One named Profile (`core`) observed with the author's Cookbook | Build, Apply from Archive, Archive observed | Activate, Freeze Draft, use, launch, deactivate observed | Two activation interruption phases observed; one post-recovery retry retained |

The matrix deliberately exposes uneven evidence. Filling a cell requires a new
observation artifact; implementation similarity alone does not change it.

After the Registry-refactoring Build, Apply, and Archive pass, macOS Activate
and Deactivate were repeated for all five built-in Platforms. Each Activate
imported the physical Host into its `origin.<platform>` Distribution, selected
it through `current.<platform>`, and redirected the application User and
Extension paths through managed symlinks. Each Deactivate removed that
Selection and restored physical Host directories. Visual Studio Code was then
activated again and returned to its pre-test `vscode-default` Selection; the
other four Platforms remained unselected.

## Platform-specific evidence

### Visual Studio Code

Visual Studio Code remains the primary Platform and the baseline for the shared
Runtime Adapter. The macOS Registry-refactoring pass built a six-Profile Dist
from the author's `vscode-default.macos.yaml`, applied it with 126 completed and
no unresolved or failed operations, archived it, and reconstructed a second
Dist from that Archive with 88 completed and no unresolved or failed
operations.

The selected Runtime was changed to the Recipe-built Dist while the
Archive-reconstructed Dist was launched independently with its own `.data` and
`.ext`. Deactivation removed `current.code` and restored physical Host `User`
and Extension directories. The independently launched second Dist remained
running, as expected: Deactivate owns the active CodeVenv selection, not every
isolated Dist launched separately.

After that restoration, Activate was repeated with the Registry build. It
imported the physical Host into `origin.code`, selected it through
`current.code`, and recreated the managed Host symlinks. The independently
launched second Dist remained outside Activate's process ownership.

The final reduced validation Recipe separately completed Build and Apply with
9 operations and no unresolved or failed operations, and Archive completed.
Its exact `emilast.logfilehighlighter@3.5.2` artifact was retained from Visual
Studio Marketplace.

The installed application updated from 1.132.1 to 1.133.0 during the observation
window. The versions are retained as evidence snapshots; no behavior boundary
was inferred from the update. This pass did not exercise interruption recovery.
The Windows row therefore remains at baseline granularity.

### Kiro

Kiro established several reusable boundaries before the newer matrix-shaped
observations:

- Host paths and process identity differ while the Runtime model remains
  VS Code-compatible.
- The current Go Windows process strategy excludes `--type=` helpers; the
  broader process-identity observation is recorded in
  [Platform Differences as Boundary Evidence](../../doc/note/note.platform-boundary-evidence.md).
- Platform command output may contain CRLF and is normalized at the Adapter
  boundary.
- Open VSX installation may fail independently of the Extension ID. In the
  observed TLS-inspecting environment, process-local
  `NODE_OPTIONS=--use-system-ca` allowed the Platform CLI to use the Windows
  trust store while retaining certificate and Extension signature
  verification.
- The Pool uses Open VSX first and Visual Studio Marketplace second.

The Registry-refactoring pass used the reduced public validation Cookbook on
macOS. Its final form completed Build and Apply with 9 operations and no
unresolved or failed operations, and Archive completed. This covered isolated
Runtime paths, one named Profile, one Profile Extension, and exact Open VSX Pool
acquisition. CodeVenv lifecycle and interruption recovery were not repeated in
this pass.

The initial Windows observation used `emilast.logfilehighlighter`. Both its
lower-case and Registry display spelling failed with the same certificate
error, and CTK then attempted its secondary Visual Studio Marketplace Pool
candidate. A later run with process-local `NODE_OPTIONS=--use-system-ca`
completed the Kiro Open VSX installation. The same trust-store configuration
also completed the observed Cursor Gallery installation. The earlier fallback
therefore records convergence order after a transport failure; it is not the
required resolution or evidence of an Extension identity problem. See
[Extension Resolution](../../doc/note/note.extension-resolve.md) for the
troubleshooting boundary and the less-preferred `http.proxyStrictSSL` result.

### VSCodium 1.126.04524

Windows x64 and macOS Apple Silicon observations established `codium`, the
`VSCodium` Host data identity, `.vscode-oss/extensions`, redirected Runtime
paths, Open VSX installation, and normal CodeVenv lifecycle behavior.

Windows activation and deactivation used Junction-based Host redirection and
preserved `naterkane.gremlins@0.26.1`. macOS confirmed named Profile persistence,
symlink-based redirection, Runtime stopping, and physical Host restoration.

The final reduced macOS validation Recipe completed Build and Apply with 9
operations and no unresolved or failed operations, and Archive completed. Its
exact `emilast.logfilehighlighter@2.8.0` artifact was retained from Open VSX.

The Windows network required process-local `NODE_OPTIONS=--use-system-ca` for
Open VSX access. This remains an execution-environment observation, not a
VSCodium Adapter default.

### Cursor 3.15.6

Cursor uses a Cursor-owned Gallery at `marketplace.cursorapi.com`. CTK treats
that Gallery as the primary Repository boundary and Visual Studio Marketplace
as the secondary exact-artifact source. It does not bypass Cursor's selection
through direct Open VSX lookup.

Named Profile creation exposed an asynchronous persistence boundary: the CLI
can return before `User/globalStorage/storage.json` contains the Profile. The
shared Adapter now waits for the record before stopping the Runtime, then runs a
default-scope Extension listing to settle the CLI.

Before first-run onboarding was complete, `--profile CTK-Observation` returned
success without creating the Profile. A successful process exit was therefore
not accepted as proof of Profile existence. The observed Gallery required its
query API and exact `Microsoft.VisualStudio.Services.VSIXPackage` asset; a
Visual Studio Marketplace-shaped fixed `vspackage` URL returned `404`.

An A/B pass against the pre-Registry binary reproduced a fresh Runtime failure
with the earlier two-second persistence window. Keeping the original launch and
stop sequence while extending the bounded poll to ten seconds completed the
reduced Cursor Build. The final public validation Recipe completed Build and
Apply with 9 operations and no unresolved or failed operations, and Archive
completed. Its exact `emilast.logfilehighlighter@2.8.0` artifact was retained
from Cursor Marketplace.

The observed macOS application carried the expected team identity and a stapled
notarization ticket, but local strict signature verification reported an invalid
signature and Gatekeeper returned an internal error. The same build installed
and launched through Homebrew Cask. This remains a distribution observation,
not proof about CTK Runtime behavior or package integrity.

On Windows, `Cursor.exe` is also used by workers without `--type`. Desktop-root
selection therefore uses same-name process ancestry. The retained Windows pass
also covers Junction redirection, four Profiles, Archive Apply, Runtime-only
stopping, origin restoration, and recovery from two interrupted activation
phases.

### Devin Desktop 3.7.16

The current application command and Host identity are `devin-desktop`, `Devin`,
and `.devin`, while installed metadata retains Windsurf migration identities.
CTK does not expose the legacy identity as another built-in command.

The retained macOS metadata named `Windsurf` and `.windsurf` as old data
locations and retained `com.exafunction.windsurf` as its bundle identifier. On
Windows, the desktop root was an argument-free `Devin.exe`; helpers used
`--type=...`, while the bundled agent used a distinct lower-case executable path.

The Platform Gallery is served from `marketplace.windsurf.com`. Exact Pool
acquisition follows the Windsurf Gallery first and uses a Visual Studio
Marketplace artifact as a secondary local candidate. Direct Open VSX is not
another configured Repository for this Platform.

macOS and Windows observations cover Build, Archive, Apply, named Profile work,
and the normal CodeVenv lifecycle. Windows additionally covers Runtime-only
stopping and two interrupted activation phases. Recovery integrity was
established; after the deeper interruption, the same invocation's new Runtime
convergence failed once and a later activation succeeded. The retry boundary is
retained rather than reported as uninterrupted success.

Normal Windows deactivation restored the Setting semantically, while JSON
serialization normalized its whitespace and changed its byte hash. In the
observed TLS-inspecting environment, Marketplace installation required
process-local `NODE_OPTIONS=--use-system-ca`; this remains caller environment,
not Platform definition.

The initially long validation Distribution name amplified the temporary
Runtime path enough for Devin Desktop to exit without persisting its Profile.
The same binary, Recipe content, Profile name, and ownership strategy succeeded
from a shallow Runtime path. Go no longer repeats the Distribution identity in
its internal Build staging name, while the validation procedure keeps its
disposable Workspace shallow because an arbitrarily deep `CTK_HOME` remains
outside CTK's control. The final validation Recipe completed Build and Apply
with 9 operations and no unresolved or failed operations, Archive completed,
and exact `emilast.logfilehighlighter@2.8.0` acquisition used Windsurf
Marketplace.

## Automated coverage inventory

Shared automated coverage currently includes:

- Profile launch, persistence-before-stop, verification, and CLI settling
- default and Profile Settings paths and inheritance
- CRLF-normalized, case-preserving Extension observation and CLI operations
- Runtime Artifact paths and unsupported Profile Tasks behavior
- CodeVenv activation, health, deactivation, rollback, and managed links
- Repository order and exact Cursor/Windsurf Gallery acquisition
- macOS process matching and Windows desktop-root script generation

Platform-name-specific coverage now includes one Registry completeness test for
all five built-ins. It fixes command, both OS Host paths, macOS process
identities, Windows executable and `same-name-root` filter, ordered Repository
candidates, and existing download capability. Registration rejects unknown
filter and Repository IDs and duplicate Repository candidates. Existing focused
tests continue to exercise process selection, Pool order, acquisition, and
lifecycle behavior.

Automated completeness does not fill the real-machine observation matrix. The
targeted macOS validation above provides the recorded Build, Apply, Archive,
Profile, and Repository evidence; unexecuted CodeVenv and recovery cells, and
the corresponding Windows re-observation, remain `Partial` or `Not recorded`.

## Windows continuation after Registry refactoring

The Registry implementation, automated tests, Windows amd64 cross-build, and
macOS real-machine acceptance are complete. Windows remains the claimed-OS
continuation for this refactoring. It should be performed from the branch that
contains the Registry changes; the ignored local inventory memo is not required
to interpret or execute the pass.

Use the OS-specific Recipes in the
[Platform Validation Cookbook](../../test/platform-validation/README.md) for
all five built-in identities. Keep the Workspace path shallow. For each
Platform, retain evidence for:

1. installed command, product version, and x64 architecture;
2. isolated User Data and Extension paths;
3. named Profile persistence and Profile-local Settings;
4. Build and Apply operation counts with no unresolved or failed operations;
5. Archive creation and, for at least the baseline Visual Studio Code pass,
   reconstruction from that Archive;
6. exact `emilast.logfilehighlighter` Lock and the primary Repository-scoped
   Pool artifact;
7. Activate, managed `current.<platform>` Junctions, Deactivate, and physical
   Host restoration;
8. Runtime-only process stopping without terminating an independently running
   Host/default Runtime;
9. retained interruption-recovery scenarios where practical.

Expected primary Pool scopes are Visual Studio Marketplace for `code`, Open VSX
for `codium` and `kiro`, Cursor Marketplace for `cursor`, and Windsurf
Marketplace for `devin-desktop`. A different observed Extension version is
evidence, not automatically a failure; the Lock and exact Pool artifact must
agree.

Windows process selection must continue to exclude `--type=` helpers and apply
the declared `same-name-root` filter. `current.<platform>` and Host redirections
are Junctions rather than Unix symbolic links. In the previously observed
TLS-inspecting environment, process-local `NODE_OPTIONS=--use-system-ca`
restored Repository access while preserving strict certificate and Extension
signature verification. Record a recurrence as environment evidence rather
than changing Platform Repository order.

Update the Windows rows in the real-machine observation matrix only from the
new pass. Keep unexecuted operations explicit. After the pass, rerun the Go
suite, `go vet`, Windows cross-build, `git diff --check`, and confirm the tested
binary is the branch build rather than an older installed binary.

## Updating this inventory

When a Platform is first added or a boundary is re-observed:

1. follow the
   [VS Code Ecosystem Platform Intake](../../doc/project-knowledge/note/note.vscode-ecosystem-platform-intake.md);
2. retain the exact OS, product version, and observed operations;
3. update only cells supported by observation artifacts;
4. distinguish common Adapter coverage from Platform-name-specific tests;
5. keep unverified behavior visible as `Partial` or `Not recorded`;
6. update the concise README status only when the implementation status changes.

## Related documents

- [Platform Runtime I/O Contract](../../doc/contract/contract.platform-runtime-io.md)
- [Code Environment Integration Contract](../../doc/contract/contract.code-environment.md)
- [Go Code Environment Contract](contract/contract.code-environment.md)
- [Platform Differences as Boundary Evidence](../../doc/note/note.platform-boundary-evidence.md)
- [VS Code Ecosystem Platform Intake experiment](../../doc/project-knowledge/experiment/experiment.vscode-ecosystem-platform-intake.md)
