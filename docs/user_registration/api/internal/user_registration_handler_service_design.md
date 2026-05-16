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

- 認証済み・再送インターバル内・新規を問わず常に同一の201を返す
- トークン生成・DB更新・Outbox登録の有無に関わらず、メール存在を外部に漏らさない

---

### ■ 再送制御

- last_sent_at を使用してインターバルを判定
- インターバル内はトークン生成・DB更新・Outbox登録を行わない
- レスポンスは成功（201）として返す（隠蔽仕様）

---

### ■ トークン

- 送信必要と確定した場合のみ生成（認証済み・再送インターバル内は生成しない）
- SHA-256ハッシュのみDBに保存

---

## 6. 処理フロー

1. バリデーション
2. トランザクション開始
   1. FindByEmailForUpdate（FOR UPDATE）
   2. 認証済みの場合（verified_at IS NOT NULL）
      → トークン生成なし・DB更新なし・Outbox登録なしで即return（201として隠蔽）
   3. 再送インターバル内の場合（last_sent_at + interval > now）
      → トークン生成なし・DB更新なし・Outbox登録なしで即return（201として隠蔽）
   4. （送信必要と確定）トークン生成・ハッシュ化・URL生成
   5. DB保存（新規 INSERT または UpdateToken）
   6. Outbox登録
3. 201返却

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
2. クエリパラメータ `token` を取得する（空なら即 `400 INVALID_REGISTRATION_TOKEN`）
3. JSONバインド（display_name / password / password_confirmation / agreed_to_terms のみ）
4. **IPアドレス単位のレートリミットチェック**（Redis、5回/分）
5. `service.Verify()` を呼び出す
6. 成功時 `201 Created` + `SuccessResponse{code, message}` を返す
7. エラー時 `app_error.Normalize()` で HTTP ステータスを決定しレスポンスを返す
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
5. txManager.WithinTransaction 開始
   a. FindByTokenHashForUpdate（FOR UPDATE 排他ロック）
   b. req == nil → 400 INVALID_REGISTRATION_TOKEN
   c. verified_at IS NOT NULL → 409 USED_REGISTRATION_TOKEN
   d. now.After(expires_at) → 400 EXPIRED_REGISTRATION_TOKEN
   e. userEmailRepo.FindByEmail（重複チェック）
   f. existing != nil → 409 USER_ALREADY_REGISTERED
   g. password の bcrypt ハッシュ化（DefaultCost）— 不正トークン大量投入時に bcrypt が走る DoS を避けるため、全検証後に実行
   h. uuidGenerator.NewV7() → users INSERT
   i. uuidGenerator.NewV7() → user_emails INSERT（is_primary=true, verified_at=now）
   j. userCredentialRepo.Create → user_credentials INSERT
   k. userRegistrationRequestRepo.UpdateVerifiedAt
6. VerifyUserRegistrationOutput{Code: USER_REGISTRATION_VERIFIED} を返す
```

#### 設計上の判断

| 判断 | 理由 |
|-----|------|
| token 空チェックをバリデーションより前に行う | token はルーティングに相当する必須値のため、フィールドバリデーションと分離する |
| bcrypt ハッシュをトランザクション内・全検証後に実行する | 不正トークン大量投入時に bcrypt が走る DoS を防ぐ。トークン存在・使用済み・期限切れ・メール重複の確認後に実行することで、正当なリクエストのみ bcrypt コストを負担させる |
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
| GET /api/v1/user-registrations/verify（トークン確認） | 5回/分 | なし |

- 全エンドポイントの IP制限カウンタは共有（Redis キー `rate_limit:ip:{ip}`、環境変数 `RATE_LIMIT_IP_PER_MINUTE`）
- 本登録系APIはメールアドレスではなく、仮登録メールに含まれる32バイト暗号乱数トークンを前提とする。そのためメール単位のレートリミットではなく、IP単位のレートリミットで総当たり試行を抑制する

---

## 10. トークン状態確認API（GET Verify）設計

### 10.1 Handler の責務

```
GET /api/v1/user-registrations/verify
```

1. `Accept-Language` ヘッダを取得する（`ja` / `en` に対応。未指定または未対応言語は `ja` にフォールバック）
2. クエリパラメータ `token` を取得する（空なら即 `400 INVALID_REGISTRATION_TOKEN`）
3. **IPアドレス単位のレートリミットチェック**（Redis、5回/分、共有カウンタ `rate_limit:ip:{ip}`）
4. `service.CheckToken()` を呼び出す
5. 成功時 `200 OK` + `SuccessResponse{code, message}` を返す
6. エラー時 `app_error.Normalize()` で HTTP ステータスを決定しレスポンスを返す

---

### 10.2 Service の責務（CheckToken）

副作用なし。トランザクション・FOR UPDATE・bcrypt処理を一切行わない。

#### 処理順序

```
1. token 空チェック（空なら 400 INVALID_REGISTRATION_TOKEN）
2. token の SHA-256 ハッシュ化（tokenHasher.Hash）
3. FindByTokenHash（読み取り専用、FOR UPDATEなし）
4. req == nil → 400 INVALID_REGISTRATION_TOKEN
5. verified_at IS NOT NULL → 409 USED_REGISTRATION_TOKEN
6. now.After(expires_at) → 400 EXPIRED_REGISTRATION_TOKEN
7. userEmailRepo.FindByEmail（重複チェック）
8. existing != nil → 409 USER_ALREADY_REGISTERED
9. CheckTokenOutput{Code: REGISTRATION_TOKEN_VALID} を返す
```

#### 設計上の判断

| 判断 | 理由 |
|-----|------|
| トランザクションなし | 書き込みがないため不要 |
| FOR UPDATEなし | 読み取り専用なのでロック不要。並行リクエストへの影響もない |
| bcryptなし | パスワード処理は本登録（POST）でのみ行う |
| 共有IP rate limit | POST と同一のカウンタ・制限値（5回/分）を使用する。専用カウンタを設けないことで設定の複雑化を防ぐ |

---

### 10.3 Repository の責務（CheckToken）

#### FindByTokenHash

```sql
SELECT id, email, token_hash, expires_at, verified_at, last_sent_at, created_at
FROM user_registration_requests
WHERE token_hash = $1
LIMIT 1
```

- `FOR UPDATE` なし
- 読み取り専用のため、同時実行しても他のリクエストに影響しない
