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

## Package-manager distribution and CTK Workspace

CTK is implemented as a self-contained Go executable and produces versioned
macOS and Windows Release Artifacts with checksums. The remaining distribution
candidate is publication through package managers such as:

- Homebrew
- Scoop
- WinGet

This direction includes signing, notarization, upgrade and rollback policy, and
the boundary between an installed binary and the CTK Workspace. Alternative
language implementations are not required for this candidate.

### Current Workspace resolution

The Go implementation can produce a repository-local `bin/ctk` and macOS and
Windows Release Artifacts. It currently resolves the Workspace in this order:

```text
CTK_HOME
    ↓
current directory or its ancestors
    ↓
repository-local executable position
```

This is sufficient for current development and manual distribution. A package
manager would separate the binary from CTK state, requiring the relationship
between the CLI and Workspace to be observed again.

### Questions for package-manager distribution

- Should the CTK root become an explicit `Workspace` Concept?
- Should `CTK_HOME` mean the default Workspace, a fixed root, or a compatibility
  environment variable?
- Is an explicit option such as `--home` or `--workspace` needed?
- How should users select among multiple Workspaces?
- How far should discovery walk from the current directory?
- How independently may the binary, Cookbook, Dist, Archive, and Pool be
  located?
- May a Cookbook live in an independent repository?
- May multiple Workspaces share an Extension Pool or Archive?
- Is a distinction between Workspace-local and user-global configuration
  needed?
- Must Direct Launchers remain stable after Homebrew or Scoop updates the
  binary?
- Should generated Direct Launchers continue to depend only on their Dist
  rather than on a Workspace?
- Which signing, macOS notarization, checksum, upgrade, and rollback concerns
  belong to CTK distribution responsibility?

### Platform expansion as an observation point

Adding another Platform would also test this boundary. The current Platform
Adapter still contains assumptions inherited from the VS Code family:

- `.data` and `.ext`
- Profiles and the Platform database
- JSONC Settings
- Extension IDs and VSIX Artifacts
- `--user-data-dir` and `--extensions-dir`
- launching and host integration through a Platform command

CTK should not generalize these assumptions by imagining another Platform.
When the first non-VS Code-family Platform is added, observe:

- which capabilities actually belong to Core
- which details are only VS Code-family Adapter representations
- which Platform capabilities belong in Distribution metadata
- where the CLI must understand a Platform difference
- which state should remain separate or shared between Platforms in one
  Workspace

### Current direction

Keep the current simple Workspace for now:

```text
CTK Workspace
├── cookbook/
├── dist/
├── archive/
└── .vsix/
```

Do not make every location configurable in advance. Use observations from an
actual package-manager distribution, multiple-Workspace use, or another
Platform to promote only the necessary boundaries into a Contract or
Configuration.

This is a Future theme for testing current location assumptions, not a current
specification or roadmap.

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

## AI documentation bundle

After CTK's documentation roles and contents become stable enough, the Release
process may produce a documentation bundle that can be shared with an AI
assistant without sharing the entire repository.

`for-your-ai.zip` is a provisional name for this possible Artifact. It does not
exist in the current Release and is not part of the present onboarding flow.

Questions to resolve include:

- which curated documents belong in the bundle
- whether Notes, Design Notes, Experiments, Future, or Raw should be included
- how the bundle communicates document role and maturity
- how it remains aligned with a particular CTK Release
- whether `release.sh` should generate it with the binary Artifacts

This would be a convenience for future AI-assisted workflows, not a requirement
for using CTK or its Documentation Resolver.
