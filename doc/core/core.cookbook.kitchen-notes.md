# Knowledge.core.cookbook.kitchen-notes.md
============================================================

## Kitchen Notes

### Why

Some Cookbook interpretation rules are implementation-specific.

Kitchen Notes provide a place to record those local rules without changing the Core concepts.

### Definition

Kitchen Notes are optional implementation-specific notes associated with a Cookbook.

They supplement how a Cookbook is interpreted during the Build process.

### Responsibility

Kitchen Notes belong to the Cookbook.

They may be consulted while building from the Cookbook.

They supplement Cookbook interpretation only.

Kitchen Notes do not redefine Core concepts or the contracts outside the Cookbook.

Kitchen Notes are adopted independently by each language implementation. An
implementation documents the Kitchen Notes it applies in that language's
README. That declaration is the public boundary of its additional Cookbook
interpretation.

An implementation is not required to adopt, reject, or emulate Kitchen Notes
declared by another implementation. Notes it does not declare are ignored, and
the implementation is understood to use its direct, unextended behavior for
that concern.

### Notes

Whether Kitchen Notes exist, how they are represented, and what they contain are implementation-specific.

Some implementations may not require Kitchen Notes at all.

A Kitchen Note does not require a special filename or directory. An ordinary
Note may explicitly declare that it describes a Kitchen Note. If the collection
grows, a repository may organize those Notes into a Kitchen Notes category or
directory without changing the concept.

At present, Merge Rules are the only defined example of Kitchen Notes. Variant
selection and precedence remain Cookbook Core behavior and are not Kitchen
Notes.

See [Merge Rules as a Kitchen Note](../note/note.merge-rules.md) for the
reusable rationale and the boundary between Core resolution and
implementation-specific combination semantics.

Project documentation may identify a currently recommended language
implementation based on its adopted Kitchen Notes and overall maturity. That
recommendation is guidance, not Cookbook Core behavior and does not create an
adoption obligation for other implementations.
