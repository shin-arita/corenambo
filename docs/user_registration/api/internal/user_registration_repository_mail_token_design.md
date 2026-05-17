# ユーザ仮登録 repository / mail / token 設計

## 1. 概要

本設計は、DB / メール / トークンの責務分割を定義する。

---

## 2. 全体方針

- repository：DB操作
- mail：Outbox送信
- token：生成・ハッシュ

---

## 3. repository

### UserRegistrationRequest

- FindByEmail
- Create
- UpdateToken

---

### MailOutbox

- Create（登録）
- FetchPending
- MarkProcessing
- MarkSent
- MarkRetry（送信失敗・再送可能）
- MarkFailed（最終失敗・再送不可）

---

## 4. Outbox Pattern

### フロー

1. service が mail_outboxes に登録
2. worker が pending レコードを取得
3. SMTP送信
4. 成功時: MarkSent（status=sent）
5. 失敗時: MarkRetry または MarkFailed で状態更新

---

### 特徴

- APIはメール送信を待たない
- 再試行可能
- 障害耐性あり

---

### retry ロジック

| 条件 | 呼び出し | status | retry_count |
|------|---------|--------|-------------|
| SMTP送信失敗（接続エラー等） | MarkRetry | pending | +1 |
| max retry 到達（WorkerMaxRetryCount 以上） | MarkFailed | failed | 変更なし |
| payload 不正（JSON parse 失敗） | MarkFailed | failed | 変更なし |
| 送信前に確定できる失敗（NonRetryableMailError） | MarkFailed | failed | 変更なし |

- `MarkRetry` は `status = 'pending'`, `retry_count = retry_count + 1`, `next_attempt_at = 現在 + 5分` に更新する
- `FetchPending` は `status = 'pending' AND next_attempt_at <= NOW()` で取得するため、次回のポーリング時に再送対象となる
- `MarkFailed` は `status = 'failed'` のみを設定し、`retry_count` は増やさない（最終状態なので変更不要）
- invalid payload（JSON parse 失敗）は retry しても解消しないため、即 MarkFailed（非retryable failure）
- max retry 到達時は `next_attempt_at = 24時間後` で MarkFailed し、FetchPending の対象外となる

### NonRetryableMailError

`mail.NonRetryableMailError` は送信前に確定できる非retryable なエラーを表す型。
worker は `errors.As` でこの型を検出し、`MarkRetry` の代わりに `MarkFailed` を呼ぶ。

非retryable とみなす条件（`mail` パッケージ内で返す）：

| 条件 | エラーメッセージ |
|------|-------------|
| payload の `url` フィールドが空 | `registration URL is empty` |
| メールテンプレートの parse 失敗 | テンプレートエラーメッセージ |
| メールテンプレートの execute 失敗 | テンプレートエラーメッセージ |

SMTP接続失敗・タイムアウト等の送信エラーは retryable として `MarkRetry` が呼ばれる。

---

### payload スキーマ

`mail_outboxes.payload` は JSON 文字列として保存する。

```json
{
  "email": "user@example.com",
  "url": "https://example.com/registration/verify?token=xxx",
  "lang": "ja"
}
```

| フィールド | 型     | 説明                          |
|---------|--------|-------------------------------|
| email   | string | 送信先メールアドレス            |
| url     | string | 本登録URL（生トークン付き）     |
| lang    | string | メール送信言語（`ja` / `en`）  |

`expires_minutes` は payload には含まない。worker が環境変数 `REGISTRATION_TOKEN_EXPIRES_MINUTES` の設定値を使用する。

---

## 5. token

- Generator：ランダム生成
- Hasher：SHA256

---

## 6. URL

```text
/registration/verify?token=xxx
```

---

## 7. worker

`cmd/worker` が outbox の正式処理経路。本番環境で動作し、実際にメールを送信する。

- 1秒間隔でポーリング
- pending を取得
- retry制御あり（NonRetryableMailError / retryable error / max retry の3分岐）

---

## 8. メリット

- DBとメールの分離
- 再送可能
- 安定性向上
