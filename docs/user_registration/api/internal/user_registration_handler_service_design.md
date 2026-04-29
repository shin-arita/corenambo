# ユーザ仮登録 handler / service 設計

## 1. 概要

本設計は、ユーザ仮登録APIにおける handler / service の責務を定義する。

---

## 2. 全体方針

- handler：HTTP制御
- service：業務ロジック
- repository：DBアクセス
- mail：Outbox登録

---

## 3. handler

### 責務

- JSON bind
- Accept-Language取得
- rate limit
- service呼び出し
- レスポンス生成

---

### 注意

- Accept-Language はそのまま取得
- translator が fallback を担当

---

## 4. service

### 責務

- バリデーション
- 仮登録レコード取得
- トークン生成
- ハッシュ化
- DB保存
- Outbox登録

---

## 5. 業務ルール

### ■ メール存在隠蔽

- 既存ユーザでも成功を返す

---

### ■ 再送制御

- last_sent_at を使用
- 一定時間内は再送しない

---

### ■ トークン

- 毎回再生成
- ハッシュ保存

---

## 6. 処理フロー

1. バリデーション
2. 仮登録取得
3. 再送制御チェック
4. トークン生成
5. DB保存（tx）
6. Outbox登録
7. 201返却

---

## 7. メール送信

- service は送信しない
- mail_outboxes に登録する
- worker が送信

---

## 8. レート制限

- IP単位
- email単位

（handlerで実装済み）

---

## 9. Verify API

- 現状は未実装（stub）
