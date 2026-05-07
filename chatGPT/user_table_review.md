# ユーザ関連テーブル設計レビュー

## 前提

現在のDB設計（users / user_* 系テーブル）について、以下の観点でレビューしてください。

- セキュリティ
- 整合性
- 将来拡張性
- 不整合リスク
- 制約不足

各指摘について以下を必ず出力してください。

- 問題点
- なぜ問題か
- 改善案
- 必要ならDDL

---

## 指摘事項

### 1. users.display_name

#### 現状
- NOT NULL

#### 懸念
- 仮登録〜本登録の間で未入力になる可能性がある
- SNSログイン拡張時に未設定になる可能性がある

#### 確認事項
- NULL許容に変更すべきか
- default値で対応すべきか

---

### 2. user_emails.is_primary と verified_at

#### 懸念
- 未認証メールが主メールになる可能性がある

#### 確認事項
- 以下の制約を追加すべきか

CHECK (is_primary = false OR verified_at IS NOT NULL)

---

### 3. user_passwords

#### 懸念
- ハッシュアルゴリズム変更に対応できない

#### 確認事項
- password_algo カラム追加の必要性

---

### 4. user_sessions.user_agent

#### 懸念
- text型で無制限のため肥大化リスク

#### 確認事項
- length制限 or 別テーブル分離の必要性

---

### 5. user_verifications（KYC）

#### 懸念
- 同一ユーザで複数の pending が作成できる

#### 確認事項
- 以下の制約を追加すべきか

CREATE UNIQUE INDEX uq_user_verifications_pending
ON user_verifications(user_id)
WHERE status = 'pending';

---

### 6. user_login_histories

#### 懸念
- success=false の場合に failure_reason がNULLになる可能性

#### 確認事項
- 制約追加の必要性

---

### 7. user_registration_requests

#### 懸念
- 同一メールで未完了レコードが複数作成される可能性

#### 確認事項
- 以下のインデックスを追加すべきか

CREATE INDEX idx_user_registration_requests_active
ON user_registration_requests(email)
WHERE verified_at IS NULL AND expires_at > now();

---

## 追加検討事項

### 1. プロフィール分離

#### 懸念
- display_name のみでは情報が不足する

#### 確認事項
- user_profiles テーブル分離の必要性

---

### 2. 権限（role）

#### 懸念
- 管理者権限などの概念が未定義

#### 確認事項
- usersに持たせるか別テーブルにするか

---

### 3. オークション連携

#### 懸念
- seller / buyer の概念が未定義

#### 確認事項
- userにroleを持たせるか
- 出品者を別テーブルにするか

---

## 最終アウトプット要求

以下を必ず出力してください：

1. 現設計の評価（10点満点）
2. 優先度付き修正一覧
3. 修正DDL（必要な場合）
4. 「今すぐやるべき修正」と「後回しでよい修正」の分類
