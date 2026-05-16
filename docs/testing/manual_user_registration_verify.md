# 手動テスト手順書 — 本会員登録

## 目次

1. [テスト環境](#1-テスト環境)
2. [前提条件](#2-前提条件)
3. [事前準備 — tokenの取得](#3-事前準備--tokenの取得)
4. [正常系 — GETトークン確認API](#4-正常系--getトークン確認api)
5. [正常系 — 本登録完了](#5-正常系--本登録完了)
6. [異常系 — tokenなし・token不正](#6-異常系--tokenなしtoken不正)
7. [異常系 — token期限切れ](#7-異常系--token期限切れ)
8. [異常系 — token使用済み](#8-異常系--token使用済み)
9. [異常系 — メール重複（USER_ALREADY_REGISTERED）](#9-異常系--メール重複user_already_registered)
10. [異常系 — バリデーションエラー](#10-異常系--バリデーションエラー)
11. [異常系 — レートリミット](#11-異常系--レートリミット)
12. [セキュリティ確認 — tokenマスク](#12-セキュリティ確認--tokenマスク)
13. [APIログ確認](#13-apiログ確認)
14. [確認観点まとめ](#14-確認観点まとめ)

---

## 1. テスト環境

| 項目 | 値 |
|------|---|
| フロントエンド | http://localhost:5173 |
| API | http://localhost:8080 |
| Mailpit UI | http://localhost:8025 |
| GETエンドポイント | `GET /api/v1/user-registrations/verify?token={token}` |
| POSTエンドポイント | `POST /api/v1/user-registrations/verify?token={token}` |
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
- 仮登録テストが完了し、Mailpit に本登録URLが届いていること。
- 本手順書は `manual_user_registration_request.md` のTC-REQ-001実施後を前提とする。

---

## 3. 事前準備 — tokenの取得

本会員登録テストには有効なtokenが必要である。以下の手順でtokenを取得する。

### 手順

1. 仮登録を実施する。

```bash
curl -s -X POST http://localhost:8080/api/v1/user-registration-requests \
  -H "Content-Type: application/json" \
  -d '{"email":"verify01@example.com","email_confirmation":"verify01@example.com"}'
```

2. Mailpit APIでメッセージを取得する。

```bash
# メッセージ一覧を取得する
curl -s http://localhost:8025/api/v1/messages | python3 -m json.tool

# 最新メッセージIDを取得する
MSG_ID=$(curl -s http://localhost:8025/api/v1/messages \
  | python3 -c "import sys,json; msgs=json.load(sys.stdin)['messages']; print(msgs[0]['ID'])")
echo "Message ID: ${MSG_ID}"
```

3. メール本文からtokenを抽出する。

```bash
# メール本文を取得してtokenを表示する
TOKEN=$(curl -s "http://localhost:8025/api/v1/message/${MSG_ID}" \
  | python3 -c "
import sys, json, re
msg = json.load(sys.stdin)
text = msg.get('Text', '')
match = re.search(r'token=([A-Za-z0-9_-]+)', text)
print(match.group(1) if match else 'NOT FOUND')
")
echo "Token: ${TOKEN}"
```

4. 取得したtokenを環境変数に設定する。

```bash
export VERIFY_TOKEN="${TOKEN}"
echo "VERIFY_TOKEN=${VERIFY_TOKEN}"
```

---

## 4. 正常系 — GETトークン確認API

### TC-VER-001 — 有効tokenでGETが成功すること

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}"
```

**期待レスポンス（HTTP 200）：**

```json
{
  "code": "REGISTRATION_TOKEN_VALID",
  "message": "本登録トークンは有効です"
}
```

---

### TC-VER-002 — GETが副作用を持たないこと（冪等性）

#### 手順

1. 同じtokenで3回GETを実行する。

```bash
for i in 1 2 3; do
  echo -n "GET $i: "
  curl -s -w "HTTP: %{http_code}\n" \
    "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
    -o /dev/null
done
```

2. DB状態を確認する（更新されていないこと）。

```sql
SELECT token_hash, verified_at, updated_at
FROM user_registration_requests
WHERE email = 'verify01@example.com';
```

#### 期待結果

- 3回とも HTTP 200 が返ること。
- `verified_at` が NULL のままであること（GETで本登録は完了しないこと）。
- `updated_at` が変化しないこと。

---

### TC-VER-003 — 画面アクセス時にGETが呼ばれること

#### 画面操作

1. ブラウザで以下のURLを開く。

```
http://localhost:5173/registration/verify?token={VERIFY_TOKEN}
```

（`{VERIFY_TOKEN}` は手順3で取得した実際のtokenで置き換える。）

2. 「確認中...」が一瞬表示された後、本会員登録フォームが表示されることを確認する。

#### 期待結果（画面）

- 「確認中...」が表示されること（短時間）。
- フォームが表示されること。
  - 「表示名」入力欄。
  - 「パスワード」入力欄。
  - 「パスワード（確認）」入力欄。
  - 「利用規約に同意する」チェックボックス。
  - 「本登録を完了する」ボタン。
- URLバーから `?token=...` が消えること（URLクリーニング）。

---

## 5. 正常系 — 本登録完了

### TC-VER-101 — 正常な入力で本登録が完了すること

#### 画面操作

1. ブラウザで以下のURLを開く（TC-VER-003の続き）。

```
http://localhost:5173/registration/verify?token={VERIFY_TOKEN}
```

2. フォームが表示されるまで待つ。
3. 「表示名」欄に `テストユーザー` を入力する。
4. 「パスワード」欄に `Password123` を入力する。
5. 「パスワード（確認）」欄に `Password123` を入力する。
6. 「利用規約に同意する」チェックボックスをチェックする。
7. 「本登録を完了する」ボタンをクリックする。

#### 期待結果（画面）

- http://localhost:5173/registration/success へ遷移すること。
- 完了メッセージが表示されること。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: ja" \
  -d '{
    "display_name": "テストユーザー",
    "password": "Password123",
    "password_confirmation": "Password123",
    "agreed_to_terms": true
  }'
```

**期待レスポンス（HTTP 201）：**

```json
{
  "code": "USER_REGISTRATION_VERIFIED",
  "message": "本登録が完了しました"
}
```

#### DB確認

```sql
-- users テーブル
SELECT id, display_name, status FROM users ORDER BY created_at DESC LIMIT 1;

-- user_emails テーブル
SELECT user_id, email, is_primary, verified_at FROM user_emails ORDER BY created_at DESC LIMIT 1;

-- user_credentials テーブル（password_hashが保存されていること）
SELECT user_id, LEFT(password_hash, 10) || '...' AS password_hash_preview
FROM user_credentials ORDER BY created_at DESC LIMIT 1;

-- user_registration_requests テーブル（verified_atが設定されていること）
SELECT email, verified_at FROM user_registration_requests
WHERE email = 'verify01@example.com';
```

**期待DB状態：**

| テーブル | 確認項目 |
|----------|----------|
| users | `display_name='テストユーザー'`、`status='active'` |
| user_emails | `is_primary=true`、`verified_at IS NOT NULL` |
| user_credentials | `password_hash` が `$2a$` で始まること（bcrypt） |
| user_registration_requests | `verified_at IS NOT NULL` |

---

### TC-VER-102 — 送信中はボタンが無効化されること

#### 画面操作

1. 本登録フォームに正しい値を入力する。
2. 「本登録を完了する」ボタンをクリックする直後にボタンを観察する。

#### 期待結果（画面）

- クリック後にボタンが「登録中...」に変わること。
- ボタンが無効化（グレーアウト）されること。
- 二重送信できないこと。

---

## 6. 異常系 — tokenなし・token不正

### TC-VER-201 — tokenパラメータなしでアクセスする

#### 画面操作

1. ブラウザで http://localhost:5173/registration/verify を開く（tokenなし）。

#### 期待結果（画面）

- 「本登録リンクが無効です」が表示されること。
- フォームが表示されないこと。

#### API確認

```bash
# GET（tokenなし）
curl -s -w "\nHTTP: %{http_code}\n" \
  "http://localhost:8080/api/v1/user-registrations/verify"

# POST（tokenなし）
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"test","password":"Password123","password_confirmation":"Password123","agreed_to_terms":true}'
```

**期待レスポンス（HTTP 400）：**

```json
{
  "code": "INVALID_REGISTRATION_TOKEN",
  "message": "トークンが不正です"
}
```

---

### TC-VER-202 — 存在しないtokenでアクセスする

#### 画面操作

1. ブラウザで以下のURLを開く。

```
http://localhost:5173/registration/verify?token=invalidtoken123456789012345678901234
```

#### 期待結果（画面）

- 「確認中...」が短時間表示された後、エラーメッセージが表示されること。
- フォームが表示されないこと。

#### API確認

```bash
# GET（不正token）
curl -s -w "\nHTTP: %{http_code}\n" \
  "http://localhost:8080/api/v1/user-registrations/verify?token=invalidtoken123456789012345678901234"
```

**期待レスポンス（HTTP 400）：**

```json
{
  "code": "INVALID_REGISTRATION_TOKEN",
  "message": "トークンが不正です"
}
```

**確認観点：**

- DB に存在しないトークンと、ハッシュ不一致のトークンで同一のエラーコードが返ること（存在の区別を行わないこと）。

---

## 7. 異常系 — token期限切れ

### TC-VER-301 — 期限切れtokenでアクセスする

#### 手順

1. 有効なtokenを取得する（手順3参照）。
2. DBで `expires_at` を過去に設定する。

```sql
UPDATE user_registration_requests
SET expires_at = NOW() - INTERVAL '1 hour'
WHERE email = 'verify01@example.com';
```

3. 期限切れtokenでアクセスする。

```bash
# GET（期限切れtoken）
curl -s -w "\nHTTP: %{http_code}\n" \
  "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}"
```

#### 期待結果（画面）

- エラーメッセージ「トークンの有効期限が切れています」が表示されること。
- フォームが表示されないこと。

**期待レスポンス（HTTP 400）：**

```json
{
  "code": "EXPIRED_REGISTRATION_TOKEN",
  "message": "トークンの有効期限が切れています"
}
```

---

## 8. 異常系 — token使用済み

### TC-VER-401 — 使用済みtokenでGETアクセスする

**前提：** TC-VER-101 で本登録を完了していること（`verified_at IS NOT NULL`）。

#### 画面操作

1. ブラウザで使用済みtokenのURLを開く。

```
http://localhost:5173/registration/verify?token={使用済みのVERIFY_TOKEN}
```

#### 期待結果（画面）

- 「会員登録済み」の見出しが表示されること。
- 「このメールアドレスは既に登録されています」が表示されること。
- 「ログインページへ」リンクが表示されること。
- フォームが表示されないこと。

#### API確認

```bash
# GET（使用済みtoken）
curl -s -w "\nHTTP: %{http_code}\n" \
  "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}"
```

**期待レスポンス（HTTP 409）：**

```json
{
  "code": "USED_REGISTRATION_TOKEN",
  "message": "既に登録が完了しています"
}
```

---

### TC-VER-402 — 使用済みtokenでPOSTアクセスする

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "テストユーザー",
    "password": "Password123",
    "password_confirmation": "Password123",
    "agreed_to_terms": true
  }'
```

**期待レスポンス（HTTP 409）：**

```json
{
  "code": "USED_REGISTRATION_TOKEN",
  "message": "既に登録が完了しています"
}
```

**確認観点：**

- フロントエンドで「会員登録済み」画面が表示されること（フォームからSubmitした場合）。

---

## 9. 異常系 — メール重複（USER_ALREADY_REGISTERED）

### TC-VER-501 — user_emailsに同一メールが存在する場合

**前提：** 本登録が完了したメールアドレスと同一のメールで別途仮登録トークンを取得していること。

#### 手順

1. DB上で別の `user_registration_requests` レコードを直接作成し、有効なtokenを設定する。

```sql
-- 既存の user_emails に登録済みメールの別仮登録レコードを確認する
SELECT urr.email, ue.email AS registered_email
FROM user_registration_requests urr
LEFT JOIN user_emails ue ON LOWER(urr.email) = LOWER(ue.email)
WHERE ue.email IS NOT NULL;
```

2. 有効なtokenでPOSTを送信する（DB上で verified_at が NULL のレコード）。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "テストユーザー2",
    "password": "Password123",
    "password_confirmation": "Password123",
    "agreed_to_terms": true
  }'
```

**期待レスポンス（HTTP 409）：**

```json
{
  "code": "USER_ALREADY_REGISTERED",
  "message": "入力されたメールアドレスは既に登録されています"
}
```

**確認観点（画面）：**

- フロントエンドで「会員登録済み」画面が表示されること。
- 「ログインページへ」リンクが表示されること。

---

## 10. 異常系 — バリデーションエラー

以下のテストはすべて有効なtokenを使用する。事前に手順3でtokenを取得すること。

### TC-VER-601 — 表示名が未入力

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "",
    "password": "Password123",
    "password_confirmation": "Password123",
    "agreed_to_terms": true
  }'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "display_name": [{"code": "DISPLAY_NAME_REQUIRED", "message": "ユーザ名を入力してください"}]
  }
}
```

#### 期待結果（画面）

- 「ユーザ名を入力してください」が「表示名」欄の下に表示されること。
- 入力欄の枠が赤くなること。

---

### TC-VER-602 — パスワードが未入力

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "テストユーザー",
    "password": "",
    "password_confirmation": "",
    "agreed_to_terms": true
  }'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "password": [{"code": "PASSWORD_REQUIRED", "message": "パスワードを入力してください"}]
  }
}
```

---

### TC-VER-603 — パスワード強度不足（境界系）

#### 手順

以下のパターンをそれぞれ確認する。

**7文字（8文字未満）：**

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"テスト","password":"Pass123","password_confirmation":"Pass123","agreed_to_terms":true}'
```

**英字のみ（数字なし）：**

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"テスト","password":"PasswordOnly","password_confirmation":"PasswordOnly","agreed_to_terms":true}'
```

**数字のみ（英字なし）：**

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"テスト","password":"12345678","password_confirmation":"12345678","agreed_to_terms":true}'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "password": [{"code": "PASSWORD_TOO_WEAK", "message": "パスワードは8文字以上で、英字と数字をそれぞれ1文字以上含めてください"}]
  }
}
```

**境界値（8文字・英字1文字・数字1文字を含む — 成功）：**

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"テスト","password":"Passw0rd","password_confirmation":"Passw0rd","agreed_to_terms":true}'
```

- HTTP 201 が返ること（境界値を満たすこと）。

---

### TC-VER-604 — パスワード不一致

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "テストユーザー",
    "password": "Password123",
    "password_confirmation": "Different456",
    "agreed_to_terms": true
  }'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "password_confirmation": [{"code": "PASSWORD_CONFIRMATION_NOT_MATCH", "message": "パスワードが一致しません"}]
  }
}
```

#### 期待結果（画面）

- 「パスワードが一致しません」が「パスワード（確認）」欄の下に表示されること。

---

### TC-VER-605 — 利用規約に同意しない

#### 画面操作

1. フォームに有効な入力を行う。
2. 「利用規約に同意する」チェックボックスをチェックしない。
3. 「本登録を完了する」ボタンをクリックする。

#### API確認

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "テストユーザー",
    "password": "Password123",
    "password_confirmation": "Password123",
    "agreed_to_terms": false
  }'
```

**期待レスポンス（HTTP 422）：**

```json
{
  "code": "VALIDATION_ERROR",
  "errors": {
    "agreed_to_terms": [{"code": "AGREED_TO_TERMS_REQUIRED", "message": "利用規約への同意が必要です"}]
  }
}
```

#### 期待結果（画面）

- 「利用規約への同意が必要です」がチェックボックス下に表示されること。

---

## 11. 異常系 — レートリミット

### TC-VER-701 — GETエンドポイントのレートリミット（5回/分超過）

**前提：** `RATE_LIMIT_IP_PER_MINUTE=5`（POSTとGETで共有カウンタ）。

#### 手順

1. Redisをリセットする。

```bash
docker compose exec redis redis-cli FLUSHALL
```

2. 同一IPから6回GETを送信する。

```bash
for i in $(seq 1 6); do
  echo -n "GET $i: "
  curl -s -w "HTTP: %{http_code}\n" \
    "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
    -o /dev/null
done
```

#### 期待結果

- 1〜5回目：HTTP 200 が返ること。
- 6回目：HTTP 429 が返ること。

**期待レスポンス（HTTP 429）：**

```json
{
  "code": "TOO_MANY_REQUESTS",
  "message": "リクエストが多すぎます。しばらく待ってから再試行してください"
}
```

---

### TC-VER-702 — GET/POSTでレートリミットカウンタが共有されること

**前提：** GETとPOSTのレートリミットは同一Redisキー `rate_limit:ip:{ip}` を使用する。

#### 手順

1. Redisをリセットする。

```bash
docker compose exec redis redis-cli FLUSHALL
```

2. GETを3回送信する。

```bash
for i in 1 2 3; do
  curl -s -o /dev/null \
    "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}"
done
```

3. POSTを3回送信する（合計6回）。

```bash
for i in 1 2 3; do
  echo -n "POST $i: "
  curl -s -w "HTTP: %{http_code}\n" \
    -X POST "http://localhost:8080/api/v1/user-registrations/verify?token=${VERIFY_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"display_name":"test","password":"Password123","password_confirmation":"Password123","agreed_to_terms":true}' \
    -o /dev/null
done
```

#### 期待結果

- POST 1・2回目（合計5・6回目）のうち、6回目が429になること。
- GETとPOSTでカウンタが共有されていること。

---

## 12. セキュリティ確認 — tokenマスク

### TC-VER-801 — URLからtokenが除去されること

#### 画面操作

1. ブラウザで本登録URLを開く。

```
http://localhost:5173/registration/verify?token={VERIFY_TOKEN}
```

2. フォームが表示された後、ブラウザのアドレスバーを確認する。

#### 期待結果

- アドレスバーのURLが `http://localhost:5173/registration/verify` になること（`?token=...` が消えること）。
- ブラウザの履歴にtokenが残らないこと（`history.replaceState` によるURLクリーニング）。

---

### TC-VER-802 — tokenがブラウザ画面上に表示されないこと

#### 画面操作

1. 本登録URLを開く。
2. フォームが表示された状態でページのソースを確認する（右クリック → ページのソースを表示）。

#### 期待結果

- ページソースに `token=` が含まれないこと。
- tokenの値がHTMLに埋め込まれていないこと。

---

### TC-VER-803 — ログにtokenが出力されないこと

#### 手順

1. 別ターミナルでAPIログをリアルタイム監視する。

```bash
docker compose logs -f api
```

2. TC-VER-001（GET）とTC-VER-101（POST）を実施する。

#### 確認観点

| 観点 | 確認内容 |
|------|----------|
| token漏洩 | ログに平文tokenが含まれないこと |
| password漏洩 | ログにpasswordが含まれないこと |
| password_hash漏洩 | ログにbcryptハッシュが含まれないこと |
| 「You trusted all proxies」 | この警告が出ていないこと |

**「You trusted all proxies」が出ていないことの確認：**

```bash
docker compose logs api 2>&1 | grep -i "trusted all proxies" | wc -l
# 0 であること
```

**tokenがログに出力されていないことの確認：**

```bash
# VERIFY_TOKENの値をログ内で検索する（存在しないこと）
docker compose logs api 2>&1 | grep -c "${VERIFY_TOKEN}"
# 0 であること
```

---

### TC-VER-804 — エラーレスポンスに内部情報が含まれないこと

#### 確認観点

- 400/422/429 エラーレスポンスに `cause` フィールドが含まれないこと。
- 500 エラー時もレスポンスに `cause` は含まれないこと（ログにのみ出力される）。
- tokenが存在しない場合と、ハッシュ不一致の場合で同一エラーコード（`INVALID_REGISTRATION_TOKEN`）が返ること。

---

## 13. APIログ確認

### TC-VER-901 — 正常完了時のログフォーマット

#### 確認観点

正常リクエスト（GET/POST）のログに以下が含まれること。

**GETログ：**

```
method=GET path=/api/v1/user-registrations/verify code=REGISTRATION_TOKEN_VALID
```

**POSTログ：**

```
method=POST path=/api/v1/user-registrations/verify code=USER_REGISTRATION_VERIFIED
```

---

### TC-VER-902 — エラー時のログフォーマット

#### 確認観点

クライアントエラー（4xx）のログに以下が含まれること。

- `method=` / `path=` / `code=` が含まれること。
- `cause=` が含まれないこと（4xxは内部原因を出力しない）。

サーバエラー（5xx）のログに以下が含まれること。

- `method=` / `path=` / `code=` / `cause=` が含まれること。

---

### TC-VER-903 — レートリミット超過時のログ

#### 手順

1. TC-VER-701 を実施する（6回目で429発生）。
2. APIログを確認する。

```bash
docker compose logs api 2>&1 | grep "TOO_MANY_REQUESTS"
```

#### 確認観点

- `code=TOO_MANY_REQUESTS` がログに出力されること。
- tokenがログに含まれないこと。

---

## 14. 確認観点まとめ

| カテゴリ | 確認項目 | TC番号 |
|----------|----------|--------|
| GET正常系 | 200とREGISTRATION_TOKEN_VALIDが返ること | TC-VER-001 |
| GET正常系 | 複数回GETしても副作用がないこと | TC-VER-002 |
| GET正常系 | 画面アクセス時にGETが呼ばれフォームが表示されること | TC-VER-003 |
| POST正常系 | 201とUSER_REGISTRATION_VERIFIEDが返ること | TC-VER-101 |
| POST正常系 | 送信中にボタンが無効化されること | TC-VER-102 |
| tokenなし | tokenなしで400が返ること | TC-VER-201 |
| token不正 | 存在しないtokenで400が返ること | TC-VER-202 |
| token期限切れ | 期限切れtokenで400が返ること | TC-VER-301 |
| token使用済み | 使用済みtokenで409（USED）が返ること | TC-VER-401 |
| token使用済み | 使用済みtokenでPOSTしても409が返ること | TC-VER-402 |
| メール重複 | USER_ALREADY_REGISTEREDで409が返ること | TC-VER-501 |
| バリデーション | 表示名未入力で422が返ること | TC-VER-601 |
| バリデーション | パスワード未入力で422が返ること | TC-VER-602 |
| バリデーション | パスワード強度不足で422が返ること（境界含む） | TC-VER-603 |
| バリデーション | パスワード不一致で422が返ること | TC-VER-604 |
| バリデーション | 利用規約未同意で422が返ること | TC-VER-605 |
| レートリミット | GET/POSTで共有カウンタにより429が返ること | TC-VER-701 |
| レートリミット | GETとPOSTのカウンタが共有されること | TC-VER-702 |
| セキュリティ | URLからtokenが除去されること | TC-VER-801 |
| セキュリティ | tokenが画面に表示されないこと | TC-VER-802 |
| セキュリティ | tokenがログに出力されないこと | TC-VER-803 |
| セキュリティ | 「You trusted all proxies」が出ないこと | TC-VER-803 |
| セキュリティ | エラーレスポンスに内部情報が含まれないこと | TC-VER-804 |
| ログ | 正常時のログフォーマットが正しいこと | TC-VER-901 |
| ログ | エラー時のcauseが4xxで出力されないこと | TC-VER-902 |
| ログ | レートリミット超過時にtokenがログに出ないこと | TC-VER-903 |

---

> **Playwright E2E 化メモ：**
> 各TCの「画面操作」ブロックが `page.fill()` / `page.click()` / `page.waitForURL()` に対応する。
> TC-VER-003の「確認中...」表示確認は `page.waitForSelector('[text="確認中..."]')` に対応する。
> TC-VER-801のURLクリーニングは `expect(page.url()).not.toContain('token=')` に対応する。
> curl確認ブロックは E2E には含めず、API単体テストとして分離する。
> tokenの取得は `手順3. 事前準備` を Playwright の `beforeAll` フックで実装する。
