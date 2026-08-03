# project-knowledge.experiment.ai-collaboration-bootstrap.md
============================================================

# Experiment: CTK as an AI Collaboration Bootstrap

## Observation

CTKはIDE環境だけでなく、AIとの協業を始められる状態もbootstrapしている、と見ることが
できる。

これはCTKの目的を定義し直す提案ではない。CTK自身をKnowledgeとAIを使って開発した
結果、既存の構成を説明する別の視点が見つかった、という観察である。

## Starting point

CTKは当初、VS CodeなどのAI IDE環境を再現・管理するツールとして設計を始めた。
主な関心は次のようなものだった。

- Runtime
- Profile
- Extension
- Recipe
- Build / Apply

Project Knowledgeも、開発中に生まれた設計メモや観察記録という位置付けだった。

## What changed through Knowledge

Knowledgeを使った日常的な運用を続けるうちに、CTKを別の角度から見られるように
なった。Knowledgeには設計思想や運用方法だけでなく、AIとの壁打ちやRawの蓄積も
含まれる。

この運用の中で、AIは単なるコード生成ツールではなく、Knowledgeを利用しながら
共同作業を行う参加者として振る舞うようになった。

その結果、CTKが準備しているものはIDE設定だけではなく、次を含む協業環境そのもの
であると気付いた。

- AIが理解できるKnowledge
- AIが利用できるRecipe
- AIが案内に使えるREADME
- 人とAIが共有するworkspace

## Onboarding as a growing environment

この見方では、CTKの対象を既存のVS Code利用者だけに限定する必要はない。
例えば新人の参加時にも、次のような流れが成立する。

1. Basic環境を展開する。
2. AIがREADMEを読み、オンボーディングを案内する。
3. 担当業務が分かった時点で、必要なRecipeを追加する。

「Java担当です」と分かった時点でJava Recipeを追加すればよく、最初から巨大な
テンプレートを用意する必要はない。オンボーディング自体がRecipeを育てる作業にも
なる。

この流れは開発者だけに限定されない。営業でも管理職でも、AIとの相談から始め、
必要ならRawへ残し、Knowledgeへ育てられる。扱うKnowledgeは違っても、協業の流れは
共通になり得る。

## Project Knowledge as a case study

この視点では、Project Knowledgeは単なる開発メモではなく、CTKを使ってAIと協業した
実践例としても読める。

- AIとどのように壁打ちしたか
- Knowledgeをどう育てたか
- なぜ現在の構成になったのか

これは「こうすべき」という仕様ではなく、「このプロジェクトではこの方法で進め
られた」というケーススタディである。

READMEが「AIと協業しながらCookbookを育てる」という入口を示すなら、Project
Knowledgeは「実際にはどのような運用だったか」を示せる。理想論だけでなく、CTK
自身がこの方法で育ったという事実が説明を支えている。

README、Knowledge、Author's Recipe、Project Knowledgeは、それぞれ異なる役割を持つ
リファレンス実装とも考えられる。

> Project Knowledgeの価値を作ったのではなく、価値を説明する視点が見つかった。

## Nature of the observation

これは最初から設計していた価値ではない。Knowledge運用とAIとの壁打ちを繰り返し、
CTK自身を使って開発した結果として自然に見えてきた。

したがって「CTKの目的を定義し直した」というより、「CTKを別のレンズから眺められる
ようになった」という表現が近い。新機能を発見したのではなく、README、Knowledge、
Cookbook、Recipeという既存の構成に別の説明が与えられた。

> プロダクトを作るための方法論が、プロダクトのユースケースを発見した。

今すぐREADMEやcurated Knowledgeへ反映する必要はない。棚卸し時に、READMEの短い
説明にするか、Knowledgeへ役割として反映するか、Project Knowledgeのケーススタディ
として残すかを判断できる。

## Candidate expressions

> CTK helps bootstrap a shared development environment for both people and AI.

> CTK is a tool for bootstrapping an environment, and also a working example of
> growing that tool together with AI.

Longer candidate:

> CTK started as a toolkit for managing AI IDE environments.
>
> As we continued developing with AI, we realized something unexpected.
> Managing the environment was only the beginning. What actually mattered was
> giving both people and AI a shared starting point: a common workspace, shared
> knowledge, and a way to grow them together.
>
> Today, CTK is designed to bootstrap not only an IDE, but AI collaboration
> itself.

More observational candidate:

> CTK was originally created to reproduce IDE environments.
>
> Through building CTK itself, Knowledge naturally emerged. As Knowledge
> evolved, it became clear that CTK was doing more than reproducing editor
> settings: it was providing the bootstrap for AI-assisted collaboration.
>
> That observation continues to shape the project today.

Compact candidate:

> We built CTK to manage AI IDE environments.
>
> While building CTK, we started collaborating with AI through Knowledge.
>
> Eventually, we realized something unexpected: CTK was not just the product we
> were building; it had become the workspace where we were building it.

Related statement:

> CTK itself is developed using the workflow it encourages.

## Related experiment

[`Role-Oriented AI Collaboration`](experiment.role-oriented-ai-collaboration.md)
examines the design structure that may have made this use case emerge.
