# Go.contract.cookbook-resolution.md
============================================================

# Go Cookbook Resolution Contract

This Contract specializes the shared
[Cookbook Resolution Contract](../../../doc/contract/contract.cookbook-resolution.md).

## Purity

Go resolves Cookbook Source into an in-memory Runtime Plan without launching a
Platform, mutating a Distribution, acquiring Extensions, or changing Source.

## Settings composition

Go preserves shared Resource resolution order and materializes one Settings
document using:

- recursive object merge
- later-value replacement for arrays, scalars, and `null`
- optional exact-path array `union` declared by the Cookbook-wide Merge Rules
  Kitchen Note

Merge Rules are loaded from `cookbook/kitchen-notes/go.merge-rules.yaml`. They
are not selected per Recipe. Unsupported versions, malformed paths, and Union
rules resolving to non-array values are errors.

## Extension resolution

Go composes exact non-empty Extension IDs, removes duplicates, and emits
deterministic ordering. It retains enough Repository and Pool source identity
for later Platform installation policy.

## Ambiguity

When multiple compatible Ingredient layouts define the same logical Resource,
Go rejects the ambiguity. A Cookbook must remove the duplicate definition
instead of depending on physical search order.

## Runtime Plan

The Runtime Plan retains Recipe provenance, OS, Platform, ordered default and
Profile Artifacts, Profile ownership and inheritance, Extension source policy,
and lifecycle strategy required by later orchestration. It is internal Go data,
not a persisted compatibility format.
