# ユーザ仮登録 repository / mail / token 設計

## 1. 概要

本設計は、ユーザ仮登録機能における repository / mail / token の責務分割を定義する。  
service から下位レイヤへの依存を明確化し、DBアクセス、メール送信、トークン生成を分離することで、保守性・テスト容易性・差し替え容易性を高めることを目的とする。

---

## 2. 全体方針

本機能では、以下の責務で分離する。

- repository
  - DBアクセス専用
- mail
  - メール送信専用
- token
  - トークン生成 / ハッシュ化専用

service はこれらを組み合わせて業務処理を実行する。

---

## 3. レイヤ責務

### 3.1 repository の責務

- SQL実行
- DBレコード取得
- DBレコード登録
- DBレコード更新
- transaction 内での DB処理

### 担当しないこと

- 業務ルール判定
- メール送信
- トークン生成
- 文言解決
- HTTPレスポンス生成

---

### 3.2 mail の責務

- メール送信依頼の受付
- 宛先、件名、本文の構築
- SMTP / 外部メールサービスへの送信
- テンプレート適用
- 言語切替

### 担当しないこと

- DB更新
- トークン保存
- 業務ルール判定
- HTTP制御

---

### 3.3 token の責務

- ランダムトークン生成
- トークンハッシュ化

---

## 4. 依存関係

Handler → Service → Repository / Mail / Token → DB / Mail

---

## 5. repository 設計

### UserRepository

```go
type UserRepository interface {
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}
```

---

### UserRegistrationRequestRepository

```go
type UserRegistrationRequestRepository interface {
    FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error)
    Create(ctx context.Context, req model.UserRegistrationRequest) error
    UpdateToken(ctx context.Context, req model.UserRegistrationRequest) error
}
```

---

### TxManager

```go
type TxManager interface {
    WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
```

---

## 6. モデル

```go
type UserRegistrationRequest struct {
    ID         string
    Email      string
    TokenHash  string
    ExpiresAt  time.Time
    VerifiedAt *time.Time
    CreatedAt  time.Time
}
```

---

## 7. mail 設計

### Mailer

```go
type Mailer interface {
    SendUserRegistrationMail(ctx context.Context, mail UserRegistrationMail) error
}
```

---

### Mail構造

```go
type UserRegistrationMail struct {
    To   string
    URL  string
    Lang string
}
```

---

## 8. token 設計

### TokenGenerator

```go
type TokenGenerator interface {
    Generate() (string, error)
}
```

---

### TokenHasher

```go
type TokenHasher interface {
    Hash(value string) (string, error)
}
```

---

## 9. 補助コンポーネント

### UUIDGenerator

```go
type UUIDGenerator interface {
    NewV7() (string, error)
}
```

---

### Clock

```go
type Clock interface {
    Now() time.Time
}
```

---

### URLBuilder

```go
type RegistrationURLBuilder interface {
    Build(token string) string
}
```

---

## 10. service 利用イメージ

```go
plainToken, _ := tokenGenerator.Generate()
tokenHash, _ := tokenHasher.Hash(plainToken)
url := urlBuilder.Build(plainToken)

mailer.SendUserRegistrationMail(ctx, mail.UserRegistrationMail{
    To: input.Email,
    URL: url,
})
```

---

## 11. テスト観点

### repository

- 取得・登録・更新が正しく動作する

### mail

- 宛先・URL・言語が正しく渡る

### token

- 毎回異なる値が生成される
- ハッシュ化される

---

## 12. 最終方針

- repository → DB責務のみ
- mail → 送信責務のみ
- token → 生成責務のみ

service はこれらを統合するが、詳細には依存しない。
