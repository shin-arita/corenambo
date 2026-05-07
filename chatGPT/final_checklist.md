# 最終確認チェックリスト（Claude Code対応後）

## 目的

Claude Codeによる修正が正しく反映されているかを確認する。

---

## ① フォーマット確認

以下が正常に実行できること：

docker compose exec api sh -lc '$(go env GOPATH)/bin/goimports -w .'

確認：

- エラーが出ない
- フォーマット差分が発生しない

---

## ② Lint / Test

docker compose exec api golangci-lint run
docker compose exec api go test ./... --cover

確認：

- lint: 0 issues
- test: 全件成功
- カバレッジ: internal配下 100%

---

## ③ DB構造確認

以下を実行：

docker compose exec db psql -U app_user -d app_db -c "\d+ users"
docker compose exec db psql -U app_user -d app_db -c "\d+ user_emails"
docker compose exec db psql -U app_user -d app_db -c "\d+ user_verifications"

確認：

- check制約が存在する
- unique indexが存在する
- user_agent length制限が入っている

---

## ④ Migration確認

docker compose exec db psql -U app_user -d app_db -c "SELECT * FROM schema_migrations;"

確認：

- 最新versionが適用されている
- dirty=false

---

## ⑤ Git確認

git status
git diff --stat

確認：

- 不要な変更が含まれていない
- docsディレクトリが正しい（docs/ai）
- cluade という誤字が存在しない

---

## ⑥ セキュリティ確認

確認：

- tokenがログ出力されていない
- passwordがログ出力されていない
- emailがhash化されている（rate limiter）

---

## ⑦ 重要ロジック確認

確認：

- Redis TTLがリセットされない（ExpireNX）
- rate limitがIP / emailで分かれている
- emailが正規化されている（lowercase）

---

## ⑧ 最終判断

以下をすべて満たしたらOK：

- lint 0件
- test 全成功
- カバレッジ100%
- DB制約反映済み
- migration適用済み
- Git差分問題なし

→ OKならコミット可能

---

## 出力要求（Claude Codeへ）

以下を必ず出力：

1. 各チェック結果
2. 問題があれば修正案
3. コミット可否（YES / NO）
