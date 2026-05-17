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
| UUID | UUID v7 を Go 側で生成 |
| Go Module | app-api |

---

## Development Flow

必ずこの順序で進める。

1. 方針決定（ユーザ + Claude Code）
2. 設計（Claude Code） → ユーザレビュー
3. 実装 + テスト（Claude Code）
4. lint / test 通過確認
5. 動作確認（ユーザ）
6. ドキュメント更新
7. コミット（ユーザのみ）

---

## Absolute Rules

### 禁止事項

- 方針未決定で設計しない
- ユーザレビュー前に実装しない
- Claude Codeはコミットしない
- `latest` / `@latest` を使用しない
- 指示なしのリファクタリング禁止
- コメント・インデント・クォートを勝手に変更しない
- bootstrap script と migration を混在させない
- migration に連番形式を使用しない

### 必須事項

- 実装後にテストコード作成・更新
- カバレッジ100%達成
- lint / test 全通過
- ユーザの最終動作確認完了
- フロントエンドは TypeScript + TSX に統一する

---

## Architecture Rules

### Migration

- migration名は `YYYYMMDDHHMMSS_name` を使用する
- timestamp migration を必須とする
- migration は schema evolution のみを扱う
- extension作成は bootstrap 側で扱う

### Bootstrap Script

```text
db/migrations/000000_init_user.sql
```

これは migration ではなく bootstrap script として扱う。

役割:

- role作成
- DB作成
- extension作成

### UUID

- UUID は Go 側で v7 を生成する
- DB側で UUID を生成しない

### Database

- PostgreSQL を唯一の正とする
- FK を必ず使用する
- 物理削除より論理削除を優先する
- DBは最低限の整合性を守る
- 複雑な validation は application layer で行う

### Application Layer

- validation は Go 側を主とする
- repository は DB責務のみを持つ
- service に business logic を集約する
- handler に business logic を書かない

### Testing

- 開発DBとテストDBを分離する
- 開発RedisとテストRedisを分離する
- test は必ずテスト用DB/Redisを使用する

---

## Quality Gate

### Backend

```bash
make fmt
make test-cover
```

### Frontend

```bash
make frontend-lint
make frontend-test
make frontend-typecheck
```

### latest混入チェック

```bash
grep -RIn --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=vendor "latest\|@latest" .
```

---

## Documentation

作業前に参照し、作業後に更新する。

```text
docs/ai/
├── development.md
├── backend.md
├── frontend.md
├── security.md
├── database.md
├── migration.md
└── testing.md
```

---

## Naming

正式名称:

```text
コレナンボ・オークション
```

「コレナンボ」と省略しない。