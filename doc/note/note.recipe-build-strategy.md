# Knowledge.note.recipe-build-strategy.md
============================================================

# Recipe Build Strategy

## Purpose

A Recipe selects Ingredients. Its optional `config` section describes how those
selected Ingredients are assembled into a Dist.

```yaml
os:
platform:
runtime:
profile:

config:
  dist-strategy:
  profile-strategy:
```

The top level remains an intentionally readable declaration of composition.
Strategy changes the resulting Runtime behavior; it is not an additional
Ingredient layer.

---

## Dist strategy

### Extension installation and Pool acquisition

Extension installation and VSIX artifact acquisition are independent choices.

```yaml
config:
  dist-strategy:
    extension-marketplace: true # true | false
    extension-pool: reuse       # reuse | refresh
```

`extension-marketplace` permits CTK to pass an Extension ID to the Platform's
normal install operation. The Platform owns its Repository lookup and install
behavior. This remains `true` by default and is the normal online installation
route.

`extension-pool` controls CTK's own acquisition of exact VSIX artifacts:

| Mode | Behavior |
| --- | --- |
| `reuse` | Use local Pool artifacts but do not download new VSIX files. This is the default. |
| `refresh` | After observing exact installed versions, permit CTK to download and store missing VSIX artifacts. Archive creation may also acquire a missing exact artifact. |

These modes deliberately allow a normal Build to use a Platform Repository
without turning every Build, Apply, Lock, Recovery, or Archive into a CTK-owned
VSIX download. A Recipe that needs offline preparation can explicitly choose
`refresh`; a later `reuse` operation can consume the resulting local Pool.

The strategy describes technical acquisition only. It does not assert that an
Extension may be used with another IDE, copied to another person, or
redistributed. Review the Repository terms, the individual Extension license,
and the target Platform before using a cached artifact outside the Platform's
normal path.

### Lock mode

`lock-mode` chooses which Lock snapshot CTK trusts after Build, Apply, Freeze,
or Archive operates on a Dist.

```yaml
config:
  dist-strategy:
    lock-mode: refresh # refresh | reuse | ask
```

| Mode | Behavior |
| --- | --- |
| `refresh` | Create a fresh Lock from the current Dist. This is the default. |
| `reuse` | Require and reuse the existing trusted Lock. |
| `ask` | Require a TTY and choose `refresh`, `reuse`, or `no` (fail). |

`reuse` and `ask: reuse` trust only a state captured by an explicit `lock`
command. They fail when no complete prior Lock exists. An explicit `lock`
command always creates a fresh Lock.

The earlier Bash names `auto` and `no-lock` were replaced because `no-lock`
actually reused a Lock rather than disabling Lock behavior. They are not part
of the current Recipe format.

### Default Profile extensions

The unprofiled Platform state normally receives resolved Runtime extensions.

```yaml
config:
  dist-strategy:
    default-profile:
      extensions: clean # runtime | clean
```

`runtime` is the default. `clean` removes all Extensions from the Default
Profile while named Profiles continue to receive their resolved
`runtime + profile` extensions.

Default Profile ownership is available per Runtime Artifact:

```yaml
config:
  dist-strategy:
    default-profile:
      settings: runtime
      keybindings: runtime
      snippets: runtime
      tasks: runtime
      mcp: unmanaged
```

`runtime` constructs and observes resolved content. `clean` is a deliberately
managed empty state. `unmanaged` leaves existing Platform content untouched and
unobserved. MCP alone defaults to `unmanaged`; other omitted Artifact modes
remain `runtime` for compatibility.

---

## Profile strategy

Named Profiles inherit Platform defaults unless their strategy declares that an
artifact belongs to the Profile itself.

```yaml
config:
  profile-strategy:
    review:
      settings: profile
      keybindings: profile
```

Each item accepts `default` (the compatibility default), `profile`, or
`unmanaged`.
Supported Artifact choices are `settings`, `keybindings`, `tasks`, `mcp`, and
`snippets`.

CTK writes `useDefaultFlags.<item>: true` for `default`; for `profile`, it
removes that key. The absence of a flag is the Platform's Profile-local mode.

For `unmanaged`, CTK leaves the corresponding Platform flag and content
untouched. It does not construct, observe, verify, recover, Freeze, or Archive
that named Profile Artifact. New Profiles retain the state created by the
Platform. This does not disable the Runtime's shared default Artifact.

### Profile-local Settings lifecycle

`settings: profile` is fully managed by the current Bash implementation.

```text
Recipe profile-strategy
  ↓
Profile-local settings.json
  ↓
Lock: <profile>.settings.jsonc
  ↓
Archive / internal activation Recovery
  ↓
Freeze Draft / Commit
  ↓
profile/<profile>.settings.json Ingredient
```

The Profile-local file contains resolved default settings plus that Profile's
settings and extension settings. A Lock or Archive Recipe that declares
`settings: profile` requires the matching Profile Settings Lock file.

Freeze and Inspect keep default and Profile Settings in one review file:

```text
## runtime.draft.settings.json
...

## profile/review.settings.json
...
```

### Other Profile artifacts

Keybindings, MCP, and Snippets now participate in Profile-local Cookbook
resolution, Runtime I/O, Lock, Recovery, Freeze/View/Sync, Commit, and Archive
through the VS Code-family Adapter. Their physical placement remains an Adapter
representation rather than a Cookbook path contract.

Profile-local capability is still Platform-specific. An Adapter must report an
unsupported Artifact instead of redirecting it into a different scope.

Tasks requires additional Platform observation. On the currently observed
macOS VS Code build, removing `useDefaultFlags.tasks` did not cause `Tasks: Open
User Tasks` to create `profiles/<location>/tasks.json`; it continued to write
the Default `User/tasks.json`, including after restart and from another named
Profile. The Profile model still exposes Tasks ownership, so this is recorded
as unresolved Platform behavior rather than removed from the Strategy. An
Adapter must not claim Profile-local Tasks support until native read and write
behavior is repeatably confirmed.

---

## Activation

CodeVenv activation creates an origin Recipe from the host Platform state.
For every named Profile whose `useDefaultFlags.settings` key is absent, the
origin Recipe declares `profile-strategy.<profile>.settings: profile`.
The moved Profile-local Settings then enter the normal Lock lifecycle.

---

## Future

Default Profile Settings may later gain an explicit strategy, for example a
default state derived from a named Profile. The native Platform default remains
the compatibility behavior until then.
