# 手動テスト手順書 — 仮会員登録

## 目次

1. [テスト環境](#1-テスト環境)
2. [前提条件](#2-前提条件)
3. [正常系](#3-正常系)
4. [異常系 — バリデーションエラー](#4-異常系--バリデーションエラー)
5. [異常系 — 同一メール連続送信](#5-異常系--同一メール連続送信)
6. [異常系 — レートリミット](#6-異常系--レートリミット)
7. [Mailpit 確認](#7-mailpit-確認)
8. [APIログ確認](#8-apiログ確認)
9. [確認観点まとめ](#9-確認観点まとめ)

---

## 1. テスト環境

| 項目 | 値 |
|------|---|
| フロントエンド | http://localhost:5173 |
| API | http://localhost:8080 |
| Mailpit UI | http://localhost:8025 |
| SMTP | mail:1025（コンテナ内） |
| DB名 | `app_db` |

```bash
# 起動確認
make up
make ps

# DB接続（手順書内のSQLを実行する場合）
docker compose exec db psql -U postgres -d app_db
```

---

## 2. 前提条件

- Docker Compose が起動済みであること。
- `make migrate-up` でマイグレーションが完了していること。
- テスト実施前に Mailpit の受信箱をクリアすること（http://localhost:8025 → Delete All）。
- テスト実施前にデータをリセットする場合は以下を実行すること。

```bash
# DBリセット（必要な場合のみ）
docker compose exec db psql -U postgres -d app_db -c "
  DELETE FROM mail_outboxes;
  DELETE FROM user_registration_requests;
"
```

```bash
# Redisリセット（レートリミットを初期化する場合）
docker compose exec redis redis-cli FLUSHALL
```

---

## 3. 正常系

### TC-REQ-001 — 新規メールアドレスで仮登録する

**前提：** 対象メールアドレスが未登録であること。

#### 画面操作

1. ブラウザで http://localhost:5173/registration を開く。
2. 「メールアドレス」欄に `test01@example.com` を入力する。
3. 「確認用メールアドレス」欄に `test01@example.com` を入力する。
4. 「仮登録メールを送信する」ボタンをクリックする。

#### 期待結果（画面）

- http://localhost:5173/registration/complete へ遷移すること。
- 画面に `test01@example.com` が表示されること。
- 「有効期限は60分です」と表示されること（`expires_minutes=60` の場合）。
- 「迷惑メールフォルダをご確認ください」の注意書きが表示されること。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -H "Accept-Language: ja" \
  -d '{"email":"test01@example.com","email_confirmation":"test01@example.com"}'
```

**期待レスポンス（HTTP 201）：**

```json
{
  "code": "USER_REGISTRATION_REQUEST_CREATED",
  "message": "仮登録メールを送信しました。メールをご確認ください",
  "expires_minutes": 60
}
```

---

### TC-REQ-002 — 認証済みメールアドレスで仮登録する（存在隠蔽）

**前提：** 対象メールアドレスが本登録済みであること（`user_registration_requests.verified_at` が設定済みであること）。

本登録フロー（TC-VER-101相当）を `registered@example.com` で実施済みであればスキップ。  
未実施の場合は下記 SQL でセットアップする。

```bash
docker compose exec db psql -U postgres -d app_db -c "
INSERT INTO user_registration_requests (id, email, token_hash, expires_at, verified_at)
VALUES (
  gen_random_uuid(),
  'registered@example.com',
  md5('setup-dummy-token'),
  NOW() + INTERVAL '1 hour',
  NOW()
)
ON CONFLICT ((LOWER(email))) DO UPDATE SET verified_at = NOW();
"
```

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"registered@example.com","email_confirmation":"registered@example.com"}'
```

**期待レスポンス（HTTP 201）：**

```json
{
  "code": "USER_REGISTRATION_REQUEST_CREATED",
  "message": "仮登録メールを送信しました。メールをご確認ください",
  "expires_minutes": 60
}
```

**確認観点：**

- 登録済みであっても 201 が返ること（ステータスコードで存在を漏らさないこと）。
- メールが送信されないこと（Mailpit に届かないこと）。

---

### TC-REQ-003 — 大文字メールアドレスが正規化される

**前提：** `TEST01@EXAMPLE.COM` は `test01@example.com` と同一扱いになること。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"TEST01@EXAMPLE.COM","email_confirmation":"TEST01@EXAMPLE.COM"}'
```

**期待レスポンス（HTTP 201）：**

```json
{
  "code": "USER_REGISTRATION_REQUEST_CREATED"
}
```

**確認観点：**

- レスポンスが 201 であること。
- DB で `email` カラムが小文字で保存されていること。

```sql
SELECT email FROM user_registration_requests ORDER BY created_at DESC LIMIT 1;
```

---

## 4. 異常系 — バリデーションエラー

### TC-REQ-101 — メールアドレス未入力

#### 画面操作

1. ブラウザで http://localhost:5173/registration を開く。
2. 両欄を空のまま「仮登録メールを送信する」ボタンをクリックする。

#### 期待結果（画面）

- 「メールアドレスを入力してください」がフォーム直下に表示されること。
- 画面遷移しないこと。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"","email_confirmation":""}'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "email": [{"code": "EMAIL_REQUIRED", "message": "メールアドレスを入力してください"}]
  }
}
```

---

### TC-REQ-102 — メールアドレス形式が不正

#### 画面操作

1. 「メールアドレス」欄に `notanemail` を入力する。
2. 「確認用メールアドレス」欄に `notanemail` を入力する。
3. 「仮登録メールを送信する」ボタンをクリックする。

#### 期待結果（画面）

- 「正しいメールアドレス形式で入力してください」が表示されること。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"notanemail","email_confirmation":"notanemail"}'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "email": [{"code": "EMAIL_FORMAT_INVALID", "message": "正しいメールアドレス形式で入力してください"}]
  }
}
```

---

### TC-REQ-103 — 確認用メールアドレスが不一致

#### 画面操作

1. 「メールアドレス」欄に `test01@example.com` を入力する。
2. 「確認用メールアドレス」欄に `different@example.com` を入力する。
3. 「仮登録メールを送信する」ボタンをクリックする。

#### 期待結果（画面）

- 「メールアドレスが一致しません」が確認欄直下に表示されること。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"test01@example.com","email_confirmation":"different@example.com"}'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "email_confirmation": [{"code": "EMAIL_CONFIRMATION_NOT_MATCH", "message": "メールアドレスが一致しません"}]
  }
}
```

---

### TC-REQ-104 — 確認用メールアドレスが未入力

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"test01@example.com","email_confirmation":""}'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "email_confirmation": [{"code": "EMAIL_CONFIRMATION_REQUIRED", "message": "確認用メールアドレスを入力してください"}]
  }
}
```

---

### TC-REQ-105 — リクエストボディが不正（JSONパース失敗）

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d 'not json'
```

**期待レスポンス（HTTP 400）：**

```json
{
  "code": "BAD_REQUEST",
  "message": "リクエストが不正です"
}
```

---

## 5. 異常系 — 同一メール連続送信

### TC-REQ-201 — メールレートリミット内に同一メールで再送する

**前提：** `EMAIL_RATE_LIMIT_MINUTES=5`（デフォルト）。TC-REQ-001 を実施済みであること。

> **注記：** ハンドラ層のメールレートリミット（Redis）がサービス層の再送インターバルより先に評価される。
> 同一メールアドレスで5分以内に再送した場合は Redis キーで 429 が返る。
> サービス層の `last_sent_at` チェックは Redis が利用不可の場合のフォールバックである。

#### 手順

1. 以下のコマンドで1回目を送信する（TC-REQ-001 を実施済みの場合はスキップ）。

```bash
curl -s -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"test01@example.com","email_confirmation":"test01@example.com"}'
```

2. 5分以内に同じメールアドレスで再送する。

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -H "Accept-Language: ja" \
  -d '{"email":"test01@example.com","email_confirmation":"test01@example.com"}'
```

#### 期待結果

- HTTP 429 が返ること。
- Mailpit に **2通目が届かない**こと。

```json
{
  "code": "TOO_MANY_REQUESTS",
  "message": "リクエストが多すぎます。しばらく待ってから再試行してください"
}
```

---

### TC-REQ-202 — 再送インターバル経過後に再送する

**前提：** TC-REQ-001 を実施済みであること。

#### 手順

1. Redis のレートリミットキーをリセットする（メールレートリミットを解除する）。

```bash
docker compose exec redis redis-cli FLUSHALL
```

2. DBで `last_sent_at` を強制的に過去に更新する（再送インターバルを経過済みにする）。

```bash
docker compose exec db psql -U postgres -d app_db -c "
UPDATE user_registration_requests
SET last_sent_at = NOW() - INTERVAL '10 minutes'
WHERE email = 'test01@example.com';
"
```

3. 同じメールアドレスで再送する。

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"test01@example.com","email_confirmation":"test01@example.com"}'
```

#### 期待結果

- HTTP 201 が返ること。
- Mailpit に新しいメールが届くこと。
- DBで `token_hash` が更新されていること（新しいトークンが発行されること）。
- DBで `expires_at` が更新されていること。

```sql
SELECT token_hash, expires_at, last_sent_at
FROM user_registration_requests
WHERE email = 'test01@example.com';
```

---

## 6. 異常系 — レートリミット

### TC-REQ-301 — IPレートリミット（5回/分超過）

**前提：** `RATE_LIMIT_IP_PER_MINUTE=5`（デフォルト）。

#### 手順

1. Redisをリセットする。

```bash
docker compose exec redis redis-cli FLUSHALL
```

2. 同一IPから6回リクエストを送信する。

```bash
for i in $(seq 1 6); do
  echo -n "Request $i: "
  curl -s -w "HTTP: %{http_code}\n" \
    -X POST http://localhost:8080/api/v1/user-registration-requests \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"ratelimit${i}@example.com\",\"email_confirmation\":\"ratelimit${i}@example.com\"}" \
    -o /dev/null
done
```

#### 期待結果

- 1〜5回目：HTTP 201 が返ること。
- 6回目：HTTP 429 が返ること。

**期待レスポンス（HTTP 429）：**

```json
{
  "code": "TOO_MANY_REQUESTS",
  "message": "リクエストが多すぎます。しばらく待ってから再試行してください"
}
```

---

### TC-REQ-302 — メールアドレス単位レートリミット（5分に1回）

**前提：** `RATE_LIMIT_EMAIL_PER_5MIN=1`（デフォルト）。

#### 手順

1. Redisをリセットする。

```bash
docker compose exec redis redis-cli FLUSHALL
```

2. 同一メールアドレスで1分以内に2回送信する。

```bash
# 1回目
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"ratelimit@example.com","email_confirmation":"ratelimit@example.com"}'

# 2回目（すぐに送信）
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"ratelimit@example.com","email_confirmation":"ratelimit@example.com"}'
```

#### 期待結果

- 1回目：HTTP 201 が返ること。
- 2回目：HTTP 429 が返ること。

---

## 7. Mailpit 確認

### TC-REQ-401 — メール受信確認

#### 手順

1. TC-REQ-001 を実施する。
2. ブラウザで http://localhost:8025 を開く。
3. 受信箱に `test01@example.com` 宛のメールが届いていることを確認する。

#### 確認観点

| 観点 | 確認内容 |
|------|----------|
| 宛先 | `test01@example.com` であること |
| 件名 | 本登録用の件名が含まれること |
| 本文 | 本登録URLが含まれること |
| URL形式 | `http://localhost:5173/registration/verify?token=` で始まること |
| token長 | URLのtokenが40文字以上の英数字であること |
| token内容 | URLクエリパラメータに `token=` が含まれること |

#### tokenの取得

Mailpit UI でメール本文をクリックし、URLを確認する。またはAPIで取得する。

```bash
# Mailpit API でメッセージ一覧取得
curl -s http://localhost:8025/api/v1/messages | python3 -m json.tool | head -50

# メッセージIDを確認してメール本文を取得
MSG_ID=$(curl -s http://localhost:8025/api/v1/messages | python3 -c "import sys,json; print(json.load(sys.stdin)['messages'][0]['ID'])")
curl -s "http://localhost:8025/api/v1/message/${MSG_ID}" | python3 -c "
import sys, json
msg = json.load(sys.stdin)
print(msg.get('Text', msg.get('HTML', '')))
"
```

---

### TC-REQ-402 — tokenがメール本文のURLにのみ存在すること

#### 確認観点

- メール件名に `token=` が含まれないこと。
- メールヘッダに `token=` が含まれないこと。
- URLのクエリパラメータ以外に平文tokenが露出しないこと。

---

### TC-REQ-403 — mail_outboxes.payloadが送信後に空になること

#### 手順

1. TC-REQ-001 でメールを送信する。
2. workerがメールを送信するまで待つ（約1〜3秒）。
3. DBで `payload` を確認する。

```sql
SELECT id, status, payload
FROM mail_outboxes
ORDER BY created_at DESC
LIMIT 1;
```

#### 期待結果

- `status` が `sent` であること。
- `payload` が `{}` であること（平文URLが削除されていること）。

---

## 8. APIログ確認

### TC-REQ-501 — ログにtokenが出力されないこと

#### 手順

1. 別ターミナルでAPIログをリアルタイム監視する。

```bash
docker compose logs -f api
```

2. TC-REQ-001 を実施する。

#### 確認観点

| 観点 | 確認内容 |
|------|----------|
| token漏洩 | ログに `token=` が含まれないこと |
| メールアドレス | ログにメールアドレスの平文が含まれないこと |
| パスワード | このフェーズにパスワードは存在しないが、パスワード関連文字列がないこと |
| 「You trusted all proxies」 | この警告が出ていないこと |

**「You trusted all proxies」が出ていないことの確認：**

```bash
docker compose logs api 2>&1 | grep -i "trusted all proxies" | wc -l
# 0 であること
```

---

### TC-REQ-502 — ログフォーマットの確認

#### 確認観点

正常リクエスト時のログに以下が含まれること。

- `method=POST`
- `path=/api/v1/user-registration-requests`
- `code=USER_REGISTRATION_REQUEST_CREATED`

エラー時のログに以下が含まれること（5xxのみ）。

- `cause=` （内部エラー詳細）

---

## 9. 確認観点まとめ

| カテゴリ | 確認項目 | TC番号 |
|----------|----------|--------|
| 正常系 | 201レスポンスと正しいJSONが返ること | TC-REQ-001 |
| 正常系 | 認証済みメールで存在を隠蔽すること | TC-REQ-002 |
| 正常系 | メールアドレスが小文字に正規化されること | TC-REQ-003 |
| バリデーション | 必須エラーが適切に返ること | TC-REQ-101 |
| バリデーション | 形式エラーが適切に返ること | TC-REQ-102 |
| バリデーション | 不一致エラーが適切に返ること | TC-REQ-103 |
| バリデーション | 400 Bad Requestが正しく返ること | TC-REQ-105 |
| 業務ルール | 再送インターバル内でメールが重複送信されないこと | TC-REQ-201 |
| 業務ルール | インターバル経過後にトークンが更新されること | TC-REQ-202 |
| セキュリティ | IPレートリミットで429が返ること | TC-REQ-301 |
| セキュリティ | メール単位レートリミットで429が返ること | TC-REQ-302 |
| メール | Mailpitで本登録URLが確認できること | TC-REQ-401 |
| セキュリティ | tokenがURL以外に漏洩しないこと | TC-REQ-402 |
| セキュリティ | payloadが送信後に空になること | TC-REQ-403 |
| ログ | tokenがログに出力されないこと | TC-REQ-501 |
| ログ | 「You trusted all proxies」が出ていないこと | TC-REQ-501 |
| ログ | ログフォーマットが正しいこと | TC-REQ-502 |

---

> **Playwright E2E 化メモ：**
> 各TCの「画面操作」ブロックが `page.fill()` / `page.click()` / `page.waitForURL()` に対応する。
> 「期待結果（画面）」ブロックが `expect(page.locator(...)).toBeVisible()` に対応する。
> curl確認ブロックは E2E には含めず、API単体テストとして分離する。
