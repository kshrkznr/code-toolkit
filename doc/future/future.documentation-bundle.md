# Knowledge.future.documentation-bundle.md
============================================================

# Future: Packaged Documentation Bundle Follow-ups

The Go executable now has a versioned Documentation Bundle, progressive
navigation, safe Export, Release verification, and explicit local-source
comparison. Those accepted behaviors are no longer Future material.

Reusable navigation and provenance observations belong to [Packaged
Documentation Navigation](../note/note.packaged-documentation-navigation.md).
Exact Go behavior and representations belong to the [Go Documentation Bundle
Contract](../../go/doc/contract/contract.documentation-bundle.md).

This Future retains only the distribution and navigation questions that remain
unsettled after that implementation.

## Current evidence

The implemented Go path has shown that:

- a standalone executable can carry and verify version-matched Knowledge;
- people and AI assistants can move from Bootstrap or Node context through
  metadata Resolve, one-document TOC, and bounded heading Show;
- Export provides ordinary full-text investigation without expanding Resolve
  into body or semantic search;
- packaged Knowledge can remain the default while one local source is selected
  explicitly and compared across independent provenance dimensions;
- one deterministic Bundle can be reproduced from a source tag and appended to
  every Release target.

This evidence supports the current implementation. It does not by itself settle
package-manager publication, every executable signing system, or whether later
use justifies more convenience around local selection.

## Candidate: Package-manager publication

The Bundle was motivated by binary-only installation, but building the feature
does not decide when CTK should publish it through Homebrew or another package
manager.

Reconsider publication when:

- the current Documentation Bundle Contract and implementation have completed
  review;
- the intended Release version is explicit;
- native packaged-binary checks cover every claimed operating system;
- install, upgrade, and uninstall behavior can be evaluated without relying on
  a source checkout.

Publication should consume the already verified Release artifacts rather than
regenerating documentation inside a package-manager recipe.

## Candidate: Persisted local-source selection

The current Go boundary requires explicit per-invocation selection:

```text
ctk docs --source <repository> <operation>
```

A future configuration field could reduce repetition for documentation
development, but it could also make a stale or dirty clone silently replace
version-matched packaged Knowledge. Workspace configuration is especially
questionable because `docs` is intentionally useful without a Workspace.

Reconsider persistence only after repeated use shows that explicit selection is
material friction. Any proposal should preserve:

- packaged Knowledge as an obvious default and recovery path;
- visible revision, Definition, content, and selected-path comparison;
- a way to bypass or inspect the persisted choice;
- home-path masking in output likely to be shared in an Issue;
- self-description without requiring Workspace validation.

The owning configuration surface remains open. No `.config` field is currently
accepted for this purpose.

## Candidate: More precise missing-reference diagnosis

A failed Show currently reports that a reference was not found and routes the
caller to Resolve and the exact-version repository. The Bundle generator knows
which source links were intentionally repository-only, but a free-form Show
reference does not distinguish an excluded repository path from an arbitrary
typo.

Reconsider a more precise diagnostic when real navigation attempts repeatedly
confuse those cases. It should not require packaging repository-only stubs or a
complete inventory of excluded repository files unless that additional index
proves useful independently.

## Candidate: Concise human Bootstrap

The current Bootstrap deliberately carries enough Concept Domain, Concept API,
and Documentation Resolver vocabulary for AI-assisted navigation. It is larger
than a conventional terminal help page.

A concise human-oriented view may be useful if terminal feedback shows that the
current default obscures rather than helps the next action. A second view should
remain generated from the same sources and should not become another maintained
Resolver or documentation authority.

## Candidate: Signing and notarization

Appending a deterministic ZIP works for the current unsigned macOS and Windows
executables. Signing or notarization may impose ordering or integrity
constraints on bytes appended to an executable.

Before a distribution claims signed or notarized artifacts, validate the full
sequence for each target: build, append, sign, package, install, execute, and
reopen the carried Bundle. If append is incompatible with a required signing
system, the existing versioned sidecar fallback remains available without
changing the documentation model.

## Settled ownership

The following topics were explored here but are now current Go behavior:

- selected Knowledge and explicit repository-only exceptions;
- the generated Bootstrap, Bundle Definition, Manifest, and aggregate digest;
- metadata-only Resolve and exported full-text investigation;
- Node shortcuts, one-document TOC, heading Show, and relative depth ranges;
- exact-version repository routing for excluded material;
- safe directory Export;
- appended-binary transport and reproducible Release verification;
- explicit local source selection and independent mismatch or dirty status;
- Workspace-independent `help`, `version`, and `docs` dispatch in the Go CLI.

Change those behaviors through the responsible Go Contract and implementation,
not by expanding this Future back into a second specification.

## Boundary

This Future does not propose:

- a second Documentation Resolver;
- generated answers to arbitrary natural-language questions;
- body or semantic search inside Resolve;
- silent network fetching;
- silent discovery of a nearby clone;
- packaging every repository or Project Knowledge file;
- a release or package-manager publication without separate approval.

## Revisit when

Revisit one candidate when its stated evidence appears. Remove it when the
responsible Contract accepts the behavior or observation shows that the added
surface is unnecessary.

Track the broader documentation-bundle work in
[Issue #20](https://github.com/kshrkznr/code-toolkit/issues/20).
