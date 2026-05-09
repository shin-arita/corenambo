# 仮メンバー登録 — セキュリティ・性能設計

コレナンボ・オークションの仮メンバー登録機能に施したセキュリティ対策および性能上の考慮をまとめる。

---

## 機能概要

メールアドレスを受け取り、確認メールを送信する仮登録 API。

```
POST /api/v1/user-registration-requests
```

登録フロー:

1. クライアントがメールアドレスを送信
2. API がトークンを生成し、ハッシュを DB に保存
3. 送信ジョブを Outbox テーブルに積む（API レスポンスはここで返す）
4. バックグラウンドワーカーが Outbox を処理し、確認メールを送信

---

## セキュリティ対策

### 1. トークン生成に `crypto/rand` を使用

```go
// token/token.go
b := make([]byte, 32)
if _, err := randReader.Read(b); err != nil { ... }
return base64.RawURLEncoding.EncodeToString(b), nil
```

- OS の乱数源（`/dev/urandom`）を使用し、数学的乱数は使わない
- 32 バイト（256 bit）のエントロピーで総当たり攻撃が現実的に不可能

### 2. トークンは SHA-256 ハッシュのみ DB に保存

```go
// token/token.go
sum := sha256.Sum256([]byte(value))
return hex.EncodeToString(sum[:]), nil
```

- DB が漏洩しても平文トークンは復元できない
- メール送信後は Outbox の payload を `'{}'` に上書きし、URL の残留を防ぐ

### 3. IP・メールアドレス別のレートリミット（Redis）

```go
// handler/rate_limiter.go
func (r *rateLimiter) AllowIP(ctx, ip, limit) bool   // 1分あたり 5 回
func (r *rateLimiter) AllowEmail(ctx, email, limit) bool // 5分あたり 1 回
```

- **固定ウィンドウ方式**を Redis で実装。`INCR` + `ExpireNX` をパイプラインで実行
- `ExpireNX`（Redis 7.0+）を使用することで、ウィンドウ先頭の TTL をリセットしない（誤った実装では Expire のたびに TTL がリセットされ、永続的に送信できる状態になる）
- メールキーは SHA-256 ハッシュ化して Redis に保存し、メールアドレスの平文を残さない

### 4. メールヘッダインジェクション対策

```go
// mail/smtp_mailer.go
func sanitizeHeader(s string) string {
    return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}
```

- 宛先アドレスおよび件名から `\r`・`\n` を除去
- SMTP ヘッダを分断して偽のヘッダを挿入する攻撃を防ぐ

### 5. 本番 SMTP は SMTPS（ポート 465）で TLS 直接接続

```go
// mail/smtp_tls.go
func defaultSendWithTLS(addr string, auth smtp.Auth, ...) error {
    conn, err := dialTLSFunc("tcp", addr, tlsCfg)
    client, err := smtp.NewClient(conn, host)
    ...
}
```

- STARTTLS（平文で開始後にアップグレード）ではなく、接続時点から TLS を確立
- 設定ミスによる平文送信を構造的に防ぐ

### 6. メールアドレスの大文字小文字を正規化

```go
// service/user_registration_service.go
email := strings.ToLower(strings.TrimSpace(input.Email))
```

- サービス層で小文字に正規化してから DB 操作を行う
- DB のユニークインデックスも `LOWER(email)` で定義し、大文字小文字の違いによる重複登録を防ぐ

```sql
CREATE UNIQUE INDEX uq_user_registration_requests_email
    ON user_registration_requests (LOWER(email));
```

### 7. リクエストボディサイズ制限

```go
// cmd/server/main.go
c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
```

- 1 MB を超えるリクエストを拒否し、巨大 payload による DoS を防ぐ

### 8. フィールドの最大長バリデーション

```go
// handler/user_registration_handler.go
Email             string `json:"email" binding:"max=255"`
EmailConfirmation string `json:"email_confirmation" binding:"max=255"`
```

- DB の `VARCHAR(255)` 制約と対応し、バインディング時点で拒否

### 9. Accept-Language をホワイトリスト方式で正規化

```go
// handler/user_registration_handler.go
func normalizeLanguage(lang string) string {
    switch {
    case strings.HasPrefix(lang, "en"):
        return "en"
    default:
        return "ja"
    }
}
```

- 許可リスト以外の言語コードを `ja` にフォールバックし、不正な値を処理系に流さない

### 10. CORS は環境変数で指定したオリジンのみ許可

```go
// cmd/server/main.go
c.Header("Access-Control-Allow-Origin", allowOrigin) // ワイルドカードを使わない
```

- 起動時に `CORS_ALLOW_ORIGIN` が未設定の場合は `panic` して安全側に倒す

### 11. panic リカバリと構造化ログ

```go
router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
    logger.Error("method=%s path=%s panic=%v stack=%s", ...)
    c.JSON(http.StatusInternalServerError, ...)
}))
```

- ミドルウェアで全 panic を補足し、スタックトレースを構造化ログに出力
- クライアントには内部情報を含まない固定メッセージを返す

---

## 性能上の考慮

### 1. Outbox パターンによる非同期メール送信

メール送信を API のリクエストパスから完全に切り離す。

```
API  ──→  DB (mail_outboxes INSERT)  ──→  202 レスポンス返却
Worker  ──→  mail_outboxes FETCH  ──→  SMTP 送信
```

- API はメール送信の完了を待たずにレスポンスを返すため、SMTP レイテンシの影響を受けない
- ワーカーが落ちていてもデータは DB に残り、再起動後に送信される

### 2. UUIDv7 による時系列順 ID

```go
// uuid/uuid.go（google/uuid ライブラリ使用）
uuid.NewV7()
```

- タイムスタンププレフィックスにより INSERT 順序がインデックスと一致する
- B-tree インデックスの分割（ページスプリット）を最小化し、書き込み性能を維持

### 3. `FOR UPDATE SKIP LOCKED` によるワーカー競合回避

```sql
-- repository/mail_outbox_repository.go
UPDATE mail_outboxes
SET status = 'processing', updated_at = NOW()
WHERE id IN (
    SELECT id FROM mail_outboxes
    WHERE status = 'pending'
      AND next_attempt_at <= NOW()
    ORDER BY created_at
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING ...
```

- 複数ワーカーが同じレコードを取得しない
- `SKIP LOCKED` によりロック待ちが発生せず、スループットが落ちない

### 4. `UPDATE...RETURNING` によるアトミックなフェッチ

fetch と status 更新を 1 クエリで実行する。

- `SELECT → UPDATE` の 2 ステップでは TOCTOU（Time Of Check To Time Of Use）問題が起きうる
- 1 クエリにまとめることでレコードの二重処理を構造的に防ぐ

### 5. DB インデックス設計

| インデックス | 目的 |
|---|---|
| `UNIQUE (token_hash)` | トークン照合の高速化・重複防止 |
| `UNIQUE LOWER(email)` | 大文字小文字を問わない一意制約 |
| `(status, next_attempt_at)` | ワーカーの pending フェッチ高速化 |
| `(expires_at)` | 期限切れレコードの削除バッチ |
| `(verified_at)` | 認証済みユーザの絞り込み |
| `(created_at)` on mail_outboxes | 送信順序の保証 |

### 6. 再送間隔をサービス層で制御

```go
// service/user_registration_service.go
resendAvailableAt := req.LastSentAt.Add(
    time.Duration(s.config.RegistrationResendIntervalMinutes()) * time.Minute,
)
if now.Before(resendAvailableAt) {
    return &CreateUserRegistrationOutput{Code: ...}, nil
}
```

- Redis のレートリミットとは独立して、メール再送を DB 側でも制限
- メールが多重送信されても DB への余分な write が発生しない

---

---

## 本登録API（Verify）セキュリティ設計

### 12. パスワードの bcrypt ハッシュ化

```go
// service/user_registration_service.go
var bcryptGenerate = bcrypt.GenerateFromPassword

var hashPassword = func(password string) (string, error) {
    b, err := bcryptGenerate([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(b), nil
}
```

- `bcrypt.DefaultCost`（コスト係数 10）でハッシュ化
- `user_credentials.password_hash` に保存。平文パスワードは一切保存しない
- bcrypt はレインボーテーブル攻撃に耐性がある（ソルトが内部に含まれる）
- `bcryptGenerate` をパッケージ変数に分離することでテスト時にモック可能にしている

### 13. パスワード・トークンのログ出力禁止

```go
// handler/user_registration_handler.go（Verify）
logger.Error("method=%s path=%s code=%s cause=%v",
    c.Request.Method, c.FullPath(), appErr.Code, appErr.Cause)
logger.Warn("method=%s path=%s code=%s",
    c.Request.Method, c.FullPath(), appErr.Code)
logger.Info("method=%s path=%s code=%s",
    c.Request.Method, c.FullPath(), output.Code)
```

- ログに含まれるのは `method` / `path` / `code` / `cause`（内部エラーのみ）のみ
- `token`（平文）/ `password` / `password_hash` はログに出力しない
- `cause` はサーバ内部エラー（5xx）時のみ出力する

### 14. エラーレスポンスによる情報漏洩防止

本登録APIのエラーレスポンスはコードとメッセージのみを返す。以下の情報は漏洩しない。

| 漏洩させない情報 | 対処 |
|-----------|------|
| トークンが DB に存在するかどうか | 存在しない / ハッシュ不一致の両方に同じコード `INVALID_REGISTRATION_TOKEN` を返す |
| パスワードの内容 | エラーコードは `PASSWORD_REQUIRED` / `PASSWORD_CONFIRMATION_NOT_MATCH` のみ |
| bcrypt 処理の失敗 | `WrapInternal(err)` で 500 に変換し、原因は外部に出さない |

### 15. 本登録APIのレートリミット

```go
// handler/user_registration_handler.go（Verify）
if h.rateLimiter != nil {
    if !h.rateLimiter.AllowIP(c.Request.Context(), c.ClientIP(),
        h.rateLimitConfig.RateLimitIPPerMinute()) {
        c.JSON(http.StatusTooManyRequests, ...)
        return
    }
}
```

- IPアドレス単位で 5回/分（環境変数 `RATE_LIMIT_IP_PER_MINUTE`）
- 仮登録API（Create）と同じ Redis キー `rate_limit:ip:{ip}` を共有する
- 本登録APIはメールアドレスではなく、仮登録メールに含まれる32バイト暗号乱数トークンを前提とする。そのためメール単位のレートリミットではなく、IP単位のレートリミットで総当たり試行を抑制する

### 16. mail_outboxes.payload を送信後に空にする設計意図

```go
// repository/mail_outbox_repository.go（MarkSent）
SET payload = '{}'
```

worker がメール送信に成功した後、`payload` フィールドを `'{}'` で上書きする。

- `payload` には本登録 URL（トークン平文を含む）が格納されている
- 送信後も平文 URL が DB に残ると、DB 漏洩時に未使用トークンが流出するリスクがある
- 送信完了後は不要なため即座に削除する

### 17. E2EテストでMailpitからトークンを取得する理由

```bash
# scripts/e2e_user_registration.sh
get_token_from_mailpit() {
  ...
  curl -s "${MAILPIT_URL}/api/v1/message/${msg_id}" | \
    grep -o '"Text":"[^"]*"' | grep -o 'token=[^\\]*' | ...
}
```

`mail_outboxes.payload` ではなく Mailpit API を使用する理由：

- worker は仮登録 API 呼び出しから約 1 秒以内にメール送信を完了する
- 送信後、`mail_outboxes.payload` は `'{}'` に上書きされる（設計 §16 参照）
- そのため E2E テスト実行時点では `payload` からトークンを取得できない
- Mailpit の REST API（`GET /api/v1/messages`, `GET /api/v1/message/{id}`）を通じて送信済みメールのテキスト本文を取得し、URL中のトークンを抽出する

---

## テスト方針

| 対象 | 手法 |
|---|---|
| SMTP 送信パス全体 | `sendMail` 関数変数をテスト用クロージャに差し替えてモック |
| TLS ダイヤル | パッケージ変数 `dialTLSFunc` を入れ替えてモック |
| Redis レートリミット | `KeyedStore` インターフェースを満たすモックで検証 |
| DB 操作 | `go-sqlmock` でクエリの引数・結果を検証 |
| サービスロジック | インターフェースを満たすダミーで全依存を差し替え |
| bcrypt エラー | `bcryptGenerate` パッケージ変数をモックしてエラーパスを検証 |

全パッケージでカバレッジ 100% を維持。
