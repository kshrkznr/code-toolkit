# Knowledge.contract.runtime-artifacts.md
============================================================

# Runtime Artifact Contract

This Contract defines composition, ownership, observation, and safety boundaries
for Runtime content beyond Settings and Extensions.

Implementations may resolve and apply these Artifact contents through Runtime
I/O, Lock observation, Workbench editing, Recovery, and Archive. Unsupported
Platform capabilities remain distinguishable from valid empty content.

## Artifact Composition Principle

JSON syntax does not imply one shared merge semantic.

Each Artifact defines a **Composition Unit** matching the unit managed by its
Platform representation. CTK composes those units deterministically but does
not infer their Platform-specific meaning.

- Unit contents are opaque unless the Artifact Contract explicitly says
  otherwise.
- Semantic identity is not guessed from fields such as `key`, `command`,
  `label`, or `id`.
- Composition never performs unrequested semantic deduplication.
- Runtime convergence applies the complete resolved Artifact. Composition
  append does not mean appending repeatedly to the existing Runtime file.
- Artifact-specific exceptional semantics may be declared later as a Kitchen
  Note instead of making the base resolver increasingly intelligent.

Settings remains intentionally different: its Composition Unit is a property
path and its existing recursive merge plus declared Merge Rules continue to
apply only to Settings.

## Shared Cookbook Source Resolution

### Accepted Source Compatibility

Keybindings, Snippets, and other JSON Runtime Artifacts follow the same
Cookbook source principles as Settings:

- the existing three Ingredient layout families
- `.json` and `.jsonc` using the accepted VS Code-family JSONC boundary
- Base, OS Variant, then Platform Variant resolution
- deterministic Cookbook Core layer and Recipe array order
- Runtime, Profile, Extension, OS, and Platform Ingredient participation when
  applicable
- missing Resources are valid empty participation
- malformed present Resources are errors
- competing compatible layouts remain an implementation-defined ambiguity

Examples for Keybindings:

```text
ingredient/runtime.golang.keybindings.jsonc
ingredient/runtime/golang.keybindings.jsonc
ingredient/runtime/golang/keybindings.jsonc
```

Examples of variants:

```text
ingredient/runtime.golang.macos.keybindings.jsonc
ingredient/runtime.golang.code.keybindings.jsonc
```

## Profile Ownership

Keybindings, Snippets, Tasks, and MCP follow the same three-state Profile
ownership model used by Settings and the other managed Artifacts:

```yaml
config:
  profile-strategy:
    review:
      keybindings: profile
      snippets: default
      mcp: unmanaged
```

- `default` uses the Platform's inherited default Artifact.
- `profile` requires a Profile-local resolved Artifact and corresponding
  Runtime observation.
- `unmanaged` leaves that Profile Artifact outside CTK ownership.
- The Platform Adapter owns the physical inheritance representation, currently
  VS Code-family `useDefaultFlags`.
- Changing ownership may require Platform database coordination even when the
  Artifact file itself is safe to write while the IDE is running.

For `unmanaged`, CTK does not:

- resolve or construct Profile-local content
- change the corresponding Platform inheritance or ownership flag
- observe it into Lock
- compare or recover it during semantic verification
- expose it as managed Freeze, View, Sync, or Archive content

Existing Platform state is left untouched. For a newly created Profile, the
Platform-created state remains authoritative. Missing strategy values continue
to mean `default`; `unmanaged` must be explicit. A named Profile whose MCP
strategy is omitted therefore follows the Default MCP ownership described
below.

This strategy applies only to the named Profile Artifact. It does not suppress
construction or observation of a shared default Artifact that may also be used
by other Profiles. A future need to exclude an Artifact for the entire Runtime
requires a separate Runtime-level strategy rather than overloading
`profile-strategy`.

### Default Profile ownership

Default Runtime Artifacts use `config.dist-strategy.default-profile`:

```yaml
config:
  dist-strategy:
    default-profile:
      extensions: clean
      settings: runtime
      keybindings: runtime
      snippets: runtime
      tasks: runtime
      mcp: unmanaged
```

- `runtime` resolves, constructs, and observes the Runtime Artifact.
- `clean` deliberately converges the Artifact to a managed empty state.
- `unmanaged` does not resolve, construct, mutate, or observe the existing
  Default Artifact.

`clean` is managed empty state and is not equivalent to `unmanaged`.

The compatibility default is `runtime` for existing Artifact kinds except MCP.
The MCP default is `unmanaged`: MCP content enters CTK ownership only through
an explicit Recipe choice. A named Profile using `default` follows that Default
Artifact whether it is managed or unmanaged. A named Profile using `profile`
may still explicitly manage Profile-local content even when the Default
Artifact is unmanaged.

The exact physical Profile paths remain Platform Adapter representation, not
Cookbook Source Contract.

## Keybindings

### Composition Unit

The Keybindings Composition Unit is one root-array element. Each element is an
opaque JSON value, normally a binding object.

Resolved Resources are concatenated in Cookbook resolution order:

```text
Resource A entries
    + Resource B entries
    + Resource C entries
    = resolved Keybindings array
```

Composition is plain append:

- preserve Resource order
- preserve entry order within each Resource
- preserve exact duplicates
- do not infer identity from `key`, `command`, `when`, or other fields
- do not interpret apparent conflicts

Build and Apply replace the managed Runtime Artifact with this complete
resolved array. Repeated Apply therefore does not accumulate additional copies
beyond those already present in the Cookbook resolution result.

Profile-local Keybindings follow the same ownership model and composition
direction as Profile-local Settings. Physical placement and Lock recovery
belong to the applicable Platform Adapter.

## Snippets

### File Collection

Snippets are a named file collection rather than one monolithic Artifact.

The first identity is the complete Platform filename, preserving the
difference between language snippet files and global `.code-snippets` files:

```text
snippets/go.json
snippets/markdown.json
snippets/project.code-snippets
```

Compatible Ingredient representations extend the Settings layout families by
placing the filename after the `snippets` Resource kind:

```text
ingredient/runtime.golang.snippets.go.json
ingredient/runtime/golang.snippets.go.json
ingredient/runtime/golang/snippets/go.json

ingredient/runtime.golang.snippets.project.code-snippets
ingredient/runtime/golang.snippets.project.code-snippets
ingredient/runtime/golang/snippets/project.code-snippets
```

Variants remain before the Resource kind, as with Settings:

```text
ingredient/runtime.golang.macos.snippets.go.json
ingredient/runtime.golang.code.snippets.go.json
```

The exact filename suffix is data, not a Variant selector.

### Snippet Composition Unit

Within one file, the Composition Unit is one top-level Snippet dictionary
entry:

```text
filename
    -> Snippet name
        -> complete Snippet definition
```

For the same filename and Snippet name, the later Resource replaces the entire
Snippet definition. CTK does not recursively merge `prefix`, `body`,
`description`, or other definition fields. A later Ingredient must therefore
contain the complete definition it intends to own.

- Different Snippet names coexist.
- The same Snippet name is not a conflict error.
- The later complete definition wins.
- Snippet definition internals remain opaque JSON.
- Language `.json` and global `.code-snippets` files never collapse into one
  identity.

### Workbench Direction

Workbench keeps Snippets as separate files instead of flattening the whole
collection into one `snippets.draft.md`.

Lock, View, and Sync observe at least:

- filename addition
- filename absence
- complete file content change
- Snippet-entry change within one file

Rename is observable as old-name absence plus new-name addition. Runtime
observation does not invent intent. Whether accepting only the removal side
keeps an empty Resource or deletes its file is an implementation-specific
Commit decision and must be documented.

## Tasks

### Scope

CTK manages only VS Code-family User and Profile Tasks as Runtime Artifacts.

Workspace-owned Tasks are outside this Contract, including:

- `.vscode/tasks.json`
- the `tasks` section of a `.code-workspace` file

Those files belong to a project or Workspace rather than an IDE Runtime. A
future Project or Workspace Artifact may define their lifecycle separately,
but Runtime Build, Apply, Lock, Freeze, Recovery, and Archive do not discover
or modify them.

### Composition Units

A Tasks document remains a root object. Within it, CTK recognizes only the
minimum structural envelope needed to compose the two root arrays:

- one `tasks` array element is one opaque Composition Unit
- one `inputs` array element is one opaque Composition Unit

Both arrays use plain append in Cookbook resolution order:

- preserve Resource order
- preserve element order within each Resource
- preserve exact duplicates
- do not infer Task identity from `label`, `type`, `command`, or other fields
- do not infer Input identity from `id` or its references
- do not validate or resolve Task variables as semantic identities

CTK does not understand, deduplicate, override, or reconcile the meaning of
individual Tasks and Inputs. Platform-specific Task types and fields remain
opaque input supplied by the user and interpreted by the Platform.

Build and Apply write the complete resolved Tasks document. Append is a
Cookbook composition rule and does not cumulatively append to the existing
Runtime document across repeated operations.

### Root and Platform-specific Content

The root `version` is document metadata rather than an append unit. Input
Resources must provide a compatible value; CTK must not combine conflicting
versions by resolution order.

Tasks comparison is semantic at the empty boundary. An absent document, an
empty root object, and a document containing only the default `2.0.0` version
with empty `tasks` and `inputs` arrays represent the same empty Tasks state.
Empty `tasks` or `inputs` arrays are also ignored when comparing an otherwise
non-empty document. This normalization affects Workbench, Recovery, and Archive
verification only; Lock observation retains the Platform's actual document and
Build or Apply may still write the explicit VS Code-compatible envelope.

Other root content, including `windows`, `osx`, and `linux` blocks, is accepted
as opaque Platform configuration. CTK does not reinterpret those blocks as
Cookbook Variants and does not inspect their Task semantics. Authors may use
Cookbook OS and Platform Variants to organize source files, but choosing and
editing that representation is a Workbench and user responsibility.

### Workbench Direction

Workbench represents one resolved User or Profile Tasks document as one
`tasks.draft.md` Artifact. It exposes the document for human or AI editing but
does not turn labels, Input IDs, commands, or OS blocks into CTK merge
identities.

### Observed VS Code-family placement

The Default User Tasks document is observed at:

```text
<User>/tasks.json
```

Workspace Tasks at `.vscode/tasks.json` are a different project-owned
Artifact and remain outside this Contract.

VS Code's Profile model exposes Tasks as inheritable Profile content and its
internal Profile representation has a Profile-specific Tasks Resource.
However, current macOS observation did not produce
`<User>/profiles/<location>/tasks.json`: `Tasks: Open User Tasks` continued to
write `<User>/tasks.json` after removing Tasks inheritance, including after an
IDE restart and while using another named Profile.

Therefore Profile-local Tasks placement and effective Platform behavior remain
**observation required**. CTK must not assume that the nominal Profile Resource
is honored by the installed Platform version. The Platform Adapter may report
Profile-local Tasks as unsupported or unavailable until repeatable native I/O
is confirmed. This uncertainty does not change the accepted Cookbook
composition or Workbench representation.

## MCP Secret Boundary

See [Secret Management Design Note](../design-note/design-note.secret-management.md)
for the responsibility rationale behind this Contract boundary.

Initial MCP support does not provide Secret management.

MCP is also opt-in for CTK ownership: omitted
`dist-strategy.default-profile.mcp` resolves to `unmanaged`. This prevents CTK
from discovering and propagating existing Default MCP configuration merely
because MCP Artifact support exists.

- MCP configuration is handled as ordinary Cookbook and Runtime data.
- CTK does not interpret URLs, user names, headers, keys, tokens, or other
  values as credentials.
- CTK does not encrypt, redact, mask, separate, or automatically convert those
  values into environment-variable or external-Secret references.
- Lock, Freeze, View, Sync, Archive, and other observation paths may preserve
  and expose values exactly as supplied or observed.
- A Secret written directly into Cookbook content may consequently appear in
  plain text in Git history, Drafts, Locks, Archives, logs, or other generated
  Artifacts.

Users are responsible for choosing and managing any reference mechanism
supported by their MCP implementation and execution environment. CTK must not
imply that ordinary Artifact handling protects embedded Secrets.

When a Recipe explicitly selects managed Default or Profile-local MCP content,
the user accepts that the ordinary Artifact lifecycle may reproduce its values.
CTK remains responsible for honoring `unmanaged` and not observing that scope;
credential secrecy within an explicitly managed scope remains the user's
responsibility.

This boundary does not prohibit future integration with Secret references or
providers. It does not promise that CTK will become a Secret manager. The
preferred direction is for CTK to preserve opaque references while the target
Platform, Extension, execution environment, or external provider resolves and
protects their values. Any future integration should be a deliberate shared
extension point or Adapter rather than an implicit MCP-only interpretation.

## Lock Compatibility

Adding Runtime Artifact fields does not inherently require a Lock format-version
increase when the representation can distinguish absent, empty, unsupported,
and unobserved content.

- An absent field in an old Lock means that Artifact was not observed by that
  implementation.
- Absence alone does not invalidate every old Lock.
- If the current Recipe requires a Profile-local Artifact, a trusted Lock must
  contain that observation; otherwise it is incomplete for that Plan.
- Valid empty content must remain distinguishable from unobserved or
  unsupported content.
- New Artifact observations participate in Recovery, semantic verification,
  Freeze/View/Sync, and Archive completeness once implemented.

## Open Questions

- equivalent Runtime placement on additional Platform and OS combinations
- unsupported versus unavailable Runtime I/O reporting
- MCP Composition Units and Platform differences
- whether any Artifact later needs a declared Kitchen Note beyond these simple
  base semantics

## Implementation-specific resolution

The primary implementation's supported Artifact matrix, Runtime paths, Freeze
Commit behavior, and trusted Lock representation are defined by the
[Go Runtime Artifact Contract](../../go/doc/contract/contract.runtime-artifacts.md).
