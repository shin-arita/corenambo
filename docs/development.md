# 開発環境セットアップ（コレナンボ・オークション）

## ■ 概要

本プロジェクトは Docker Compose によるマルチコンテナ構成で動作します。

- Frontend: React (Vite)
- API: Go (Gin + Air)
- DB: PostgreSQL + PGroonga
- Mail: Mailpit
- Cache: Redis
- Worker: 非同期メール送信

---

## ■ 前提

以下がインストールされていること：

- Docker
- Docker Compose
- Make

---

## ■ 初回セットアップ

### 1. 環境変数設定

```bash
cp .env.example .env
```

### 2. JWTシークレット生成（必須）

```bash
openssl rand -base64 32
```

`.env` に設定：

```env
JWT_ACCESS_SECRET=生成した値
JWT_REFRESH_SECRET=生成した値
```

---

### 3. コンテナ起動

```bash
make up
```

---

### 4. DBマイグレーション

```bash
make migrate-up
```

---

## ■ アクセス先

| サービス | URL |
|----------|-----|
| Frontend | http://localhost:5173 |
| API | http://localhost:8080 |
| Mailpit | http://localhost:8025 |
| DB | localhost:5432 |

---

## ■ 日常操作

### 起動

```bash
make up
```

---

### 停止

```bash
make down
```

---

### 完全削除（DB含む）

```bash
make destroy
```

---

### ログ確認

```bash
make logs
```

---

### コンテナ状態

```bash
make ps
```

---

## ■ 開発用コマンド

### APIコンテナに入る

```bash
make api
```

---

### Frontendコンテナに入る

```bash
make front
```

---

### DB接続

```bash
make db
```

---

### マイグレーション実行

```bash
make migrate-up
```

---

## ■ API開発

### フォーマット

```bash
docker compose exec api gofmt -w .
```

---

### Lint

```bash
docker compose exec api golangci-lint run
```

---

### テスト

```bash
docker compose exec api go test ./...
```

---

## ■ システム構成

以下のコンテナが起動します：

- frontend（React）
- api（Go）
- db（PostgreSQL + PGroonga）
- mail（Mailpit）
- redis（Redis）
- worker（非同期処理）

---

## ■ メール仕様

- Mailpit を使用（開発用）
- SMTP: mail:1025
- Web UI: http://localhost:8025

---

## ■ 注意事項

- `.env` は絶対に Git にコミットしない
- DBデータは `make destroy` で削除される
- JWT_SECRET 未設定の場合は起動エラーになる
- APIは Air によりホットリロードされる

---

## ■ トラブルシュート

### コンテナが起動しない

```bash
make destroy
make up
```

---

### DBが壊れた場合

```bash
make destroy
make up
make migrate-up
```

---

### ポート競合

`.env` を修正：

```env
API_PORT=8081
FRONTEND_PORT=5174
```

---

## ■ 補足

本プロジェクトは以下の設計思想に基づいています：

- Outbox Pattern によるメール送信
- Token Hash によるセキュリティ
- Rate Limit による攻撃対策
- トランザクションによる整合性担保
