# ユーザ仮登録 テーブル定義

## テーブル

`user_registration_requests`

## 概要

ユーザ仮登録情報を管理するテーブルです。

仮登録メール送信用トークンは平文では保存せず、ハッシュ化した値のみ保存します。
また、仮登録メールの再送間隔制御のため、最終送信日時を保持します。

---

## カラム

| カラム名 | 説明 |
|---|---|
| id | 仮登録ID。UUID v7をGo側で生成する |
| email | 仮登録対象メールアドレス |
| token_hash | 本登録用トークンのハッシュ値 |
| expires_at | トークン有効期限 |
| verified_at | 本登録完了日時。未完了の場合はNULL |
| last_sent_at | 仮登録メールの最終送信日時 |
| created_at | 作成日時 |

---

## 方針

- UUIDはDBで生成しない
- UUID v7をGo側で生成する
- トークンは平文保存しない
- トークンは `token_hash` としてハッシュ保存する
- 仮登録メールの再送制御には `last_sent_at` を使用する
- `updated_at` は持たない

---

## セキュリティ方針

- トークン漏洩時の被害を抑えるため、DBにはハッシュ値のみ保存する
- メールアドレスの存在有無をAPIレスポンスから推測できないようにする
- 短時間での大量送信を防ぐため、再送間隔を制御する

---

## 関連実装

- `api/internal/model/user_registration_request.go`
- `api/internal/repository/user_registration_request_repository.go`
- `api/internal/repository/user_registration_request_repository_impl.go`
- `api/internal/service/user_registration_service.go`
