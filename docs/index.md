# ドキュメント索引

コレナンボ・オークションの設計・実装ドキュメント一覧。

新しくドキュメントを追加したときはここを更新する。

---

## 最初に読むべき docs

設計判断の背景を知りたい場合、この順番で読むと文脈がつながります。

1. **[Outbox Pattern 採用理由](architecture/outbox_pattern_decision.md)** — DB とメールをどうつなぐか。なぜ SMTP 直送が壊れるか
2. **[リトライ戦略](architecture/retry_strategy_decision.md)** — 失敗を分類しないと何が起きるか。stuck mail とは何か
3. **[at-least-once delivery と冪等性](architecture/at_least_once_and_idempotency.md)** — exactly-once を諦めた理由と、本登録の二重防御
4. **[セキュリティ・性能設計](user_registration/security_and_performance.md)** — 19 項目の実装レベル対策
5. **[テスト方針](ai/testing.md)** — dev/E2E 分離・Mailpit polling・カバレッジ 100% の理由

---

## こんな人はここから

| 知りたいこと | 最初に読む doc |
|---|---|
| 設計判断の理由・ADR | [Architecture — 設計判断記録](#architecture--設計判断記録)（このページ内） |
| API のリクエスト・レスポンス仕様 | [ユーザ登録 API仕様](user_registration/api/user_registration_api.md) |
| 開発環境の立ち上げ方 | [開発環境セットアップ](development.md) |
| E2E / テストの実行方法 | [テスト方針](ai/testing.md) |
| セキュリティ実装の詳細 | [セキュリティ・性能設計](user_registration/security_and_performance.md) |

---

## Architecture — 設計判断記録

設計の「なぜ」を記録する。実装内容ではなく判断の根拠・代替案・副作用を扱う。

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| Outbox Pattern 採用理由 (Why Outbox Pattern) | [docs/architecture/outbox_pattern_decision.md](architecture/outbox_pattern_decision.md) | SMTP直送を避けた理由・pollingが必要な理由・payload clearの意図 |
| リトライ戦略 — retryable / non-retryable / stuck (Retry Strategy) | [docs/architecture/retry_strategy_decision.md](architecture/retry_strategy_decision.md) | 失敗分類の根拠・stuck mailの発生原因と回復設計・リトライ上限 |
| at-least-once delivery と冪等性方針 (At-Least-Once Delivery and Idempotency) | [docs/architecture/at_least_once_and_idempotency.md](architecture/at_least_once_and_idempotency.md) | exactly-onceを採用しない理由・重複許容の根拠・本登録の二重防御 |

---

## Architecture / Decisions — 機能要件と意思決定

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| オークション仕様書 (Auction Specification) | [docs/auction/auction_spec.md](auction/auction_spec.md) | 価格降下・ホールド・同時購入制御・冪等性の仕様 |
| ユーザ仮登録 要件定義書 (User Registration Requirements) | [docs/user_registration/requirements/user_registration_requirements.md](user_registration/requirements/user_registration_requirements.md) | トークン再生成方式・Outbox採用・存在隠蔽の意思決定 |

---

## API

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| ユーザ登録 API仕様 (User Registration API Spec) | [docs/user_registration/api/user_registration_api.md](user_registration/api/user_registration_api.md) | 仮登録・本登録・トークン確認の全エンドポイント仕様（リクエスト・レスポンス・エラーコード） |

---

## Backend

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| handler / service 設計 (Handler and Service Design) | [docs/user_registration/api/internal/user_registration_handler_service_design.md](user_registration/api/internal/user_registration_handler_service_design.md) | 処理フロー・存在隠蔽ルール・bcrypt DoS対策の設計判断 |
| repository / mail / token 設計 (Repository, Mail, and Token Design) | [docs/user_registration/api/internal/user_registration_repository_mail_token_design.md](user_registration/api/internal/user_registration_repository_mail_token_design.md) | Outbox Pattern詳細・リトライ制御・payload スキーマ |
| エラー / i18n 設計 (App Error and i18n Design) | [docs/user_registration/api/internal/user_registration_app_error_i18n_design.md](user_registration/api/internal/user_registration_app_error_i18n_design.md) | code中心設計・serviceはmessageを持たない・handler責務 |

---

## Database

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| データベース方針 (Database Guidelines) | [docs/ai/database.md](ai/database.md) | UUID生成・正規化・会員ステップ成立条件 |
| マイグレーション方針 (Migration Guidelines) | [docs/ai/migration.md](ai/migration.md) | タイムスタンプ命名・up/down必須・最小分割 |
| 仮登録テーブル定義 (Provisional Registration Table) | [docs/user_registration/db/db_user_registration_requests.md](user_registration/db/db_user_registration_requests.md) | カラム定義・token_hash保存方針・last_sent_at |

---

## Frontend

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| フロントエンド方針 (Frontend Guidelines) | [docs/ai/frontend.md](ai/frontend.md) | React+Vite・テストフレームワーク・セキュリティ |
| 仮登録フォーム定義 (Registration Form Design) | [docs/user_registration/design/ui_user_registration_form.md](user_registration/design/ui_user_registration_form.md) | フォーム項目・完了画面・state渡し・リダイレクト設計 |

---

## Infrastructure

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| 開発環境セットアップ (Development Environment Setup) | [docs/development.md](development.md) | Docker Compose・初回セットアップ・E2Eテスト実行・ポート競合対処 |

---

## Security

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| セキュリティ方針 (Security Guidelines) | [docs/ai/security.md](ai/security.md) | アクセスログマスク・プロキシ設定・トークンクリーニング |
| セキュリティ・性能設計 (Security and Performance Design) | [docs/user_registration/security_and_performance.md](user_registration/security_and_performance.md) | 19項目の実装レベル対策（crypto/rand・bcrypt・SMTPS・CORS等） |

---

## Testing

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| テスト方針 (Testing Guidelines) | [docs/ai/testing.md](ai/testing.md) | テストDB分離・Backend/Frontend/E2Eテスト方針 |
| テスト一覧・カバレッジ分析 (Test Inventory and Coverage) | [docs/ai/test-inventory.md](ai/test-inventory.md) | 仮会員登録フロー全テスト一覧（130件以上） |
| 手動テスト手順書 — 仮会員登録 (Manual Test: Provisional Registration) | [docs/testing/manual_user_registration_request.md](testing/manual_user_registration_request.md) | TC 20件。Mailpit確認・Redisクリーン・ログ確認 |
| 手動テスト手順書 — 本会員登録 (Manual Test: Full Registration) | [docs/testing/manual_user_registration_verify.md](testing/manual_user_registration_verify.md) | TC 30件以上。トークンマスク・URLクリーニング・bcrypt確認 |

---

## AI / Development Rules

| 日本語タイトル（English Title） | パス | 概要 |
|---|---|---|
| 開発方針 (Development Guidelines) | [docs/ai/development.md](ai/development.md) | 開発フロー・禁止事項・完了条件 |
| バックエンド方針 (Backend Guidelines) | [docs/ai/backend.md](ai/backend.md) | CLAUDE.md に統合済みのため縮小管理 |
