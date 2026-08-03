# project-knowledge.note.development-style.md
============================================================

# Note: Conversation-first Concept Development

This project unexpectedly evolved into an interesting development process.

Initially, the goal was simply to build a small personal tool.

However, over time, the conversation itself became part of the design process.

## Concepts came first

Looking back, most core concepts were present from the beginning.

- Cookbook
- Runtime
- Freeze
- Recipe
- Ingredient

The concepts themselves rarely changed.

Instead, we repeatedly searched for better ways to express them.

Many discussions were not about changing the design.

They were about finding language that accurately represented an existing concept.

---

## Implementation was used to observe concepts

Implementation was not the final step.

It became a way to observe the concept.

A typical cycle looked like this:

```
Concept
    ↓
Implementation
    ↓
Observation
    ↓
Conversation
    ↓
Better Language
    ↓
Clearer Concept
```

Once the concept became clearer, missing capabilities naturally appeared.

Examples include:

- Inspect
- Variant
- Documentation Resolver
- Workbench

Rather than inventing new ideas, implementation often revealed concepts that had been implicit all along.

---

## Conversation became part of the design

One unexpected observation was that conversation itself became a design activity.

The implementation produced observations.

The observations produced questions.

The conversations searched for language.

The new language clarified the concept.

The concept then influenced the next implementation.

The goal of the conversation was rarely to "find an answer."

More often it was:

> "That is close... but not quite."

Repeated refinement eventually produced language that felt natural.

---

## AI acted more like an editor than a designer

The AI rarely introduced entirely new concepts.

Instead, it repeatedly tried different levels of abstraction, terminology, and explanations.

The role gradually shifted from:

- reviewer
- implementation assistant

toward

- concept editor
- language editor

The concepts already existed.

The collaboration focused on discovering language that accurately represented them.

---

## Why this worked

This approach probably depends on both the project and the people involved.

Several characteristics seemed important:

- trusting a concept before it could be fully explained
- validating concepts through implementation
- treating observations as design input
- allowing terminology to evolve
- prioritizing concepts over names

None of these guarantee success.

However, together they produced a surprisingly stable design process.

---

## Reflection

One interesting outcome is that the implementation itself changed less than expected.

Instead, the understanding of the concepts became progressively clearer.

As concepts solidified, missing pieces became obvious.

Features were not added because they seemed useful.

They appeared because the existing concepts naturally demanded them.

This feels less like feature-driven development and more like concept-driven evolution.

Whether this approach generalizes is unknown.

However, for a small project with continuous implementation, observation, and conversation, it was an unexpectedly enjoyable and productive way to develop a product.


============================================================

# Notes

Most design discussions were about representation, not concepts.

---

AI contributed less by inventing concepts than by discovering language that matched existing concepts.

---

AI accelerated the discovery of language rather than the discovery of ideas.

---

Concept
   │
   ▼
Implementation
   │
   ▼
Observation
   │
   ▼
Language
   │
   └──────────┐
              ▼
         Better Concept

---

**The minimal unit was not a team.

The minimal unit was a conversation capable of preserving concepts across implementation, observation, and language.**

---

We did not simply reduce misunderstandings. We created an environment where
the concept itself could mature.


---

**Conversation-driven Concept Development**

We assumed that concepts exist before their final expression does.

**Concept Cultivation** — growing concepts through implementation,
observation, conversation, and language.

---

This note records one successful experience, not a recommended methodology.

Whether this approach works depends on the project, the participants, and the type of problem being explored.

---
