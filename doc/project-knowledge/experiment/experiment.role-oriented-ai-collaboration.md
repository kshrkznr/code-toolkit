# project-knowledge.experiment.role-oriented-ai-collaboration.md
============================================================

# Experiment: Role-Oriented AI Collaboration

## Observation

CTKはAI対応を目指した結果、この形になったのではない。

> 工程を分離した結果、AIが自然に参加できる構造になった。

この観察が扱うのはAI機能ではなく、担当者から独立した役割として工程を設計すること
と、そこへ人やAIが参加できることの関係である。

## Starting point: splitting Freeze

CTKは最初からAIとの連携を目的に設計したわけではない。

転換点になったのは、Freezeを **Draft** と **Commit** に分割したことだった。当初の
目的は、自動化できない部分を人間がレビューしやすくすることだった。

しかし工程を分離すると、レビューは人間だけでなくAIにも任せられると気付いた。
この発想は後にKnowledgeやInspectにも広がった。

## Separate the work instead of embedding AI

CTKはAI機能を持たない。LLMを呼び出すことも、特定のAIサービスに依存することも
ない。それでもAIと相性が良いのは、AIを組み込んだからではなく、知的作業を工程と
して分離したからである。

AIはReviewだけでなく、Recipeを書き、Ingredientsを整理し、Knowledgeを編集し、
Buildすることもできる。CommitやApplyも、オーナーが許可するなら担当できる。

重要なのは担当者の種類ではない。誰にどの工程を任せるかは、オーナーが決めれば
よい。

## Roles rather than actors

CTKは「人間だからレビューする」「AIだから生成する」とは考えない。代わりに、
次のような役割を定義する。

- Inspect
- Design
- Build
- Commit
- Apply

それを人間、AI、チームメンバー、将来の自分の誰が担当するかは運用で決められる。

> CTKが設計しているのはActorではなくRoleである。

この意味での「AI Collaboration」はAI機能を意味しない。人間とAIが自然に協力できる
工程を提供する、という見方である。人間だけでも、その混在でも成立し、AIは特別な
存在ではなく知的な参加者の一人になる。

## Growing the IDE and Cookbook

例えば「このテーマは見づらい」とAIに相談し、Recipeを更新し、Buildし、Applyする
体験もこの延長にある。

重要なのはAIが直接IDEを書き換えることではない。変更がCookbookやRecipeとして
管理され、再現可能なプロダクトとして育っていくことである。

この考え方はCookbookを提供する側にも使える。会社が標準Cookbookを提供し、Project
Knowledgeで設計意図を共有する。利用者はAIと相談しながら環境を調整し、AIはその
Knowledgeを読んで標準を理解した上で提案する。

「Kiroを使ってください」だけでなく、次のようなオンボーディングが可能になる。

> このProject Knowledgeを読んで、このCookbookを育ててください。

## Reflection

Freezeを分割したことで、責務を通して協業する発想が生まれた。そこからAIは「機能」
ではなく「参加者」として扱えると気付いた。この違いは、その後のKnowledge、Inspect、
Resolverなどの設計にも共通して現れている。

## Candidate welcome

> Welcome, collaborator.
>
> Whether you are a human, an AI, or something else is not important.
>
> Understand the project.
>
> Respect the responsibilities.
>
> Improve the product.
>
> Leave it better than you found it.

## Related experiment

[`CTK as an AI Collaboration Bootstrap`](experiment.ai-collaboration-bootstrap.md)
examines the broader product and onboarding lens that emerged from this structure.
