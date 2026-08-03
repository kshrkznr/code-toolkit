# Knowledge.workbench.md
============================================================

# Concept Domain: Workbench

Workbench provides temporary review contexts for understanding, comparing, and
evolving Cookbook state.

Its Artifacts are disposable. An accepted result may enter a durable Cookbook
change only through the lifecycle responsible for that change.

## Responsibility

Workbench is responsible for making observed or proposed state reviewable
without treating the review representation as canonical state.

It is intentionally separate from the Core responsibilities that define
Cookbook composition and persistence.

Two workbenches serve different moments of the workflow:

| Workbench | Purpose | Lifecycle |
| --- | --- | --- |
| Draft | Edit a known Runtime change and commit it | `freeze draft` → `freeze commit` |
| Inspect | Understand or compare arbitrary completed states | `view` / `sync` |

## Navigate by question

Start with the question closest to the current work:

- How do I enter an existing review context without changing it?
  → [Opening a Workbench](#opening-a-workbench)
- How do I review a known Runtime change that should return to the Cookbook?
  → [Draft](#draft)
- How do I inventory one completed Recipe, Dist, Layer, or Ingredient?
  → [Inspect: View](#view)
- How do I compare two completed Recipe or Dist states?
  → [Inspect: Sync](#sync)

For practical review guidance, see [Workbench
Review](../note/note.workbench-review.md). Behavioral and Commit boundaries
remain in the [Workbench Contract](../contract/contract.workbench.md).

## Opening a Workbench

Workbench generation and Workbench review are separate operations.

- `freeze draft`, `view`, and `sync` generate or replace review material.
- `workbench` opens review material that already exists.

`ctk workbench` is the shared review entry point for the Draft and Inspect
areas. Omitting a choice uses the implementation's Selector without changing
the identity or lifecycle of the selected Workbench.

```text
ctk workbench
ctk workbench draft
ctk workbench inspect
ctk workbench inspect <viewpoint>
ctk workbench inspect <viewpoint> --editor <command>
```

Draft identifies the single commit Workbench at `cookbook/draft`. Inspect may
contain multiple disposable viewpoints, so selecting Inspect is followed by
viewpoint selection when one is not explicit.

The command opens the Workbench directory rather than one generated Artifact.
This keeps `summary.md`, Recipe material, Settings, and Extensions together as
one review context. Opening a Workbench does not generate, replace, commit, or
otherwise mutate its Artifacts.

The Go and Bash implementations resolve the editor in this order:

1. explicit `--editor <command>`
2. the `EDITOR` environment variable
3. `code`, when available
4. `vim`

Editor selection and directory placement are implementation integration
choices. The Concept API responsibility is the common ability to enter an
existing Draft or Inspect review context without changing its lifecycle.

---

## Concept APIs

### Draft

`freeze draft` is the everyday synchronization path.

It starts from a Dist, automatically resolves its Recipe, then produces a
commit-ready Draft under `cookbook/draft/`.

```text
Dist
  ↓
freeze draft
  ↓
Draft
  ↓
review / edit
  ↓
freeze commit
  ↓
Cookbook
```

A single `settings.draft` may contain both default and Profile-local settings.
Its `##` targets keep the Commit route explicit:

```text
## runtime.draft.settings.json
...

## profile/review.settings.json
...
```

---

### Inspect

Inspect is a disposable workbench under `cookbook/inspect/`. It does not modify
Cookbook, Dist, or Platform resources.

#### View

`view <recipe.yaml | dist-dir | ingredient-dir>` answers: “what is in
this now?”

Recipe and Ingredient views create canonical inventories. Every item is written
as a `+` line under its Ingredient-relative `##` target, so the inventory can be
edited and moved into `cookbook/draft/` for Freeze Commit.

Dist views read the existing Lock with an empty reference. Every observed item
therefore appears as an addition.

```text
cookbook/inspect/
└── recipe.<name>/
    ├── summary.md
    ├── recipe.draft.yaml
    ├── extensions.draft.md
    └── settings.draft.md
```

#### Sync

`sync <left> <right>` answers: “what changed between these completed
states?”

Go also permits either side to be omitted interactively. It selects Recipe or
Dist first, then selects the concrete source for that side.

The accepted inputs are Recipe and Dist. They are normalized before comparison:

```text
Recipe → resolved Reference
Dist   → observed Lock

Recipe ↔ Recipe = Reference ↔ Reference
Recipe ↔ Dist   = Reference ↔ Lock
Dist ↔ Dist      = Lock ↔ Lock
```

Resolved Settings and Extensions are the completed-state comparison. Recipe
YAML is also retained as right-side provenance and shown diagnostically in the
Summary; it does not replace the resolved comparison. Generated typed Drafts
use the same Commit syntax as Freeze Draft.

```text
cookbook/inspect/
└── sync.<left>.<right>/
    ├── summary.md
    ├── recipe.draft.yaml
    ├── extensions.draft.md
    └── settings.draft.md
```

The Summary records both typed sources and their Recipe identities. Settings
differences use the same first-property headings and `+/-` CTK Flat assignments
as Freeze Draft. Default and Profile Settings remain in one typed Artifact.

---

## Working style

```text
Normal update         What differs?              What is here?
─────────────         ─────────────              ─────────────
freeze draft          sync                       view
```

The three commands share the Draft representation, but differ in intent:

- `freeze draft` is the normal path from a Runtime back to Cookbook.
- `sync` is exploratory comparison with explicit left and right sides.
- `view` is an inventory that supports discovery and housekeeping.

This makes the workbench a place for both human review and AI-assisted review
without turning temporary inspection data into Core persistence.
