# project-knowledge.design-note.documentation-onboarding.md
============================================================

# Why README Became a Navigator

## Background

README was initially approached as a place to summarize CTK's features.

That approach felt unnatural as the number of concepts increased. A complete
feature explanation duplicated Knowledge, while a short feature list did not
help a new reader decide what mattered to their workflow.

## Observation

The explanation became more natural when it started from the reader's problem:

- What is difficult about the current environment?
- What would the reader like to reproduce, separate, review, or evolve?
- Which CTK concepts are relevant to that situation?

This changed README from a compressed architecture document into an entry point
and navigator.

## Resulting shape

README now introduces the problem and value, presents the available Concept
Domains, and directs detailed questions toward the Documentation Resolver and
relevant Knowledge.

```text
Reader's situation
      ↓
README
      ↓
Relevant concept or workflow
      ↓
Documentation Resolver / Knowledge
```

The same shape supports both people and AI. Neither needs to read the entire
Knowledge Pool before beginning a relevant discussion.

## Boundary

README does not replace Knowledge and does not need to explain every feature.

Its responsibility is to make CTK's value and available entry points visible
enough for the reader to choose the next useful context.
