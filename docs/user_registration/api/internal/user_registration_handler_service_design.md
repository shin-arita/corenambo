# ユーザ仮登録 handler / service 設計

## 1. 概要

本設計は、ユーザ仮登録APIにおける handler / service の責務を定義する。

---

## 2. 全体方針

- handler：HTTP制御
- service：業務ロジック
- repository：DBアクセス
- mail：Outbox登録

---

## 3. handler

### 責務

- JSON bind
- Accept-Language取得
- rate limit
- service呼び出し
- レスポンス生成

---

### 注意

- Accept-Language はそのまま取得
- translator が fallback を担当

---

## 4. service

### 責務

- バリデーション
- 仮登録レコード取得
- トークン生成
- ハッシュ化
- DB保存
- Outbox登録

---

## 5. 業務ルール

### ■ メール存在隠蔽

- 既存ユーザでも成功を返す

---

### ■ 再送制御

- last_sent_at を使用
- 一定時間内は再送しない

---

### ■ トークン

- 毎回再生成
- ハッシュ保存

---

## 6. 処理フロー

1. バリデーション
2. 仮登録取得
3. 再送制御チェック
4. トークン生成
5. DB保存（tx）
6. Outbox登録
7. 201返却

---

## 7. メール送信

- service は送信しない
- mail_outboxes に登録する
- worker が送信

---

## 8. レート制限

- IP単位
- email単位

（handlerで実装済み）

---

## 9. 本登録API（Verify）設計

### 9.1 Handler の責務

```
POST /api/v1/user-registrations/verify
```

1. `Accept-Language` ヘッダを取得する（`ja` / `en` に対応。未指定または未対応言語は `ja` にフォールバック）
2. JSONバインド（フィールド最大長チェックを含む）
3. **IPアドレス単位のレートリミットチェック**（Redis、5回/分）
4. `service.Verify()` を呼び出す
5. 成功時 `201 Created` + `SuccessResponse{code, message}` を返す
6. エラー時 `app_error.Normalize()` で HTTP ステータスを決定しレスポンスを返す
   - 500以上はサーバログに `cause` を出力（token / password は含めない）
   - 4xx は `code` のみをログ出力

---

### 9.2 Service の責務（Verify）

#### 処理順序

```
1. token 空チェック（空なら 400 INVALID_REGISTRATION_TOKEN）
2. validateVerifyInput（display_name / password / password_confirmation / agreed_to_terms）
3. display_name のトリム
4. token の SHA-256 ハッシュ化（tokenHasher.Hash）
5. password の bcrypt ハッシュ化（DefaultCost）
6. txManager.WithinTransaction 開始
   a. FindByTokenHashForUpdate（FOR UPDATE 排他ロック）
   b. req == nil → 400 INVALID_REGISTRATION_TOKEN
   c. verified_at IS NOT NULL → 409 USED_REGISTRATION_TOKEN
   d. now.After(expires_at) → 400 EXPIRED_REGISTRATION_TOKEN
   e. userEmailRepo.FindByEmail（重複チェック）
   f. existing != nil → 409 USER_ALREADY_REGISTERED
   g. uuidGenerator.NewV7() → users INSERT
   h. uuidGenerator.NewV7() → user_emails INSERT（is_primary=true, verified_at=now）
   i. userCredentialRepo.Create → user_credentials INSERT
   j. userRegistrationRequestRepo.UpdateVerifiedAt
7. VerifyUserRegistrationOutput{Code: USER_REGISTRATION_VERIFIED} を返す
```

#### 設計上の判断

| 判断 | 理由 |
|-----|------|
| token 空チェックをバリデーションより前に行う | token はルーティングに相当する必須値のため、フィールドバリデーションと分離する |
| bcrypt ハッシュをトランザクション外で計算する | bcrypt は処理が重く、DB ロック保持時間を短くするためトランザクション前に実行する |
| FOR UPDATE で排他ロック取得 | 同一トークンへの並行リクエストによる二重登録を防ぐ |
| user_emails に LOWER インデックス + FindByEmail で二重防御 | アプリ層とDB層の両方でメール重複を防ぐ |

---

### 9.3 Repository の責務（Verify）

#### FindByTokenHashForUpdate

```sql
SELECT id, email, token_hash, expires_at, verified_at, last_sent_at, created_at
FROM user_registration_requests
WHERE token_hash = $1
LIMIT 1
FOR UPDATE
```

- `FOR UPDATE` によりトランザクション中の排他ロックを取得する
- ロック取得により、同一トークンへの並行リクエストは直列化される

#### UpdateVerifiedAt

```sql
UPDATE user_registration_requests SET verified_at = $2 WHERE id = $1
```

#### UserRepository.Create

```sql
INSERT INTO users (id, display_name, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
```

#### UserEmailRepository.Create

```sql
INSERT INTO user_emails (id, user_id, email, is_primary, verified_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
```

#### UserCredentialRepository.Create

```sql
INSERT INTO user_credentials (user_id, password_hash, password_changed_at, created_at)
VALUES ($1, $2, $3, $4)
```

---

### 9.4 トランザクション範囲

以下の4操作が同一トランザクション内に含まれる。1つでも失敗すれば全てロールバックする。

```
BEGIN
  ├─ SELECT ... FOR UPDATE  （user_registration_requests）
  ├─ INSERT INTO users
  ├─ INSERT INTO user_emails
  ├─ INSERT INTO user_credentials
  └─ UPDATE user_registration_requests SET verified_at = now
COMMIT
```

---

### 9.5 レートリミット

| 対象API | IP制限 | メール制限 |
|--------|--------|---------|
| POST /api/v1/user-registration-requests（仮登録） | 5回/分 | 1回/5分 |
| POST /api/v1/user-registrations/verify（本登録） | 5回/分 | なし |

- IP制限のカウンタは両エンドポイントで共有（同じ Redis キー `rate_limit:ip:{ip}`）
- 本登録APIはメールアドレスではなく、仮登録メールに含まれる32バイト暗号乱数トークンを前提とする。そのためメール単位のレートリミットではなく、IP単位のレートリミットで総当たり試行を抑制する
