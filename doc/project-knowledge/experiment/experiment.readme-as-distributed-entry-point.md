# project-knowledge.experiment.readme-as-distributed-entry-point.md
============================================================

# Experiment: README as a Distributed Entry Point

## Starting question

Can familiar, directory-local README files provide a lower-cost path into CTK's
documents while preserving the responsibilities of Knowledge and the
Documentation Resolver?

The root README already acts as a navigator. This experiment asks whether that
shape can extend into directories such as `doc/` and `note/`, allowing each
README to introduce the responsibility, culture, and useful reading paths of
its local area.

## Hypothesis

- Entering through a local README may require less initial context than
  entering directly through `doc/README.md`.
- Reusing the familiar convention of reading a README may make exploration
  natural for any collaborator, including people and AI assistants.
- Knowledge may be easier to explain as a documentation practice and concept
  than as the name of one particular file.
- A network of local entry points may distribute part of the Resolver's
  navigation work without distributing or duplicating accepted responsibility.

## Trial 1: Core

The first bounded trial uses Core because `core.md` already acted as the Core
Concept Domain document.

The trial moves that document to `core/README.md` while preserving its canonical
identity, `Knowledge.core.md`. The Core README explains only:

- the responsibility and boundary shared by the Core layer
- question-oriented entry points into Core
- where to return when the question does not belong to Core
- the common documentation guidance for accepted Core responsibilities

The Repository Map points to the Core node rather than cataloging every Core
document. Detailed documents retain their own semantic relationships through
local links such as `See also` when those relationships are useful.

This creates three navigation levels:

```text
Repository Map
    routes to a responsibility node
        ↓
Node README
    supplies shared context and the first local route
        ↓
Detailed document
    may supply useful concept-specific relationships
```

Then observe whether a short request such as this provides enough context:

> Read `doc/core/README.md` and start from the relevant Core responsibility.

The test is not merely whether the request becomes shorter. It is whether the
reader reaches an appropriate document role without losing the current
question or treating the node's navigation as a substitute for the detailed
concept responsibilities.

## What to observe

- whether collaborators reach the intended document layer without extra
  explanation
- whether instructions can become shorter without becoming ambiguous
- whether local READMEs reduce repeated navigation guidance
- whether each README can describe its directory's culture, responsibility,
  and exploration paths without becoming a duplicate specification
- whether node-level Repository Map entries remain sufficient without a central
  detailed-document catalog
- whether optional local relationships are enough for navigation after the
  first node entry, and when a `See also` section is actually useful
- whether repeated navigation patterns reveal a natural boundary between local
  README guidance and the Documentation Resolver

## Trial 2: Notes

The second trial replaces an obsolete mixed `note.md` with `note/README.md`.

Unlike Core, the Note layer does not define accepted Concept APIs. Its node
therefore emphasizes:

- the responsibility and authority boundary of a Note
- practical entry questions rather than a conceptual hierarchy
- flexible local structure rather than a shared heading profile

The obsolete material is retained as a Reviewed Experiment with links to the
durable observations that already moved elsewhere. This tests whether a README
node can clarify a layer without silently promoting its legacy contents.

## Trial 3: Project Knowledge Notes

The third trial adds `project-knowledge/note/README.md` to a directory that
preserves reusable observations about CTK's evolution.

This Node makes the distinction between Note and Experiment locally visible,
then provides question-based routes across the current observations. It does
not treat the Note shelf as a detailed-document catalog or give its contents
the authority of accepted CTK Knowledge.

The trial also inventories the shelf at the same time: unresolved proposals
and hypotheses move to individual Experiments, while reusable observations
remain Notes. This tests whether a Node can clarify a local responsibility by
showing the result of that boundary rather than merely describing it.

## Trial 4: Project Knowledge Design Notes

The fourth trial adds `project-knowledge/design-note/README.md` as the local
entry point for rationale about CTK's Knowledge and documentation model.

Unlike the main Design Note Node, this Node distinguishes meta-level rationale
about documentation structure from rationale about CTK's product and concept
design. The existing Concept Domain rationale receives a subject-specific
physical path so the Node can own the directory-level canonical identity.

## Risks and open questions

- Will responsibilities be repeated or drift between README files?
- How much navigation should remain in the Documentation Resolver?
- Should READMEs share a small common shape, or should each document layer
  choose its own structure?
- Does every document directory need an entry point, or only directories with a
  distinct responsibility and enough internal choice?
- How should a local README defer to accepted Knowledge without presenting a
  competing explanation?

## Boundary

This experiment does not propose that every directory must contain a README or
that the Documentation Resolver should be replaced by a directory tree.

The README is being tested as an entry point, not as a new specification layer.
Any shared template should emerge from repeated use rather than being defined
before the first bounded trials.

## Related documents

- [`Why README Became a Navigator`](../design-note/design-note.documentation-onboarding.md)
  — established observation about the responsibility of the root README
- [`Resolver Reading Strategy`](experiment.resolver-reading-strategy.md) —
  ongoing observation about how entry paths affect interpretation
- [`Discussion Context Resolver`](experiment.discussion-context-resolver.md) —
  separate experiment about preparing topic-specific context before discussion
