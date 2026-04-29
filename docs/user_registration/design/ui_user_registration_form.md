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

以下のメッセージを常に表示する：

- 仮登録メールを送信しました。メールをご確認ください。
- A temporary registration email has been sent. Please check your email.

---

## セキュリティ方針

- メールアドレスの存在有無は表示しない
- 登録済み・未登録を区別しない
- 常に同一の成功メッセージを表示する

---

## 備考

- 実際の登録状態はサーバ側で判定する
- フロント側では状態を分岐しない
