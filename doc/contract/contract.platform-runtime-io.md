# Knowledge.contract.platform-runtime-io.md
============================================================

# Platform Runtime I/O Contract

Platform Runtime I/O is the capability boundary between resolved Cookbook
content and a Platform-owned Runtime.

It exposes operations required by later Build, Apply, and Lock orchestration
without embedding Cookbook parsing or Distribution layout assumptions in a
Platform adapter.

```text
Runtime Plan
    │
    ▼
Platform Runtime I/O
    │
    ▼
Platform-owned Runtime
```

## Required Capabilities

### Runtime scopes

- Address the default Runtime scope.
- Discover and address named Profile scopes when the Platform supports them.
- Create or ensure a named Profile required by the Runtime Plan.
- Preserve the distinction between default-inherited and Profile-local content.

### Settings

- Read Settings from an applicable Runtime scope.
- Provide resolved Settings to an applicable Runtime scope in a format the
  Platform can read.
- Report malformed or rejected Settings as an error.

### Extensions

- Observe installed Extension IDs in an applicable Runtime scope.
- Install an Extension requirement through a Platform-owned operation.
- Uninstall an Extension through a Platform-owned operation.
- Preserve Extension ID spelling and letter case at the boundary.
- Report partial or rejected Extension operations as errors.

### Observation

- Expose the Platform-owned state required for later Lock observation.
- Distinguish unavailable content from valid empty content.

## Safety Invariants

- Cookbook Source is never modified by Platform Runtime I/O.
- Installed Extension directories are not treated as CTK-owned canonical Source.
- Platform-owned Extension metadata is changed through the Platform adapter,
  not edited as if it were a Cookbook Resource.
- Reapplying the same desired content is supported.
- A partial failure is not reported as successful convergence.
- Default Runtime and named Profile scopes are not silently mixed.
- Platform-specific representation does not leak into Cookbook Resolution.

## Extensible Content Model

The boundary allows Runtime content kinds to evolve without restructuring
Cookbook or lifecycle orchestration. Applicable content includes:

- Keybindings
- Snippets
- Tasks
- MCP configuration
- other Platform-supported Runtime resources

Their composition and Cookbook Source boundaries are defined by the
[Runtime Artifact Contract](contract.runtime-artifacts.md).

Platform Runtime I/O should therefore model content scope and capability rather
than define one monolithic Settings-and-Extensions function.

## Current VS Code-family Strategy

The current adapter may satisfy these capabilities through the Platform CLI and
its supported user-data representation, for example:

- create or open a Profile
- list Extensions
- install or uninstall Extensions
- read or write Settings for the default Runtime or a named Profile
- configure Profile inheritance

Specific CLI flags, process commands, Profile database fields, and file paths
are adapter representation details rather than Required Capabilities.

## Open Questions

- The common capability vocabulary for Platforms without named Profiles.
- How an adapter reports unsupported optional content kinds.
- How structured operation reports evolve as new content kinds are added.
- How future packaged Distributions expose mutable and observable Runtime scopes.
