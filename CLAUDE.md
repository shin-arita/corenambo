# CLAUDE.md

## Project

**コレナンボ・オークション** — 価格降下型オークションシステム。

- 時間経過で価格が下がる
- ユーザはホールド（一時確保）できる
- 複数ユーザが競合すると価格が上昇する
- 実サービス運用 + 技術ポートフォリオ公開が目的

---

## Tech Stack

| 領域 | 技術 |
|------|------|
| Frontend | React + Vite + TypeScript |
| Backend | Go + Gin |
| DB | PostgreSQL + PGroonga |
| Cache / Rate Limit | Redis |
| Tokenizer | MeCab |
| Environment | Docker Compose |
| Migration | golang-migrate |
| Hot Reload | air |
| Dev Mail | Mailpit |
| UUID | v7をGo側で生成（DBでは生成しない） |
| Go Module | app-api |

---

## Development Flow

**必ずこの順序で進める。省略禁止。**

1. 方針決定（ユーザ + Claude Code）
2. 設計（Claude Code） → ユーザレビュー
3. 実装 + テスト（Claude Code） → lint/test通過確認
4. 動作確認（ユーザ）
5. ドキュメント作成（Claude Code） → ユーザレビュー
6. コミット（ユーザのみ）

---

## Absolute Rules

**禁止事項（即停止）**

- 方針未決定で設計しない
- ユーザレビュー前に実装しない
- Claude Codeはコミットしない
- `latest` / `@latest` を使用しない
- 指示なしのリファクタリング禁止
- コメント・インデント・クォートを勝手に変更しない

**必須事項（完了条件）**

- 実装後にテストコード作成・更新
- カバレッジ100%達成
- lint / test 全通過
- ユーザの最終動作確認完了
- フロントエンドは長期運用前提で TypeScript + TSX に統一する
- 既存 JSX は段階的に TSX へ移行する
- TSX移行は1ファイルずつ行い、各ステップで lint / test を通す

---

## Quality Gate

**Backend（必須）**

```bash
docker compose exec api gofmt -w internal cmd
docker compose exec api golangci-lint run
docker compose exec api go test ./... -cover
```

**Frontend（必須）**

```bash
docker compose exec frontend npx eslint src --fix
docker compose exec frontend npm run lint
docker compose exec frontend npm run test
```

**latest混入チェック（必須）**

```bash
grep -RIn --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=vendor "latest\|@latest" .
```

---

## Documentation

**作業前に必ず参照。作業後に必ず更新。**

docs/ai/
├── development.md
├── backend.md
├── frontend.md
├── security.md
├── database.md
├── migration.md
└── testing.md

---

## Naming

正式名称：**コレナンボ・オークション**

「コレナンボ」と省略しない。