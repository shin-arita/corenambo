# バックエンド方針

## アーキテクチャ

以下のレイヤ構成を厳守する

handler
service
repository
model
config
app_error
i18n
mail
token
uuid
clock
logger

---

## ルール

- handlerで生エラーを返さない
- businessロジックはservice層に集約する
- DBアクセスはrepositoryのみで行う
- 共通処理は適切な層に分離する

---

## エラーハンドリング

- code / status を持つエラーを使用する
- 内部エラーはログのみ出力
- レスポンスに詳細を含めない