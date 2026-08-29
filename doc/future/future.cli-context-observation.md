# Knowledge.future.cli-context-observation.md
============================================================

# Future: CLI Context Observation

CTK resolves a Workspace and, in a future packaged documentation experience,
may select between Knowledge carried by the binary and an explicitly configured
local repository. A person or AI currently has to infer those paths and sources
from environment variables, the working directory, or later command failures.

A concise dynamic context summary could make the current invocation easier to
understand before performing an operation.

## Candidate experience

The Go CLI help could append an observation such as:

```text
Current context:
  Workspace:      ~/local/my-ctk
                  source: CTK_HOME
  Documentation:  embedded v0.4.0 @ abc1234
  Local source:   ~/local/dev/code-toolkit
                  configured, revision differs
```

This summary would answer two immediate questions:

- Which Workspace would a Workspace-dependent command use, and why?
- Which documentation source is selected, and does a configured local source
  represent the same revision as the binary?

The summary should remain short enough to accompany ordinary Help. Detailed
documentation Definition, digest, revision, and selected-path comparison could
belong to a later `ctk docs status` form rather than expanding Help into a full
diagnostic report.

## Observation without Workspace dependency

Current Go Help is dispatched before Workspace discovery so a missing or
invalid Workspace cannot prevent the CLI from explaining itself. Dynamic
context should preserve that property.

Help could observe a Workspace candidate and its selection source without
loading Cookbook Source, Dist, Archive, Extension Pool, Workbench, or Host
integration state. Outcomes such as these remain displayable context rather
than Help failures:

```text
Workspace: ~/local/my-ctk
           source: CTK_HOME

Workspace: unavailable
           CTK_HOME path is not a CTK Workspace

Workspace: not selected
           run inside a Workspace or set CTK_HOME
```

A Workspace-dependent command would continue to validate the selected context
strictly before operating. Help would describe the resolution outcome without
granting an invalid candidate operational authority.

Useful selection-source vocabulary may include:

- explicit `CTK_HOME`;
- working-directory discovery;
- executable-relative development fallback;
- no selected Workspace.

The exact representation remains a Go implementation candidate. The important
distinction is between the resolved path and the evidence that selected it.

## Home-relative display

Absolute local paths are operationally useful but often contain a user name
when Help or diagnostics are copied into an Issue. Display can preserve useful
structure while masking that identity:

```text
/Users/alice/local/my-ctk
    → ~/local/my-ctk

C:\Users\alice\local\my-ctk
    → ~\local\my-ctk
```

The implementation should derive the actual home directory through the host
environment rather than matching hard-coded `/Users`, `/home`, or `C:\Users`
patterns. Only the home path itself and true descendants become home-relative.
A similar prefix or an unrelated path remains unchanged.

This is a presentation rule only. Workspace selection, configuration loading,
validation, and filesystem operations continue to use the resolved absolute
path. Paths outside home may still expose organization, repository, or project
names, so contribution guidance should ask reporters to review diagnostic
output before pasting it publicly.

## Documentation source relationship

The [Packaged Documentation Bundle](future.documentation-bundle.md) defaults to
version-matched Knowledge carried by the binary. A nearby clone must not replace
that source merely because it can be discovered from the current directory.

If a local documentation repository is explicitly configured, the context
summary should distinguish at least:

- embedded documentation selected;
- local documentation configured but not selected;
- local documentation selected;
- local and binary revisions match, differ, or are unknown.

The concise summary need not compute or display every selected-path difference.
Detailed local revision, Bundle Definition, aggregate content, and dirty-path
status can remain the responsibility of documentation-specific status output.

## Release coordination

This candidate and the Packaged Documentation Bundle have separate Issues and
responsibilities:

- Documentation Bundle owns Knowledge selection, provenance, navigation,
  Export, and binary transport;
- CLI Context Observation owns concise dynamic visibility of the sources and
  paths selected for one invocation.

If adopted, they should first ship in the same CTK Release. A binary should not
introduce configurable embedded/local documentation without also making the
selected source observable. Sharing a Release version does not merge their
Contracts or implementation boundaries.

## Open questions

- Should the dynamic summary appear on every `ctk help`, or behind a concise
  `ctk help context` or `ctk status` form?
- Which optional configuration can Help observe without turning missing or
  malformed Workspace configuration into a Help failure?
- Should an explicit diagnostic option reveal absolute paths, or is the
  home-relative form sufficient for all presentation?
- Which fields remain stable enough for scripts, and which are human-oriented
  Help text only?
- How should a configured but unavailable local repository be summarized?

## Boundary

This Future does not define:

- accepted Help output or a stable machine-readable status schema;
- a global active-Workspace registry;
- silent local-repository discovery or selection;
- relaxed validation for Workspace-dependent commands;
- path masking for secrets or arbitrary organization names;
- implementation of packaged documentation or local-source comparison.

The implementation discussion is tracked separately in [Issue
#23](https://github.com/kshrkznr/code-toolkit/issues/23). Release coordination
with packaged documentation is tracked through [Issue
#20](https://github.com/kshrkznr/code-toolkit/issues/20).
