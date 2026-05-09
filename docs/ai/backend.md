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

---

## 実装済みAPI

### ユーザ仮登録

```
POST /api/v1/user-registration-requests
```

- メールアドレスを受け取り、本登録用トークンを含むメールを送信する
- トークンは SHA-256 ハッシュで DB に保存する（平文保存禁止）
- メール送信は Outbox Pattern（worker による非同期処理）

### ユーザ本登録

```
POST /api/v1/user-registrations/verify
```

- トークン（平文）を受け取り、SHA-256 ハッシュで `user_registration_requests` を照合する
- 以下のテーブルを単一トランザクションで書き込む
  - `users` — ユーザ本体（display_name, status='active'）
  - `user_emails` — メールアドレス（is_primary=true, verified_at=now）
  - `user_credentials` — bcrypt ハッシュ化したパスワード（DefaultCost）
  - `user_registration_requests.verified_at` — 本登録完了日時
- `FOR UPDATE` による排他ロックで同一トークンの二重処理を防ぐ

---

## テーブル責務

| テーブル | 責務 |
|--------|------|
| user_registration_requests | 仮登録トークンの管理。token_hash のみ保存し平文は持たない |
| users | ユーザ本体。display_name と status を持つ |
| user_emails | メールアドレス管理。LOWER インデックスで大文字小文字を区別しない |
| user_credentials | パスワードハッシュ管理。bcrypt ハッシュのみ保存し平文は持たない。現時点ではパスワード認証情報のみを保持する。将来的に認証方式を拡張する場合も、users 本体ではなく認証情報として責務を分離する |
| mail_outboxes | 非同期メール送信キュー。送信後 payload を `'{}'` で上書きする |