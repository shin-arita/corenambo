# マイグレーション方針

## 命名規則

タイムスタンプ形式を使用する

例:
20260506120000_create_users_table.up.sql

---

## 実行

docker compose exec api sh -lc 'migrate -path /db/migrations -database "$DATABASE_URL" up'

---

## ルール

- up / down を必ず用意する
- 変更内容は最小単位に分割する
- 既存データへの影響を考慮する