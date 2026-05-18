# テスト方針

## テスト設計の考え方

「CI を通すためのテスト」ではなく「再現性のある動作確認」を目的として設計しています。

### flaky test を避ける

E2E テストは専用の Docker Compose プロジェクト（`COMPOSE_PROJECT_NAME=corenambo-e2e`）で dev 環境と完全に分離して実行します。
dev 環境が起動中でも停止中でも結果が変わらず、ホストのポート競合も起きません。
テスト終了後はボリュームごと削除するため、前回のテストデータが今回の結果に影響しません。

### Mailpit polling を使う理由

E2E テストでは `mail_outboxes.payload` ではなく Mailpit REST API からメールを取得します。
worker はメール送信後に `payload` を `'{}'` で上書きします（payload clear）。
`payload` から取得しようとすると、テスト実行のタイミング次第で取れないことがあります。
Mailpit API を 1 秒間隔でポーリングすることで、送信完了後のメールを確実に取得できます。

### Quality Gate の目的

カバレッジ 100% は「未テストのコードをなくすこと」が目的です。
実装と同時にテストを書くことで、後から「テストしにくいコード」が増えることを防ぎます。
`make fmt` / `make lint` / `make frontend-typecheck` など、動作確認以外のチェックコマンドは「動くコード」だけでなく「読めるコード」を維持するために設けています。

---

## 必須要件

- カバレッジ100%
- 全テスト成功

---

## テスト用DB（app_db_test）のセットアップ

### 概要

- ユニットテストは `app_db_test` データベースを使用する（開発用 `app_db` は破壊しない）
- テストDBが存在しない場合は `make test-db-setup` で作成する

### セットアップ手順

```bash
make test-db-setup
```

`make test-db-setup` は以下を順に実行する：

1. 既存の `app_db_test` に接続中のセッションを切断する
2. `app_db_test` を DROP する
3. `app_db_test` を CREATE する（オーナー: `app_user`、エンコード: UTF8、ロケール: ja_JP.UTF-8）
4. `pgcrypto` / `pgroonga` 拡張を作成する（postgres スーパーユーザで実行）
5. 最新 migration を `app_db_test` に適用する

### 実行タイミング

- `make test-cover` を初めて実行する前
- migration を追加・変更したとき
- `app_db_test` が壊れたとき

### 注意事項

- `make test-db-setup` は `app_db`（開発用DB）に影響を与えない
- テスト実行時は `DATABASE_URL` が `app_db_test` を向くよう `Makefile` で設定済み

---

## Backend (Go)

### 対象

- handler
- service
- repository
- その他ロジック

### 実行

```bash
make fmt
make lint
make test-cover
```

### ルール

- テストなしで実装しない
- 境界値を考慮する
- 異常系を必ず含める

---

## Frontend (React)

### フレームワーク

- Vitest（テストランナー）
- @testing-library/react（コンポーネントテスト）
- jsdom（DOM シミュレーション）
- @vitest/coverage-v8（カバレッジ計測）

### 対象

- コンポーネント（レンダリング・props による表示変化・ユーザ操作）
- カスタムフック（状態変化・副作用）
- ユーティリティ関数（純粋関数のロジック）

### 実行

```bash
make frontend-lint
make frontend-test
make frontend-typecheck
```

### ルール

- テストなしで実装しない
- 境界値を考慮する
- 異常系を必ず含める

---

## E2Eテスト

### 対象機能

ユーザ登録フロー（仮登録 → 本登録 → DB反映確認）

### 実行

```bash
make e2e
```

`make e2e` は専用の Docker Compose プロジェクト（`COMPOSE_PROJECT_NAME=corenambo-e2e`）を起動し、Playwright コンテナ内でテストを実行する。dev 環境の起動状態に関係なく単独で実行できる。終了時にコンテナ・ボリュームを自動削除する。

### 補助スクリプト（手動確認用）

`scripts/e2e_user_registration.sh` は curl ベースの API レベル確認スクリプトで、`make e2e` からは呼ばれない。手動でのデバッグや補助確認に使用する旧手順であり、標準手順ではない。

使用する場合は dev 環境が起動済みであること。

```bash
# 手動確認の場合のみ
docker compose up -d db api redis worker mail
bash scripts/e2e_user_registration.sh
```

### テストファイル

```
e2e/tests/user_registration.spec.js
```

### テストケース（現在 1 件）

| テスト | 検証内容 |
|--------|--------|
| 仮登録正常系：フォーム送信 → 完了画面遷移 → メール到着 → token確認 | フォーム送信 → 完了ページ表示 → Mailpit にメール到着 → 本文に `/registration/verify` と `token=` が含まれること |

### トークン取得の仕組み

仮登録 API 呼び出し後、worker が約 1 秒以内にメールを送信し `mail_outboxes.payload` を `'{}'` に上書きする。
そのため `mail_outboxes` ではなく Mailpit REST API からメールを取得する。

Playwright では `pollForEmail()` ヘルパーが 1 秒間隔で最大 30 秒 Mailpit API をポーリングし、テスト開始以降に届いたメールを確認する。

```javascript
// e2e/tests/user_registration.spec.js（概要）
const res = await apiContext.get(
  `${MAILPIT_API}/search?query=${encodeURIComponent('to:' + toEmail)}&limit=10`
);
// 取得したメッセージIDで本文を取得
const msgRes = await apiContext.get(`${MAILPIT_API}/message/${message.ID}`);
```

### クリーンアップ

`make e2e` は `trap` により成功・失敗問わず終了時に以下を実行する。

```bash
docker compose -f docker-compose.yml -f docker-compose.e2e.yml down -v --remove-orphans
```

DB ボリュームが毎回削除・再作成されるため、テストデータは蓄積しない。

### 結果

1 passed（2026-05-18 確認済み）