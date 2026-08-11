# VS Code Ecosystem Platform Intake

This experiment asks how CTK can add a VS Code-family application without
assuming that every fork behaves exactly like Visual Studio Code.

Cursor and Kiro are the first useful comparison. They expose familiar CLI and
Runtime shapes, but each has already shown product-specific behavior at a
different boundary. The goal is to retain a repeatable investigation path,
not to turn the current observations into a universal compatibility claim.

## Candidate intake sequence

Observe a new Platform in layers:

1. Record the installed application, Platform command, version, architecture,
   app or executable identity, default User data path, and Extension path.
2. Use temporary `--user-data-dir` and `--extensions-dir` paths to test CLI
   isolation before changing the real host environment.
3. Exercise default Settings and Extension observation, then exact Extension
   install and uninstall behavior against the Platform's actual marketplace.
4. Create a named Profile and inspect its storage identity, local artifacts,
   inheritance representation, and Extension scope.
5. Observe root and helper process command lines for both the default Runtime
   and an isolated Runtime. Process stopping must select the intended root and
   must not infer ownership from the application name alone.
6. Only after those observations, exercise `ctk activate`, `freeze`, `build`,
   `apply`, `use`, `launch`, and `deactivate` with a disposable Cookbook and a
   recoverable host environment.
7. Repeat the OS-specific observations on every OS claimed as supported.

This sequence separates four questions that are easy to conflate: whether the
command exists, whether Runtime paths can be redirected, whether content
semantics are compatible, and whether CTK can safely manage the host lifecycle.

## Cursor observation: macOS 3.15.6

Observed on Apple Silicon after installing the official Cursor build through
Homebrew Cask:

- The Platform command is `cursor` and resolves into the application bundle.
- The default User data root is `~/Library/Application Support/Cursor`; CTK's
  managed User boundary is its `User` directory.
- The default Extension root is `~/.cursor/extensions`.
- The root process is `/Applications/Cursor.app/Contents/MacOS/Cursor`.
  Helpers use `Cursor Helper` identities and `--type=...` arguments.
- `--user-data-dir`, `--extensions-dir`, `--list-extensions`,
  `--show-versions`, and `--profile` are present in the desktop CLI.
- An isolated Extension listing created only temporary Runtime metadata such as
  `CachedProfilesData`, logs, `machineid`, and `extensions.json` beneath the
  supplied paths.

The first-run application opened a Cursor login or signup surface. Before that
onboarding was completed, invoking `--profile CTK-Observation` returned
success but did not create the Profile, and a subsequent Profile-scoped
Extension listing reported `Profile 'CTK-Observation' not found`. Named Profile
compatibility was therefore not established by this first-run observation; CTK
must not reinterpret a successful process exit as proof that the requested
Profile now exists. A later post-onboarding Build established the settling
handshake described below.

The downloaded 3.15.6 application carried a stapled notarization ticket and
the expected Cursor Team identifier, but local strict code-signature and
Gatekeeper assessment did not complete normally: strict verification reported
an invalid signature and Gatekeeper returned an internal error. The same build
installed and launched through Homebrew Cask, while strict verification still
reported the failure. This is retained as a distribution observation, not
treated as a CTK Runtime behavior or as proof of package integrity.

### Activation and Extension repository

A Default-only `ctk activate cursor` completed successfully in the author's
macOS Workspace. It created `origin.cursor`, selected it through
`current.cursor`, redirected the Host User and Extension paths, and reproduced
the observed Setting and two Anysphere Extensions without leaving an unfinished
transaction journal.

Cursor's bundled product metadata names a Cursor-owned Gallery at
`marketplace.cursorapi.com`. Cursor describes this Gallery as Open VSX-based,
but it is not equivalent to direct Open VSX access: it contains
Anysphere-specific Extensions and may synchronize, filter, or lag upstream
versions. The Platform boundary is therefore the Cursor Gallery, not its
upstream implementation.

The observed Gallery exposes a query API that accepts a full Extension ID and
returns versioned assets. Exact VSIX acquisition selects the requested version's
`Microsoft.VisualStudio.Services.VSIXPackage` asset. The Visual Studio
Marketplace-style fixed `vspackage` path returned `404` and is not used for
Cursor.

The resulting local Pool lookup order is deliberately explicit:

```text
code          → visual-studio-marketplace
codium        → open-vsx → visual-studio-marketplace
kiro          → open-vsx → visual-studio-marketplace
cursor        → cursor-marketplace → visual-studio-marketplace
devin-desktop → windsurf-marketplace → visual-studio-marketplace
```

VSCodium and Kiro use this same order for both local Pool lookup and CTK-owned
acquisition: Open VSX first, then Visual Studio Marketplace.

Cursor does not fall back directly to Open VSX. Doing so would bypass the
Platform's own compatibility and distribution controls while adding another
source whose relationship to the Cursor mirror may vary over time.

### Profile creation needs a settling handshake

Cursor 3.15.6 accepts the VS Code-compatible `--profile` argument, but the CLI
may return before the new Profile has been persisted to
`User/globalStorage/storage.json`. Stopping Cursor immediately can therefore
lose alternating Profile creations in a multi-Profile Build.

The earlier Bash implementation already encoded the operational workaround:

1. launch the Platform with `--profile <name>`;
2. wait for the Profile record to appear;
3. stop the Platform process completely;
4. run `--list-extensions` once against the default scope to settle the CLI.

The Go adapter must preserve this lifecycle ordering. Waiting only after the
process has been stopped is too late because the state may already have been
discarded. With the pre-stop persistence wait restored, an isolated Cursor
Build created `core`, `inspect`, `review`, and `struct` and completed all 88
convergence operations without failures.

After this correction, the author's macOS environment completed the intended
Cursor lifecycle from activation and Freeze Draft through Ingredient and Recipe
review, Build, Apply, and normal use. This establishes the macOS observation;
it does not substitute for the Windows pass.

## Cursor observation: Windows 3.15.6

Observed on Windows x64 after installing the per-user build and completing
first-run onboarding:

- the default managed paths are `%APPDATA%\Cursor\User` and
  `%USERPROFILE%\.cursor\extensions`;
- the desktop CLI accepts `--user-data-dir`, `--extensions-dir`,
  `--list-extensions`, `--show-versions`, and `--profile`;
- Extension listing uses CRLF output, which the reusable adapter normalizes;
- temporary User data and Extension roots isolate Extension observation from
  the default Host paths;
- Cursor uses `Cursor.exe` not only for its desktop root and `--type=...`
  helpers, but also for workers such as the bundled JSON language server and
  Git worker. Those workers do not necessarily carry a `--type` argument.

The desktop root is distinguished by process ancestry rather than by treating
every `Cursor.exe` without `--type` as a root. Stopping only the top-level
desktop root caused its helpers and workers to exit without separately forcing
them to stop. Windows process stopping therefore selects a matching process
whose parent is not another same-name Platform process, then waits for that
root to exit.

The Windows environment also completed activation, Freeze Draft, Inspect,
Archive, Build with four named Profiles, Apply from Archive, `use`, `launch`,
normal use, and deactivation with origin recovery. Runtime-only stopping during
Build left the separately running default Host intact, while Selection stopping
closed the default root and allowed all descendants to exit.

Interrupted-transaction recovery was exercised at both
`host-backups-planned` and `host-backed-up`. In the latter case, the Host User
and Extension directories had already moved to transaction backups and the
current Selection had been created before the activation process was forced to
exit. The next lifecycle invocation restored the physical Host directories and
the prior origin, removed the partial Selection and journal, left no transaction
backups, and reproduced the pre-interruption Setting hash and Extension list.

## Earlier Kiro observations

Kiro established several parts of this intake path:

- its Platform command, User data, Extension root, and process identity differ
  from VS Code even though the Runtime model is similar;
- Windows process selection must exclude `--type=` helpers and select the
  actual `Kiro.exe` root;
- CLI output may cross OS boundaries with CRLF and must be normalized at the
  adapter boundary;
- marketplace installation can depend on Platform networking behavior and
  should be observed as an operation, not inferred from Extension IDs alone.

Cursor adds a complementary warning: shared flags do not guarantee that
first-run or named Profile lifecycle behavior is equivalent.

## VSCodium observation: Windows 1.126.04524

Observed on Windows x64 after installing the current per-user package through
winget:

- the package ID is `VSCodium.VSCodium`, and the Platform command is `codium`;
- the desktop root is `VSCodium.exe`;
- the managed Host paths are `%APPDATA%\VSCodium\User` and
  `%USERPROFILE%\.vscode-oss\extensions`;
- the CLI reports version 1.126.04524 and accepts redirected User data and
  Extension paths plus Extension list, install, and uninstall operations;
- VSCodium's normal Gallery installed `naterkane.gremlins@0.26.1` from Open
  VSX into an isolated Distribution.

The Windows lifecycle completed Recipe Build, activation as `origin.codium`,
selection of the built Distribution through `use`, isolated launch, and normal
deactivation. Activation redirected both Host paths through junctions to
`current.codium`. Deactivation restored physical Host directories, preserved
the original Gremlins version, and left no selected VSCodium Runtime process.

In the observed TLS-inspecting network, Open VSX installation failed with
`unable to verify the first certificate` until CTK was launched with
`NODE_OPTIONS=--use-system-ca`. This remains a caller environment requirement,
not a VSCodium adapter default.

VSCodium uses Open VSX as its normal Gallery. CTK uses the same Extension
resolution policy as Kiro for Pool artifacts: `open-vsx` first, followed by
`visual-studio-marketplace` when the exact artifact is unavailable there.

## VSCodium observation: macOS 1.126.04524

Observed on Apple Silicon after installing the current Homebrew Cask:

- the Platform command is `codium`, and the application executable is
  `/Applications/VSCodium.app/Contents/MacOS/VSCodium`;
- Helpers use `VSCodium Helper` identities and `--type=...` arguments;
- the managed Host paths are `~/Library/Application Support/VSCodium/User`
  and `~/.vscode-oss/extensions`;
- the CLI accepts redirected User data and Extension paths, Extension list,
  install, and uninstall operations, and named Profiles;
- the bundled product metadata points its Gallery at Open VSX;
- `naterkane.gremlins@0.26.1` installed from Open VSX and uninstalled again in
  an isolated Runtime, and the `CTK-Observation` Profile persisted in the
  redirected User data.

A disposable macOS Recipe completed Build with no unresolved or failed
operations, activation as `origin.codium`, selection of the built Distribution
through `use`, and isolated launch. Normal deactivation stopped the running
VSCodium root and its Helpers, removed `current.codium`, and restored the Host
User and Extension paths as physical directories without a remaining Runtime
process or transaction artifact.

## Devin Desktop observation: macOS 3.7.16

[Windsurf was renamed to Devin Desktop in June 2026](https://devin.ai/blog/windsurf-is-now-devin-desktop).
The current application preserves backwards compatibility with Windsurf, but
its executable and new storage identity changed. CTK follows the current
Platform command rather than introducing a legacy command that is not present
in a fresh installation.

Observed on Apple Silicon from the current official Homebrew Cask:

- the Platform command is `devin-desktop` and reports Devin Desktop 3.7.16
  with editor 1.126.0;
- the root process is `/Applications/Devin.app/Contents/MacOS/Devin`, while
  helpers use `Devin Helper` identities and `--type=...` arguments;
- current Host paths are `~/Library/Application Support/Devin/User` and
  `~/.devin/extensions`;
- installed product metadata names `Windsurf` and `.windsurf` as the old data
  identities, and retains `com.exafunction.windsurf` as the bundle identifier;
- `--user-data-dir`, `--extensions-dir`, `--list-extensions`,
  `--show-versions`, `--install-extension`, `--uninstall-extension`, and
  `--profile` are present in the desktop CLI;
- a named Profile launched in temporary Runtime paths was persisted to
  `User/globalStorage/storage.json`, and stopping the root allowed its helpers
  to exit;
- the product-owned Gallery is
  `https://marketplace.windsurf.com/vscode/gallery`.

An isolated install of `editorconfig.editorconfig` resolved through the
Windsurf Gallery, installed version 0.18.2, appeared in the versioned
Extension list, and uninstalled normally. The Gallery query API returned the
exact `Microsoft.VisualStudio.Services.VSIXPackage` asset from its selected
Open VSX source. The useful boundary is still the Windsurf Gallery response:
direct Open VSX lookup would bypass the Platform's selection and compatibility
controls.

Two Extensions absent from the Windsurf Gallery, `chrmarti.regex@0.6.0` and
`nhoizey.gremlins@0.26.0`, were already available as exact artifacts in the
Visual Studio Marketplace Pool. Devin accepted both local VSIX files, reported
their exact IDs and versions, and created the resulting Extension directories.
CTK therefore uses Visual Studio Marketplace as a secondary Pool source after
the normal Platform install fails. It does not pass the Visual Studio
Marketplace identity to Devin as another online Gallery.

The macOS CTK lifecycle was then exercised with a disposable Home and
Workspace. Recipe View, Default-only Build, Apply, Windsurf Marketplace Pool
refresh, Archive, activation with `origin.devin-desktop`, isolated launch,
`use`, Freeze Draft, and normal deactivation completed without unresolved or
failed operations. Deactivation restored physical Host directories, removed
`current.devin-desktop`, and left no transaction journal or backup. A separate
Build created a named Profile, observed its persisted storage location, and
finished with no remaining Devin process.

## Devin Desktop observation: Windows 3.7.16

Observed on Windows x64 from the current per-user winget package:

- the package ID is `CognitionAI.DevinDesktop`, and the installed Platform CLI
  is `%LOCALAPPDATA%\Programs\Devin\bin\devin-desktop.cmd`;
- the default managed paths are `%APPDATA%\Devin\User` and
  `%USERPROFILE%\.devin\extensions`;
- the desktop root is an argument-free `Devin.exe`; helpers use the same
  executable with `--type=...`, while the bundled agent has a distinct
  lower-case `resources\app\extensions\windsurf\devin\bin\devin.exe` path;
- the CLI accepts `--user-data-dir`, `--extensions-dir`, Extension management,
  and named Profile arguments;
- product metadata retains the `Exafunction.Windsurf` application identity and
  uses `devin-desktop` as the current application name.

The Windows CTK lifecycle completed activation, Freeze Draft, Distribution
View, Build with a named `core` Profile, Archive, Apply from Archive, `use`,
isolated launch, and normal deactivation. Activation redirected the Host User
and Extension paths with junctions through `current.devin-desktop`. Launch
passed the selected Distribution's `.data` and `.ext` paths to the desktop
root, and its helpers inherited the isolated User data path. Build completed
all convergence operations without unresolved or failed work, and Apply from
Archive did the same.

Deactivation restored physical Host directories, removed
`current.devin-desktop`, and left no Devin process, transaction journal, or
backup. The original Extension ID and version were restored. The original
Setting value was also restored, although JSON serialization normalized its
whitespace and therefore changed its byte hash.

In the observed TLS-inspecting network environment, the Devin CLI failed
Marketplace installation with `unable to verify the first certificate` unless
started with `NODE_OPTIONS=--use-system-ca`. With that process-local option,
the same isolated Extension install and CTK lifecycle succeeded. This records
the environment boundary rather than making the option part of the Platform
adapter.

Runtime-only stopping was exercised by leaving `origin.devin-desktop` running
while a separate suffixed Distribution was built. The original desktop root
kept the same PID and command line through all 16 successful Build operations;
only processes associated with the Build Runtime were eligible to stop.

Interrupted activation was exercised at both `host-backups-planned` and
`host-backed-up`. The deeper interruption left the journal after both physical
Host directories had moved to their transaction backups and before the current
Selection was created. The next lifecycle invocation removed the journal and
transaction backups and restored physical Host directories, the pre-test
Setting byte hash, and `naterkane.gremlins@0.26.1`, with no remaining Devin
process or partial Selection. Its attempt to continue into a new activation
reported Runtime convergence incomplete once; a separate subsequent activation
and normal deactivation completed successfully. Recovery integrity is therefore
established, while the transient post-recovery convergence remains an observed
retry boundary rather than being hidden as an uninterrupted success.

## Current implementation boundary

The Go implementation now recognizes VSCodium, Cursor, and Devin Desktop host
paths, root process identities, Platform Gallery policy, exact VSIX assets,
and Pool repository order alongside VS Code and Kiro. The reusable VS Code
Runtime adapter is still the intended capability boundary.

This does not claim that every VS Code fork belongs in the supported Platform
list.

Those claims should follow artifacts from the intake sequence rather than the
similarity of the products' interfaces.
