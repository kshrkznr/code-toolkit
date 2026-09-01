# Go.contract.cookbook.md
============================================================

# Go Cookbook Contract

This Contract specializes the shared
[Cookbook Representation Contract](../../../doc/contract/contract.cookbook.md).

## Recipe and Settings parsing

Go loads `.yaml` and `.yml` Recipes by explicit path or Cookbook selection. It
uses document `name`, `os`, and where required `platform` for identity rather
than inferring identity from the filename.

Go parses Settings as JSONC in-process. It accepts standard JSON plus line
comments, block comments, and trailing commas. It does not promise the broader
JSON5 syntax accidentally accepted by the retained Bash toolchain.

## Compatible Ingredient layouts

Go supports every flat and directory layout declared by the shared Contract.
If more than one compatible path defines the same logical Resource, Go reports
an ambiguity error rather than selecting one by search order.

## Extension Resources

Go ignores empty lines and interprets every non-empty Runtime and Profile
Extension Resource line as either an exact Extension ID or an Extension Set
declaration with the exact, case-sensitive `set:<name>` form. Extension IDs
preserve spelling and letter case. Go removes duplicate IDs after composition
and produces deterministic Extension ordering.

The current line-oriented Resource does not interpret comments.

An Extension Set name matches `[A-Za-z0-9][A-Za-z0-9._-]*`; lookup preserves
case and performs no normalization. For Set `<name>`, Go resolves exactly these
compatible candidates under the Ingredient root:

```text
extension-set.<name>.extensions
extension-set/<name>.extensions
extension-set/<name>/extensions
```

More than one matching layout is an ambiguity error. An absent or empty
Extension Set Resource is a valid empty contribution; the declaration itself
establishes Set identity and no separate existence manifest is used.

Extension Set Resources contain only concrete Extension IDs. Go rejects nested
`set:` declarations and does not resolve Extension variants. Runtime and
Profile declarations expand one level, after which concrete IDs use the normal
Extension Ingredient resolution path. The Runtime Plan remains concrete-only
and records both the declaring Runtime or Profile Resource and each present
Extension Set Resource as Sources. Its review-only Extension origin metadata
maps each concrete ID to every direct or Set Resource that declared it; this
metadata is not persisted into Runtime, Lock, Distribution, or Archive state.

CTK v0.6.2 reserved `set:` declarations as a downgrade-safety guard. Cookbook
Sources that use Extension Sets are unsupported by v0.6.1 and earlier. The
exact prefix and canonical namespace define the implicit v1 representation;
an incompatible future interpretation must use a new explicit representation
rather than silently changing these semantics.

## Unknown and empty content

Missing Resources contribute valid empty content. Present malformed Resources
are errors. Go does not fall back to an alternate interpretation of malformed
Source.

Unknown strategy fields or values are not silently treated as supported
behavior. Read-modify-write workflows preserve unknown source only where the
owning operation explicitly declares that responsibility.
