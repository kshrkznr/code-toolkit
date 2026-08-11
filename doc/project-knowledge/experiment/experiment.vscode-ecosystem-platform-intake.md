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

The resulting Pool repository order is deliberately explicit:

```text
code          → visual-studio-marketplace
kiro          → open-vsx → visual-studio-marketplace
cursor        → cursor-marketplace → visual-studio-marketplace
devin-desktop → windsurf-marketplace
```

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

The macOS CTK lifecycle was then exercised with a disposable Home and
Workspace. Recipe View, Default-only Build, Apply, Windsurf Marketplace Pool
refresh, Archive, activation with `origin.devin-desktop`, isolated launch,
`use`, Freeze Draft, and normal deactivation completed without unresolved or
failed operations. Deactivation restored physical Host directories, removed
`current.devin-desktop`, and left no transaction journal or backup. A separate
Build created a named Profile, observed its persisted storage location, and
finished with no remaining Devin process.

The current product metadata represents Windows user data as `Devin`, the
Extension data folder as `.devin`, and the desktop executable as `Devin.exe`.
Those values support a provisional Windows adapter representation, but do not
establish Windows lifecycle support. Activation, interruption recovery,
Runtime-only stopping, Build/Apply, and deactivation must still be exercised
on a real Windows host before that support is claimed.

## Current implementation boundary

The Go implementation now recognizes Cursor and Devin Desktop host paths, root
process identities, Platform Gallery queries, exact VSIX assets, and Pool
repository order alongside VS Code and Kiro. The reusable VS Code Runtime
adapter is still the intended capability boundary.

This does not claim that every VS Code fork belongs in the supported Platform
list.

Those claims should follow artifacts from the intake sequence rather than the
similarity of the products' interfaces.
