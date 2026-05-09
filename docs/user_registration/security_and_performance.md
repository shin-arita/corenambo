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

## テスト方針

| 対象 | 手法 |
|---|---|
| SMTP 送信パス全体 | `sendMail` 関数変数をテスト用クロージャに差し替えてモック |
| TLS ダイヤル | パッケージ変数 `dialTLSFunc` を入れ替えてモック |
| Redis レートリミット | `KeyedStore` インターフェースを満たすモックで検証 |
| DB 操作 | `go-sqlmock` でクエリの引数・結果を検証 |
| サービスロジック | インターフェースを満たすダミーで全依存を差し替え |

全パッケージでカバレッジ 100% を維持。
