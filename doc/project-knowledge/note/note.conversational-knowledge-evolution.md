# project-knowledge.note.conversational-knowledge-evolution.md
============================================================

# Knowledge Structure Emerged from Conversation

> Source: conversation pickup, 2026-08-02.

CTK's Knowledge structure was not originally planned as an information
architecture.

It emerged from a smaller practical problem: long design conversations were
becoming difficult to continue.

The owner often moves between related viewpoints during a discussion—for
example, from implementation to concepts, documentation, and project workflow,
then back to the original topic. As conversations grew, both the owner and the
conversational AI spent more time reconstructing context.

Rather than trying to change that conversation style, the project repeatedly
adjusted the documentation around it.

```text
Long design conversation
        ↓
Context becomes expensive to reconstruct
        ↓
Small documentation adjustment
        ↓
The next conversation becomes easier to continue
        ↓
Repeated adjustment reveals a structure
```

Documentation Resolver, Knowledge Roles, cross-cutting Concepts, and Node
READMEs became recognizable through this process. The structure was observed
after it had grown; it was not specified before the conversations began.

## Tuned for collaboration, not AI-first documentation

The immediate participants were the owner and conversational AI, so the
structure was tuned through that collaboration. This does not make it an
AI-first documentation system.

The goal was to preserve enough shared context for the particular conversation
to continue without forcing every thought into one long document.

The same organization has appeared usable with other LLMs and sometimes with
human readers. It may also help the owner's future self. These are encouraging
observations, not evidence of a general best practice.

## Roles separate responsibilities, not topics

One consequence was that directories came to represent Knowledge Roles rather
than a partition of Concepts.

A Concept can cross several roles:

- Core describes its accepted responsibility.
- Contract fixes a boundary that implementations must share.
- Design Notes preserve consequential design rationale.
- Notes provide operational or explanatory guidance.
- Future preserves unsettled directions.
- Project Knowledge records how understanding emerged.

This allows a discussion to change viewpoint without requiring one document to
carry every kind of authority.

## Navigation became a two-part movement

The resulting reading practice can be described loosely as:

```text
Question
    ↓
Documentation Resolver
    ↓
Knowledge Role
    ↓
Concept search
```

The Resolver narrows the responsibility to inspect. Searching by Concept then
finds the relevant material within and across those roles.

This appeared to reduce unnecessary reading, limit distraction from adjacent
topics, and make the reasoning scope easier to control. Those effects remain
observations from CTK rather than guaranteed properties of the structure.

## Node READMEs emerged as local entrances

Node READMEs followed the same pattern. They were not introduced because every
directory was expected to have an index.

As a directory accumulated material, the recurring question became:

> What is this a place to read about?

A local README could answer that question without turning the repository map
into a complete document inventory. The current trial suggests that these
entrances may help both discovery and maintenance, but their useful boundary is
still being observed.

## Reusable observation

CTK did not design a documentation structure and then adapt conversation to it.

It improved the conditions of conversation repeatedly, and the documentation
structure became visible afterward.

This Note records one project's origin story and a loose way of interpreting
its current layout. It does not require other projects—or future CTK
directories—to reproduce the same structure.

## Related documents

- [The Resolver Emerged from Repeated Navigation](note.documentation-resolver.md)
- [Naming](note.naming.md)
- [Conversation-first Concept Development](note.development-style.md)
- [README as a Distributed Entry Point](../experiment/experiment.readme-as-distributed-entry-point.md)

