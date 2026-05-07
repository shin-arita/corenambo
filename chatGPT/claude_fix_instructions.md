# 修正指示（最優先）

## 現状

以下の問題が発生しています。

- gofmt が実行できていない
- docs ディレクトリのパスが壊れている（cluade code）
- git の staged / unstaged が不整合
- go.mod のバージョンが意図せず変更されている可能性
- migration が未反映 or 未整理の可能性

---

## やること（必ず順番通り）

### ① gofmt 修正

問題：
- gofmt が見つからない

対応：

docker compose exec api sh -lc 'go install golang.org/x/tools/cmd/goimports@latest'
docker compose exec api sh -lc '$(go env GOPATH)/bin/goimports -w .'

---

### ② docs ディレクトリ修正（最重要）

問題：

docs/cluade code/
docs/claude code/

→ typo + スペースありで崩壊している

対応：

docs/ai/
  development.md
  backend.md
  frontend.md
  security.md
  database.md
  migration.md
  testing.md

やること：

- cluade → claude → ai に統一
- スペースを削除
- ディレクトリを整理
- 不要ファイル削除

---

### ③ Git 状態修正

対応：

git restore --staged .
git add .
git status

---

### ④ go.mod 修正

問題：

go 1.25 → 1.24 に変更されている

確認：

- 意図した変更か？

意図しない場合：

git restore api/go.mod

---

### ⑤ migration 確認

docker compose exec api sh -lc 'migrate -path /db/migrations -database "$DATABASE_URL" up'

---

### ⑥ 再実行（必須）

docker compose exec api go test ./... --cover
docker compose exec api golangci-lint run

---

## 完了条件（これを満たすまで終了禁止）

- gofmt または goimports が正常に実行できる
- docs ディレクトリが docs/ai/ に統一されている
- git status がクリーン
- go.mod が意図通り
- migration が適用済み
- lint 0件
- test 100%カバレッジ

---

## 出力要求

以下を必ず出力：

1. 修正後のディレクトリ構成
2. git status
3. 実行コマンド結果
4. 変更理由の説明（簡潔）
