# project-knowledge.experiment.observation.md
============================================================

> This observation was extracted from the
> [Freeze Draft Collaborative Design](experiment.freeze-draft-collaborative-design.md)
> experiment.

## Observation: Shared Structure Changes Collaboration

今回の実験では、AI が未知のツールや Repository を扱う場合でも、構造や責務、運用ルールが共有されていると、既知のツールを扱うような自然な共同作業へ移行する可能性が観測された。

重要なのは、AI が個別のツールを知っていたことではない。

共同編集の中で、

- 責務
- ドキュメント構造
- 運用文化
- 共通言語

が徐々に共有されることで、

個別機能の説明を減らし、本来の設計議論へ集中できる状態が生まれた。

今回のケースでは、途中からツール自体を評価する場面が減り、

ツールを前提とした設計レビューへ自然に移行した。

これは未知のツールが既知のツールになったというより、

共同編集者の間で共通のメンタルモデルが形成された結果である可能性が高い。

十分な構造・共通言語・運用文化が共有されると、未知のツールであっても共同設計の前提として扱われる状態が生まれる可能性がある。
今回の共同設計は、新しい体験というより、既存の共同設計体験を個人Repository上へ持ち込めた感覚に近かった。

## Related Observation: Inference from Shared Context

Knowledgeを共有した後、過去の会話を直接引用しなくても、以前の議論と近い提案が自然に現れる場面があった。

これは会話を記憶していたことの証明ではない。

共有された

- Concept
- Responsibility
- Authorの判断材料
- Documentation structure

から、現在の状況に合う提案が推論された可能性がある。

現時点では一つの運用テストで得られた観察であり、再現可能な仕組みや保証された挙動としては扱わない。

今後は、この現象が

- Repository の構造
- ドキュメント設計
- Onboarding
- User Context

など、どの要素にどの程度依存するのかを継続して観測したい。

---
