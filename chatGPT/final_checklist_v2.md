# 最終確認チェックリスト（コミット前・完全版）

## 目的
Claude Codeによる修正が「安全にコミットできる状態か」を最終確認する。

---

## ① Goフォーマット確認

docker compose exec api sh -lc '$(go env GOPATH)/bin/goimports -w .'

確認：
- エラーなし
- 差分なし（再実行しても変化なし）

---

## ② Lint / Test

docker compose exec api golangci-lint run
docker compose exec api go test ./... --cover

確認：
- lint: 0 issues
- test: 全成功
- internal 配下: カバレッジ100%

---

## ③ DB制約チェック（重要）

docker compose exec db psql -U app_user -d app_db -c "\d+ user_emails"
docker compose exec db psql -U app_user -d app_db -c "\d+ user_verifications"
docker compose exec db psql -U app_user -d app_db -c "\d+ user_login_histories"

確認：
- user_emails:
  - primaryはverified必須
- user_verifications:
  - pendingのユニーク制約あり
- user_login_histories:
  - failure_reason制約あり

---

## ④ Migration状態

docker compose exec db psql -U app_user -d app_db -c "SELECT * FROM schema_migrations;"

確認：
- 最新version
- dirty = false

---

## ⑤ Git整合性

git status
git diff --stat

確認：
- staged / unstaged 混在なし
- 不要な削除・追加なし
- docsディレクトリ正常

特にチェック：
- docs/ai/ に統一されている
- 「cluade」誤字が存在しない

---

## ⑥ セキュリティ確認

確認：
- メールアドレスが rate limiter で hash化されている
- token をログ出力していない
- password をログ出力していない
- SMTPヘッダインジェクション対策あり

---

## ⑦ ロジック確認（重要）

確認：
- Redis:
  - ExpireNX が使われている（TTLリセットされない）
- Email:
  - lowercase + trim 正規化
- Outbox:
  - SELECTではなく UPDATE ... RETURNING でロック取得

---

## ⑧ 重大リスクチェック

以下が1つでもあればコミット禁止：

- gofmt 未適用
- lintエラーあり
- test失敗あり
- カバレッジ未達
- migration未適用
- DB制約不足
- git状態不整合

---

## 最終判断

すべてOKなら：

→ コミットしてよい

---

## Claude Codeへの出力要求

以下を必ず出力：

1. 各チェックの結果
2. 問題があれば修正内容
3. 最終判断（YES / NO）
