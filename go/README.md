# CTK Go

This directory contains the primary native Go implementation of CTK.

The Windows release is a native executable. Using the primary Go
implementation on Windows does not require Git Bash.

This README is the current entry point for Go behavior and implementation
decisions. Historical observations from the Bash-to-Go migration are secondary
context in the Project Knowledge
[Responsibility-First Implementation Migration](../doc/project-knowledge/experiment/experiment.responsibility-first-implementation-migration.md)
Experiment. The preserved Bash implementation remains available under `bash/`
as a historical and behavioral reference.

## Implementation principles

- Write idiomatic Go rather than translating Bash functions and pipelines.
- Model domain responsibilities explicitly with Go types and packages.
- Keep filesystem and Platform integration at clear boundaries.
- Prefer deterministic behavior and explicit errors.
- Introduce interfaces only for a concrete second implementation or useful
  test boundary.
- Add tests against Contracts and observable behavior.
- Preserve unknown source data when a read-modify-write workflow requires it.
- Avoid host mutation until its recovery boundary is documented and tested.

## Applied Kitchen Notes

The Go implementation currently adopts the Merge Rules Kitchen Note for
Settings composition:

- Settings Resources enter the merge stream in Cookbook Core resolution order.
- Objects merge recursively.
- Arrays, scalars, and `null` use the later value.
- Array paths absent from Merge Rules use later-value `replace`.
- Cookbook-wide exact paths may use deterministic, canonical-value-deduplicated
  `union` through `cookbook/kitchen-notes/go.merge-rules.yaml`.
- Merge Rules are not selected per Recipe.

Notes not declared here are not part of Go Cookbook interpretation.

## Implementation Contracts

Observable behavior and persisted representations selected by the Go
implementation are indexed under
[`doc/contract`](doc/contract/README.md). Read them together with the shared
Contracts under [`../doc/contract`](../doc/contract/README.md).

The shared Contract defines the implementation-independent agreement. The Go
Contract records concrete choices such as trusted Manifest formats, Native CLI
selection, CodeVenv recovery, and Workbench Commit syntax.

Recipe Extension installation uses the Platform-owned CLI by default. Exact
VSIX acquisition into `.vsix` is separate: omitted
`config.dist-strategy.extension-pool` means `reuse`, so Go reads existing Pool
artifacts without downloading new ones. `refresh` explicitly enables Go-owned
Repository download after Runtime observation and when an Archive requires a
missing exact artifact.

## Package layout

```text
go/
├── doc/contract/          Go implementation Contracts
├── cmd/ctk/              CLI entry point
├── verify.sh             Local and CI repository verification
├── release.sh            Cross-platform Release artifact builder
└── internal/
    ├── archive/          Hashed snapshot packaging and validation
    ├── buildinfo/        Binary version and build provenance
    ├── cli/selector/     Native single-selection boundary
    ├── codevenv/         Runtime selection and transactions
    ├── converge/         Settings and Extension convergence reports
    ├── cookbook/         Pure Cookbook-to-Runtime-Plan resolution
    ├── directlauncher/   Generated host-facing command launchers
    ├── distribution/     Distribution discovery and metadata
    ├── flatformat/       Editable Workbench Settings representation
    ├── launcher/         Override and native launch resolution
    ├── lifecycle/        Recipe/Archive Build, Apply, and Lock orchestration
    ├── mergerules/       Go Kitchen Note merge-rule loading
    ├── platform/         Built-in Platform Registry and process integration
    ├── recipe/           Recipe representation and loading
    ├── recovery/         Internal trusted-Lock Runtime reconstruction
    ├── runtimeio/        Platform Runtime capability boundary
    ├── runtimelock/      Runtime observation and trusted Lock publication
    ├── settings/         JSONC parsing and VS Code merge semantics
    └── workbench/        Freeze and Inspect Artifact lifecycle
```

## Build and test

Build the repository binary:

```bash
go -C go build -o ../bin/ctk ./cmd/ctk
```

Run the same repository verification used by CI:

```bash
go/verify.sh
```

It checks Go formatting, tests, vet, the Windows cross-build, Bash syntax, the
current Documentation Bundle, generated third-party notices, and whitespace
errors. CI also runs the Go tests natively on macOS and Windows so host-specific
implementations and build-tagged tests execute on their target OS.

The repository entry points are:

```text
bin/ctk        Primary Go implementation
bash/bin/ctk   Bash reference implementation
bash/scripts/  Bash reference source
```

## Release artifacts

Create versioned macOS and Windows artifacts with checksums:

```bash
go/release.sh v0.5.1
```

Release assembly requires a clean checkout whose `HEAD` is the exact requested
tag. It generates one deterministic Documentation Bundle, compares it with a
fresh archive of that tag, appends the same verified bytes to every executable,
and verifies each packaged Manifest before creating checksums.

The builder produces:

```text
release/v0.5.1/
├── ctk_v0.5.1_darwin_arm64.tar.gz
├── ctk_v0.5.1_darwin_amd64.tar.gz
├── ctk_v0.5.1_windows_amd64.zip
└── checksums.txt
```

Each platform archive contains the executable with version-matched CTK
documentation, the CTK `LICENSE`, and generated `THIRD_PARTY_NOTICES`. The
packaged binary provides `ctk docs` navigation and safe `ctk docs export
<directory>` for repository-style full-text search. Regenerate the checked-in
notice inventory after dependency changes:

```bash
go/third-party-notices.sh
```

Pushing an annotated `vX.Y.Z` tag starts the Draft Release workflow. The
workflow accepts only a tag whose commit is contained in `main`, runs
`go/verify.sh`, invokes this builder once on macOS, verifies the same artifact
bytes and packaged CLI on Windows, and then creates a Draft GitHub Release.
Manual dispatch can verify an existing tag without creating or changing a
Release; Draft creation must be selected explicitly for a manual run.

Version selection, versioned documentation updates, the release Pull Request,
and annotated tag creation remain human-owned. After the workflow succeeds, a
maintainer reviews the generated notes, exact assets, and checksums before
publishing the Draft. The workflow does not replace a local `ctk` binary or
publish to package managers.

Repository builds remain the development path. Versioned Releases are the
intended source for future Homebrew and Scoop distribution.

The module path is `github.com/kshrkznr/code-toolkit/go`, matching the public
repository and the `go/` module root. Binary Releases and the `ctk` command name
remain independent of the module path.

## Workspace discovery

CTK resolves its workspace in this order:

1. `CTK_HOME`, when explicitly configured.
2. The current directory or its ancestors.

A workspace is discoverable when it contains both `cookbook/recipe` and
`cookbook/ingredient`, or when it contains `.config/workspace.yaml`. This
allows a Homebrew or Scoop binary to operate on Workspace state without
installing that state alongside the executable.

When interactive `ctk activate` finds no Workspace and `CTK_HOME` is not set,
it prompts for an editable Workspace path with `~/ctk` suggested. Enter accepts
the suggestion, another path replaces it, and Escape cancels without writing.
After confirmation, CTK creates the minimum footing without the executable
sample and continues activation. Non-interactive activation does not create a
Workspace implicitly. Later commands run from the new Workspace or use an
explicit `CTK_HOME`; activation bootstrap does not persist that selection.

When a binary-only installation needs a new starting point, the optional
initializer creates one at an explicit path:

```text
ctk init /path/to/my-ctk
cd /path/to/my-ctk
```

The default includes a small executable sample. Use `--exclude-sample` to
create only `cookbook/recipe` and `cookbook/ingredient`. Initialization does
not select or persist `CTK_HOME`; an already configured `CTK_HOME` continues
to take precedence over current-directory discovery.

The optional configuration can keep versioned Cookbook Source and generated
Distributions in independent locations:

```yaml
paths:
  cookbook-source: /path/to/cookbook
  dist: /path/to/dist
```

Relative values resolve from `CTK_HOME`. Recipe and Ingredient Source moves to
`cookbook-source`; generated `cookbook/draft` and `cookbook/inspect` remain
under `CTK_HOME`. Archive and Extension Pool retain their Workspace-local
defaults. See the [Go Workspace Contract](doc/contract/contract.workspace.md)
for validation and ownership details.

## Commands

```text
ctk init <path> [--exclude-sample]
ctk completion <bash|zsh|fish|powershell>
ctk activate [platform] [--force]
ctk build [recipe-or-archive] [--on-conflict suffix|abort] [--keep-staging] [--force]
ctk apply [recipe-or-archive] [dist] [--force]
ctk archive [dist] [--on-conflict suffix|replace|abort]
ctk lock [dist]
ctk freeze draft [dist] [--on-conflict abort|replace]
ctk freeze commit [--force]
ctk view [source] [--on-conflict abort|replace]
ctk view dist [dist] [--on-conflict abort|replace]
ctk view recipe [recipe] [--on-conflict abort|replace]
ctk view ingredient [all|layer|layer.name] [--on-conflict abort|replace]
ctk sync [left] [right] [--on-conflict abort|replace]
ctk list
ctk current [platform]
ctk deactivate [platform] [--force|--force-empty]
ctk use [dist]
ctk launch [dist] [--] [args...]
ctk workbench [draft|inspect] [viewpoint] [--editor command]
ctk docs [--source <repository>] [status|nodes|resolve|toc|show|export]
ctk select
ctk version
ctk help
```

Use `ctk <command> --help` (or `-h`) for concise command syntax and options.
Subcommand help is available without resolving a Workspace. Commands with a
clear conceptual owner also provide a copyable `ctk docs show <reference>`
route into the packaged Concepts and Contracts; detailed explanation remains
owned by `ctk docs` rather than being duplicated in Help.

Completion scripts contain static commands, subcommands, options, and closed
option values. They do not resolve Workspace Recipes or Distributions:

```text
ctk completion bash
ctk completion zsh
ctk completion fish
ctk completion powershell
```

`ctk help` keeps its static command list available without a valid Workspace.
It follows that list with a best-effort summary of the selected Workspace, its
selection source, and the packaged documentation version and revision. Use
`ctk docs status` for full Bundle provenance or `ctk docs --source
<repository> status` for explicit local-source comparison.

`launch` forwards trailing file, directory, and Platform arguments without
changing them. Use an empty Distribution slot or `--` to select a Distribution
interactively while still providing launch targets:

```text
ctk launch vscode-default .
ctk launch "" .
ctk launch -- .
ctk launch vscode-default -- file.go another-directory
```

Relative launch targets remain relative to the caller's working directory.

## Open a Workbench

Open an existing Draft or Inspect Workbench in an editor:

```text
ctk workbench
ctk workbench draft
ctk workbench inspect
ctk workbench inspect dist.vscode-default --editor code
```

Omitting the kind selects between the available Draft and Inspect areas.
Omitting an Inspect viewpoint selects one generated viewpoint. The editor is
resolved from `--editor`, then `$EDITOR`, then `code` when available, with
`vim` as the final fallback. CTK opens the Workbench directory so its Summary
and typed Draft Artifacts can be reviewed together.

Invoking `ctk` without a command opens the Native command Selector. Explicit
commands remain available for automation and shell history.

## Current implementation

The Go implementation provides the complete M0-M8 lifecycle:

- Native CLI selection, launch, and Runtime switching.
- Cookbook resolution with Settings variants and Merge Rules.
- Recipe and Archive Build/Apply with trusted Locks and Extension Pool support.
- CodeVenv activation and deactivation with recovery and safety gates.
- Freeze Draft/Commit, View, and Sync Workbench operations.
- Exact-version, hashed Archive creation and offline reconstruction.
- Direct Launcher generation for macOS command files and Windows `.cmd` files.
- Primary `ctk` binary and versioned Release artifact generation.

Windows builds and generated `.cmd` files have automated coverage and
cross-build validation. Host-specific behavior continues to be refined from
real-machine feedback.

## Observed Platform behavior

The implementation-specific Platform definitions, real-machine observation
matrix, automated coverage, and known evidence gaps are maintained in the
[Go Platform Support Inventory](doc/platform-support.md).

The inventory keeps product versions and OS-specific evidence separate from the
concise support status in the repository README and from behavioral Contracts.
The operational declaration and service boundary is described in the
[Built-in Platform Registry Note](../doc/note/note.platform-registry.md).

## Future

- Optional log and Operation Report presentation modes: normal, verbose,
  quiet, and JSON.
- Package-manager publication, signing, notarization, upgrade, and rollback
  questions preserved in
  [`future.candidates.md`](../doc/future/future.candidates.md).
- Workspace-defined Platform loading preserved in
  [`future.platform-registry.md`](../doc/future/future.platform-registry.md).
- Configurable no-argument behavior if a concrete non-interactive use case
  requires it.
