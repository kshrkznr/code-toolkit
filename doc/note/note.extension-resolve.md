# Knowledge.note.extension-resolve.md
============================================================

# Extension Resolution

## Purpose

Extension Resolution prepares extension artifacts when a Runtime must be constructed or recovered.

It is separate from reviewing a Runtime. Inspect and Freeze Draft compare existing representations and do not resolve repositories or download artifacts.

---

## Resolution scope

Repository resolution is used by:

- Build
- Apply
- CodeVenv activation recovery

Repository resolution is not used by:

- Inspect
- Freeze Draft

Those workflows work from the Runtime, Recipe, and Lock that already exist.

---

## Pool

The Extension Pool is a shared cache of available VSIX artifacts.

It is not a version repository and it is not an Archive.

Artifacts are stored in CTK's shared `.vsix` directory and scoped by their source Repository.

The Go implementation uses the lower-case Extension ID as the Pool storage
key. Cookbook IDs, Marketplace requests, and Lock observations retain their
original spelling; only Pool filenames and lookup keys are normalized.

```text
emilast.LogFileHighlighter@2.8.0
  → open-vsx/emilast.logfilehighlighter-2.8.0.vsix
```

Go continues to read legacy mixed-case Pool filenames. New downloads use the
lower-case key, so ordinary Pool refresh gradually converges storage without
making Marketplace spelling normalization part of Cookbook Resolution.

```text
CTK_HOME/
└── .vsix/
    ├── platform-main/
    │   └── publisher.extension-1.2.3.vsix
    └── visual-studio-marketplace/
        └── publisher.extension-1.2.3.vsix
```

The Pool keeps one version of an extension for each Repository. A newly obtained version replaces the previous artifact for the same extension and Repository.

When a newer artifact is stored, the previous artifact for the same extension and Repository is removed.

The version in the filename allows the Pool to compare an observed Lock with an available artifact without introducing another metadata format.

---

## Repository order

Each Platform defines its normal extension Repository.

When resolving an extension from the Pool, CTK uses this fixed order:

```text
Pool: Platform main Repository
    ↓ miss
Pool: Visual Studio Marketplace
```

If the Platform main Repository is Visual Studio Marketplace, it is considered only once.

When the main Pool misses and Marketplace access is permitted, the Platform's
normal CLI resolves its main Repository before CTK tries an artifact found only
in the secondary Pool. This prevents a Visual Studio Marketplace VSIX from
silently overriding a Platform such as Kiro whose main Repository is Open VSX.

Other Repositories may be introduced later when a concrete compatibility or organizational requirement exists.

Visual Studio Marketplace is not a configurable fallback Repository.

It is the fixed second candidate because VS Code is treated as the reference implementation of the VS Code ecosystem.

VSCodium uses the same Extension resolution as Kiro: `open-vsx` first and
`visual-studio-marketplace` second for both local Pool lookup and CTK-owned
acquisition. When a validated exact artifact is already present in either
local candidate, `refresh` retains it as a completed Pool operation without
contacting a Repository.

---

## Platform installation policy

Recipe Build Strategy controls whether CTK may pass an Extension ID to the
Platform's normal installation operation.

```yaml
config:
  dist-strategy:
    extension-marketplace: false
```

`extension-marketplace` defaults to `true`. Despite the historical field name,
this permits a Platform-owned lookup; it does not permit CTK to download a VSIX
into the Pool.

When Platform Repository access is disabled:

```text
Pool: Platform main Repository
    ↓ miss
Pool: Visual Studio Marketplace
    ↓ miss
warning and install skip
```

No network Repository is contacted for installation. This supports intranet
and offline Distributions while still allowing reuse of artifacts already
present in the Pool. Existing extensions remain installed when their IDs are
still expected by the Cookbook.

## VSIX acquisition policy

CTK-owned VSIX download and Pool update are a separate, opt-in Build Strategy.

```yaml
config:
  dist-strategy:
    extension-pool: reuse # reuse | refresh
```

`reuse` is the default. Build, Apply, CodeVenv recovery, and Archive may use
matching local Pool artifacts but do not contact a Repository to obtain a
VSIX. Lock observes the Runtime without updating the Pool. Archive creation
fails when an exact required artifact is missing.

`refresh` permits CTK to acquire missing exact-version VSIX artifacts through
the Platform Repository order and store them in the local Pool. Build, Apply,
Lock, and recovery perform this refresh after observing the resulting Runtime.
Archive creation may download an exact artifact that is absent locally.

This policy separates artifact handling from permission. CTK validates the
artifact identity and version, but it does not determine whether an Extension
license permits use with a target Platform or redistribution of the VSIX.

---

## Version behavior

Recipe-based Runtime construction resolves extension IDs. It may use the currently available Pool artifact or retrieve a current artifact from an allowed Repository.

Locks observe extension versions, but Inspect and Freeze Draft unlock them into extension IDs for review.

Archive reconstruction is different.

```text
Archive reconstruction
  → requires the exact versioned VSIX artifact
  → uses only the artifact contained in the Archive
  → does not use Pool or Marketplace fallback
  → fails when the required artifact is unavailable
```

This keeps ordinary Recipes flexible while preserving Archive as the responsibility for exact and offline reconstruction.

---

## Current installation flow

Recipe Build and Apply preserve the Repository identity of Pool candidates.

```text
Platform main Pool hit
  → current VSIX artifact path
  → Platform CLI install

Platform main Pool miss + extension-marketplace: true
  → extension ID
  → Platform CLI main Repository install / update
  → on failure, try the secondary Pool VSIX when present

Platform main Pool miss + extension-marketplace: false
  → try the secondary Pool VSIX when present
  → otherwise warning and install skip
```

The Platform CLI remains responsible for its normal Marketplace lookup,
installation, and extension update behavior. Uninstall always uses the
extension IDs resolved from the Cookbook, independently of the Pool resolver.
When `extension-pool: refresh` is explicit, CTK uses versions observed by Lock
to update the Pool.

```text
Platform CLI
  → extension install / update
  → Lock
  → Pool exact-version lookup
  → download only when the artifact is absent
```

With `extension-pool: refresh`, Pool download is an explicit Build, Apply,
Lock, or recovery side effect. Failure to refresh the Pool is reported as
unresolved and does not invalidate a Runtime that was successfully constructed
through the Platform Repository. With the default `reuse`, no Pool download is
attempted. Failure of every permitted installation candidate remains a Recipe
convergence failure unless the user explicitly selects `--force`.

### Forced Recipe convergence

Go Recipe Build and Apply accept `--force`. Force is a lifecycle acceptance
rule, not the Platform CLI's VSIX `--force` option. After all permitted
extension candidates fail, CTK records the extension operation as unresolved
and continues Recipe convergence.

Force does not suppress Settings, Profile, Platform database, Lock, Pool I/O,
extension ID case-conflict, or Extension uninstall failures. Archive Build and
Apply do not accept it because Archive reconstruction requires exact and
complete artifacts.

---

## Troubleshooting Platform installation

An Extension appearing in a public Registry does not prove that the Platform
CLI can install it in the current environment. Read the failed or unresolved
operation before changing the Cookbook ID or adding a Variant.

```text
[unresolved] install extension publisher.extension:
Platform Repository: ... unable to verify the first certificate
```

`Platform Repository` identifies the attempt where CTK passed the Cookbook ID
to the Platform CLI. A `secondary Pool artifact` error identifies a later VSIX
fallback attempt. `--force` preserves these failures as unresolved and allows
Recipe convergence to continue; it does not repair the installation problem.

Check failures in this order:

1. **Transport and certificate trust** — TLS verification, corporate proxy,
   custom CA, authentication, and Registry reachability.
2. **Platform Repository policy** — the Platform may use Open VSX, Visual
   Studio Marketplace, a private Registry, or an organizationally restricted
   Registry.
3. **Extension identity** — confirm publisher, name, and spelling only after
   transport succeeds.
4. **Compatibility** — confirm Platform version, extension engine range,
   target architecture, and VSIX support.
5. **Pool fallback** — inspect the Repository directory and validate that the
   cached VSIX is appropriate for the target Platform.

### Observed Kiro certificate failure

On Kiro for Windows, Open VSX installation of
`emilast.logfilehighlighter` failed with `unable to verify the first
certificate`. Both lower-case and Registry display spelling failed because the
problem was TLS validation, not Extension identity. The observed HTTPS scanner
CA was trusted by Windows but absent from the Platform's bundled Node CA set.

Prefer retaining strict TLS verification and adding the Windows trust store to
the Platform CLI's Node process:

```powershell
$env:NODE_OPTIONS = '--use-system-ca'
ctk build cookbook\recipe\kiro-default.windows.yaml
```

CTK inherits its caller's environment, and the Platform CLI inherits it from
CTK; no CTK-specific CA injection is required. Scope the variable to the shell
or operation that needs it when other Node applications should remain
unaffected. This completed the observed Kiro Open VSX installation and also
the observed Cursor Gallery installation while continuing certificate and
Extension signature verification.

The earlier Kiro observation used the VS Code-family setting below and also
completed the Open VSX installation:

```json
{
  "http.proxyStrictSSL": false
}
```

The Go convergence order writes managed Settings before installing Extensions,
so this Runtime-local setting can affect Kiro CLI during the same Build or
Apply. It did not affect the observed Cursor 3.15.6 Extension CLI. More
importantly, it disables strict certificate validation, so retain it only as an
explicit environment policy when system or additional CA trust cannot be
configured.

Do not introduce an OS or Platform Extension Variant solely from a failed
Marketplace lookup until transport and Registry policy have been ruled out.
CTK does not reinterpret Platform CLI failures or rewrite Marketplace IDs.

The Windows VSCodium observation produced the same certificate failure while
installing `naterkane.gremlins` from Open VSX. Process-local
`NODE_OPTIONS=--use-system-ca` allowed the Build and the subsequent activation,
selection, launch, and deactivation lifecycle to complete without weakening
TLS verification.

### Seeding the secondary Pool through Code

A VS Code-family Platform may be able to install a VSIX even when its main
Repository does not publish that Extension. For example, an Extension may be
absent from Open VSX but present in Visual Studio Marketplace. In that case, a
successful `code` Recipe Build or Apply with `extension-pool: refresh` can
populate CTK's
`visual-studio-marketplace` Pool after Lock observation:

```text
Recipe for code
  → code installs from Visual Studio Marketplace
  → Lock observes Extension ID and version
  → Pool update stores the VSIX

Recipe for kiro or another Code-family Platform
  → main Repository lookup misses or fails
  → secondary Visual Studio Marketplace Pool candidate
  → Platform CLI attempts local VSIX installation
```

Therefore, when extension resolution fails on a Platform whose main Repository
is not Visual Studio Marketplace, building or applying an equivalent `code`
Recipe first may make a usable secondary artifact available. This can also
support a later Marketplace-disabled Build when the required artifact is
already in the Pool.

This is an artifact acquisition technique, not a compatibility guarantee.
The target Platform may still reject the VSIX because of engine constraints,
product-specific APIs, target architecture, signatures, licensing, or
organizational policy. CTK reports the local VSIX installation result rather
than assuming that every VS Code Marketplace Extension is valid for every
Code-family Platform.

---

## Archive VSIX resolution

When Archive is created, CTK reads the versioned extension Locks and resolves
their exact artifacts through the local Pool. With the default
`extension-pool: reuse`, a missing artifact is an error. `refresh` explicitly
permits Archive creation to acquire it from a Repository.

```text
Archive
  → extension Locks
  → Pool exact-version lookup
  → when refresh is explicit, download a missing artifact
  → replace the Pool artifact for that extension
  → copy exact artifact into Archive
```

The Pool source is selected by Platform. `code` uses
`visual-studio-marketplace`; Kiro and VSCodium use `open-vsx` first and
`visual-studio-marketplace` second. It is a cache for Archive creation; Archive
reconstruction remains self-contained and never falls back to the Pool.

An unavailable artifact makes Archive creation fail. A failed download is
removed from its temporary location. CTK never creates an Archive that is
missing a locked extension artifact.

---

## Future: Pool update and sources

Pool management has no dedicated command.

Pool update may eventually use sources other than direct Marketplace download,
such as an internal Repository.

Downloads should be written to a temporary file and renamed only after success so incomplete artifacts never become Pool entries.

---

## Not yet needed

The following remain intentionally outside the current model:

- explicit Pool synchronization
- Pool cleanup commands
- multiple retained versions
- Pool distribution
- internal Repository synchronization

They should be introduced only when actual operation requires them.
