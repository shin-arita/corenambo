# テスト方針

## 必須要件

- カバレッジ100%
- 全テスト成功

---

## Backend (Go)

### 対象

- handler
- service
- repository
- その他ロジック

### 実行

docker compose exec api go test ./... -cover

### ルール

- テストなしで実装しない
- 境界値を考慮する
- 異常系を必ず含める

---

## Frontend (React)

### フレームワーク

- Vitest（テストランナー）
- @testing-library/react（コンポーネントテスト）
- jsdom（DOM シミュレーション）
- @vitest/coverage-v8（カバレッジ計測）

### 対象

- コンポーネント（レンダリング・props による表示変化・ユーザ操作）
- カスタムフック（状態変化・副作用）
- ユーティリティ関数（純粋関数のロジック）

### 実行

docker compose exec frontend npm run lint
docker compose exec frontend npm run test -- --coverage

### ルール

- テストなしで実装しない
- 境界値を考慮する
- 異常系を必ず含める

---

## E2Eテスト

### 対象機能

ユーザ登録フロー（仮登録 → 本登録 → DB反映確認）

### スクリプト

```bash
bash scripts/e2e_user_registration.sh
```

### 前提条件

以下のコンテナが起動済みであること。

```bash
docker compose up -d db api redis worker mail
```

### テストケース（全 20 項目）

#### テスト 1: 正常系（仮登録 → 本登録 → DB検証）

| ステップ | 検証内容 |
|--------|--------|
| 仮登録 POST /api/v1/user-registration-requests | HTTP 201 / code=USER_REGISTRATION_REQUEST_CREATED |
| DB確認 | user_registration_requests に1件作成されていること |
| トークン取得 | Mailpit API からメール本文のトークンを抽出できること |
| 本登録 POST /api/v1/user-registrations/verify?token={token} | HTTP 201 / code=USER_REGISTRATION_VERIFIED |
| DB確認 users | users テーブルに1件作成されていること |
| DB確認 user_emails | is_primary=true であること |
| DB確認 user_emails | verified_at がセットされていること |
| DB確認 user_credentials | 1件作成されていること |
| DB確認 user_credentials | password_hash が非空であること |
| DB確認 user_registration_requests | verified_at がセットされていること |

#### テスト 2: 異常系 — トークン不正

| 検証内容 |
|--------|
| HTTP 400 |
| code=INVALID_REGISTRATION_TOKEN |

#### テスト 3: 異常系 — トークン期限切れ

| 検証内容 |
|--------|
| DBの expires_at を過去に更新（created_at / expires_at 制約を満たすよう両方更新） |
| HTTP 400 |
| code=EXPIRED_REGISTRATION_TOKEN |

#### テスト 4: 異常系 — トークン使用済み

| 検証内容 |
|--------|
| 同一トークンで2回 Verify を呼び出す |
| 1回目 HTTP 201 |
| 2回目 HTTP 409 |
| code=USED_REGISTRATION_TOKEN |

### トークン取得の仕組み

仮登録 API 呼び出し後、worker が約 1 秒以内にメールを送信し `mail_outboxes.payload` を `'{}'` に上書きする。  
そのため `mail_outboxes` ではなく Mailpit REST API からトークンを取得する。

```bash
# メッセージIDを検索
curl "http://localhost:8025/api/v1/messages?query=to%3A{encoded_email}"

# メール本文からトークンを抽出
curl "http://localhost:8025/api/v1/message/{msg_id}" \
  | grep -o '"Text":"[^"]*"' \
  | grep -o 'token=[^\\]*' \
  | sed 's/token=//'
```

### クリーンアップ

スクリプト実行前に以下を自動実行する。

```bash
# Redis レートリミットキーを全削除
docker compose exec -T redis redis-cli \
  EVAL "local keys = redis.call('keys', 'rate_limit:*'); if #keys > 0 then return redis.call('del', unpack(keys)) else return 0 end" 0

# Mailpit のメッセージを全削除
curl -X DELETE http://localhost:8025/api/v1/messages
```

各テスト終了後に DB のテストデータ（users / user_emails / user_credentials / user_registration_requests / mail_outboxes）を削除する。

### 結果

PASS=20 / FAIL=0（2026-05-10 確認済み）