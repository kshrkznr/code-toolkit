# README.md
============================================================

# CTK(Code-Toolkit)

## What is CTK?

CTK is a toolkit for transforming existing development environments into reusable, composable Cookbooks.

## Why CTK?

You may use VS Code and Kiro side by side, but their Settings and Extensions
gradually drift apart. You may change a shared setting in one Profile and
forget to repeat it in the others.

These are everyday examples of a broader problem: modern development
environments are highly personalized but difficult to reproduce, review, and
evolve.

CTK captures those environments as reusable Cookbook responsibilities rather
than one-off configuration files. Shared parts can remain shared, while each
editor or Profile keeps its explicit differences.

You may also want to:

- start with a capable editor environment by describing how you want to work,
  while an AI assistant handles the initial Cookbook and build workflow
- keep an editor environment under explicit control
- review how an existing environment is composed
- rebuild or evolve an environment incrementally
- collaborate with AI without handing over architectural decisions

CTK can start from the environment you already use or from a description of how
you want to work. You can establish a small Cookbook first and introduce finer
Ingredients, Layers, and workflows only when they become useful.

## Explore with AI

The problems above may sound familiar, but you may not want to read a new
toolkit's documentation just to find out whether it helps. You do not have to
start there.

You can begin with the annoyance itself:

> I use VS Code and Kiro, but their settings have drifted apart. Show me how CTK
> could organize them.

Or describe only the environment you want:

> I want a lightweight editor for Markdown and Java. Use CTK to build an
> environment for me and help me refine it.

An AI assistant can navigate CTK's documentation and Cookbook, propose a small
Recipe, build a separate Distribution, and adjust it from your feedback. You do
not need to learn every setting, command, or CTK lifecycle before trying the
resulting environment. AI-assisted onboarding is optional; it is simply another
way into the same inspectable CTK workflow.

The same route also works after onboarding. Instead of providing physical
settings paths, you can ask an AI assistant to inspect a Recipe or Distribution,
create a View when useful, identify the responsible Ingredient, and prepare a
Build or Apply. Freeze remains the review path for bringing observed Runtime
changes back into the Cookbook; a clear Cookbook source change can be made at
its responsible Ingredient directly.

CTK's Concepts, Contracts, and review artifacts give the assistant a structured
workspace. They also keep the resulting environment inspectable when you later
want to understand or change it yourself.


============================================================

# Concept Domains

CTK is organized into a small set of Concept Domains.

A Concept Domain represents a major area of responsibility within CTK and provides a high-level view of the product.

Each domain exposes one or more Concept APIs, which define the stable public concepts of CTK.

The following diagram provides an overview of how these domains relate to each other.

## Conceptual diagram

```text
      CTK
Concept Domains
    │
    ├── Core
    │      ├── Cookbook API
    │      ├── Build Lifecycle API
    │      └── Persistence Lifecycle API
    │
    ├── Workbench
    │      ├── Draft API
    │      └── Inspect API
    │
    └── Integration
           ├── Documentation Resolver API
           ├── Project Structure API
           └── Code Environment API
```

============================================================

# Concept APIs

Each Concept Domain exposes one or more Concept APIs.

A Concept API represents a stable public contract of CTK.

The CLI, repository structure, and internal implementation exist to realize these APIs rather than define them.

Each Concept API is documented independently.

Use the Documentation Resolver to navigate to the appropriate Knowledge document for the task at hand.

============================================================

# Core Concept APIs

The Core Concept APIs define the fundamental concepts of CTK.

Together, they describe how reusable environments are organized, generated, and evolved.

Enter this Concept Domain through the [Core README](doc/core/README.md).

These APIs form the conceptual foundation of CTK and remain independent of specific implementations such as the CLI, repository structure, or supported platforms.

The following diagram illustrates the relationships between the Core Concept APIs.

## Conceptual diagram

```text
                  Core Concept API
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
    Cookbook      Build Lifecycle   Persistence Lifecycle
        │                 │                 │
        ▼                 ▼                 ▼
Recipe / Ingredient  Build / Apply   Freeze / Archive
```

## Cookbook
Reusable environment composition.

See: [Cookbook](doc/core/core.cookbook.md)

Primary concepts:
- Recipe
- Ingredient
- Variant
- Ingredient Layers

Implementation-specific Cookbook interpretation is documented by Kitchen
Notes. See: [Cookbook Kitchen Notes](doc/core/core.cookbook.kitchen-notes.md)

---

## Build Lifecycle
Generate consumable environments.

See: [Build Lifecycle](doc/core/core.build-lifecycle.md)

Primary concepts:
- Build
- Apply
- Dist

---

## Persistence Lifecycle
Review and evolve environments.

See: [Persistence Lifecycle](doc/core/core.persistence-lifecycle.md)

Primary concepts:
- Lock
- Freeze Draft
- Freeze Commit
- Archive

---

============================================================

# Workbench Concept APIs

The Workbench Concept APIs define the interactive work areas used to inspect, review, and evolve generated environments.

Unlike the Core Concept APIs, these APIs focus on supporting day-to-day workflows rather than defining the environment model itself.

The following APIs provide dedicated workspaces for reviewing changes and preparing persistent updates.

Enter this Concept Domain through the [Workbench README](doc/workbench/README.md).

---

## Draft

A review workbench for preparing changes before they become part of the
Cookbook.

See: [Workbench](doc/workbench/README.md)

Primary concepts:
- Draft Directory
- Draft Engine

---

## Inspect

A disposable workbench for understanding generated environments through inventories, reports, and comparisons.

See: [Workbench](doc/workbench/README.md)

Primary concepts:
- Recipe Inspection
- Dist Inspection
- Ingredient Inventory
- Difference Report
- Synchronization Review

============================================================

# Integration Concept APIs

The Integration Concept APIs define how CTK connects with external environments, projects, and documentation.

Rather than introducing new concepts, these APIs make the Core and Workbench concepts accessible across different workflows and development environments.

The following APIs provide integration points between CTK and its surrounding ecosystem.

Enter this Concept Domain through the [Integration README](doc/integration/README.md).

## Documentation Resolver

Guides users and AI agents to the appropriate Knowledge based on the task being performed.

See: [Documentation Resolver](doc/README.md)

Primary concepts:
- Concept Domains
- Concept APIs
- Knowledge
- Notes

---

## Project Structure

Maps CTK concepts and document responsibilities to discoverable repository
documents without prescribing a canonical project or Cookbook layout.

See: [Project Structure](doc/integration/integration.project-structure.md)

Primary concepts:
- Canonical Document Identity
- Concept Domain and Concept API Headings
- Repository Map
- Role-sensitive Heading Profiles

---

## Code Environment

Applies CTK concepts to development environments such as editors and IDEs.

See: [Code Environment](doc/integration/integration.code-venv.md)

Primary concepts:
- Platform
- Runtime
- Profile
- Distribution

---

============================================================

# Installation

The current macOS and Windows binaries are published on the
[GitHub Releases page](https://github.com/kshrkznr/code-toolkit/releases/latest).
Start from a repository checkout so the executable can discover the accompanying
Cookbook and documentation. Create `release/current` in the checkout, then
download `checksums.txt` and the archive for your platform from the release
Assets into that directory.

## macOS

Choose the archive for the Mac's architecture, verify the published checksums,
and place `ctk` on `PATH`:

```bash
cd release/current
shasum -a 256 ctk_v0.2.0_darwin_arm64.tar.gz
tar -xzf ctk_v0.2.0_darwin_arm64.tar.gz
mkdir -p "$HOME/.local/bin"
install -m 0755 ctk "$HOME/.local/bin/ctk"
```

Compare the displayed SHA-256 value with the corresponding entry in
`checksums.txt`. Use `ctk_v0.2.0_darwin_amd64.tar.gz` instead on an Intel Mac.
Add `$HOME/.local/bin` to `PATH` when it is not already available.

## Windows

From PowerShell, verify that the displayed hash matches the Windows entry in
`checksums.txt`, then extract the executable into a directory on `PATH`:

```powershell
Set-Location release/current
Get-FileHash .\ctk_v0.2.0_windows_amd64.zip -Algorithm SHA256
Expand-Archive .\ctk_v0.2.0_windows_amd64.zip -DestinationPath "$env:LOCALAPPDATA\CTK\bin"
```

Add `%LOCALAPPDATA%\CTK\bin` to the user `PATH` when necessary.

Confirm the installed binary:

```text
ctk version
```

Run CTK from the repository root, or set `CTK_HOME` to that directory. CTK
uses the workspace containing `cookbook/recipe` and `cookbook/ingredient`.

============================================================

# Getting Started

If you already use a VS Code-family editor, the quickest way to start is to
import that environment and let CTK generate the initial Cookbook structure.

If you are starting with a new editor or prefer not to operate CTK yourself,
begin with [Explore with AI](#explore-with-ai): describe the work you want the
editor to support and let an AI assistant prepare a small, separate
Distribution. The import workflow below remains available when there is an
existing environment you want to preserve.

You do not need to understand Ingredient Layers or Variants before getting started.

The generated proposals can be committed as they are and refined gradually as your Cookbook evolves.

---

## 1. Activate a Platform

Choose the Platform command whose current environment you want CTK to manage.

A Platform command is the CLI entry point for a VS Code-family application.
For example, use `code` for Visual Studio Code or `kiro` for Kiro.

```bash
ctk activate code
```

Platform activation is an explicit user action. Activating one Platform command
does not activate or change another Platform command.

Activation is reversible through `ctk deactivate <platform>`. See
[Leaving CTK](doc/note/note.leaving-ctk.md) for the boundary between restoring a
managed Platform and removing generated environments or the CLI.

CTK imports the current environment for that Platform and creates its initial
`origin.<platform>` Distribution.

Results:

- A temporary Recipe is generated.
- The Platform's User data and extensions become managed by CTK.
- An initial Distribution named `origin.<platform>` is created.
- The origin Runtime is recorded as a Lock.

---

## 2. Freeze Draft

Generate a Draft workbench.

```bash
ctk freeze draft
```

CTK analyzes the active Recipe and generates proposed Ingredients for your environment.

---

## 3. Review the generated proposals

Review the generated Ingredient proposals.

CTK generates draft Ingredients such as:

- `runtime.draft.settings.json`
- `runtime.draft.extensions`

These draft Ingredients can be committed as they are, or renamed and reorganized before committing.

The initial goal is simply to establish a reusable Cookbook.

Ingredient Layers and finer organization can be introduced later as your Cookbook evolves.

> **Note**
>
> Comments and formatting from the original files may not be preserved during the initial import.
---

## 4. Freeze Commit

Persist the reviewed Draft into the Cookbook.

```bash
ctk freeze commit
```

The proposed Ingredients become reusable Cookbook resources.

---

## 5. Build or Apply

Generate or activate environments from the Cookbook.

- **Apply** updates the current Distribution.
- **Build** generates a new Distribution.

After activation and Freeze Commit, both the committed Recipe and imported
Distribution use the `origin.<platform>` identity. To update the existing
Visual Studio Code environment:

```bash
ctk apply origin.code origin.code
```

The first `origin.code` selects the committed Recipe. The second selects the
existing Distribution created by `ctk activate code`. Use the corresponding
Platform identity, such as `origin.kiro`, when another Platform was activated.

To build a separate Distribution from the same Recipe instead:

```bash
ctk build origin.code
```

Once your environment is managed by CTK, you can continue refining the Cookbook incrementally without repeating the initial import process.

---

> ### From Comments to Concepts
>
> Your existing comments are not wasted.
>
> In traditional configuration files such as `settings.json`, comments are often the only way to express intent:
>
> ```jsonc
> // Git
> // Java
> // UI
> ```
>
> These comments describe concepts that the configuration format itself cannot represent.
>
> CTK provides a richer model.
>
> As your Cookbook evolves, those comments can naturally become Ingredients or other named responsibilities.
>
> Rather than preserving comments verbatim, CTK encourages promoting reusable intent into named concepts.
>
> You do not need to organize everything immediately.
>
> A single `runtime.draft.settings.json` is a perfectly valid starting point.
>
> As patterns emerge, you can split them into focused Ingredients and use Layers or Variants when those distinctions become useful.

============================================================

# Try the Sample Recipe

The executable sample is a separate path for exploring CTK without importing
the editor environment you already use.

Build the Recipe matching the current OS:

```bash
# macOS
ctk build cookbook/recipe/vscode-sample.macos.yaml

# Windows
ctk build cookbook/recipe/vscode-sample.windows.yaml
```

Then launch the generated Distribution without changing the Runtime selected
by CodeVenv:

```bash
ctk launch vscode-sample
```

This sample remains intentionally small. The main Getting Started path above
continues to begin with the user's existing environment.

============================================================

# AI Collaboration

> Welcome, collaborator.

CTK organizes environment work into explicit, inspectable responsibilities.

A collaborator may be a human, an AI assistant, a teammate, or your future
self. CTK does not assign work based on who performs it. Instead, it provides
shared concepts and reviewable artifacts through which collaborators can:

- navigate the project through the Documentation Resolver
- understand responsibilities through stable Concept APIs
- inspect Recipes, Distributions, Views, and Drafts
- identify and edit the Ingredient responsible for a clear source change
- build, apply, and evolve a Cookbook through explicit lifecycle operations

CTK does not embed an AI service or require AI-assisted operation.

It provides a workspace in which different participants can understand the
same responsibilities, review the same artifacts, and improve the environment
together.

============================================================

# Examples

The executable Cookbook includes a small text-editor Recipe for Visual Studio
Code:

- [macOS](cookbook/recipe/vscode-sample.macos.yaml)
- [Windows](cookbook/recipe/vscode-sample.windows.yaml)

Current Recipe selectors use `macos` or `windows` for `os`, and `code` or
`kiro` for the Platform command. This sample selects `code`.

It composes three focused Runtime Ingredients: `common`, `text`, and
`markdown`. The sample intentionally omits Profiles and environment-specific
Settings Variants until a concrete difference needs them. Its Distribution
strategy is written explicitly so the executable defaults are visible in the
Recipe itself. Extension IDs use the Platform's normal installation route,
while `extension-pool: reuse` keeps CTK-owned VSIX download disabled unless a
Recipe opts into artifact acquisition.

The Author's Recipe demonstrates how the Concept APIs are composed in a real environment.

See [Author's Recipe](doc/authors-recipe/README.md). Its generated Inspect Views
are provisional and do not define a canonical Cookbook layout.

============================================================

# Supported Platforms

The current Go implementation provides host integration on macOS and Windows
for these VS Code-family applications:

| Application | Platform command | Current status |
| --- | --- | --- |
| Visual Studio Code | `code` | Primary supported Platform |
| Kiro | `kiro` | Implemented |
| Cursor | `cursor` | Implemented |

Other compatible IDEs remain an intended area of exploration rather than a
current compatibility or release commitment.

The [Code Environment Concept API](doc/integration/integration.code-venv.md) and
[Code Environment Integration Contract](doc/contract/contract.code-environment.md)
define the Platform boundary independently from this current support list.

============================================================

# Primary CLI Commands

The Go and Bash implementations are independent implementations of CTK. Go is
the current primary implementation; Bash is retained as historical and
behavioral evidence, not as a helper layer used by Go.

The commands below are the primary interfaces used by the onboarding and
review workflows. See the [Go Language README](go/README.md#commands) for the
complete current command syntax. The retained Bash implementation documents
its own boundary in the [Bash Language README](bash/scripts/README.md).


| Command | Purpose | Related Concept API | Notes |
|---------|---------|---------------------|-------|
| activate | Activate a Platform command for CodeVenv management | Code Environment | Requires an explicit Platform command |
| freeze draft | Generate a draft workbench | Draft | |
| freeze commit | Persist cookbook changes | Draft | |
| build | Generate distributions | Build Lifecycle | |
| apply | Apply a distribution | Build Lifecycle | |
| view | Inventory one completed source | Inspect | |
| sync | Compare two completed sources | Inspect | |
| workbench | Open an existing Draft or Inspect workbench | Workbench | Selects the area and Inspect viewpoint when omitted |

============================================================

# License

CTK is available under the [MIT License](LICENSE).

Copyright (c) 2026 kshrkznr.

SPDX license identifier: `MIT`.

Third-party software included in the release binaries is documented in
[`THIRD_PARTY_NOTICES`](THIRD_PARTY_NOTICES). Release archives include both
the CTK License and these notices.
