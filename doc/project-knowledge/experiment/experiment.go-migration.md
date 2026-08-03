# Experiment: CTK Go Migration

> **Inventory status: Reviewed (2026-07-29)**
>
> This file remains the primary record of the completed CTK Bash-to-Go
> migration experiment. Current Go behavior and implementation decisions belong
> to `../../../go/README.md` and the relevant Knowledge / Contracts.
>
> The focused hypothesis extracted from this record remains experimental in
> `experiment.responsibility-first-implementation-migration.md`. The broader
> Concept-development observation is in
> `../note/note.development-style.md`.

## What happened

CTKのBash実装を保存しながら、Goによる新しい実装を約2日間で構築した。

ただし、実際に行ったことはBashコードの逐語的な移植ではなかった。

READMEから既存のConceptを読み、Contractを作り、実装で観測し、実機で試し、
違和感があればContractの抽象度まで戻って考え直した。

結果として、Cookbookの解決からBuild、Apply、Lock、CodeVenv、Workbench、
Archive、Direct Launcher、Release buildまでがGo実装として接続された。

この記録は完成した仕様を説明するものではない。

短期間の共同作業で、何が理解と実装を加速させたのかを振り返るための観測である。

### What the migration was testing

Go版を作ること自体が、実験の中心ではなかった。

当時の問いは、まとまり始めたKnowledgeから、既存のBash実装に引っ張られずに別実装を
組み立てられるか、というものだった。

```text
Knowledgeを読む
      ↓
判断が揺れそうな境界を見つける
      ↓
Knowledgeからraw Contractを導出する
      ↓
オーナーが判断する
      ↓
Goで実装して観測する
```

この意味でGo Migrationは、Knowledgeだけで別実装を導けるかを試し、足りない境界を
Contractとして発見する実験でもあった。

AIへの依頼も、単に「Goを書いて」ではなかった。各MilestoneでKnowledgeを読み、必要な
Contractと判断が必要な点を提示し、オーナーの判断を受けてから実装した。AIはコード
生成器というより、Concept Reviewerとして参加していた。

この実験で使ったContractがrawだったことは、単に未完成だったことを意味しない。
それらはKnowledgeから実験用に導出され、Go実装によって観測されたContractでもある。

実験後に残る問いは、各raw Contractをそのまま正式化することではなく、どの部分が
共有Knowledge Contract、Go固有Contract、Note、または不要な実験足場に属するかを
責務ごとに棚卸しすることである。この記録はその再分類を確定しない。

The later inventory promoted shared agreements under
[`doc/contract`](../../contract/README.md) and separated observable Go choices
under [`go/doc/contract`](../../../go/doc/contract/README.md). This Experiment
retains the question that led to that outcome rather than duplicating the
resulting Contracts.

---

## Overall impression

最初の印象は「移植対象としては小さくない」だった。

Bashのコード量だけでなく、CTKには次のような性質があった。

- Host環境を書き換える操作がある。
- VS Code-family固有の挙動がある。
- Cookbookという独自の入力形式がある。
- Buildした状態をLock、Freeze、Archiveとして異なる目的で観測する。
- 既存実装の偶然と、本来のConceptを区別する必要がある。

それにもかかわらず、途中から実装速度は落ちず、むしろ速くなった。

これは単にGoが書きやすかったからではない。

共同作業の中で判断基準が育ち、後半では新しい問題を既存の言葉で分類できるように
なったことが大きかった。

例えば、新しい挙動を見つけたとき、次のように考えられるようになった。

- これはCoreが要求する能力か。
- 現在のDistribution表現に固有の制約か。
- Go実装が選んだ戦略か。
- Kitchen Noteとして宣言する特殊な解釈か。
- Workbenchの中だけで使う一時的な語彙か。
- Bashを読む必要がある問題か。

分類できる問題は、実装判断も速かった。

---

## What worked well

### Documentation was read before Bash

Go実装では、Documentを優先し、Bashは最後のreferenceとした。

この順序は非常に有効だった。

Bashを先に読むと、関数構造、ファイル配置、外部CLI、shell上の都合をそのまま
必要条件として解釈しやすい。

先にDocumentから責務を理解したことで、Bashは「何をしているかを確認する資料」に
なり、「Goをどう書くかの設計図」にはならなかった。

その結果、Goではpackage、型、Adapter、staging、Report、検証境界として自然に
表現できた。

### Contract and implementation corrected each other

Contractを先に書いたことはよかったが、最初から完全だったわけではない。

特に重要だったのは、Distribution Contractを具体的なディレクトリ構成として
書きすぎていると気づいた場面だった。

本当に必要だったのは、例えば次のような能力だった。

- 起動できる。
- BuildやApplyの由来を取得できる。
- Lockとして観測できる。
- Launch Overrideを持てる。

`.data`、`.ext`、`.meta/recipe.yaml`のような構成は、現在のVS Code-family Runtimeを
実現するためには重要でも、あらゆるDistributionに要求するConceptではなかった。

実装がContractの過剰な具体性を発見し、会話がContractを修正し、その修正が以降の
実装自由度を上げた。

Contract-firstだったが、Contract-fixedではなかったことがよかった。

### Small real executions were part of design

各Milestoneで、実際のコマンド出力や生成物を早い段階で確認した。

例えば次のような観測が、そのまま設計入力になった。

- Go CLIはBash版より体感上の応答が速かった。
- VSIXのファイル名は正しくても、内部metadataが大文字を保持する場合があった。
- 相対Dist pathがDist rootと二重結合される問題があった。
- Extension directoryの単純なcopyやmoveではIDE側のfingerprintと整合しない。
- Direct LauncherはCTKを経由せず、Platform commandを直接呼べる。

これらは大規模な検証工程ではなく、会話の途中にある短い実行だった。

短いFeedback Loopが、長い推測を置き換えた。

### The user owned meaning, the AI explored representation

今回の共同作業では、Conceptの意味は主にユーザー側にすでに存在していた。

AIは、次のような仕事を繰り返した。

- 暗黙の意図を複数の候補へ展開する。
- Contract、Note、Current Strategyのどこへ置くか提案する。
- 実装可能な境界へ翻訳する。
- 実装後の違和感を言語へ戻す。
- 一度決めた言葉が問題を複雑にしていないか見直す。

ArrayとSetの議論を、最終的にreplaceとunionという操作へ戻した場面は象徴的だった。

型のイメージは思考を助けたが、実現したい責務は二つのmerge operatorだった。

言葉が設計を助ける一方で、言葉自身が設計を過剰に複雑化することもある。

その兆候を会話で検出できた。

### A completed slice made the next slice faster

初期のMilestoneでは、ひとつの機能を作るたびにCTKの責務を確認する必要があった。

しかし、Selector、Cookbook Resolution、Platform Runtime I/O、Lockができると、後続の
機能はそれらの組み合わせとして考えられるようになった。

```text
Cookbook
    ↓
Build / Apply
    ↓
Distribution
    ↓
Lock
    ↓
Freeze / View / Sync
    ↓
Cookbook
```

```text
Distribution
    ↓
Archive
    ↓
Build / Apply
    ↓
Distribution
```

後半の速度は、コード生成の速度というより、前半で作られたConceptual Infrastructureの
効果だった可能性が高い。

---

## What could be improved

### Contract levels could have been labeled earlier

途中から、次の違いをかなり意識するようになった。

- Required capability
- Recommended behavior
- Current Go strategy
- Bash reference behavior
- Future possibility

この分類を最初からDocument templateとして持っていれば、DistributionやLaunchの
Contractを具体化しすぎる往復は減らせたかもしれない。

ただし、最初から分類だけを厳密にすると、まだ見えていないConceptを形式へ押し込む
危険もある。

改善案は、分類を必須フォーマットにすることではなく、議論が実装形式へ寄りすぎた
ときの確認用lensとして用意することである。

### Milestone completion checks could be more uniform

各Milestoneでは十分なテストと実機確認を行ったが、完了条件の表現はMilestoneごとに
少し異なっていた。

次回は、各sliceの最後に次を短く記録すると追跡しやすい。

- Document updated
- Unit behavior verified
- Cross-platform build verified
- Real operation observed
- Remaining host-specific validation
- Future intentionally deferred

これは重いchecklistである必要はない。

「自動テストで完了したこと」と「実機で今後も観測すること」を分けるだけでもよい。

### Generated and authored artifacts should be distinguished immediately

Direct Launcherの終盤で、Applyが同名ファイルをどう扱うかという問題が出た。

生成物であることを示すmarker、未知のファイルを上書きしない規則、Launch Overrideとの
名前衝突は、生成Artifactを導入する時点で共通に確認できる論点だった。

今後、新しい生成Artifactを作るときは早い段階で次を確認したい。

- 誰が所有するファイルか。
- 再生成してよいか。
- 人間の編集をどう検出するか。
- 衝突時に保持、警告、失敗のどれを選ぶか。
- stagingとrollbackの境界はどこか。

### Release distribution exposed workspace discovery late

HomebrewやScoopを視野に入れたとき、binary位置からproject rootを求める従来の前提が
成立しなくなることに気づいた。

これはRelease buildそのものより重要な発見だった。

配布を考えるときは、binary formatだけでなく、次も同時に確認すべきである。

- 状態はどこにあるか。
- executableとworkspaceは同居する必要があるか。
- current directoryから発見できるか。
- 明示的なworkspace指定はあるか。
- upgrade後も生成Launcherが動くか。

Packagingは、単なるcompile targetの追加ではない。

実行時のlocation assumptionsを観測する作業でもある。

### Some discussions could have used a tiny example sooner

抽象度の高い議論が長くなりかけた場面では、小さな入力と期待出力を置くと急に
収束した。

CTK JSON Flat Format、配列selector、Direct Launcher、Archive Extension versionなどが
その例だった。

次回は、用語の議論が二往復ほど続いた時点で、次の最小例を置くとよさそうである。

```text
input
current state
operation
expected output
```

例はContractの代わりではない。

言葉が同じ対象を指しているか確認するprobeとして使える。

---

## Knowledge likely to remain useful

### Preserve concepts, not implementation ancestry

Migrationで保存すべきものは、古い実装の形ではなく、その実装が満たしていた責務である。

Source compatibilityが必要なCookbookはそのまま利用できるようにした。

一方で、Bashの関数構造、外部CLI依存、Distribution内部表現までGoへ保存することは
要求しなかった。

互換性を一語で扱わず、何の互換性かを分解することが重要だった。

### Reference implementation needs a language declaration

Bashをreferenceとして保存するだけでは、そのコードがCoreなのかBash固有なのか判断しにくい。

言語ごとのREADMEが、適用するKitchen Notes、実装戦略、既知の差分を宣言する層として
機能した。

```text
Core / Contract
    ↓
Language declaration
    ↓
Reference source
```

この一層があることで、sourceを読む前にその実装の立場を理解できる。

### Workbench can be deliberately less strict than Runtime mutation

FreezeやInspectでは、すべてを自動検証するより、人間とAIが読んで編集できることを
優先した。

一方、Activate、Archive publish、Host path切替では、staging、trusted Lock、hash、
rollback、Safety Gateを重視した。

すべての機能へ同じ安全性を適用するのではなく、失敗時の損失に応じて厳密さを変える。

```text
Workbench
    → visible, editable, recoverable

Runtime convergence
    → observable, repeatable, reports partial failure

Host integration / Archive publication
    → staged, verified, rollback-aware
```

この区別は他のtoolingでも再利用できそうである。

### A warning can be a first-class result

キャンセル、unresolved operation、未知の生成Artifact、Launch Overrideの非復元など、
成功か失敗だけでは表現しにくい状態が何度も現れた。

これらを例外的な文字列ではなく、Report、Warning、Safety Gateとして扱ったことで、
処理責務とCLI表示を分けられた。

「続行できるが、無視してはいけない結果」を設計上の状態として持つことは有効だった。

### Fast collaboration depends on accumulated shared language

2日間ずっと同じ速度だったわけではない。

初期は、言葉の意味と境界を確認する時間が多かった。

後半は、過去の議論で作った言葉を使って、新しい問題を短く説明できた。

速度は個々の返答の速さではなく、会話が捨てられずに次の判断基準として蓄積された
ことから生まれた可能性が高い。

```text
Conversation
    ↓
Shared language
    ↓
Documented boundary
    ↓
Smaller next discussion
    ↓
Faster implementation
```

---

## Hypotheses for future work

今回の経験から、次の仮説が残った。

1. Conceptが比較的安定しているMigrationでは、code-firstより
   responsibility-firstの方が最終的な実装速度も速くなる可能性がある。
2. AIとの共同実装では、完全な仕様書より、判断の優先順位が明記されたDocumentの方が
   未知の問題へ対応しやすい可能性がある。
3. 実装ごとのREADMEは使用方法だけでなく、Coreに対するその言語の立場を宣言する
   layerとして機能する可能性がある。
4. Contractを固定物ではなく、実装によって反証可能な仮説として扱うと、過剰な抽象化と
   過剰な具体化の両方を修正しやすい可能性がある。
5. 小さな実機観測を会話へ頻繁に戻すことで、長い設計と最後の統合試験を分離するより
   早く安全に収束できる可能性がある。

これらは一般的な方法論として確立されたものではない。

CTKというConcept中心の個人tool、継続した同一Conversation、すぐに実行できるRepository、
そしてユーザーが意味の最終判断を持っていたという条件に強く依存している可能性がある。

---

## Closing reflection

最もよかったのは、短期間で大量のコードが書けたことだけではない。

実装が進むほど、CTKを説明する言葉が増えるのではなく、少ない言葉でより多くの挙動を
説明できるようになったことだった。

Go Migrationは、既存のBash版を捨てる作業ではなかった。

Bash版が実用の中で見つけていたConceptを保存し、そのConceptと実装上の偶然を分離し、
別の言語で再び組み立てられるかを試す実験だった。

結果として、Bashは読みやすいreferenceとして残り、GoはBashに引っ張られない独立した
実装になった。

そして、Documentは実装前の説明でも実装後の報告でもなく、Conversationと実装の間で
判断を保存するworking memoryとして機能した。

この点が、今回もっとも次へ残したい観測である。
