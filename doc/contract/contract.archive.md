# Knowledge.contract.archive.md
============================================================

# Archive Contract

Archive preserves one completed CTK Runtime as an immutable,
self-contained package that can reconstruct the same applicable Runtime
without its original Cookbook or an external Extension source.

## Capability boundary

An Archive combines:

```text
trusted Lock
+ Lock Recipe or equivalent provenance
+ exact external Runtime artifacts
+ integrity and provenance metadata
= offline-reconstructable Runtime
```

Archive is not a Recipe, editable Workbench, Runtime directory backup, or
packaged IDE application. Freeze evolves Cookbook Source; Archive preserves one
completed state.

The compatible Platform executable and its host prerequisites may remain host
dependencies. The Archive must make those dependencies distinguishable from
the Runtime content it claims to preserve.

## Source Distribution preconditions

Archive creation accepts only a Distribution that provides the Recipe
provenance and Runtime observation capabilities needed for a complete trusted
Lock. A metadata-only or launch-only input is not archivable.

These preconditions are validated before Lock observation or Archive
publication changes durable Archive state.

## Trusted Lock

- Archive creation follows the source Recipe's declared Lock mode.
- The selected trusted Lock is the Archive source of truth.
- Every observation required for reconstruction is complete.
- Every external artifact required for reconstruction is resolved exactly.
- Missing, invalid, or mismatched required content prevents publication.

Ordinary Build or Apply may report unavailable Extension acquisition as
unresolved. Archive must not publish an incomplete package.

## Exact artifact snapshot

Archive preserves every distinct artifact identity required by the observed
default Runtime and selected Profiles. If different scopes observe different
versions of the same Extension, the Archive preserves those versions rather
than collapsing them.

Reconstruction applies the exact archived artifacts through the Platform
boundary, observes the resulting scopes, and verifies the identity and version
claimed by the Archive. A target Platform that cannot reproduce the archived
combination does not produce a successful reconstruction.

## Integrity

An Archive records enough version, provenance, inventory, and content-integrity
information to validate all content before Runtime mutation.

Validation includes:

- Archive format compatibility
- source and Recipe provenance
- trusted Lock completeness
- required file presence and type
- cryptographic content hashes
- external package identity and version

Archive content is immutable after publication. Symlink substitution or other
indirection is not accepted where it would weaken the claimed content
integrity.

## Reconstruction through Build and Apply

Build and Apply may accept Archive input alongside Recipe input when the input
kind is explicit or can be distinguished without precedence guessing.

Archive reconstruction:

- uses the Archive's Lock, provenance, and artifacts as its complete source
- does not fall back to Marketplace, Pool, or current Cookbook content
- uses the ordinary staged Build publication and retryable Apply failure model
- publishes a fresh trusted Lock after reconstruction
- verifies the reconstructed Runtime against the archived observation

A verification mismatch is not successful reconstruction.

## Failure and immutability

- Acquisition or validation failure publishes no Archive.
- Replacement preserves the previous Archive unless the new Archive is fully
  validated and published.
- Build failure does not publish a partial Distribution.
- Apply may leave completed Platform operations in place, reports failure, and
  remains retryable.
- Freeze Commit and Cookbook operations do not edit a published Archive.
- Archive is not an Inspect source; reconstruct it as a Distribution before
  Runtime observation.

## Implementation-specific resolution

The primary implementation's Archive directory, Manifest, SHA-256 inventory,
collision behavior, Launch Override provenance, and exact reconstruction rules
are defined by the
[Go Archive Contract](../../go/doc/contract/contract.archive.md).
