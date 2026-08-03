# Knowledge.contract.runtime-convergence.md
============================================================

# Runtime Convergence Contract

Runtime Convergence turns a resolved Runtime Plan into Platform-owned state and
observes the resulting state as a Lock.

```text
Runtime Plan
    │
    ├── Build ──► new Distribution
    │
    └── Apply ──► existing Distribution
                         │
                         ▼
                        Lock
```

## Build and Apply

### Required Capabilities

- Build creates a new Runtime from resolved source.
- Apply converges an existing Runtime without recreating it.
- Both use Platform Runtime I/O rather than modifying Platform-owned Extension
  metadata as Cookbook Source.
- Desired Extensions converge by installing `desired - installed`, removing
  `installed - desired`, and retaining or updating `desired ∩ installed`.
- Recipe provenance remains available from the resulting Distribution.
- Repeating Apply can continue convergence after a partial failure.

### Failure model

Platform operations may partially change a Runtime. Build and Apply do not
claim success when an attempted Settings, Profile, or Extension operation
fails. They report the completed, failed, and unresolved operations so the
result can be inspected and Apply can be retried.

Full rollback of Platform-owned state is not required. In particular, an
Extension successfully installed before another Extension fails need not be
uninstalled merely to recreate the earlier state.

An Extension unavailable because Marketplace access is disabled and no Pool
artifact exists is reported as `unresolved`, not as an application error. The
Runtime may complete with this warning and Lock records the actual installed
state. A Platform operation that was attempted and rejected is a failure.

### Profile preservation

Apply must converge Profiles selected by the Recipe. Removing a Profile absent
from the Recipe is not required. Preserving such Profiles is the recommended
strategy because implicit Profile deletion can destroy user state.

### Platform process coordination

Stopping the Platform is required only around operations that mutate or replace
Platform database state which is unsafe to edit concurrently, currently the
VS Code-family Profile `storage.json` representation. Settings and Extension
operations do not require a global stop solely because they are part of Apply.

Profile creation may temporarily launch a Platform window. The Platform adapter
owns creation, readiness detection, required stop, and creation verification so
the lifecycle does not depend on those representation details.

### Extension identity

Extension IDs use exact, case-sensitive identity during convergence. Cookbook
Source is responsible for spelling an ID in the form understood by its target
Platform; CTK does not silently normalize it.

An implementation must not let an underlying case-insensitive Platform turn an
apparent case-only replacement into accidental removal. It reports the
conflict or otherwise preserves both the Source identity and existing Runtime.

## Lock

### Required Capabilities

- Lock observes actual Platform-owned Runtime state, not only Recipe intent.
- A complete Lock contains every observation required to reconstruct the
  applicable Runtime.
- Required observations depend on the Recipe and Runtime capabilities.
- Named Profile Extension observations are required for selected Profiles.
- Profile-local Settings observations are required when the Profile does not
  inherit default Settings.
- An incomplete Lock is rejected as a trusted reconstruction source.
- Lock refresh does not destroy the previous complete Lock when observation or
  validation fails.

The required Profile observation boundary follows the Recipe. An implementation
may record additional observed Profiles, but reconstruction must not depend on
observations outside the declared completeness boundary.

Lock persistence uses staging and publishes the new snapshot only after
capability-aware validation. Transient observation failures receive a bounded
number of retries. Retry count and delay are implementation policy.

If Lock fails after Apply, the Runtime may already be changed. Apply reports
that state explicitly and does not claim complete success. If Lock fails during
staged Build, the new Distribution is not published.

### Lock mode

Recipe `config.dist-strategy.lock-mode` uses these values:

- `refresh` — observe and publish a fresh Lock; the default
- `reuse` — require and retain a previously complete trusted Lock
- `ask` — require an interactive choice between refresh, reuse, and failure

The names describe the trusted snapshot action. Earlier names `auto` and
`no-lock` are not part of the current Recipe format.

### Representation boundary

Lock completeness is a capability Contract. It does not require a directory,
specific filenames, or a particular serialization.

An implementation documents the representation it trusts. A representation
that cannot prove capability-aware completeness must be treated as untrusted
until the Runtime is observed again.

## Runtime Recovery boundary

Runtime Recovery may be an internal primitive used by CodeVenv activation. It
is not a public Lock Apply Concept API and does not require a public
`apply --lock` command. Public Build and Apply continue to accept their declared
Recipe or Archive sources.

Recovery must use a trusted complete Lock, reconstruct only the intended
Runtime scopes, observe the result again, and make semantic differences visible
before publication or host activation. Exact Extension version reconstruction
belongs to Archive rather than ordinary recovery.

## Implementation-specific resolution

The primary implementation's staging publication, trusted Lock Manifest,
Runtime Recovery verification, Force boundary, and Extension Pool publication
are defined by the
[Go Runtime Convergence Contract](../../go/doc/contract/contract.runtime-convergence.md).
