# ユーザ仮登録 フォーム定義

## 概要

ユーザがメールアドレスを入力し、仮登録を行うフォームです。

本フォームではセキュリティ上の理由により、  
メールアドレスの存在有無はユーザに通知しません。

---

## 項目

- email
- email_confirmation

---

## エラーメッセージ（日本語）

- メールアドレスを入力してください
- 正しいメールアドレス形式で入力してください
- メールアドレスが一致しません

---

## エラーメッセージ（英語）

- Please enter your email address
- Please enter a valid email address
- Email addresses do not match

---

## 正常時の表示

API成功後、`/registration/complete` に遷移する。  
フォーム自体にはメッセージを表示しない。

---

## 完了画面（/registration/complete）

### 遷移条件

- フォーム送信成功後に `/registration/complete` へ遷移する
- 直接アクセスした場合は `/registration` へリダイレクトする

### 受け取るstate

`useNavigate` で渡される `location.state` に以下を含む：

| フィールド          | 型                | 説明                                  |
|---------------|------------------|-------------------------------------|
| email         | string           | 送信対象メールアドレス                        |
| expiresMinutes | integer \| null | 本登録URLの有効期限（分）。`null` の場合は非表示      |

### 表示内容

- 送信先メールアドレス（`state.email`）
- 本登録URLの有効期限（`state.expiresMinutes` が非nullの場合のみ）
- 迷惑メールフォルダの注意書き
- 主要な指示文

---

## セキュリティ方針

- メールアドレスの存在有無は表示しない
- 登録済み・未登録を区別しない
- 常に同一の完了画面を表示する

---

## 備考

- 実際の登録状態はサーバ側で判定する
- フロント側では状態を分岐しない
