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

Go ignores empty lines, treats every non-empty line as an exact Extension ID,
preserves spelling and letter case, removes duplicate IDs after composition,
and produces deterministic Extension ordering.

The current line-oriented Resource does not interpret comments.

## Unknown and empty content

Missing Resources contribute valid empty content. Present malformed Resources
are errors. Go does not fall back to an alternate interpretation of malformed
Source.

Unknown strategy fields or values are not silently treated as supported
behavior. Read-modify-write workflows preserve unknown source only where the
owning operation explicitly declares that responsibility.
