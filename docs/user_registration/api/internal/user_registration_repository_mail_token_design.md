# ユーザ仮登録 repository / mail / token 設計

## 1. 概要

本設計は、DB / メール / トークンの責務分割を定義する。

---

## 2. 全体方針

- repository：DB操作
- mail：Outbox送信
- token：生成・ハッシュ

---

## 3. repository

### UserRegistrationRequest

- FindByEmail
- Create
- UpdateToken

---

### MailOutbox

- Create（登録）
- FetchPending
- MarkProcessing
- MarkSent
- MarkFailed

---

## 4. Outbox Pattern

### フロー

1. service が mail_outboxes に登録
2. worker が取得
3. SMTP送信
4. 状態更新

---

### 特徴

- APIはメール送信を待たない
- 再試行可能
- 障害耐性あり

---

## 5. token

- Generator：ランダム生成
- Hasher：SHA256

---

## 6. URL

```text
/user-registration/verify?token=xxx
```

---

## 7. worker

- 1秒間隔でポーリング
- pending を取得
- retry制御あり

---

## 8. メリット

- DBとメールの分離
- 再送可能
- 安定性向上
