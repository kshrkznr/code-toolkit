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

---

## Marketplace connection policy

Recipe Build Strategy controls whether Marketplace access is permitted for a Distribution.

```yaml
config:
  dist-strategy:
    extension-marketplace: false
```

When Marketplace access is disabled:

```text
Pool: Platform main Repository
    ↓ miss
Pool: Visual Studio Marketplace
    ↓ miss
warning and install skip
```

No network Repository is contacted. This supports intranet and offline Distributions while still allowing reuse of artifacts already present in the Pool. Existing extensions remain installed when their IDs are still expected by the Cookbook.

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

The Platform CLI remains responsible for its normal Marketplace lookup, installation, and extension update behavior. Uninstall always uses the extension IDs resolved from the Cookbook, independently of the Pool resolver. After Build or Apply refreshes its Lock, CTK uses the observed extension versions to update the Pool.

```text
Platform CLI
  → extension install / update
  → Lock
  → Pool exact-version lookup
  → download only when the artifact is absent
```

Pool download is a Build / Apply side effect. Failure to refresh the Pool is
reported as unresolved and does not invalidate a Runtime that was successfully
constructed through the Platform Repository. Failure of every permitted
installation candidate remains a Recipe convergence failure unless the user
explicitly selects `--force`.

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
problem was TLS validation, not Extension identity. In that environment,
applying the VS Code-family setting below before extension convergence allowed
Kiro CLI to complete the Open VSX installation:

```json
{
  "http.proxyStrictSSL": false
}
```

The Go convergence order writes managed Settings before installing Extensions,
so a Runtime-local proxy setting can affect the Platform CLI during the same
Build or Apply. Disabling strict certificate validation weakens transport
security and is not a general recommendation. Prefer installing or configuring
the required CA trust when possible, and use this setting only as an explicit
environment policy.

Do not introduce an OS or Platform Extension Variant solely from a failed
Marketplace lookup until transport and Registry policy have been ruled out.
CTK does not reinterpret Platform CLI failures or rewrite Marketplace IDs.

### Seeding the secondary Pool through Code

A VS Code-family Platform may be able to install a VSIX even when its main
Repository does not publish that Extension. For example, an Extension may be
absent from Open VSX but present in Visual Studio Marketplace. In that case, a
successful `code` Recipe Build or Apply can populate CTK's
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

## Archive VSIX download

When Archive is created, CTK reads the versioned extension Locks and resolves their artifacts through the Pool.

```text
Archive
  → extension Locks
  → Pool exact-version lookup
  → download missing artifact from Visual Studio Marketplace
  → replace the Pool artifact for that extension
  → copy exact artifact into Archive
```

The Pool source is selected by Platform. `code` uses `visual-studio-marketplace`; `kiro` uses `open-vsx` first and `visual-studio-marketplace` second. It is a cache for Archive creation; Archive reconstruction remains self-contained and never falls back to the Pool.

An unavailable artifact is removed from its temporary download location and makes Archive creation fail. CTK never creates an Archive that is missing a locked extension artifact.

---

## Future: Pool update and sources

Pool management has no dedicated command.

Pool update may eventually use sources other than direct Marketplace download, such as a Master Runtime or an internal Repository.

Downloads should be written to a temporary file and renamed only after success so incomplete artifacts never become Pool entries.

---

## Not yet needed

The following remain intentionally outside the current model:

- explicit Pool synchronization
- Pool cleanup commands
- multiple retained versions
- Pool distribution
- Master Runtime
- internal Repository synchronization

They should be introduced only when actual operation requires them.
