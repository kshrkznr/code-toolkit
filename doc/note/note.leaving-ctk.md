# Knowledge.note.leaving-ctk.md
============================================================

# Leaving CTK

Using CTK does not require keeping every generated Distribution forever, and a
generated Runtime does not become irreplaceable merely because it contains a
complete editor environment.

Leaving CTK has several independent scopes. Decide whether you want to stop
CodeVenv management, remove generated environments, remove the CLI, or discard
Cookbook and persistence material. These actions do not need to happen
together.

## Restore an activated Platform first

When a Platform command is managed by CodeVenv, leave through its lifecycle:

```text
ctk current
ctk deactivate <platform>
```

Normal deactivation restores the default Runtime imported during activation
and removes CodeVenv management for that Platform. Deactivate each active
Platform separately.

Do not manually remove `current.<platform>`, `origin.<platform>`, transaction
journals, recovery backups, or host connections while the Platform is managed.
Those artifacts may be required to identify the selected Runtime and restore
the imported environment safely. If CTK reports an unhealthy integration, use
the recovery direction it provides rather than improvising filesystem changes.

## When normal deactivation cannot proceed

Do not delete managed state to make deactivation appear complete. Use
`ctk deactivate <platform> --force` only when CTK reports that trusted recovery
is available.

If normal deactivation cannot use the existing origin, but another completed
Distribution already represents the exact host state that should be restored,
the current Go implementation has a specialized manual-maintenance route. See
[Go CodeVenv Origin Recovery](note.go-codevenv-origin-recovery.md) before
changing any reserved path. It requires a complete trusted Lock and is not the
normal deactivation workflow.

If no trusted recovery input remains and an empty host environment is
acceptable, `ctk deactivate <platform> --force-empty` is the explicit
destructive last resort. It does not restore the imported Runtime.

## Remove unused Distributions

A Distribution produced by Build is generated output. An ordinary Distribution
that is not selected by any `current.<platform>` connection can be removed when
you no longer need to launch, inspect, apply, lock, freeze, or archive that
Runtime.

Before removing one, use `ctk current` and confirm that no active Platform
selects it. The `current.<platform>` connection itself is managed CodeVenv
state, not a disposable Distribution.

`origin.<platform>` needs additional care. While its Platform is active, origin
is trusted restoration input even when `current.<platform>` selects another
Distribution. Keep it until normal deactivation has completed. After the
Platform is inactive, an origin left in `dist` is no longer selected CodeVenv
state and may be treated according to whether you still want that Runtime as a
generated artifact.

Removing a Distribution also removes the Lock stored inside that Distribution.
It does not remove the source Recipe, Ingredients, or a separately created
Archive. Preserve or Archive a Runtime first when its observed state will still
matter after cleanup.

## Remove the CLI or workspace

After all managed Platforms have been deactivated, the `ctk` executable can be
removed from `PATH` without changing the restored editor environment.

A generated Distribution with its Direct Launcher can continue to launch
without `ctk` on `PATH`, but it remains a generated environment outside future
CTK lifecycle operations. Confirm the launcher and Runtime you intend to keep
before removing the workspace they live in.

The workspace named by `CTK_HOME` may contain several kinds of material:

- Cookbook source such as Recipes and Ingredients
- generated Distributions
- Locks stored with those Distributions and Archives preserved separately
- documentation and the CTK repository itself

Keep or remove these by the value you want to retain, not as one mandatory
uninstall bundle. A Cookbook can remain useful source history after the CLI and
generated Distributions are gone. A separately created Archive can preserve a
Runtime without requiring its original Distribution to remain active.

## Boundary

This Note is cleanup guidance, not a promise that all Platform-owned state is
contained in CTK's Runtime paths. Authentication, cloud state, caches, or other
IDE data may exist outside the User and Extensions paths managed by CodeVenv.

The accepted activation, selection, deactivation, and safety boundary remains
owned by the [CodeVenv Concept API](../integration/integration.code-venv.md) and
the [Code Environment Contract](../contract/contract.code-environment.md).
