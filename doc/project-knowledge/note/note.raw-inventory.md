# project-knowledge.note.raw-inventory.md
============================================================

# Raw Inventory by Theme and Responsibility

This Note records an editing approach that has worked during the current CTK
Raw inventory.

It is a loose Project Knowledge observation, not a required documentation
process. Different material may need different review lenses or a different
order.

## Observation

Early Raw documents often contain several unrelated kinds of information in one
file:

- concept explanations
- design rationale
- operational knowledge
- development observations
- temporary thought experiments
- superseded language

Reviewing the whole file as one unit makes both promotion and deletion
unnecessarily difficult.

The Analogy inventory showed that smaller cross-document themes were easier to
compare with current Knowledge and easier to discuss independently.

## Inventory Workbench

The current inventory uses a temporary review view between source material and
durable documents.

```text
Raw / Experiments / Observations
              │
              ▼
   Temporary Inventory Workbench
              │
       Human + AI review
              │
              ▼
 Appropriate document / Preserve / Drop
```

`out/raw-inventory/` is one local example. It gathers fragments and proposed
decisions without becoming part of CTK Knowledge itself. Once the decisions are
reflected in tracked documents or reviewed sources, the temporary view can be
discarded.

This resembles a CTK Workbench: it is a place to observe and prepare changes,
not the authoritative source. The comparison explains the working relationship;
it does not make Raw inventory part of the CTK Workbench Concept API.

Earlier discussions used `Inbox` for several roles at once. In the current
workflow, its closest remaining role is this temporary review context rather
than a permanent storage location or required document type.

A temporary review context also lets work continue without forcing every
adjacent idea into a premature decision. Deferred material remains inspectable
when its theme becomes relevant again.

## Review lenses

The following lenses have been useful:

### Current responsibility

Read the current Knowledge first. It provides the accepted responsibility and
prevents historical wording from silently redefining the concept.

### Theme

Collect fragments that discuss the same subject, even when they live in
different Raw files.

### Function

Ask what each fragment is doing. It may explain a concept, record a design
decision, provide an operational tip, preserve an observation, or capture a
temporary experiment.

### Document responsibility

Use the Documentation Resolver to decide which document role, if any, should
own the durable result.

### Destination

Adopt the fragment into an appropriate document, preserve the source as a
thought record, keep it for later discussion, or Drop it.

## Not a fixed sequence

Theme, Function, Responsibility, and Destination are review lenses rather than
mandatory consecutive stages.

```text
                 Current Knowledge
                        │
                        ▼
Mixed Raw ──► Inventory Workbench
                      │
                      ▼
             Theme ◄──► Function
                │            │
                └─────┬──────┘
                      ▼
              Responsibility
                      │
                      ▼
          Destination / Preserve / Drop
```

A likely destination may reveal that the function was misunderstood. A
responsibility boundary may split one apparent theme into two. Review may move
back and forth between these lenses.

## Working decisions

The current inventory uses four provisional decisions:

- **Adopt** — integrate a reusable result into an appropriate document
- **Preserve** — retain the source as evidence or a thought record
- **Discuss** — keep the boundary open
- **Drop** — do not promote the fragment

Adopt and Preserve may both apply. An Experiment can remain after its reusable
observation has been promoted.

When a Raw file mixes themes, remove only settled fragments. Leave unrelated or
uncertain material for a later review.

## Scope of the observation

This approach has been useful in the CTK documentation inventory. It has not
been established as a general documentation method, and CTK contributors do
not need to apply it mechanically.
