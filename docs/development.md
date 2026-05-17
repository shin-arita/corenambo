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
- Docker Compose **v2.24.0 以上**（`!reset` タグ対応版。`docker compose version` で確認）
- Make

> **Note:** `docker-compose.e2e.yml` では `ports: !reset []` / `container_name: !reset ""` を使用しており、Docker Compose v2.24.0 未満では E2E 起動が失敗する。CI 環境も同バージョン以上を確保すること。

---

## ■ 初回セットアップ

### 1. 環境変数設定

```bash
cp .env.example .env
```

### 2. コンテナ起動

```bash
make up
```

### 3. DBマイグレーション

```bash
make migrate-up
```

> **JWTシークレットについて:** 現時点では認証機能未実装のため `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` は未設定でも起動する。認証機能実装時に `.env` へ設定すること（`openssl rand -base64 32`）。

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

### 停止

```bash
make down
```

### 完全削除（コンテナ + DB ボリューム含む全ボリューム削除）

```bash
make destroy
```

### ログ確認

```bash
make logs
```

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

### Frontendコンテナに入る

```bash
make front
```

### DB接続

```bash
make db
```

### マイグレーション実行

```bash
make migrate-up
```

---

## ■ API開発

### フォーマット

```bash
make fmt
```

> 内部的に `PATH="/usr/local/go/bin:$PATH" gofmt -w ...` を実行する。`sh -lc` のログインシェルは Go の PATH を持たないため、Makefile がパスを明示的に追加している。

### Lint

```bash
make lint
```

### テスト（カバレッジ100%必須）

```bash
make test-cover
```

---

## ■ E2Eテスト

### 実行

```bash
make e2e
```

### 動作概要

| 項目 | 内容 |
|------|------|
| 実行環境 | Docker コンテナ内 Playwright（ホストに Node.js 不要） |
| プロジェクト分離 | `COMPOSE_PROJECT_NAME=corenambo-e2e` により dev 環境と完全分離 |
| ホストポート | `ports: !reset []` で全サービスのホストバインドを無効化。dev 環境起動中でも port conflict しない |
| クリーンアップ | `trap ... down -v --remove-orphans` で成功・失敗問わずコンテナ・ボリュームを自動削除 |
| DB初期化 | 毎回 migration を再適用（`corenambo-e2e_db_data` ボリュームは `down -v` で削除済み） |
| メール確認 | Mailpit REST API（`http://mail:8025/api/v1/search`）を 1 秒間隔で最大 30 秒 polling |
| 通信経路 | E2E コンテナ → Docker 内部 DNS（`frontend`, `api`, `mail`）→ 各サービス。ホスト経由なし |

### 手動で停止せずに実行する場合

通常開発環境（`make up` 済み）との共存：
- ホストポートは競合しない（E2E 側がバインドしない）
- ただし E2E は専用 DB を持つため、dev 環境が起動したまま実行しても安全

> DB ボリュームは毎回削除・再作成されるため、E2E を繰り返し実行してもデータが蓄積しない。

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
- `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` は現時点では任意（認証機能未実装）
- APIは Air によりホットリロードされる

---

## ■ トラブルシュート

### コンテナが起動しない

```bash
make destroy
make up
```

### DBが壊れた場合

```bash
make destroy
make up
make migrate-up
```

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
