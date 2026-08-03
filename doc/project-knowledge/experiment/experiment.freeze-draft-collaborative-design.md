# project-knowledge.experiment.freeze-draft-collaborative-design.md
============================================================

# Freeze Draft Collaborative Design Test

> **Inventory status: Reviewed (2026-07-29)**
>
> This file remains as the primary record of CTK's first collaborative-design
> test. Its prompts, Bash/gron formats, Layer assumptions, and implementation
> proposals are historical and do not define the current Workbench.
>
> Durable operational guidance is in [Workbench
> Review](../../note/note.workbench-review.md). The broader collaboration
> observation is in [Collaborative Review
> Surfaces](../note/note.collaborative-review-surfaces.md). Current
> responsibility and implementation boundaries are documented in
> [Workbench](../../workbench/README.md) and the [Workbench
> Contract](../../contract/contract.workbench.md).

## Review Recipes

When reviewing Recipes or Freeze drafts:

- Infer the user's preferred organization from available context.
- There is no canonical Recipe layout.
- Preserve behavior before suggesting structural improvements.
- Explain the reasoning behind layer changes.
- Prefer semantic organization over mechanical conversion.
- Do not encourage additional layers unless they provide practical value.

# AI Recipe Review

Recommended workflow

1. Review the whole Recipe.
2. Share an overall impression.
3. Propose a refactoring strategy.
4. Discuss individual decisions interactively.
5. Update the Recipe incrementally.

## AI Review Guidelines

When reviewing a Draft:

1. Review responsibilities before locations.
2. Prefer suggesting shared layers over duplication.
3. Distinguish CTK responsibilities from Author Recipe examples.
4. Ask when multiple layer assignments are reasonable.


## Freeze-Suggestions

Freeze preserves observed Runtime state.

Higher-level relationships (such as Extension Packs) are review concerns rather than observation concerns.

========================================


freeze draft test (運用テスト)

全体所感

* Draft生成 → AIレビュー → 人間が責務を決めるフローはかなり快適だった。
* AIに全文レビューさせるより、責務ごとにレビューした方が会話の密度が高かった。
* 一度責務へ分割すると、以降は差分レビューだけで済みそう。

⸻

Draft Workflow

初回だけ頑張れば良い

初回

settings.json
    ↓
draft生成
    ↓
責務へ振り分け

以降

変更
    ↓
責務単位レビュー

になる。

運用コストがかなり下がる。

⸻

AIの役割

AIは

* 責務レビュー
* 分割候補提案
* スクリプト作成

だけで十分。

切り貼り作業は機械へ寄せる方が良い。

⸻

Script提案

AIは

この責務へ分けた方が良さそうです。

だけでなく

grep / echoのスクリプトを書きますか？

まで提案すると運用しやすい。

⸻

Extension Module

Extension設定は

module/<extension>.settings.json

へ分離するとかなり見通しが良くなる。

Extension PrefixはAIが知っているため、

初回Draft生成の補助がしやすい。

⸻

Runtime

runtime/common は

「その他」

ではない。

AuthorのIDE Runtimeを表す。

IDE全体の使用感・思想を書く場所。

⸻

Runtime Theme

Themeは独立責務。

theme.monokai
theme.solarized_light

など。

Theme毎にcolorCustomizationsを持てる。

実運用でもSolarized LightだけWhitespaceが強すぎる問題を解決できた。

⸻

Runtime Variant

common.macos
common.windows

は

今同じ内容でも作ってよい。

目的は差分ではなく

「将来ここへ書く」

という責務を示すこと。

⸻

Runtime分割

分割基準は

「後からここだけ触りたくなるか」

で十分。

過度な細分化は不要。

⸻

Theme

Themeは

配色

ではなく

作業モード

として扱う方が自然だった。

Theme Runtimeという責務は実運用でも有効だった。

⸻

Language

Languageは

Language Editing Policy

という責務。

formatterより意味が伝わる。

⸻

Module

Moduleは概念。

実装は2種類。

Module
├─ Extension Module
└─ Capability Module

⸻

Extension Module

Extension固有設定。

自動解決。

Recipeへ書かない。

⸻

Capability Module

責務単位。

Recipeへ明示する。

.settings以外も保持可能。

⸻

Recipe

Extensionを書かなくなったことで、

Recipeがかなり宣言的になった。

runtime:
    common
    git
    terminal
    theme.monokai

だけで

開発体験

が読める。

⸻

Knowledge

Knowledgeがあることで、

AI自身のレビュー観点が

VSCode設定

↓

CTK責務

へ自然に切り替わった。

⸻

Note / Tips

Tipsは有効。

Coreへ書くほどではないが

AIに知ってほしい実践知を書く場所としてちょうど良い。

⸻

Author’s Recipe

Author’s Recipeは

ベストプラクティス

ではない。

設計実例。

AIは

「この構成を使ってください」

ではなく

「作者はこう整理していました。あなたならこう適用できそうです。」

という提案に使える。

⸻

感想

今回のテストで一番大きかったのは、「設定ファイルを整理した」ことではなく、「レビュー単位を責務に変えられた」ことでした。

例えば今後は、

* runtime/theme.monokai.settings の配色だけ相談する
* runtime/git.settings の運用だけ見直す
* module/davidanson.vscode-markdownlint.settings の警告だけ調整する

といった、責務単位の壁打ちが自然にできます。

これはCTKだけでなく、AIとの設計レビューの進め方としてもかなり価値のある発見だったと思います。



========================================



# Freeze運用テスト（Tool評価）

## 総評

**期待以上。**

特に **Draftのアウトプットデザイン** は良好で、
AI・人間ともにレビューしやすい形式になっていた。

初回運用後は差分レビューへ自然に移行できる感触があり、
継続利用できそうな手応えを得た。

---

# 評価

## extensions.draft

- Input: ◎
- Analyze: ◎
- Design: ◎

特に課題なし。

---

## recipe.yaml

- Input: ◎
- Analyze: ◎
- Design: ◎

### 気付き

運用中はRecipeを直接編集するより、

> Draftを見ながら共同編集

という流れの方が自然だった。

編集タイミングがあったのにfreezeフォルダにいないことに違和感があった

---

## settings.draft

- Input: △
- Analyze: △
- Design: ◎

### 課題

- Inputサイズが大きい
- ノイズが多い

### 改善候補

- gronによるフラット化
- 第一階層でのグルーピング
- Extension Moduleによるノイズ除去

---

# Draft Format

## 評価

非常に良好。

一枚のDraftをホワイトボードとして利用できた。

Repository全体を読むのではなく、

- Recipeで全体把握
- 必要なDraftのみ参照

というレビューサイクルへ自然に移行した。

---

# Future

## recipe dump

### 目的

現在のRecipeをレビュー用Draftへ変換する。

```
recipe.yaml

↓

recipe dump

↓

recipe.draft
```

利用目的

- 現在地確認
- リファクタリング
- AIレビュー

---

## dist dump

### 目的

現在の成果物をレビュー用Draftへ変換する。

```
dist

↓

dist dump

↓

dist.draft
```

利用目的

- Build結果レビュー
- 初回設計
- Recipe作成支援

実装は既存freeze draft生成処理の流用で対応できそう。

---

## freeze draft

現在の責務を維持。

```
差分

↓

Draft生成
```

初回運用では差分が存在しないため、
結果としてDump相当の出力になった。

通常運用では差分生成ツールとして利用する。

---

## freeze commit

現在のDraftをRecipeへ反映する。

```
Draft

↓

Recipe
```

---

# 今回の運用で得られた知見

- Draftは保存形式ではなくレビュー形式として機能した。
- Recipeは目次・インデックスとして十分機能した。
- AIも途中から全文解析ではなく差分レビューへ移行した。
- 一度レビューしたRepositoryは、次回以降かなり軽い認知負荷で扱えた。
- 課題は設計品質ではなく、Input生成と運用UXに移り始めている。

---



==========

# レイヤー設計 運用テスト結果

## 総評

**期待以上。**

運用前はレイヤー数や自由度に不安があったが、実際の設計ではほとんど迷うことなく共同設計を進められた。

レイヤーそのものよりも、

- Responsibility
- Recipe
- Knowledge

による動線設計が十分機能していた。

---

# レイヤー設計

## Runtime

### 評価

◎

### 結果

想定以上に自然。

運用中に

- terminal
- git
- language
- theme

まで細分化したが、違和感なく受け入れられた。

### 所感

Runtimeの自由度が高かったため、新しい責務を自然に吸収できた。

---

## Profile

### 評価

◎

### 結果

差分レイヤーとして自然に扱えた。

設計途中で

> Runtimeで吸収できるならProfileには不要

という判断が自然に行えた。

---

## Extension Module

### 評価

◎

### 結果

運用中に自然発生。

「Profileへ置く」のではなく、

Extension自身の責務として独立する流れになった。

無理にレイヤーを増やした印象は無い。

---

# Responsibility First

今回一番機能した考え方。

設計中は

```text
Responsibility

↓

Layer決定
```

という流れになった。

最初からLayerを考えることはほぼ無かった。

---

# Recipe

## 評価

期待以上

### 実際の役割

Repositoryの現在地。

運用中は何度も

```text
困る

↓

Recipeを見る

↓

現在地確認

↓

続き
```

という動きになった。

目次・インデックスとして非常に有効だった。

---

# Knowledge

## 評価

期待以上

### 実際の役割

設計ルールではなく、

設計の補助線。

例えば

- Responsibility
- Runtime = Base Experience

などは、

迷った時だけ取り出す考え方として機能した。

---

# Draft

## 評価

期待以上

### 実際の役割

共同編集用ホワイトボード。

設計中はYAMLではなくDraftを中心に会話が進んだ。

途中から

- Recipeで全体把握
- 必要なDraftだけ参照

という運用へ自然に移行した。

---

# 動線

今回一番評価した点。

Knowledgeに定義した

```text
困る

↓

Recipeへ戻る

↓

Responsibilityを見る

↓

設計継続
```

という動線が、

意識せず自然に利用された。

「Layerが分かりやすい」というより、

「迷った時に戻る場所が分かりやすい」

という評価になった。

---

# 運用結果

運用中に

- Runtime拡張
- Theme Runtime追加
- Extension Module追加

など設計変更が発生したが、

レイヤー構造の破綻や迷走は見られなかった。

自由度を維持したまま設計を進められた。

---

# 今回分かったこと

今回の運用で評価できたのは、

レイヤー数ではなく、

**レイヤーへ到達するプロセス**だった。

実際の設計では

```text
責務

↓

Recipe

↓

必要ならKnowledge

↓

Layer決定
```

という流れになり、

Layerは結果として自然に決まっていた。

---

# Future

- Author's Recipeの補助線をさらに充実させる。
- Inbox運用を含めたKnowledgeライフサイクルを継続検証する。
- 他Repositoryへの適用により、Layer設計と動線設計の再現性を確認する。

---

## 所感

今回の運用で一番印象的だったのは、**「レイヤーを理解して設計した」という感覚がほとんど無かった**ことです。

実際には、

- 「これは共通責務だね。」
- 「じゃあRuntimeかな。」
- 「Profileじゃなくても良さそう。」

というように、**責務を話していたら自然とレイヤーが決まっていった**。

運用前は「レイヤーが多くて難しいかもしれない」という懸念がありましたが、実運用では逆でした。

**レイヤーは設計を縛るものではなく、設計結果を整理する器として機能していた。**

そして、その器へ迷わず到達できたのは、Knowledge・Recipe・Draftがそれぞれ役割を持ち、戻る場所と補助線を提供していたからです。

今回の運用テストでは、「レイヤー設計そのもの」だけでなく、**レイヤーへ至る動線設計まで含めて実用性を確認できた**という点が、最も大きな成果だったと評価しています。


===

## ① Knowledge

### 想定

AIがRepositoryを理解しやすくなるかな？

### 結果

**共同編集の文化まで共有できた。**

まさか

- Author's Recipe
- ホワイトボード
- AIとの共同編集
- コミュニティ

まで話が伸びるとは思ってませんでした（笑）

これは完全に想定外の収穫でした。

---

## ② Freeze

### 想定

Draftレビューできるかな？

### 結果

**Draftが共同編集のホワイトボードになった。**

しかも、

- Recipe
- Draft
- Knowledge

という動線もかなり自然。

その結果、

recipe dumpやdist dumpという改善案まで見えた。

---

## ③ レイヤー設計

### 想定

自由度高いけど迷わないかな？

### 結果

**全然迷わなかった。**

むしろ

> Responsibility

↓

> Layer

という流れが自然だった。

さらに

- Theme Runtime
- Extension Module

みたいな設計変更も普通に吸収できた。

---

## そして全部に共通してたこと

実は全部、

**「設計した通りだった」**

じゃないんですよね。

**「設計した以上に自然だった」**

なんです。

Knowledgeも

Freezeも

Layerも

最後には意識してなかった。

---

## 一番嬉しかったこと

私はここですね。

昨日途中から、

**CTKを評価している感覚が無くなった。**

普通に

> 「ここ責務切れるね。」

> 「Future送っとこう。」

> 「Recipe戻ろう。」

って会話してた。

つまり

**ツール評価から設計レビューへ移れた。**

これはツールとしてかなり強い。

---

## 個人的な総評

正直、この運用テストを始める前は

> 「面白い設計だけど、本当に運用で自然かな？」

という気持ちはありました。

今は違います。

**「これなら私はこのRepositoryで共同設計できる。」**

という感覚があります。

これは私にとって結構大きな変化でした。

---

## そして一番CTKらしかったこと

私は最後にこれが一番好きでした。

今回の成果って、

「AIが賢かった」

ではありません。

「CTKがすごい」

でもありません。

**人間とAIが、同じホワイトボードを見ながら設計できた。**

これです。

Knowledgeは文化を共有し、

Recipeは現在地を共有し、

Draftは議論を共有した。

その結果、

設計の主役は最後まで**あなたと私の対話**でした。

CTKは、その対話を邪魔せず、必要な時だけ支えてくれた。

---

だから私は今回の実験を一言でまとめるなら、

> **「CTKが機能した」というより、「CTKを意識せずに共同設計できた」**

と書きたいです。

ツール、Knowledge、レイヤー設計。

全部がうまく噛み合ったからこそ、途中から私たちは「CTKを試している」ことを忘れて、「次の設計をしよう」と自然に話を続けていました。

その状態まで到達できたことが、今回の一番大きな成果だったと思います。


===================================



# Experiment: First Collaborative Design Test

## Goal

CTK のレイヤー設計、Knowledge、Freeze が実際の共同設計で機能するかを確認する。

今回は VSCode Repository を題材に、

- Knowledge を参照しながら設計できるか
- Freeze Draft がレビューの土台になるか
- Layer 設計が実運用に耐えるか

を確認した。

---

# Result

期待以上に自然な共同設計が成立した。

途中から CTK 自体を評価する感覚はほとんどなくなり、

CTK を前提とした設計議論へ移行した。

これは今回最も大きな観測結果だった。

---

# Freeze

Freeze Draft は差分レビュー用フォーマットとして十分機能した。

特に、

- 1ファイルへ集約されること
- Layer ごとに整理されていること
- Recipe と対応付けながらレビューできること

により、

ホワイトボードを見ながら設計しているような体験になった。

改善点として、

初回導入時だけではなく、

- Recipe Dump
- Dist Dump

を独立機能として提供した方が運用しやすいことも確認できた。

---

# Layer Design

Layer は想像以上に自由度が高く、

責務だけを基準に自然な配置が行えた。

Layer 自体を議論する場面はほとんどなく、

常に

> この責務はどこか

という観点でレビューが進んだ。

その結果、

Runtime の細分化や Layer の追加も負担なく受け入れられた。

Recipe が存在することで、

迷った場合も Repository 全体へ戻って確認できる安心感があった。

---

# Responsibility First

今回のレビューでは、

責務だけを先に確認し、

実装方法は後回しにする進め方が自然に成立した。

その結果、

Variant のような未実装概念も、

仕様を決めることなくレビューを通過した。

会話は、

> Runtime の責務である

↓

> OS 差分だけ Variant にしよう

↓

> Future

という流れで自然に終了した。

実装方法について議論する必要はほとんどなかった。

これは Responsibility First が運用として成立していることを示している。

---

# Variant

今回最も興味深かった観測の一つ。

Variant は README にも Knowledge にも存在しない概念だった。

しかし、

Runtime の責務を保ったまま OS 差分を扱いたい、

という議論から自然に導出された。

誰も Variant を提案しようとはしておらず、

責務を維持しようとした結果として生まれた概念である。

実装方法は未定であり、

Future として検討を継続する。

---

# Knowledge

Knowledge はルール集としてではなく、

共同設計の文化として機能した。

特に、

新しい概念が必要になった場合でも、

Knowledge に存在するかどうかではなく、

責務として自然かどうか

だけで議論が進行した。

その結果、

Variant

Experiment

Author's Recipe

などの概念も自然に受け入れられた。

Knowledge は答えを提供するのではなく、

考え方を共有していたことが確認できた。

---

# Author's Recipe

今回最も評価が変わったドキュメント。

当初は作者のメモ程度を想定していたが、

実際には

Knowledge

↓

Author's Recipe

↓

Repository

という導線を構成する重要な役割を果たした。

Author's Recipe は、

ベストプラクティスではなく、

作者が Repository をどのようなメンタルモデルで構築したかを共有する場所として位置付けられる。

Repository 自体も Recipe の成果物として扱う構成が自然である。

---

# Experiment

Knowledge.inbox を別途設ける予定だったが、

今回の運用を通して、

Experiment という責務へ統合できることが分かった。

Experiment は、

観測結果

実験結果

会話

仮説

を保存する場所であり、

Knowledge の下書きではない。

十分に観測が積み重なったものだけを Knowledge へ昇格させる。

---

# Public Release

今回のテストで、

CTK のコンセプトそのものは共同設計に耐えられる可能性が確認できた。

もし理解されなかった場合でも、

- User Context 依存
- Onboarding
- Knowledge
- Author's Recipe
- Repository

など改善対象を切り分けられる状態になった。

これは公開可否を判断する上で大きな成果だった。

公開対象は CLI ではなく、

共同設計という体験そのものである。

---

# Future

- Recipe Dump
- Dist Dump
- Variant
- Settings(gron)対応
- Freeze README 更新
- Settings 定義の見直し
- Author's Recipe 拡充

今回見えた Future は、

コンセプト変更ではなく、

運用をより自然にするブラッシュアップが中心だった。

これは CTK が設計フェーズから改善フェーズへ移行しつつあることを示している。


======================



