# Knowledge.integration.project-structure.md
============================================================

# Concept API: Project Structure

## Concept Domain

Integration

## Definition

Project Structure maps CTK concepts and document responsibilities to
discoverable repository documents.

It connects three views:

```text
Concept responsibility
        ↓ represented by
Canonical document identity
        ↓ located through
Repository path and headings
```

The API helps people, AI agents, and repository tools answer:

- What kind of document is this?
- Which Concept Domain or Concept API does it represent?
- What responsibility should its contents have?
- Where is the next relevant document?

Project Structure does not define one canonical layout for every CTK project,
Cookbook, Recipe, or source implementation.

## Responsibility

Project Structure defines the navigation-facing shape of CTK documentation:

- a canonical document identity on the first content line
- explicit Concept Domain and Concept API headings for documents at those
  semantic layers
- directory-local Node READMEs that expose one responsibility boundary and
  question-oriented routes into detailed documents
- use of the authoritative Repository Map to resolve document responsibilities
  to current locations
- role-sensitive heading profiles
- enough consistency for concept search, review, and AI-assisted navigation

It does not define the concepts described by those documents. Core, Workbench,
and Integration documents remain responsible for their own Concept APIs.

## Document identity and repository path

A repository path is the physical location used by the current checkout.

A canonical document identity is a path-like conceptual name used inside the
document and in documentation references.

For example:

```text
Repository path
doc/core/core.cookbook.md

Canonical document identity
Knowledge.core.cookbook.md
```

The repository path may omit context already supplied by its directory. The
canonical identity restores that context and stays useful when a document is
quoted, copied, packaged, or read without its containing directory.

### First-line identity

The first content line of a CTK Knowledge document should be a level-one
heading containing its canonical document identity.

```md
# Knowledge.core.cookbook.md
============================================================
```

`First content line` allows a format-required preamble when one is genuinely
needed. Ordinary Markdown documents should place the identity on line one.

The identity is metadata expressed through normal Markdown. It is not a
filesystem path, import path, URL, or promise that the physical file will never
move.

## Concept layer headings

Documents that represent a Concept Domain or Concept API should say so in a
semantic heading. A filename alone should not be the only evidence of its
conceptual level.

### Concept Domain document

```md
# Knowledge.core.md
============================================================

# Concept Domain: Core
```

A Concept Domain document provides a high-level responsibility boundary and
routes readers to the Concept APIs it contains.

### Concept API document

```md
# Knowledge.core.cookbook.md
============================================================

# Concept API: Cookbook

## Concept Domain

Core
```

A Concept API document defines one public conceptual interface. It declares its
Concept Domain explicitly so the relationship remains visible outside README's
catalog.

### Combined documents

One physical document may currently introduce a Concept Domain and more than
one closely related Concept API. In that case, it should expose both levels
through headings rather than pretending they are one concept.

```md
# Concept Domain: Workbench

## Concept APIs

### Draft
### Inspect
```

Splitting the document is optional. Discoverable responsibility is the goal;
one-file-per-concept is not a Project Structure requirement.

## Node README

A documentation directory exposes its local entry node through `README.md`.

The Node README connects a broad Repository Map entry to the detailed documents
inside one Concept Domain or document-role collection:

```text
Repository Map
    routes to a responsibility node
        ↓
Node README
    states the local responsibility and boundary
    routes by the reader's question
        ↓
Detailed document
    owns the concept, agreement, rationale, or observation
```

A Node README is responsible for:

- declaring the Canonical identity of the local node when the documentation
  role uses one
- stating the responsibility and authority boundary shared by the directory
- providing a small set of question-oriented entry routes
- returning the reader to the Documentation Resolver when the question belongs
  elsewhere
- describing local documentation guidance only when that guidance helps readers
  interpret the category

A Concept Domain README may also be the Domain document, as in Core, Workbench,
or Integration. A supporting-role README, such as Contract, Design Note, Note,
or Future, introduces the authority and reading style of that document
collection without claiming to define a Concept API.

The README is an entry node, not a second specification or exhaustive document
catalog. Detailed documents retain ownership of their contents and may expose
additional local relationships when those links provide a useful next step.

`README.md` is the adopted entry-point convention for CTK documentation nodes.
This does not require every repository, implementation, generated-data, or
temporary-work directory to become a documentation node.

## Supporting document roles

Not every document represents a Concept Domain or Concept API.

Supporting documents declare a canonical identity and use headings appropriate
to their role. They should not adopt a `Concept API` heading merely to appear
more authoritative.

### Core, Workbench, and Integration

These documents describe accepted public responsibilities. Their leading
headings should make the Concept Domain or Concept API explicit. Definition,
responsibility, boundary, and related-concept headings should remain relatively
stable once the concept is accepted.

### Contract

Contract documents describe explicit behavioral or representation agreements.
Their leading semantic heading names the subject and `Contract`.

Capability boundaries, required behavior, current resolution, and open
questions should remain easy to distinguish. Contract heading changes deserve
the same deliberate review as changes to agreed wording.

### Design Note

Design Notes explain rationale, alternatives, trade-offs, and design evolution.
Their title should expose the decision or question being explained.

Background, observation, decision, trade-off, and boundary are useful shapes,
but a Design Note does not need to imitate a Concept API specification.

### Note

Notes preserve reusable observations, operational knowledge, and loose
guidance. Their title should expose the topic a reader is likely to search for.

Context, observation, guidance, and boundary are useful shapes. The local
subject matters more than a fixed template.

### Experiment

Experiments preserve a question and its thought path. Their title should make
the hypothesis, observation, or experiment recognizable.

Status, question, hypothesis, procedure, observations, discarded paths, and
outcome may be used when they help reconstruct the work. A Reviewed Experiment
may link to durable outcomes without rewriting its history as a specification.

### Future

Future documents describe unsettled candidate concepts or directions. Their
heading should make the candidate visible without implying an implementation
plan or accepted responsibility.

Questions, current evidence, boundary, and reconsideration conditions are
useful shapes.

### Author's Recipe

Author's Recipe documents show how one author applies CTK concepts. Their
headings should distinguish personal interpretation, trade-offs, and examples
from accepted CTK responsibility.

## Heading strength follows responsibility

Heading consistency is not applied equally to every category.

```text
Accepted responsibility / agreement
Core · Workbench · Integration · Contract
        ↓ comparatively stable

Rationale / reusable observation
Design Note · Note · Author's Recipe
        ↓ deliberately structured but flexible

Exploration / candidate direction
Experiment · Future
        ↓ structure follows the local question
```

This is a review gradient, not a hierarchy of importance.

Core and Contract headings act as a stronger navigation interface because
readers use them to find accepted responsibilities and agreements. Note and
Experiment headings can evolve with the observation or thought record.

Consistent headings are a means of making responsibility discoverable. They
are not a versioned compatibility contract and do not make prose executable.

## Repository Map

The [Documentation Resolver Node](../README.md) owns the authoritative
Repository Map. Project Structure does not duplicate that list.

This API describes how to interpret and expand one Map entry. For example:

```text
Repository Map entry
doc/core/README.md
    Stable conceptual model and Core responsibilities
        ↓ opens

Node README
    Canonical identity: Knowledge.core.md
    Concept Domain: Core
    question-oriented routes
        ↓ selects a detailed document

Repository path
doc/core/core.cookbook.md
        ↓ declares

Canonical document identity
Knowledge.core.cookbook.md
        ↓ exposes

Concept API heading
Concept API: Cookbook
        ↓ belongs to

Concept Domain
Core
```

The same shape can be applied to another category without requiring the same
heading profile. A Note still declares its canonical identity and searchable
topic, but it does not claim to define a Concept API.

When the authoritative Repository Map changes, canonical identities, links,
and Resolver routes should be reviewed together. Project Structure examples
need updating only when the convention itself changes.

## Navigation and search

Project Structure supports several equivalent entry paths:

```text
Reader question ──► Documentation Resolver ──► Node README ──► detailed document
Map entry ────────► Node README ─────────────► question-oriented route
Concept name ─────► repository search ───────► heading / identity
Repository path ──► first-line identity ─────► conceptual context
README catalog ───► Concept API link ────────► accepted responsibility
```

No path requires reading the repository from beginning to end.

Canonical identities and semantic headings should use the vocabulary readers
are likely to encounter in README, the Resolver, discussions, and source
reviews.

## Current adoption

Node READMEs are the adopted local entry points for CTK documentation domains
and role collections. Core, Workbench, Integration, Contract, Design Note,
Note, Future, and Author's Recipe currently expose this structure. Project
Knowledge also uses a directory README as its local entry point.

Many detailed documents already declare a path-like identity and descriptive
title. Concept Domain and Concept API labels are not yet explicit in every
accepted document, and some supporting documents may still have migration
gaps. Correct those gaps during a relevant category or document review rather
than through unrelated bulk normalization.

## Boundaries

Project Structure does not define:

- a canonical layout for user projects or Cookbooks
- how Recipes organize Ingredients
- one source-code architecture
- a one-to-one mapping between concepts and files
- automatic generation of source or repositories from Knowledge
- identical heading templates for every document role
- permanent physical paths

Project Structure defines how the current repository exposes responsibility
and navigation. The concepts remain authoritative because of their document
roles, not because of a directory name.

## Related documents

- [`Knowledge.md`](../README.md) — Documentation Resolver and compact
  Repository Map
- [`README.md`](../../README.md) — Concept Domain and Concept API catalog
- [README Node experiment](../project-knowledge/experiment/experiment.readme-as-distributed-entry-point.md)
  — the bounded experiment that preceded Node README adoption
- [`note.documentation-review.md`](../project-knowledge/note/note.documentation-review.md)
  — loose guidance for headings as a navigation interface
- [`design-note.concept-api.md`](../project-knowledge/design-note/design-note.concept-api.md)
  — why CTK is viewed through Concept APIs
- [`design-note.concept-domain.md`](../project-knowledge/design-note/design-note.concept-domain.md)
  — why Concept Domains were introduced
