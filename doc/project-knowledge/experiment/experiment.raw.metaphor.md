# project-knowledge.experiment.raw.metaphor.md
============================================================

# Metaphor

> **Inventory status: Reviewed (2026-07-27)**
>
> The durable observations from this experiment were promoted to
> `../note/note.analogy.md` and `../note/note.analogy-design-review.md`.
> This file remains as a thought record. Its exploratory mappings do not define
> the current CTK responsibility model.

==========================================================================================

Analogy is not only a communication technique. It also guides exploration by encouraging AI to anchor new concepts to familiar ones before abstracting differences.

Knowledge is not a rulebook. It is a gentle bias.

----

Analogy is a bridge, not a requirement.

Use it when introducing or recovering a concept. Skip it when the shared mental model is already established.

==========================================================================================

# メタファーは違和感を観察するための道具

CTKではメタファー（Metaphor）やアナロジー（Analogy）を多く利用している。

しかし、その目的は「おしゃれな命名」や「世界観を作ること」ではない。

メタファーは、設計の違和感を観察するための道具である。

## 命名のためではなく、観察のため

設計を進めていると、

- なんとなく違和感がある
- 説明はできないが落ち着かない
- 責務が混ざっている気がする

という状態になることがある。

この段階では、設計そのものを眺めても原因を言語化しづらい。

そこで一度、別のメンタルモデルへ写像してみる。

例えば今回のCTKでは、`Recipe` という名称を使用しているが、
名称をきっかけに料理という既存の概念へ写像して比較した。

すると、

```text
Cookbook
├── Recipes
├── Templates
├── Files
└── Attributes
```

という構造よりも、

```text
Cookbook
├── Recipes
└── Ingredients
    ├── Templates
    ├── Files
    └── Attributes
```

の方が自然に感じられた。

この時、「Ingredientsという名前が好き」だったわけではない。

料理として眺めたときに、

> 「Recipeが使う素材」という責務が自然に見えた

ことが重要だった。

## メタファーは責務を浮かび上がらせる

同じことはRuntimeの構造でも起こった。

設計上は、

```text
Runtime
└── Module
    ├── Name
    ├── Settings
    └── Extensions
```

という構造を考えていた。

しかし料理へ写像すると、

```text
Ingredients
└── Runtime
    └── Git
        ├── Name
        ├── Settings
        └── Extensions
```

と考えた方が自然だった。

Git Runtimeという「素材」があり、その素材には

- 名前
- 設定
- 拡張

といった付属情報がある。

これは

> 鶏肉には部位や下味がある

という関係とよく似ている。

この視点では、「Module」という中間概念は不要になる。

つまりメタファーは、命名を変えたのではなく、不要な概念を発見した。

## メタファーは設計を決めない

重要なのは、メタファーが設計の正しさを保証するわけではないことである。

料理に似ているから正しい、ではない。

料理へ写像した結果、

- 責務が自然に説明できるか
- 境界が明確になるか
- 不自然な概念が残っていないか

を観察する。

もし写像後の世界でも違和感が残るなら、元の設計にも改善の余地がある可能性が高い。

## メタファーは設計レビューの道具

今回の検討では、

- Component
- Ingredients
- Cookbook
- Module

といった命名を考えていたように見える。

しかし実際には、

- Cookbookという責務が見えた
- RecipeとIngredientsという境界が見えた
- Moduleという抽象概念が不要になった
- ディレクトリ構成やライブラリ構成まで整理できた

という設計上の発見につながった。

これは命名を考えた結果ではない。

メタファーを利用して、設計をレビューした結果である。

## CTKでの考え方

CTKでは、メタファーを世界観として採用することを目的としない。

まず責務を整理し、違和感を観察する。

その過程で、あるメタファーが自然に当てはまるのであれば、それを採用する。

つまり、

> メタファーに設計を合わせるのではなく、設計を観察するためにメタファーを利用する。

これがCTKにおけるメタファーとアナロジーの役割である。

```text
現実の設計
      │
      ▼
別のメンタルモデルへ写像
      │
      ▼
違和感を観察
      │
      ▼
現実の設計へ戻す
```

### メタファーは設計を説明するためだけのものではない。

一度別の世界へ設計を写像し、その世界で違和感を観察してから元の設計へ戻ることで、責務や境界の不自然さを発見できる。

CTKでは、このような「設計レビューのためのアナロジー」も重要な設計手法として扱う。

======================================

## Metaphor Review

責務や境界に違和感を感じたら、一度別のメンタルモデルへ設計を写像する。

その世界の責務分割と比較し、不自然な概念や境界がないかを観察する。

最後に元の設計へ戻し、得られた気付きを反映する。

### 手順

```text
設計
    ↓
メタファーへ写像
    ↓
その世界の責務で眺める
    ↓
違和感を抽出
    ↓
元の設計へ戻す
```

### 注意

> **良いメタファーは、新しい名前を与えるよりも、不要な概念を見つける。**

目的はメタファーへ合わせることではない。

メタファーを利用して設計の違和感を観察することである。

```text
| CTK | 料理 |
|------|------|
| Cookbook | Cookbook |
| Recipe | Recipe |
| Ingredient | Ingredient |
| Runtime | 食材の種類 |
| Git Runtime | 鶏肉 |
| Settings/extensions | 下味・切り方・保存方法など素材の属性 |
```

---

既存OSSは実装例だけでなく、責務分割のサンプルでもある。

採用することが目的ではない。

自分の設計を投影し、責務や境界を比較するための観察対象として利用できる。

---

メタファーは成果物ではない。

メタファーは設計中に利用する思考補助であり、利用者へ公開する概念とは限らない。

また、メタファーを利用した思考実験の成果物にもレイヤーが存在する。

### 1. 公開するメタファー

これはユーザーにも価値がある。

- Cookbook
- Recipe
- Ingredients

このくらいなら理解を助ける。

---

### 2. 設計レビュー用メタファー

これは設計者だけが使う。

- Repository = Restaurant
- Build = 調理
- Archive = 配送
- Freeze = 試作メモ

READMEには書かない。

でも、

> 「責務合ってる？」

を確認するときにはすごく役立つ。

---

### 3. 思考実験

もっと一時的なもの。

例えば今回は、

> Runtime = 肉
> Profile = 野菜
> Extension = 調味料


---

| Layer | 役割 |
|--------|------|
| Platform | 実行基盤 |
| Runtime | 実行環境 |
| Workspace | 作業環境 |
| Profile | 利用シーン |
| **Cookbook** | 組み立て知識 |


---

Cookbook
├── Ingredients
├── Recipes
├── Draft
└── Ingredients Layer
    ├── Platform
    ├── Runtime
    ├── Workspace
    ├── Profile
    ├── Extension
    ├── Settings
    ├── Snippets
    ├── Workspace
    ├── DevContainer
    ...

---


repository:restaurant
├──dist
├──archive
├──Cookbook
    ├── Ingredients
    ├── Recipes
    ├── Draft
