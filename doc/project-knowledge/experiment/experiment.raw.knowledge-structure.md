# project-knowledge.experiment.raw.knowledge-structure.md
============================================================

### Observation: Knowledge structure as externalized thinking

> During discussions, we observed that the Knowledge structure may serve another purpose beyond documentation organization.
>
> The author naturally moves between different levels of abstraction—from implementation details to architectural concepts and back—at a relatively high frequency.
>
> Rather than constraining this thinking style, the Knowledge layers appear to provide shared landmarks for collaborators (human or AI), allowing them to identify the current level of discussion and rejoin the conversation more easily.
>
> In that sense, the Documentation Resolver and Knowledge hierarchy may function as a guide for navigating the author's thinking process, not merely for locating documents.
>
> Whether this observation generalizes beyond this project remains an open question.


---

Documentation Resolver was not conceived as a documentation architecture.

It emerged because conversations repeatedly lost their current context as discussions rapidly moved across abstraction levels.

The term “Resolver” was adopted almost casually, inspired by concurrent implementation work around recipe resolution, because the role felt similar: resolving where to continue.

---

Documentation Resolver does not appear to encourage re-reading all project knowledge.

Instead, it provides enough context to activate the relevant knowledge domain, allowing discussions to continue without reconstructing the entire project state.

--|

Projects are organized by roles, not species

During discussions about AI collaboration, we noticed that “human vs AI” was not the most useful distinction.

A project naturally assigns responsibilities and authority to participants, regardless of whether they are human, AI, or any other future intelligent agent.

The important questions were not “Who are you?” but:

* Can you participate in constructive discussion?
* Can you respect project governance?
* Can you stay within your assigned responsibilities?


---

# Observation: Knowledge Architecture beyond CTK

CTK currently operates at a scale where a single Documentation Resolver is sufficient.

For an individual project, this may already be enough.
Additional abstraction would likely increase complexity without providing proportional benefit.

However, an interesting question emerged during discussion.

What happens when the same ideas are applied to a team rather than an individual?

We do not expect CTK itself to scale unchanged.
Instead, we expect the surrounding Knowledge architecture to evolve.

For example, a team may already have independent knowledge sources such as:

- Team workflow
- Product domain model
- Architecture guidelines
- Organization design guide

Rather than consolidating these into a single knowledge base, each may continue to evolve independently while the team's entry point simply declares its dependencies.

In that world, a team's README may no longer be a document that explains everything.
Instead, it becomes a resolver that establishes context by saying:

> "This team assumes the following knowledge."

This resembles dependency management in software more than traditional documentation.

An equally interesting possibility is that Documentation Resolvers themselves become hierarchical or interconnected.

A project resolver may reference:

- Organization knowledge
- Product knowledge
- Team knowledge

while the team knowledge may itself reference other independent concept domains.

The resulting structure is no longer a simple document hierarchy.
It begins to resemble a knowledge graph.

Whether this actually emerges remains an open question.

Another observation follows naturally.

As knowledge becomes increasingly distributed, maintaining these relationships manually becomes expensive for humans.

This may be an area where AI provides significant practical value.

Rather than generating documents, AI could assist with navigating knowledge relationships:

- identifying relevant concept domains
- discovering missing dependencies
- suggesting related knowledge
- maintaining consistency between independently evolving knowledge bases

Interestingly, this expectation does not originate from AI.

It originates from a scalability problem.

Only after recognizing the scalability challenge does AI appear to be a natural participant.

At the moment, these remain expectations rather than conclusions.

The next interesting observation will be whether similar structures emerge naturally when these ideas are applied within a real development team.
