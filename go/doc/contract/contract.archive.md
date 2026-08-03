# Go.contract.archive.md
============================================================

# Go Archive Contract

This Contract specializes the shared
[Archive Contract](../../../doc/contract/contract.archive.md) for the primary Go
implementation.

## Source and trusted Lock

Go Archive creation requires a complete directory Distribution containing a
readable `.meta/recipe.yaml`, `.data/`, and `.ext/`. A metadata-only or
launch-only directory is not archivable.

Archive follows the Recipe's `lock-mode`:

- `refresh` observes and publishes a fresh trusted Go Lock
- `reuse` requires the existing trusted Go Lock
- `ask` selects Refresh, Reuse, or Abort interactively

A line-oriented legacy Lock without a valid supported Manifest is not trusted
Go Archive input. Every observed Extension must include a non-empty version.

## Exact Extension snapshot

Go preserves every distinct `Extension ID@version` observed across the default
Runtime and selected Profiles. Different versions of one ID in different
scopes remain distinct.

Creation resolves exact VSIX artifacts from the Pool and may use the Platform
repository downloader to fill a miss. Missing, invalid, or mismatched artifacts
are hard errors. Reconstruction installs the archived VSIX per scope and
verifies exact ID and version through fresh Platform observation.

## Directory representation

```text
archive/<name>/
├── manifest.json
├── lock/
│   ├── manifest.json
│   ├── recipe.yaml
│   └── observed scope files
├── vsix/
│   └── <extension-id>-<version>.vsix
└── launch-override/
    ├── run.sh
    └── run.cmd
```

Only present Launch Override files are stored. Creation uses staging and
validates the complete tree before publication.

## Manifest and integrity

The Manifest records the format version, source Distribution, Recipe identity,
creation time, trusted Lock provenance, required Extension IDs and versions,
and SHA-256 hashes for required Lock, Recipe, VSIX, metadata, and stored Launch
Override files.

Validation rejects missing files, hash mismatches, malformed Locks, invalid
VSIX ZIPs, package identity or version mismatches, and symlink substitution. At
the VSIX metadata boundary only, publisher and name comparison is ASCII
case-insensitive; version comparison remains exact.

## Launch Override provenance

Archive warns and stores a source `run.sh` or `run.cmd`, but Build and Apply do
not restore or execute it. Reconstructed Distributions use their normal Go
metadata and Platform Adapter launch path. `.data` and `.ext` are validated as
capabilities and are not copied as directory backups.

## Command and collision behavior

```text
ctk archive [dist] [--on-conflict suffix|replace|abort]
```

Omitted Distribution uses the Native Selector. `suffix` selects the next name;
`replace` validates staging before replacing the old Archive; `abort` preserves
the old Archive. Interactive omission asks, while non-interactive omission
defaults to `suffix`.

## Reconstruction

Go Build and Apply accept Recipe YAML or a directory with a valid Archive
Manifest. Ambiguous input is an error. Reconstruction never falls back to the
Marketplace, normal Pool, or current Cookbook.

Apply without `dist` matches existing Distributions by Recipe `name + os +
platform`. One match is automatic, multiple interactive matches use the Native
Selector, multiple non-interactive matches are an error, and no match directs
the caller to Build. An explicit target must match the Archive identity.

After reconstruction, Go publishes a fresh trusted Lock and verifies Settings,
Profile topology and inheritance, and exact per-scope Extension IDs and
versions.
