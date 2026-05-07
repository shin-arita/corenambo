# テスト方針

## 必須要件

- カバレッジ100%
- 全テスト成功

---

## Backend (Go)

### 対象

- handler
- service
- repository
- その他ロジック

### 実行

docker compose exec api go test ./... -cover

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

docker compose exec frontend npm run lint
docker compose exec frontend npm run test -- --coverage

### ルール

- テストなしで実装しない
- 境界値を考慮する
- 異常系を必ず含める