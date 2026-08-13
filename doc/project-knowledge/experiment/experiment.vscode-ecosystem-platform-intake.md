# project-knowledge.experiment.vscode-ecosystem-platform-intake.md
============================================================

# VS Code Ecosystem Platform Intake

This experiment asked how CTK could add a VS Code-family application without
assuming that every fork behaves exactly like Visual Studio Code.

Cursor and Kiro first made the question concrete. Their familiar CLI and Runtime
shapes suggested reuse, but each exposed a product-specific boundary: Kiro made
OS output and Marketplace behavior visible, while Cursor exposed first-run,
Profile persistence, Gallery, and process-ancestry differences. Later VSCodium
and Devin Desktop observations tested whether those discoveries formed a
repeatable intake path.

## What the experiment established

Four questions that initially looked like one compatibility question need to be
observed separately:

1. Does the Platform command exist and identify the intended application?
2. Can User data and Extensions be redirected into an isolated Runtime?
3. Are Settings, Extensions, Profiles, and other content compatible with the
   shared VS Code-family Runtime Adapter?
4. Can CTK safely stop the correct processes and manage the Host lifecycle on
   each supported OS?

Shared flags are useful evidence, but they do not establish all four. Support
claims should follow concrete CLI, file, process, Gallery, lifecycle, and
recovery artifacts.

## Durable results

The reusable observation sequence has moved to the
[VS Code Ecosystem Platform Intake Note](../note/note.vscode-ecosystem-platform-intake.md).
It preserves the intake practice without making it a required compatibility
specification.

The current built-in Platform definitions, OS/version observation matrix,
implementation incorporation points, automated coverage, and visible gaps have
moved to the
[Go Platform Support Inventory](../../../go/doc/platform-support.md).
That inventory owns current implementation status and can change as Platforms
are re-observed.

The supporting observations include:

- CRLF normalization at the command-output boundary
- Junction and symbolic-link resolution as managed links
- Cursor Profile persistence before Runtime stop
- same-name process ancestry for desktop-root selection
- Platform Gallery identity as a compatibility and distribution boundary
- caller-owned system certificate configuration in TLS-inspecting environments
- interrupted CodeVenv transaction recovery and explicit retry boundaries

## Outcome

The Registry refactoring resolved the experiment's declaration boundary for
Built-in Platforms. Host paths, process identities and registered filters, and
Repository policy now resolve from centralized data. Process-selection
algorithms, Profile handshakes, and Gallery download validation remain named Go
behavior rather than arbitrary logic loaded from configuration.

External Workspace definition loading remains a Future. It reuses this boundary
but does not turn a user definition into a CTK-supported Platform or expose an
arbitrary executable extension API.
