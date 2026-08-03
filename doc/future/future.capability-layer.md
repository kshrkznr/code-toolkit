# Knowledge.future.capability-layer.md
============================================================

# Future: Capability Layer

Capability was explored as a generic Ingredient Layer for reusable feature
sets such as Git, AI, or other development experiences.

Conceptually, a Capability Layer could allow the same named responsibility to
participate across different Runtime and Profile compositions.

## Why it is not Core today

Current Recipes can express the observed needs through:

- Runtime and Profile Ingredients
- explicit Recipe composition
- moving or renaming Ingredient Resources as responsibilities become clearer

A generic Capability Layer would increase freedom, but its responsibility and
composition path are not currently clearer than these simpler operations. The
current Go Recipe also has no separate `capability:` selection.

Capability is therefore not an accepted Ingredient Layer or a current public
Concept API.

## Revisit when

Reconsider Capability only after repeated use shows a responsibility that:

- cannot be expressed clearly as Runtime or Profile
- must be reused across those compositions
- becomes awkward when represented by explicit Recipe selection or ordinary
  file movement
- has a clear resolution and review model

This is a low-priority Future. It is not a roadmap item.
