# ユーザ仮登録 API設計

## 1. 概要

本APIは、ユーザの仮登録を受け付け、本登録用トークンを含むメールを送信するためのAPIである。  
入力されたメールアドレスに対して仮登録情報を作成または再発行し、本登録導線を提供する。

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

- `Accept-Language` は任意
- 未指定時は `ja` をデフォルトとする

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

## 5. レスポンス仕様

### 5.1 正常時

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
| USER_REGISTRATION_REQUEST_CREATED | 仮登録メール送信完了 |

---

### 5.2 異常時

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

#### 409 Conflict

```json
{
  "code": "USER_ALREADY_REGISTERED",
  "message": "入力されたメールアドレスは既に登録されています。"
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

## 6. バリデーション仕様

### 6.1 email

- 必須
- メールアドレス形式であること
- 既存ユーザテーブルに同一メールアドレスが存在しないこと

#### エラーメッセージ

- メールアドレスを入力してください。
- 正しいメールアドレス形式で入力してください。
- 入力されたメールアドレスは既に登録されています。

---

### 6.2 email_confirmation

- 必須
- `email` と一致すること

#### エラーメッセージ

- メールアドレスを入力してください。
- メールアドレスが一致しません。

---

## 7. 業務ルール

### 7.1 新規仮登録

以下の条件をすべて満たす場合、新規仮登録として登録する。

- users に対象メールアドレスが存在しない
- user_registration_requests に有効な未認証データが存在しない

実施内容：

- UUID v7 を Go 側で生成する
- ランダムトークンを生成する
- トークンはハッシュ化して保存する
- 有効期限を設定する
- 認証日時は null とする
- 仮登録メールを送信する

---

### 7.2 既存ユーザ

users に同一メールアドレスが存在する場合はエラーとする。

- Status Code: `409 Conflict`
- Error Code: `USER_ALREADY_REGISTERED`

---

### 7.3 仮登録済みデータがある場合

user_registration_requests に同一メールアドレスが存在する場合、状態に応じて処理を分岐する。

#### 7.3.1 未認証かつ有効期限内

- 再送扱いとする
- 新しいトークンを生成する
- token_hash を更新する
- expires_at を更新する
- 仮登録情報は残す
- 仮登録メールを再送する

#### 7.3.2 有効期限切れ

- 再発行扱いとする
- 新しいトークンを生成する
- token_hash を更新する
- expires_at を更新する
- 仮登録情報は残す
- 仮登録メールを再送する

#### 7.3.3 認証済み

- 再発行扱いとする
- 新しいトークンを生成する
- token_hash を更新する
- expires_at を更新する
- verified_at の扱いは別API設計と整合を取ること
- 仮登録メールを再送する

---

## 8. 処理フロー

1. リクエスト受信
2. JSONバインド
3. リクエスト形式不正なら 400 を返却
4. 入力バリデーションを実施
5. バリデーションエラーがある場合は 422 を返却
6. users を検索し、同一メールアドレスの既存ユーザ有無を確認
7. 既存ユーザが存在する場合は 409 を返却
8. user_registration_requests をメールアドレスで検索
9. 対象データなしの場合は新規作成
10. 対象データありの場合は状態に応じて更新
11. ランダムトークンを生成
12. トークンをハッシュ化
13. DBへ保存
14. メール送信
15. 正常終了時は 201 を返却

---

## 9. 処理詳細

### 9.1 トークン生成

- トークンはランダム値とする
- トークン自体に意味を持たせない
- 十分な長さと推測困難性を持たせる

---

### 9.2 トークン保存

- 平文では保存しない
- ハッシュ化して保存する
- メール送信時のみ平文トークンを使用する

---

### 9.3 有効期限

- 仮登録トークンには有効期限を設定する
- 具体的な期限時間は別途定数化する

例：

- 24時間
- 60分

※ 本設計では値は未確定とし、実装時に定数で管理する

---

### 9.4 メール送信

- 本登録用URLをメール本文に含める
- URLには平文トークンを含める
- メール送信失敗時は 500 を返却する
- DB更新とメール送信の整合性は実装設計で別途整理する

---

## 10. 想定レスポンスメッセージ

### 10.1 正常系

- USER_REGISTRATION_REQUEST_CREATED
  - 仮登録メールを送信しました。メールをご確認ください。

---

### 10.2 エラー系

- BAD_REQUEST
  - リクエストが不正です。
- VALIDATION_ERROR
  - 入力内容に誤りがあります。
- EMAIL_REQUIRED
  - メールアドレスを入力してください。
- EMAIL_FORMAT_INVALID
  - 正しいメールアドレス形式で入力してください。
- EMAIL_CONFIRMATION_REQUIRED
  - メールアドレスを入力してください。
- EMAIL_CONFIRMATION_NOT_MATCH
  - メールアドレスが一致しません。
- USER_ALREADY_REGISTERED
  - 入力されたメールアドレスは既に登録されています。
- INTERNAL_SERVER_ERROR
  - システムエラーが発生しました。

---

## 11. エラーコード案

| コード                          | 説明                             |
|------------------------------|--------------------------------|
| BAD_REQUEST                  | リクエスト形式不正                      |
| VALIDATION_ERROR             | 入力チェックエラー                      |
| EMAIL_REQUIRED               | email 必須                       |
| EMAIL_FORMAT_INVALID         | email 形式不正                     |
| EMAIL_CONFIRMATION_REQUIRED  | email_confirmation 必須          |
| EMAIL_CONFIRMATION_NOT_MATCH | email と email_confirmation 不一致 |
| USER_ALREADY_REGISTERED      | 既存ユーザあり                        |
| INTERNAL_SERVER_ERROR        | システムエラー                        |

---

## 12. i18n連携方針

- API内部では code を返却制御の基準とする
- 表示文言は `internal/i18n` で解決する
- エラー構造は `internal/app_error` で統一する
- 正常系・異常系ともに `code + message` で統一する
- service は表示文言を直接保持しない

想定：

- `code`: システム判定用
- `message`: 多言語化済みユーザ向け文言
- `errors`: 項目単位の詳細エラー

---

## 13. 今後の設計対象

本API設計の次工程では以下を定義する。

- handler 設計
- service 設計
- repository 設計
- メール送信インターフェース設計
- トークン生成インターフェース設計
- 本登録API設計
- verified_at の再発行時ルール確定
- トランザクション境界の整理

---

## 14. 検討メモ

### 14.1 認証済みデータ再発行時の扱い

「有効期限切れ or 認証済みは再発行」とあるため、認証済みレコードを再利用するか、新規レコードを作るかは今後詳細化が必要である。  
現時点では「同一レコードを更新して再発行」を前提とする。

### 14.2 成功レスポンスの扱い

正常系も異常系と同様に `code + message` で返却する。  
表示文言は i18n で解決し、service にはベタ書きしない。

### 14.3 レート制限

短時間の大量送信対策として、将来的にはメールアドレス単位やIP単位でレート制限を追加する余地がある。  
本設計では対象外とする。
