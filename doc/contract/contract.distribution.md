# Knowledge.contract.distribution.md
============================================================

# Distribution Contract

A Distribution is a completed Runtime artifact.

The Distribution contract is capability-based. It does not require a
Distribution to be a directory or prescribe its internal files, executable
format, metadata encoding, or launch command.

A directory, application bundle, executable, script, package, or another future
artifact may be a Distribution when it provides the required capabilities.

## Required Capabilities

### Launch

- The Distribution's Platform can be launched.
- Its launch capability can be consumed from supported CTK CLI and GUI surfaces.
- `ctk launch` can launch the Distribution without making it the default
  Runtime.
- Additional Platform arguments can cross the launch boundary when the launch
  surface accepts arguments.
- The launch representation may be replaced through Launch Override.

### Recipe provenance

- CTK can obtain the Recipe used by Build or Apply, or an equivalent provenance
  representation that preserves the information required by later lifecycles.
- Provenance remains available without requiring the source Recipe to remain at
  its original repository path.

### Runtime observation

- Lock can observe the completed Runtime state exposed by the Distribution.
- The Distribution exposes enough Runtime state for the applicable Platform
  adapter to produce a Lock-compatible snapshot.

### Lifecycle interoperability

- Build, Apply, Launch, Lock, Freeze, and Archive exchange Distribution
  capabilities rather than depending on one implementation's internal layout.
- A Distribution created by one lifecycle remains a valid input wherever the
  Concept API accepts a completed Runtime.

---

## Safety Invariants

- Launching a Distribution must not implicitly change the default Runtime.
- Observing a Distribution must not mutate its Runtime state unless the user
  explicitly invokes a mutating lifecycle.
- Missing launch or provenance capabilities are reported rather than guessed
  from an artifact name.
- A failed Launch Override must not silently start a different launch strategy.

---

## Launch Strategies

Launch Strategy is an implementation choice used to satisfy the Launch
capability. It is not itself the Distribution contract.

Possible strategies include:

- application bundles such as `.app`
- native executables such as `.exe`
- scripts such as `.sh` or `.cmd`
- a native Platform adapter
- a generated CLI or GUI launcher
- a Distribution-provided Launch Override

An implementation may support more than one strategy and define a deterministic
resolution order.

---

## Required Source Compatibility

Distribution representation compatibility is not a general project
requirement.

Unlike Cookbook Source, Distributions created by different implementations do
not need to share internal files when both satisfy the capability Contract.

---

## Implementation-specific resolution

The primary implementation's directory representation, Launch resolution,
launch-only input, Direct Launcher, and retained Bash interoperability are
defined by the
[Go Distribution Contract](../../go/doc/contract/contract.distribution.md).

---

## Open Questions

- The long-term normalized provenance model.
- How artifact capabilities are discovered without assuming a directory layout.
- How packaged `.app` and `.exe` Direct Launchers are represented when a future
  Build Strategy selects them.
- How Lock observes packaged or native executable Distributions.
- Whether a formal Distribution format version becomes useful.
