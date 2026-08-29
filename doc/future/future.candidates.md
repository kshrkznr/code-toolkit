# Knowledge.future.candidates.md
============================================================

# Collected Future Candidates

This document collects candidate directions that do not yet need separate
Future documents.

The candidates have different priorities and levels of evidence. None of them
is an accepted responsibility or implementation commitment.

## Master Runtime

A Master Runtime could be an internal CTK Runtime rather than a Runtime selected
by a user's Recipe.

Possible responsibilities include:

- resolving Extension versions
- synchronizing the Extension Pool
- collecting Extension metadata
- maintaining a future cache
- keeping CTK-related Extensions available in one internal environment
- supporting Extension mode or source management

The candidate remains underspecified. CTK does not yet need a separate Runtime
identity for these operations, and existing Build, Lock, and Pool behavior
should not be moved behind a Master Runtime without a concrete operational
need.

Reconsider it when several internal operations require one shared, managed
Runtime whose lifecycle is clearly different from a user Runtime.

## Documentation maturity and glossary

Knowledge may intentionally lag behind implementation while a concept is still
stabilizing. This is preferable to presenting provisional implementation
behavior as an accepted responsibility.

A glossary may eventually help readers and AI assistants resolve recurring CTK
terms without loading an entire Concept document. Its responsibility and
maintenance boundary are not yet clear: it could become a useful navigation
aid, or it could duplicate definitions already owned by Concept APIs.

Reconsider a glossary when repeated onboarding or review work shows that term
resolution is a distinct problem rather than a symptom of unclear documents.

## Workspace and Dev Container Build

Project-owned Workspace and Dev Container content may eventually become Build
targets composed through Cookbook Ingredients and Recipes.

Possible targets include:

- VS Code-family `.code-workspace` files
- Dev Container definitions

These are not current Ingredient Layers or accepted Core responsibilities. CTK
does not currently define their composition units, Build lifecycle, ownership,
or relationship to a project repository.

This candidate is also separate from CTK Workspace discovery, which locates the
Cookbook, Dist, Archive, and other CTK-owned state.

The current Runtime Artifact boundary remains unchanged: project-owned Tasks
such as `.vscode/tasks.json` and Tasks inside `.code-workspace` files are not
managed as IDE Runtime content. A future Workspace Build would need to define
its own responsibility before that boundary could be extended.

This is a low-priority candidate. Reconsider it when an actual Recipe needs to
build or maintain project-owned Workspace or Dev Container content.

## Knowledge as an implementation source

Mature Knowledge documents might eventually become the primary source for
generating repository structures, configuration files, or implementations
across multiple languages and Platforms.

The question is not whether CTK should generate code. It is whether documented
responsibilities can become complete and precise enough to serve as an
implementation contract without losing their value as readable Knowledge.

Reconsider this direction only after several implementations can be compared
against the same stable conceptual and behavioral boundaries.

## Package-manager distribution

CTK is implemented as a self-contained Go executable and produces versioned
macOS and Windows Release Artifacts with checksums. The remaining distribution
candidate is publication through package managers such as:

- Homebrew
- Scoop
- WinGet

This candidate does not reopen Workspace ownership or Direct Launcher
independence. Those boundaries are already defined by the
[Workspace Concept API](../integration/integration.workspace.md),
[Workspace integration rationale](../design-note/design-note.platform-definition-scope.md),
and [Go Distribution Contract](../../go/doc/contract/contract.distribution.md).

The remaining candidate is the publication and maintenance mechanism itself:

- Homebrew Tap, Scoop Bucket, and possible WinGet package ownership
- signing, macOS notarization, and Windows trust presentation
- checksum publication and verification through package-manager workflows
- upgrade, rollback, and version-retention policy

Reconsider it when CTK is ready to maintain one concrete package channel and
validate installation, upgrade, rollback, and removal on its target OS.

## Shared Recipes

People may eventually share complete development Recipes rather than only
sharing editor Settings.

Possible examples include:

- Review Recipe
- Java Enterprise Recipe
- Minimal Recipe
- Vim-like Recipe

Each Recipe could include Knowledge describing its design philosophy and
intended workflow.

This remains an interesting cultural possibility, not a sharing protocol or
roadmap item. Reconsider it when independently maintained Recipes expose a
concrete need for publication, discovery, or compatibility boundaries.
