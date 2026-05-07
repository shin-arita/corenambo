# フロントエンド方針

## 技術

- React + Vite

---

## ルール

- コンポーネントは責務ごとに分割する
- ロジックとUIを分離する
- 再利用性を意識する

---

## テスト

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

```
docker compose exec frontend npm run lint
docker compose exec frontend npm run test -- --coverage
```

---

## セキュリティ

- tokenをログ出力しない
- URLに含まれるtokenもログ出力しない