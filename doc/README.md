# Knowledge.md
============================================================

# Documentation Resolver

The Documentation Resolver routes a current question to the document role that
owns the next useful context.

This document is not a specification or a complete catalog. Accepted concepts,
agreements, rationale, operational guidance, and historical observations remain
owned by their respective documents.

## Responsibility

The Documentation Resolver is responsible for:

- choosing documentation by the reader's question rather than by where an idea
  or implementation detail was created
- exposing the current high-level documentation structure through one
  authoritative Repository Map
- helping readers move from a broad entry point to the smallest relevant
  context
- keeping concept, agreement, rationale, observation, and implementation roles
  distinguishable during navigation

It does not decide the answer to a concept or implementation question. It
directs the reader to the document responsible for that answer.

## Read from the current question

CTK Knowledge is not intended to be read from beginning to end.

```text
Current question
      ↓
README or Documentation Resolver
      ↓
Relevant document role
      ↓
Detailed document
      ↓
Continue working
```

Use directory READMEs as local entry nodes. Once the correct domain or document
role is clear, follow only the relationships useful to the current question.
Searching by a concept name with a repository tool such as `rg` is also an
expected navigation path.

Concrete examples can make a concept easier to understand, but they do not
replace accepted responsibility. Resolve the Concept API or Contract first
when an example, implementation, or historical record could be mistaken for
the current boundary.

## Navigate by question

Start with the question closest to the current work:

- What is CTK, why might I use it, and how do I begin?
  → [README](../README.md)
- How are reusable environments composed, built, and preserved?
  → [Core](core/README.md)
- How are Draft and Inspect review contexts used?
  → [Workbench](workbench/README.md)
- How do CTK concepts connect to documentation, repositories, and IDE
  environments?
  → [Integration](integration/README.md)
- Which behavioral, representation, compatibility, or safety agreement must an
  implementation satisfy?
  → [Contracts](contract/README.md)
- How does one author apply CTK concepts in a concrete Recipe?
  → [Author's Recipe](authors-recipe/README.md)
- Why was a consequential design direction chosen?
  → [Design Notes](design-note/README.md)
- Which reusable operational observation or loose guideline applies?
  → [Notes](note/README.md)
- How should version-matched packaged Knowledge be navigated or compared with a
  local documentation source?
  → [Packaged Documentation Navigation](note/note.packaged-documentation-navigation.md)
- Which candidate concept or direction remains unsettled?
  → [Future](future/README.md)
- How did CTK's design, documentation, or collaboration practice evolve?
  → [Project Knowledge](project-knowledge/README.md)
- How does the current Go implementation realize CTK?
  → [Go implementation](../go/README.md)
- What does the retained Bash implementation show?
  → [Bash implementation](../bash/scripts/README.md)

## Resolve implementation evidence

For an implementation question, resolve evidence in this order unless the
current task establishes a narrower source:

```text
Concept responsibility
        ↓
Shared Contract, when one applies
        ↓
Language README and implementation-specific Contract
        ↓
Source as readable behavioral evidence
```

A language implementation may choose a concrete strategy without redefining a
shared Concept or Contract. Source is useful when the relevant documents do not
resolve the question, but an implementation detail does not automatically
become a universal CTK responsibility.

## Canonical document identity

Repository paths remain concise while documents declare a path-like Canonical
identity on their first content line.

```text
Repository path
doc/core/core.cookbook.md

Canonical identity
Knowledge.core.cookbook.md
```

The surrounding directory may supply context omitted from the physical
filename. The Canonical identity restores that context when a document is
searched, quoted, copied, or read outside its directory.

The [Project Structure Concept
API](integration/integration.project-structure.md) defines this convention,
role-sensitive headings, and the relationship between concept responsibility,
Canonical identity, and repository location.

## Repository Map

This is the authoritative high-level map of CTK documentation responsibilities.
It identifies entry nodes and document categories rather than cataloging every
detailed document.

The [Project Structure Concept
API](integration/integration.project-structure.md) explains how to expand one
Map entry without duplicating this list.

```text
README.md
    Public entry point, project overview, and onboarding

doc/README.md
    Documentation Resolver and authoritative Repository Map

doc/core/README.md
    Stable conceptual model and Core responsibilities

doc/workbench/README.md
    Draft and Inspect review concepts

doc/contract/README.md
    Shared behavioral, representation, compatibility, and safety agreements

doc/integration/README.md
    Mapping CTK concepts to documentation, repositories, and environments

doc/authors-recipe/README.md
    One practical composition of CTK concepts

doc/design-note/README.md
    Design rationale, trade-offs, and design evolution

doc/note/README.md
    Operational observations and loose guidance

doc/future/README.md
    Unsettled candidate concepts and directions

doc/project-knowledge/README.md
    Observations of CTK's design, documentation, and collaboration process

go/README.md
    Current Go implementation principles and declared Kitchen Notes

go/doc/contract/README.md
    Observable agreements specific to the Go implementation

bash/scripts/README.md
    Retained Bash implementation guidance and historical operational evidence

go/internal/* and bash/scripts/*
    Readable implementation evidence used after the responsible documents do
    not resolve the question
```

## Development philosophy

### Principle

CTK prefers adapting familiar concepts rather than inventing new ones.

> Make powerful ideas approachable and adapt them to local needs.

### Inspirations

| Reference | Familiar idea that informed CTK |
| --- | --- |
| Git | Review and persistence lifecycles |
| Spring Boot | Layered configuration |
| Python `venv` | Runtime isolation |
| Docker | Runtime packaging |
| GoF Builder | Incremental construction |

These references are entry points into a CTK concept, not implementation
targets or sources of authority. Explain the useful similarity, make the point
where the analogy stops visible, and return to CTK's own responsibility.

Project Knowledge records additional loose guidance on [Analogy as a
Bridge](project-knowledge/note/note.analogy.md) and [Analogy as a Design Review
Tool](project-knowledge/note/note.analogy-design-review.md).

Prefer semantic responsibility boundaries over duplicated representation when
the current concept and Contract allow it.

## Boundary

The Documentation Resolver does not define:

- the concepts owned by Core, Workbench, or Integration
- the agreements owned by Contracts
- one mandatory reading order
- a detailed-document inventory that duplicates local README routes
- one interpretation of an Author's Recipe or implementation
- unresolved candidates as if they were accepted responsibilities
- collaboration practices recorded by Project Knowledge

When navigation repeatedly requires the same missing route, improve the
responsible README or document relationship before expanding this Resolver into
a second specification layer.
