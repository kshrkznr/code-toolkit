# project-knowledge.note.documentation-review.md
============================================================

# Documentation Navigation and Review

CTK's documents are used as a shared navigation surface by people and AI.

Titles, headings, and placement help a reader move from a current question to
the relevant responsibility or observation. In this limited sense, headings
form a **navigation interface**.

This is not a versioned compatibility contract. The amount of stability
expected from that interface depends on the document's role.

## Strength follows document role

Core and Contract documents express accepted responsibilities or explicit
agreements. Their titles, heading structure, and terminology should therefore
be changed deliberately. A structural change may affect how readers find and
interpret the accepted model even when the prose itself remains correct.

Notes and Experiments have a looser role. Their headings should make the local
topic understandable, but exploration and thought records should not be forced
into the same structural stability as Core or Contract documents.

Other document roles can choose an appropriate level between these ends. This
is a review perspective, not a complete hierarchy of documentation rules.

## Loose guidance

Useful questions when reviewing a document include:

- Does the heading describe a topic a reader is likely to look for?
- Does the section have one understandable responsibility?
- Does its placement match the document role?
- Would moving or renaming it make an accepted concept harder to find?
- Does the change affect responsibility, rationale, observation, or only
  wording?
- Can the intended change be understood from the diff?

Consistent names and hierarchy can improve navigation, but consistency is a
means rather than the objective. The objective is to make the document's role
and the next useful context easier to recognize.

## Treat failed discovery as documentation feedback

Packaged documentation resolves documents from maintained navigation metadata:
canonical identity, repository path, Node alias, title, and headings. It does
not hide missing routes behind document-body search.

When a person or AI expects a document to answer a question but cannot find it
from that metadata, the failure is useful documentation feedback. Prefer a
focused documentation-only Issue proposing a clearer title, heading, or Node
route. No implementation change is required for that Issue to improve CTK.

Full-text search remains useful for investigation after exporting or cloning
the documentation. It does not replace review of the navigation interface.

## Boundary

Reviewing documentation with source-like care does not make prose executable,
nor does it make every heading a permanent API.

It means that documentation structure is part of the project's shared design
surface and can be reviewed with attention appropriate to its responsibility.
