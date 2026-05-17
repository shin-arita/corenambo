# データベース方針

## 使用技術

- PostgreSQL
- PGroonga + MeCab

---

## ルール

- UUIDはGoで生成する（DBで生成しない）
- テーブル・カラムにコメントを付与する
- インデックスは用途を明確にする

---

## 設計方針

- 正規化を基本とする
- 必要に応じてパフォーマンス最適化を行う

---

## 会員ステップ成立条件

会員の機能解放状態は専用カラムではなく、テーブルの存在・内容から判定する。

### Step1 完了条件

`user_profiles` レコードが存在すること。

成立すると解放される機能:

- 閲覧
- ウォッチ

### Step2 完了条件

`user_profile_details` レコードが存在し、かつ以下の必須フィールドがすべて非NULLであること。

必須フィールド:

- `phone_country_code`
- `phone_number`
- `country_code`
- `postal_code`
- `region`
- `locality`
- `address_line1`

空レコードのみが存在する状態ではStep2完了とみなさない。
判定はGoのservice層で行う（DBのCHECK制約は最低限の整合性のみを担う）。

成立すると解放される機能:

- ホールド
- 入札

### Step3 完了条件

`user_verifications` を `user_id` で絞り込み、`created_at DESC` で最新1件を取得したとき、
そのレコードの `status` が `'approved'` であること。

- approved が1件でも存在すれば成立とはしない（最新レコードのみを対象とする）
- 承認期限は現時点では設けない
- 将来、承認期限や取消が必要になった場合は別途設計する

成立すると解放される機能:

- 出品
