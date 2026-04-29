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

---

## 4. リクエスト仕様

### 4.1 Headers

```http
Content-Type: application/json
Accept: application/json
Accept-Language: ja
```

- `Accept-Language` は以下のみ許可する
  - `ja`
  - `en`
- 未指定または不正値の場合は `ja` をデフォルトとする

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
  "message": "仮登録メールを送信しました。メールをご確認ください。"
}
```

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

- users に対象メールアドレスが存在しない
- 仮登録データが存在しない

---

### 8.3 既存ユーザ

users に同一メールアドレスが存在する場合：

- DB更新は行わない
- レスポンスは成功（201）を返却する

---

### 8.4 仮登録済みデータがある場合

状態に応じて処理する。

#### 8.4.1 未認証かつ有効期限内

- トークン再生成
- token_hash 更新
- expires_at 更新

#### 8.4.2 有効期限切れ

- トークン再生成
- token_hash 更新
- expires_at 更新

#### 8.4.3 認証済み

- トークン再生成
- token_hash 更新
- expires_at 更新
- verified_at は null にリセット

---

## 9. 処理フロー

1. リクエスト受信
2. サイズチェック
3. JSONバインド
4. バリデーション
5. users 確認
6. 仮登録確認
7. トークン生成
8. ハッシュ化
9. DB保存（トランザクション）
10. Outbox登録（メール）
11. 201返却

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
| INTERNAL_SERVER_ERROR        | システムエラー                        |

---

## 16. i18n連携方針

- code を判定基準とする
- message は i18n で生成する
- service 層に文言を持たせない

---

## 17. 今後の設計対象

- handler 設計
- service 設計
- repository 設計
- 本登録API設計
- トランザクション境界

---

## 18. 検討メモ

- verified_at の扱いは本登録APIと整合を取る
- レート制限は別設計で追加予定
