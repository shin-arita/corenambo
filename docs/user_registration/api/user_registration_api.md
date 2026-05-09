# ユーザ仮登録 API設計

## 1. 概要

本APIは、ユーザの仮登録を受け付け、本登録用トークンを含むメール送信を行うためのAPIである。  
入力されたメールアドレスに対して仮登録情報を作成または再発行し、本登録導線を提供する。

本APIはセキュリティ要件として以下を満たす：

- メールアドレスの存在有無を外部に漏らさない
- トークンは毎回再生成する
- メール送信は非同期で実行する（Outbox Pattern）

---

## 2. 目的

- ユーザがメールアドレスを用いて仮登録を開始できるようにする
- 本登録用トークンをメールで送信する
- 既存ユーザ、未認証ユーザ、期限切れデータを適切に判定する

---

## 3. エンドポイント

### 3.1 ユーザ仮登録

- Method: `POST`
- Path: `/api/v1/user-registration-requests`

### 3.2 ユーザ本登録

- Method: `POST`
- Path: `/api/v1/user-registrations/verify`

---

## 4. リクエスト仕様

### 4.1 Headers

```http
Content-Type: application/json
Accept: application/json
Accept-Language: ja
```

- `Accept-Language` は `ja` / `en` に対応する。未指定または未対応言語の場合は `ja` にフォールバックする

---

### 4.2 Body

```json
{
  "email": "user@example.com",
  "email_confirmation": "user@example.com"
}
```

---

### 4.3 パラメータ定義

| 項目                 | 型      | 必須 | 説明           |
|--------------------|--------|---:|--------------|
| email              | string |  ○ | 仮登録対象メールアドレス |
| email_confirmation | string |  ○ | 確認用メールアドレス   |

---

## 5. リクエスト制限

- 最大サイズ：1MB

---

## 6. レスポンス仕様

### 6.1 正常時（常に同一レスポンスを返却）

- Status Code: `201 Created`

```json
{
  "code": "USER_REGISTRATION_REQUEST_CREATED",
  "message": "仮登録メールを送信しました。メールをご確認ください。",
  "expires_minutes": 60
}
```

#### レスポンスフィールド

| フィールド           | 型       | 説明                    |
|-----------------|---------|------------------------|
| code            | string  | レスポンスコード              |
| message         | string  | メッセージ（i18n）           |
| expires_minutes | integer | 本登録URLの有効期限（分）        |

#### 正常コード

| コード                               | 説明         |
|-----------------------------------|------------|
| USER_REGISTRATION_REQUEST_CREATED | 仮登録メール送信受付完了 |

---

### 6.2 異常時

#### 400 Bad Request

```json
{
  "code": "BAD_REQUEST",
  "message": "リクエストが不正です。"
}
```

---

#### 422 Unprocessable Entity

```json
{
  "code": "VALIDATION_ERROR",
  "message": "入力内容に誤りがあります。",
  "errors": {
    "email": [
      "メールアドレスを入力してください。"
    ],
    "email_confirmation": [
      "メールアドレスが一致しません。"
    ]
  }
}
```

---

#### 429 Too Many Requests

```json
{
  "code": "TOO_MANY_REQUESTS",
  "message": "リクエスト回数が上限を超えました。しばらく待ってから再試行してください。"
}
```

---

#### 500 Internal Server Error

```json
{
  "code": "INTERNAL_SERVER_ERROR",
  "message": "システムエラーが発生しました。"
}
```

---

※ セキュリティ上の理由により、既存ユーザ存在エラー（409）は返却しない

---

## 7. バリデーション仕様

### 7.1 email

- 必須
- メールアドレス形式であること

#### エラーメッセージ

- メールアドレスを入力してください。
- 正しいメールアドレス形式で入力してください。

---

### 7.2 email_confirmation

- 必須
- `email` と一致すること

#### エラーメッセージ

- メールアドレスを入力してください。
- メールアドレスが一致しません。

---

## 8. 業務ルール

### 8.1 基本ルール

- 常に新しいトークンを生成する
- 同一メールでもレスポンスは常に同一
- 状態に応じてDBのみ更新する

---

### 8.2 新規仮登録

以下の場合、新規登録とする：

- 仮登録データが存在しない

---

### 8.3 仮登録済みデータがある場合

状態に応じて処理する。

#### 8.3.1 未認証かつ有効期限内

- トークン再生成
- token_hash 更新
- expires_at 更新

#### 8.3.2 有効期限切れ

- トークン再生成
- token_hash 更新
- expires_at 更新

#### 8.3.3 認証済み

- DB更新は行わない
- そのまま201を返す

---

## 9. 処理フロー

1. リクエスト受信
2. サイズチェック
3. JSONバインド
4. バリデーション
5. 仮登録確認
6. トークン生成
7. ハッシュ化
8. DB保存（トランザクション）
9. Outbox登録（メール）
10. 201返却

---

## 10. トークン仕様

- crypto/rand を使用
- 長さ：32バイト以上
- 推測困難であること
- URLセーフであること

---

## 11. トークン保存

- 平文保存禁止
- SHA256でハッシュ化して保存

---

## 12. 有効期限

- 60分（推奨）
- 定数または環境変数で管理

---

## 13. メール送信

- Outbox Pattern を使用
- mail_outboxes に登録
- worker が非同期送信

### 重要

- APIはメール送信成功を待たない
- DB成功時点で成功とする

---

## 14. 想定レスポンスメッセージ

### 14.1 正常系

- USER_REGISTRATION_REQUEST_CREATED
  - 仮登録メールを送信しました。メールをご確認ください。

---

### 14.2 エラー系

- BAD_REQUEST
- VALIDATION_ERROR
- EMAIL_REQUIRED
- EMAIL_FORMAT_INVALID
- EMAIL_CONFIRMATION_REQUIRED
- EMAIL_CONFIRMATION_NOT_MATCH
- TOO_MANY_REQUESTS
- INTERNAL_SERVER_ERROR

---

## 15. エラーコード案

| コード                          | 説明                             |
|------------------------------|--------------------------------|
| BAD_REQUEST                  | リクエスト形式不正                      |
| VALIDATION_ERROR             | 入力チェックエラー                      |
| EMAIL_REQUIRED               | email 必須                       |
| EMAIL_FORMAT_INVALID         | email 形式不正                     |
| EMAIL_CONFIRMATION_REQUIRED  | email_confirmation 必須          |
| EMAIL_CONFIRMATION_NOT_MATCH | email と email_confirmation 不一致 |
| TOO_MANY_REQUESTS            | リクエスト回数が上限超過                   |
| INTERNAL_SERVER_ERROR        | システムエラー                        |

---

## 16. i18n連携方針

- code を判定基準とする
- message は i18n で生成する
- service 層に文言を持たせない

---

---

## 17. 本登録API仕様（POST /api/v1/user-registrations/verify）

### 17.1 概要

仮登録メールに含まれたトークンを受け取り、ユーザ本登録を完了させるAPI。  
1回のリクエストで以下のテーブルへの書き込みをトランザクションで実行する。

- `users` — ユーザ本体の作成
- `user_emails` — メールアドレスの登録（主メール・検証済み）
- `user_credentials` — パスワードハッシュの登録
- `user_registration_requests.verified_at` — 本登録完了日時の更新

---

### 17.2 Headers

```http
Content-Type: application/json
Accept: application/json
Accept-Language: ja
```

- `Accept-Language` は `ja` / `en` に対応する。未指定または未対応言語の場合は `ja` にフォールバックする

---

### 17.3 Request Body

```json
{
  "token": "BvduupOZ9i1KuMzmH8Yx7chKgpJCjCFXy2SiVBvE2LU",
  "display_name": "テストユーザー",
  "password": "Password123!",
  "password_confirmation": "Password123!",
  "agreed_to_terms": true
}
```

#### パラメータ定義

| 項目                  | 型      | 必須 | 最大長 | 説明                     |
|---------------------|--------|---:|------|------------------------|
| token               | string |  ○ | 500  | 仮登録メールに記載されたトークン（平文）    |
| display_name        | string |  ○ | 100  | 画面表示用ユーザ名               |
| password            | string |  ○ | 255  | 設定するパスワード               |
| password_confirmation | string |  ○ | 255  | パスワード確認用                |
| agreed_to_terms     | bool   |  ○ | —    | 利用規約への同意（`true` のみ受け付け） |

---

### 17.4 リクエスト制限

- 最大サイズ：1MB（全エンドポイント共通ミドルウェア）
- レートリミット：IPアドレス単位 5回/分（Redis）

---

### 17.5 正常レスポンス

- Status Code: `201 Created`

```json
{
  "code": "USER_REGISTRATION_VERIFIED",
  "message": "本登録が完了しました。"
}
```

#### レスポンスフィールド

| フィールド   | 型      | 説明             |
|---------|--------|----------------|
| code    | string | レスポンスコード       |
| message | string | メッセージ（i18n）   |

---

### 17.6 エラーレスポンス

#### 400 Bad Request — トークン不正

```json
{
  "code": "INVALID_REGISTRATION_TOKEN",
  "message": "トークンが不正です。"
}
```

- トークンが空の場合
- SHA-256ハッシュが `user_registration_requests` に存在しない場合

#### 400 Bad Request — トークン期限切れ

```json
{
  "code": "EXPIRED_REGISTRATION_TOKEN",
  "message": "トークンの有効期限が切れています。"
}
```

- `expires_at < 現在時刻` の場合

#### 409 Conflict — トークン使用済み

```json
{
  "code": "USED_REGISTRATION_TOKEN",
  "message": "既に登録が完了しています。"
}
```

- `verified_at IS NOT NULL` の場合

#### 409 Conflict — メール重複

```json
{
  "code": "USER_ALREADY_REGISTERED",
  "message": "入力されたメールアドレスは既に登録されています。"
}
```

- `user_emails` に同じメールアドレス（大文字小文字を問わない）が存在する場合

#### 422 Unprocessable Entity — バリデーションエラー

```json
{
  "code": "VALIDATION_ERROR",
  "message": "入力内容に誤りがあります。",
  "errors": {
    "display_name": ["ユーザ名を入力してください。"],
    "password": ["パスワードを入力してください。"],
    "password_confirmation": ["パスワードが一致しません。"],
    "agreed_to_terms": ["利用規約への同意が必要です。"]
  }
}
```

#### 429 Too Many Requests

```json
{
  "code": "TOO_MANY_REQUESTS",
  "message": "リクエストが多すぎます。しばらく待ってから再度お試しください。"
}
```

#### 500 Internal Server Error

```json
{
  "code": "INTERNAL_SERVER_ERROR",
  "message": "システムエラーが発生しました。"
}
```

---

### 17.7 バリデーション仕様

| フィールド                | ルール                                          |
|----------------------|----------------------------------------------|
| token                | 空文字の場合は 400 を返す（バリデーションエラーとは区別）              |
| display_name         | 必須。空白のみは不可（TrimSpace後に空文字チェック）               |
| password             | 必須                                           |
| password_confirmation | 必須。`password` と一致すること                        |
| agreed_to_terms      | `true` であること                                 |

---

### 17.8 エラーコード一覧（本登録API）

| コード                          | HTTP  | 説明                        |
|------------------------------|-------|---------------------------|
| USER_REGISTRATION_VERIFIED   | 201   | 本登録完了                     |
| INVALID_REGISTRATION_TOKEN   | 400   | トークン不正または存在しない            |
| EXPIRED_REGISTRATION_TOKEN   | 400   | トークン有効期限切れ                |
| USED_REGISTRATION_TOKEN      | 409   | トークン使用済み（verified_at設定済み） |
| USER_ALREADY_REGISTERED      | 409   | メールアドレス重複                 |
| VALIDATION_ERROR             | 422   | 入力値エラー                    |
| DISPLAY_NAME_REQUIRED        | 422   | display_name 必須            |
| PASSWORD_REQUIRED            | 422   | password 必須               |
| PASSWORD_CONFIRMATION_REQUIRED | 422 | password_confirmation 必須  |
| PASSWORD_CONFIRMATION_NOT_MATCH | 422 | パスワード不一致                 |
| AGREED_TO_TERMS_REQUIRED     | 422   | 利用規約同意が必要                 |
| TOO_MANY_REQUESTS            | 429   | レートリミット超過                 |
| INTERNAL_SERVER_ERROR        | 500   | サーバ内部エラー                  |

---

### 17.9 i18n 対応

全エラーコードについて `ja` / `en` 両言語のメッセージを定義済み。  
`Accept-Language` ヘッダで言語を切り替え、未対応言語は `ja` にフォールバックする。

---

### 17.10 処理フロー

1. リクエスト受信（ボディサイズ制限 1MB）
2. JSONバインド
3. IPレートリミットチェック
4. バリデーション（display_name / password / password_confirmation / agreed_to_terms）
5. token の SHA-256 ハッシュ化
6. password の bcrypt ハッシュ化
7. トランザクション開始
   1. `user_registration_requests` を `FOR UPDATE` で取得（排他ロック）
   2. レコード存在チェック（不正トークン）
   3. `verified_at` チェック（使用済み）
   4. `expires_at` チェック（期限切れ）
   5. `user_emails` 重複チェック
   6. `users` INSERT
   7. `user_emails` INSERT（is_primary=true, verified_at=now）
   8. `user_credentials` INSERT（password_hash=bcrypt）
   9. `user_registration_requests.verified_at` UPDATE
8. 201 返却

---

### 17.11 DB反映内容

| テーブル | 操作 | 主要カラム |
|------|------|---------|
| users | INSERT | id, display_name, status='active' |
| user_emails | INSERT | id, user_id, email, is_primary=true, verified_at=now |
| user_credentials | INSERT | user_id, password_hash, password_changed_at |
| user_registration_requests | UPDATE | verified_at=now |
