# project-knowledge.experiment.responsibility-first-implementation-migration.md
============================================================

# Experiment: Responsibility-First Implementation Migration

## Question

Can an implementation be replaced or added by preserving documented
responsibilities rather than translating the structure of the existing source?

This question emerged from CTK's Bash-to-Go migration. It currently has one
substantial example and should remain an Experiment rather than reusable
guidance.

## Observed case

CTK already had a working Bash implementation, evolving Knowledge, Recipes,
and observable Runtime behavior.

The Go implementation was built in this order:

```text
README and relevant Knowledge
    ↓
Concept responsibilities
    ↓
Compatibility and safety Contracts
    ↓
Explicit implementation decisions
    ↓
Observable Cookbook / Distribution data
    ↓
Bash behavior when earlier evidence was insufficient
    ↓
Idiomatic Go representation
```

The Bash implementation remained valuable evidence, but its functions,
pipelines, external tools, and directory assumptions were not treated as the Go
architecture.

## What was preserved

- accepted Concept responsibilities
- Cookbook source compatibility where it was explicitly required
- observable behavior that satisfied those responsibilities
- safety boundaries discovered through real operation
- design discoveries that had not yet reached a curated document

## What was not automatically preserved

- Bash function and file organization
- shell pipelines and external CLI dependencies
- implementation-specific Distribution details outside the agreed boundary
- behavior that existed only as an accident of a parser or shell
- one language's temporary Workbench representation

## Observations

### Compatibility became easier to discuss when divided

"Compatible with Bash" was too broad. Cookbook source compatibility,
Distribution migration support, CLI behavior, and internal representation could
be decided independently.

### Implementation could challenge the Contract

Writing the Go implementation exposed Contracts that described one directory
representation more strongly than the underlying capability required.

The useful cycle was therefore not Contract-first-and-fixed:

```text
Responsibility
    ↓
Contract hypothesis
    ↓
Implementation and observation
    ↓
Refined Contract or implementation decision
```

### Language declarations reduced source ambiguity

A language README between shared documents and source made it possible to
declare applied Kitchen Notes, current strategies, and known differences before
reading implementation code.

### Small vertical slices built shared judgment

Recipe loading, Distribution discovery, selection, Build/Apply, Lock, and later
lifecycles were connected incrementally. Tests and small real executions fed
new evidence back into the next slice.

The later implementation pace may have come from accumulated shared language
and completed responsibility boundaries rather than code generation speed
alone.

## Current hypothesis

When a project's responsibilities are more stable than its implementation, a
responsibility-first migration may produce a more independent implementation
than a source-shaped port while still preserving intentionally selected
compatibility.

## Limits

This observation comes from one personal tool, one Bash-to-Go migration, a
continuous conversation, and an author who retained final authority over
Concept meaning.

It does not yet establish that the approach works when:

- accepted responsibilities are weak or absent
- undocumented source behavior is the actual public contract
- several independent authors disagree about Concept meaning
- binary or internal structural compatibility must be preserved
- the target language cannot express the same operational boundaries naturally

More examples would be needed before converting this Experiment into a Note or
recommended migration method.

## Related documents

- `experiment.go-migration.md` — Reviewed primary record of the CTK migration
- `../note/note.development-style.md` — broader Concept / implementation /
  observation cycle
- `../../../go/README.md` — current Go implementation declaration and behavior
- `../../../bash/scripts/README.md` — preserved Bash implementation declaration
