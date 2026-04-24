# ユーザ仮登録 テーブル定義

## テーブル
user_registration_requests

## カラム
- id (UUID v7, Go側生成)
- email
- verification_token_hash
- expires_at
- verified_at
- created_at

## 方針
- UUIDはDBで生成しない
- トークンはハッシュ保存
- updated_atは持たない
