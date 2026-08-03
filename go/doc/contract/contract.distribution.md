# Go.contract.distribution.md
============================================================

# Go Distribution Contract

This Contract specializes the shared
[Distribution Contract](../../../doc/contract/contract.distribution.md).

## Directory representation

```text
dist/<name>/
├── .data/
├── .ext/
├── .meta/
│   └── recipe.yaml
├── run.sh or run.cmd
└── <dist-name> or <dist-name>.cmd
```

`.data` is VS Code-family user data, `.ext` contains installed Extensions, and
`.meta/recipe.yaml` preserves Recipe provenance. A complete Go Distribution
requires the Runtime and provenance capabilities applicable to its lifecycle.

Go reads the compatible Bash directory paths and platform-appropriate
`run.sh`/`run.cmd` as a retained interoperability bridge. Bash-only metadata is
not required Go state.

## Launch resolution

```text
platform-appropriate Launch Override
        │
        ├── present ──► execute Override
        └── absent  ──► native Platform Adapter
```

A failing Override does not fall back to the Native Adapter.

For `ctk launch` only, Go accepts a directory containing only the
platform-appropriate Override. This launch-only input does not claim Recipe
provenance, observation, Lock, Build, Apply, Archive, or `use` capabilities.
Without an Override, Native Adapter launch requires the complete current Runtime
representation.

## Direct Launcher

Build and Apply may generate a transparent host-facing Direct Launcher in the
Distribution. It projects the same resolution rule: execute the Override when
present, otherwise invoke the Platform command with Distribution Runtime paths.
It does not require `ctk` to remain installed or on `PATH` after generation.

A Direct Launcher is generated output, not a Launch Override.

## Source compatibility boundary

Cookbook Source is shared directly across implementations; Distribution
internals are not. Go validates required capabilities and does not infer a
complete Distribution merely from a directory name or one launch file.
