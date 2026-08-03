# Go.contract.runtime-convergence.md
============================================================

# Go Runtime Convergence Contract

This Contract specializes the shared
[Runtime Convergence Contract](../../../doc/contract/contract.runtime-convergence.md).

## Build publication

Go constructs a new Distribution in staging and publishes it only after
construction and required trusted Lock validation succeed. Failure does not
replace an existing completed Distribution.

When the target name exists, interactive Build offers the next numeric suffix
or Abort. Non-interactive callers default to `suffix` and may request `abort`.
Failed staging is removed unless `--keep-staging` retains and reports it for
diagnosis.

## Extension identity

Go compares Extension IDs exactly and case-sensitively. If desired and
installed IDs differ only by letter case, convergence reports a conflict and
does not install or uninstall either form. This prevents a case-insensitive
Platform from turning apparent replacement into accidental removal.

## Trusted Lock representation

Go trusts a Lock only when its supported structured Manifest is present and
capability-aware validation succeeds. A Distribution containing only legacy
line-oriented Lock files is treated as unlocked until Go observes it.

Go may emit line-oriented Settings and Extension observations beside the
Manifest for inspection or reference interoperability. They do not replace the
trusted Manifest.

Go stages a refreshed Lock and publishes it only after validation. It may retry
transient observation failures a bounded number of times. A refresh failure
preserves the previous complete Lock.

Go also records discovered Profiles beyond the Recipe-selected completeness
boundary. Reconstruction and trust decisions remain based on the scopes
required by the applicable Recipe.

## Internal Runtime Recovery

Go Recovery is an internal CodeVenv primitive, not a public `apply --lock`
surface. It:

- accepts a valid trusted Go Lock and its temporary Recipe
- restores only Recipe-selected Profiles
- restores Settings, Artifact ownership, and Profile inheritance
- restores Extensions by exact ID without requiring the observed version
- starts from an empty Extension area
- never copies installed Extension directories or Platform inventory metadata
- observes and publishes a fresh Lock after recovery

Observed Extension versions are diagnostics and Pool lookup hints. Archive owns
exact-version reconstruction.

## Semantic verification and Force

Recovery compares the source observation, temporary Recipe, and recovered
observation. The report covers Profile sets, Settings and managed Runtime
Artifacts, Profile inheritance, and exact Extension IDs. Ordinary Recovery does
not compare Extension versions or observation timestamps.

Matching verification continues automatically. Interactive mismatch offers
Abort or Force; non-interactive mismatch requires explicit `--force`. The
report is retained as diagnostic evidence.

Force does not bypass malformed provenance, unreadable trusted state, staging
failure, rejected Settings or Profile mutations, ambiguous recovery, Lock
failure, or Archive completeness. For unavailable Extension acquisition only,
an explicit Force policy may retain the operation as unresolved and publish the
partially converged Runtime after fresh observation. It never reports the
Extension as completed.

## Extension Pool publication

Pool maintenance is not a Lock success condition. Go downloads into
repository-local staging, validates a VSIX ZIP and its Extension Manifest,
publishes the new artifact, and only then removes an older artifact for the same
Extension and Repository. An incomplete or rejected download never replaces a
usable Pool artifact.
