# セキュリティ方針

## 基本方針

- 情報漏洩を防ぐ
- 攻撃耐性を持つ設計にする

---

## 仮登録・認証

- 登録済みメールでも同一レスポンスを返す（列挙対策）
- IP単位のrate limit
- email単位のrate limit

---

## アクセスログ

- `gin.Logger()` は query string をそのまま出力するため、カスタム middleware でマスクする
- `token` クエリパラメータの値は `***` に置換してログ出力する（`middleware.LoggerWithTokenMask`）
- `token` 以外のクエリパラメータはそのまま出力する

---

## プロキシ・IP偽装対策

- `router.SetTrustedProxies(nil)` を設定し、外部からの `X-Forwarded-For` / `X-Real-IP` を信用しない
- `c.ClientIP()` は `RemoteAddr` から直接取得する
- 現構成（API直接公開、nginx等なし）ではプロキシを信頼しない設定が正しい

---

## query parameter token のセキュリティ

### バックエンド対応（実装済み）

- アクセスログで token をマスクする（`token=***`）

### フロント実装時の必須対応

- 本登録画面で token 取得後、URL から token を除去する（`history.replaceState` 等）
- 外部リンクには `rel="noreferrer"` を付与する
- `Referrer-Policy: no-referrer` を設定する

---

## 環境変数

- `FRONTEND_BASE_URL` は登録メールURLに使用される。本番環境では正しい HTTPS ドメインを設定すること
- `JWT_*` は認証機能実装後に必須。現時点ではAPIで未使用

---

## 禁止事項

- tokenのログ出力
- パスワードのログ出力
- 機密情報のレスポンス出力

---

## 必須要件

- tokenは使い捨て
- 有効期限を設定する
- 古いデータは削除可能にする
