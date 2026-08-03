# project-knowledge.experiment.legacy-knowledge-note-inventory.md
============================================================

# Experiment: Legacy Knowledge Note Inventory

## Review status

**Reviewed during the Note README trial.**

This file preserves the mixed source material that previously occupied
`doc/note/note.md`. It no longer represents the current responsibility of the
Note layer.

## Inventory outcome

Several durable observations already have clearer destinations:

- conversation-first design and concept evolution
  → [`Conversation-first Concept Development`](../note/note.development-style.md)
- naming as an expression of responsibility
  → [`Naming`](../note/note.naming.md)
- current Persistence responsibilities
  → [`Persistence Lifecycle`](../../core/core.persistence-lifecycle.md)
- the current Note layer responsibility and entry paths
  → [`Notes`](../../note/README.md)

The remaining collaboration reminders, analogies, punchline candidates, and
empty working sections are retained below as historical thought material. They
are not current guidance or a required documentation template.

## Preserved source material

# Notes (development style)

Project
    │
    ▼
Markdown
   ╱   ╲
Human  AI

- AI向けのプロンプトを書くのではなく、プロジェクトの暗黙知をMarkdownとして明文化する。
- HumanとAIは同じMarkdownを共有し、プロジェクトの前提を共有する。
- Projectがsource of truthであり、Markdownは共有メモリとして機能する。

## Design process

1. Start from a real problem.
2. Find familiar concepts.
3. Compare similarities and differences.
4. Borrow only the useful ideas.
5. Expand the expected use cases.
6. Generalize into a reusable design.

## Naming

Prefer familiar terminology from existing developer ecosystems.
Avoid inventing new commands unless existing terminology would be misleading.
Reuse expectations rather than create new vocabulary.

## Design evolution

Implement the smallest solution for today's problem.

Only introduce the next lifecycle step
after it naturally appears through real usage.

The lifecycle is discovered, not designed.

## Current impressions

- HumanとAIが対等に設計を議論できる感覚があり、開発体験が良い。
- 会話から設計やパンチラインが生まれることが多い。
- 暗黙知をMarkdownへ外部化することで、会話の認知負荷が下がる。
- 長期的な有効性は今後も観察する。

==========================================

---

## Document structure

```text
## XXX

Inspired by

### Notes
### Future(this Thema)
### Knowlage candidate
### README candidate
### Conceptual diagram
```

Not every section is required.
Choose the structure that best explains the concept with minimal context.

---

## 基本思想
### Inspired by

- choromium(VSCode-family)
- IaC
- OSS
- docker

### Notes

- シンプルさを優先
- カジュアルダウン
- AI協業(Creator/User both)
- VSCode Family
- platformの抽象化（ide体験の摩擦減）が前提。vs codeとkiro/macとwinを意識したくない。

### Punchlines(this Thema)

- “IDE environment as code.”
- “Build your IDE like software.”
- “Write once, use across VSCode-family IDEs.”
- Reduce the cost of changing your development environment.

- 「IDEを管理するツール」ではない。**「IDEを仮想環境として組み立てるツール」**
- “Composable recipes for VSCode-family environments.”

### Future

### Random Ideas

---

### Knowlage candidate

- Project is the source of truth. Markdown is the shared memory.

```text
Project
    ↓
Markdown
    ↓
Human ↔ AI
```

#### Analogy

```text
|元ネタ|CTKでのローカライズ|
|Git|freeze / draft / commit|
|CI|build|
|Artifact|Runtime|
|venv|codevenv|
|Spring Boot|Runtime → Profile → Dynamic Settings|
|DI|Resolver|
|OSS Package|Module|
```

```text
|CTK|Docker|
|Recipe|Dockerfile|
|Build|docker build|
|Runtime (dist)|Image|
|Launch|docker run|
|IDE Binary (VSCode/Kiro)|Docker Engine|
```

- Infrastructure as Code に対して、
  - Dev Container = Runtime as Code
  - CTK = VSCode as Code


### README candidate

#### Conceptual diagram

```text
表側（毎日触るもの）

* ctk use
* ctk launch
* code .

裏側（それを支えるもの）

* recipe
* build
* apply
* freeze
* archive
* lock
```

---

## Persistence Lifecycle

### Notes

The persistence lifecycle describes how a Runtime is observed and preserved.

Each step has a single responsibility.

- Lock observes the current Runtime and generates a Manifest.
- Freeze converts the Manifest into Recipes.
- Archive converts the Manifest into a self-contained distribution package.

This separation keeps observation, editing, and packaging independent.

---

## Collaboration Notes

AI向け運用

- 情報混濁に注意
- 作者も間違える
- 推測を明示
- 実装/設計確定と思想確定が存在
- 期待する役割 [壁打ち, PG, Review, 案出し, PRJパートナー]
- 例え話、違和感の指摘は歓迎


---

## Parking Lot

今は考えない。

- ○○
- ○○

---

## Punchlines/Memo(in a conversation)

育てたいフレーズ/気になったワード。不要であれば都度消す。

### Design Fragments
### Ideas
### Message Drafts
### Inbox
