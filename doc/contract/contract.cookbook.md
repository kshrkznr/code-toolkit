# Knowledge.contract.cookbook.md
============================================================

# Cookbook Representation Contract

This document records compatibility boundaries for representing and resolving
Cookbook resources.

The contract describes observable Cookbook behavior. It does not require a
specific parser, programming language, or repository implementation.

Cookbook is a Required Source Compatibility boundary. Implementations must be
able to use the same Recipe and Ingredient files directly; the representation
details documented here are therefore shared rather than one implementation
profile.

## Recipe

### Required Contract

- A Recipe is represented as YAML.
- A Recipe declares its identity with `name`.
- A Recipe may select an operating system with `os`.
- A Recipe may select an IDE Platform command with `platform`.
- `runtime` is an ordered list of Runtime Ingredient names.
- `profile` is an ordered list of Profile Ingredient names.
- Recipe names may be shared across operating-system variants.
- Recipe configuration under `config` controls Distribution assembly strategy;
  it does not redefine the responsibility of an Ingredient.

Example:

```yaml
name: vscode-golang
os: macos
platform: code

runtime:
  - common
  - golang

profile:
  - review
```

### Current Resolution

- Recipe files use the `.yaml` or `.yml` extension.
- A Recipe may be loaded by file path or selected from the Cookbook Recipe
  directory.
- Recipe variants are currently resolved by `name` and `os`.

### Open Questions

- Whether `platform` must also participate in Recipe variant identity.
- Which Recipe fields are required for every operation, as opposed to only a
  specific lifecycle.
- Whether unknown fields must be preserved when a Recipe is read and written.

---

## Ingredient Identity

### Required Contract

- A Recipe selects Ingredients by name.
- An Ingredient name is connected to its resources through the Layer and Name.
- Ingredient resources may be organized in flat or directory-based layouts.
- A compatible implementation must resolve all of the canonical layouts
  documented for each resource type.
- An Ingredient is valid when none of its Resource files currently exist.
- A Recipe is valid when every selected Ingredient currently resolves to zero
  Resources. Recipe composition may exist before concrete Resources are added.
- Directory-based layouts are supported so that a growing Ingredient can keep
  related resources together without changing its Recipe identity.

The following forms therefore refer to the same Runtime Ingredient resource:

```text
ingredient/runtime.golang.extensions
ingredient/runtime/golang.extensions
ingredient/runtime/golang/extensions
```

Moving between these layouts must not require changing the Recipe entry:

```yaml
runtime:
  - golang
```

### Current Resolution

- Flat layouts are commonly used for small Ingredients.
- Directory layouts allow an Ingredient to grow without crowding the root
  Ingredient directory.

### Open Questions

- None of the supported layouts has exclusive ownership of an Ingredient.
  Behavior when the same Resource exists in multiple layouts is intentionally
  implementation-defined and is not a Required Source Compatibility rule.

## Extension Ingredient

### Required Contract

- Extension resources use a line-oriented representation.
- Each non-empty line represents one Platform extension identifier.
- Extension identifiers are passed to the Platform without case normalization.
- Implementations must preserve identifier spelling, including letter case.

Example:

```text
golang.go
```

### Current Resolution

Runtime extension resources are resolved from these compatible layouts:

```text
ingredient/runtime.<name>.extensions
ingredient/runtime/<name>.extensions
ingredient/runtime/<name>/extensions
```

Profile extension resources use the equivalent `profile` layouts.

### Open Questions

- A future extension format may define comments. Comment support is not
  prohibited by the Concept API, but the current line-oriented format does not
  define or interpret comment syntax.

---

## Settings Ingredient

### Required Contract

- Settings Resources use the JSON syntax accepted by VS Code-family Settings:
  standard JSON with optional line comments, block comments, and trailing
  commas.
- Full JSON5 syntax is not required. In particular, implementations need not
  accept single-quoted strings, unquoted property names, hexadecimal numbers,
  `Infinity`, or `NaN`.
- Resolved Settings are provided in a representation the target Platform can
  read.
- Settings Resource resolution order is deterministic and preserved.
- OS and Platform variants may specialize an Ingredient without changing the
  Ingredient name selected by the Recipe.

### Current Resolution

For a Runtime Ingredient named `<name>`, implementations resolve Settings from
these layout families:

```text
ingredient/runtime.<name>.settings.json
ingredient/runtime.<name>.settings.jsonc
ingredient/runtime/<name>.settings.json
ingredient/runtime/<name>.settings.jsonc
ingredient/runtime/<name>/settings.json
ingredient/runtime/<name>/settings.jsonc
```

Variant resources insert an OS or Platform selector before `settings`, for
example:

```text
ingredient/runtime.common.macos.settings.json
ingredient/runtime.common.code.settings.json
```

The required base Settings resolution order is:

```text
OS
Platform
Runtime Extension settings
Runtime settings
default Profile settings
```

### Open Questions

- Whether JSONC comments and formatting must survive resolution and persistence.
- The representation used to provide ordered Settings to a Platform may differ
  by implementation.

### Recommended composition strategy

When the target Platform consumes one Settings document, implementations should
compose the ordered Resources using Platform-compatible later-value precedence:

- objects merge recursively
- arrays are replaced by the later value
- scalar values are replaced by the later value
- `null` is retained as the later value

The Required Contract is Platform readability plus resolution-order
preservation; a future Platform may consume the ordered Resources without using
this exact materialized merge.

---

## Distribution Strategy

### Required Contract

- Recipe `config` may influence how a Distribution is assembled.
- Omitted strategy values have defined defaults.
- Strategy fields must not change the identity or internal responsibility of an
  Ingredient.

### Current Resolution

The reference implementation currently recognizes:

- `config.dist-strategy.extension-marketplace`
- `config.dist-strategy.lock-mode`
- `config.dist-strategy.default-profile.extensions`
- `config.profile-strategy.<profile>.<content>`

Profile Artifact ownership accepts:

- `default`: use the Platform default Artifact
- `profile`: construct and observe Profile-local content
- `unmanaged`: do not construct, observe, switch, verify, recover, or archive
  that named Profile Artifact

An omitted value remains `default`. `unmanaged` is explicit and does not
disable the shared default Artifact for the Runtime.

Default Profile Artifact ownership uses
`config.dist-strategy.default-profile.<content>`:

- `runtime`: resolve, construct, and observe Runtime content
- `clean`: converge to a managed empty state
- `unmanaged`: do not resolve, construct, mutate, or observe existing content

Existing Artifact kinds retain `runtime` as their compatibility default except
MCP. Omitted `default-profile.mcp` means `unmanaged`, so MCP enters the normal
Cookbook, Lock, Workbench, Recovery, and Archive lifecycle only when explicitly
enabled by a Recipe.

This ownership boundary is not a Secret classification mechanism. See the
[Secret Management Design Note](../design-note/design-note.secret-management.md)
and the
[Runtime Artifact Contract](contract.runtime-artifacts.md#mcp-secret-boundary).

### Open Questions

- Which strategy fields are stable public Recipe format.
- How implementations should report unknown strategy fields or values.

## Implementation-specific resolution

The primary implementation's JSONC parser boundary, duplicate Extension
handling, and ambiguous Resource behavior are defined by the
[Go Cookbook Contract](../../go/doc/contract/contract.cookbook.md).
