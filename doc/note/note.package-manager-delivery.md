# Knowledge.note.package-manager-delivery.md
============================================================

# Package-manager Delivery

## Context

CTK publishes self-contained macOS and Windows archives plus `checksums.txt` in
each stable GitHub Release. Homebrew and Scoop definitions are maintained in
separate public repositories so installing the CLI does not transfer ownership
of a user's Workspace, Cookbook Source, or generated Distribution state to a
package manager.

The current channels are:

- [`kshrkznr/homebrew-tap`](https://github.com/kshrkznr/homebrew-tap), with
  `Formula/ctk.rb` for macOS arm64 and amd64, including static bash, zsh, and
  fish completion installed into Homebrew-managed directories;
- [`kshrkznr/scoop-bucket`](https://github.com/kshrkznr/scoop-bucket), with
  `bucket/ctk.json` for Windows amd64, including an install note for adding
  static CTK completion to the user-owned PowerShell profile.

## Release-to-package flow

Publishing a stable `vX.Y.Z` GitHub Release starts the Package Delivery
workflow. The workflow:

1. requires an existing published, non-prerelease Release;
2. downloads the exact macOS arm64, macOS amd64, Windows amd64, and checksum
   assets;
3. verifies every archive against the published `checksums.txt`;
4. deterministically updates only `Formula/ctk.rb` and `bucket/ctk.json`;
5. opens reviewable Pull Requests in the Tap and Bucket repositories.

Package repository checks and maintainer review remain between generation and
merge. The delivery workflow does not push a generated definition directly to
either repository's `main` branch.

Manual dispatch accepts an already published stable Release. It is
verification-only by default; Pull Request creation must be selected
explicitly. Creating package Pull Requests requires the
`PACKAGE_DELIVERY_TOKEN` repository secret to grant narrowly scoped write
access to the two package repositories.

## Failure and retry behavior

Delivery fails before opening a Pull Request when the Release is absent, is a
Draft or prerelease, has an unexpected checksum inventory, fails checksum
verification, or would change files outside the two package definitions.

The generated branch name is stable for a Release. A retry reuses an existing
matching Pull Request, accepts a package definition that is already current,
and rejects a same-named branch whose scope or generated content differs. This
makes an interrupted run inspectable without silently replacing unrelated
work.

## Verification boundary

Repository automation verifies archive integrity and package-definition
structure. Homebrew owns the installed completion files but does not edit shell
startup files. Scoop displays the PowerShell completion command but does not
edit `$PROFILE`. A channel's installation lifecycle is supported only to the
extent observed on its target OS. Installation, executable provenance,
completion activation, packaged help and documentation, Workspace-independent
initialization, upgrade, rollback, uninstall, and survival of user-owned
Workspace state are separate target device checks.

A network or VPN restriction is evidence about the test environment, not a
successful or failed package lifecycle. Record the exact stopping point and
resume on the target device when its normal package endpoints are reachable.

## Boundary

This Note does not define:

- GitHub Release assembly or publication ownership;
- package-manager internals or third-party repository policy;
- signing, notarization, or Windows trust presentation;
- WinGet publication;
- a general upgrade, rollback, or version-retention guarantee.

Release assembly remains documented in the [Go README](../../go/README.md).
Unresolved channel and trust questions remain in [Collected Future
Candidates](../future/future.candidates.md#package-manager-distribution).
