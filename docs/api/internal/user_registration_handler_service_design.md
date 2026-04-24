# ユーザ仮登録 handler / service 設計

## 1. 概要

本設計は、ユーザ仮登録APIにおける handler / service の責務分割を定義する。  
HTTP入出力と業務ルールを明確に分離し、実装・テスト・保守性を高めることを目的とする。

---

## 2. 全体方針

責務は以下のように分離する。

- handler
  - HTTPリクエスト受信
  - JSONバインド
  - ヘッダ取得
  - service 呼び出し
  - HTTPレスポンス返却

- service
  - 業務バリデーション
  - 既存ユーザ判定
  - 仮登録レコード判定
  - トークン生成
  - ハッシュ化
  - 保存処理
  - メール送信依頼

- repository
  - DBアクセス

- i18n / app_error
  - code から表示文言への変換
  - エラー構造統一

---

## 3. レイヤ責務

## 3.1 handler の責務

handler は HTTP 層専用とする。

### 担当すること

- HTTPリクエスト受信
- JSONバインド
- リクエスト形式チェック
- `Accept-Language` 取得
- service input への詰め替え
- service 呼び出し
- service結果を HTTP レスポンスへ変換
- `code + message` 形式で正常レスポンス返却
- `app_error` を使った異常レスポンス返却

### 担当しないこと

- 既存ユーザ判定
- 仮登録済みデータ判定
- トークン生成
- トークンハッシュ化
- DB保存
- メール送信
- 表示文言のベタ書き

---

## 3.2 service の責務

service は業務ルールの中核とする。

### 担当すること

- 入力値の業務バリデーション
- 既存ユーザ確認
- 仮登録レコード確認
- 新規作成 / 再送 / 再発行の判定
- UUID v7 生成
- トークン生成
- トークンハッシュ化
- 保存内容の決定
- repository 呼び出し
- メール送信依頼
- code を返す

### 担当しないこと

- HTTPステータスコードの直接返却
- JSON整形
- `gin.Context` の操作
- 日本語 / 英語文言の直接返却

---

## 4. 正常レスポンス方針

正常系も異常系と同様に `code + message` で統一する。

### 例

```json
{
  "code": "USER_REGISTRATION_REQUEST_CREATED",
  "message": "仮登録メールを送信しました。メールをご確認ください。"
}
```

### 方針

- service は `code` のみ返す
- handler で i18n 解決を行う
- 表示文言は `internal/i18n` に集約する
- service に成功メッセージをベタ書きしない

---

## 5. 処理シーケンス

```text
Client
  ↓
Handler
  ↓
Service
  ↓
Repository / TokenGenerator / TokenHasher / UUIDGenerator / Clock / Mailer
  ↓
DB / Mail
```

---

## 6. handler 設計

### 6.1 配置

```text
api/internal/handler/user_registration_handler.go
```

---

### 6.2 構造

```go
type UserRegistrationHandler struct {
    userRegistrationService service.UserRegistrationService
    translator              i18n.Translator
}
```

---

### 6.3 リクエストDTO

```go
type CreateUserRegistrationRequest struct {
    Email             string `json:"email"`
    EmailConfirmation string `json:"email_confirmation"`
}
```

---

### 6.4 正常レスポンスDTO

```go
type SuccessResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

---

### 6.5 handler メソッド

```go
func (h *UserRegistrationHandler) Create(c *gin.Context)
```

---

### 6.6 handler 処理手順

1. `Accept-Language` を取得
2. JSON を bind
3. bind失敗なら `BAD_REQUEST`
4. service input に詰め替え
5. service を呼び出し
6. service が返した `code` を i18n で翻訳
7. `201 Created` と `code + message` を返却
8. 異常時は `app_error` を返却

---

### 6.7 handler 疑似コード

```go
func (h *UserRegistrationHandler) Create(c *gin.Context) {
    lang := c.GetHeader("Accept-Language")
    if lang == "" {
        lang = "ja"
    }

    var req CreateUserRegistrationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := app_error.NewBadRequest("BAD_REQUEST")
        c.JSON(appErr.StatusCode(), appErr.ToResponse(lang))
        return
    }

    input := service.CreateUserRegistrationInput{
        Email:             req.Email,
        EmailConfirmation: req.EmailConfirmation,
        Language:          lang,
    }

    output, err := h.userRegistrationService.Create(c.Request.Context(), input)
    if err != nil {
        appErr := app_error.Normalize(err)
        c.JSON(appErr.StatusCode(), appErr.ToResponse(lang))
        return
    }

    c.JSON(http.StatusCreated, SuccessResponse{
        Code:    output.Code,
        Message: h.translator.Translate(lang, output.Code),
    })
}
```

---

## 7. service 設計

### 7.1 配置

```text
api/internal/service/user_registration_service.go
```

---

### 7.2 interface

```go
type UserRegistrationService interface {
    Create(ctx context.Context, input CreateUserRegistrationInput) (*CreateUserRegistrationOutput, error)
}
```

---

### 7.3 input / output

```go
type CreateUserRegistrationInput struct {
    Email             string
    EmailConfirmation string
    Language          string
}

type CreateUserRegistrationOutput struct {
    Code string
}
```

---

### 7.4 service 実装構造

```go
type userRegistrationService struct {
    userRepo                    repository.UserRepository
    userRegistrationRequestRepo repository.UserRegistrationRequestRepository
    txManager                   repository.TxManager
    tokenGenerator              token.Generator
    tokenHasher                 token.Hasher
    uuidGenerator               uuid.Generator
    clock                       clock.Clock
    mailer                      mail.Mailer
    registrationURLBuilder      url.Builder
}
```

---

## 8. service の責務詳細

### 8.1 入力バリデーション

service では業務バリデーションを実施する。

#### チェック内容

- email 必須
- email_confirmation 必須
- email 形式
- email と email_confirmation 一致

#### エラーコード例

- EMAIL_REQUIRED
- EMAIL_CONFIRMATION_REQUIRED
- EMAIL_FORMAT_INVALID
- EMAIL_CONFIRMATION_NOT_MATCH

---

### 8.2 既存ユーザ確認

`users` に同一 email が存在するか確認する。

存在する場合：

- `USER_ALREADY_REGISTERED`
- 409 相当エラー

---

### 8.3 仮登録レコード確認

`user_registration_requests` を email で検索する。

判定パターン：

- 該当なし → 新規作成
- 未認証かつ期限内 → 再送
- 有効期限切れ → 再発行
- 認証済み → 再発行

---

### 8.4 トークン関連

service が以下を実施する。

- 平文トークン生成
- トークンハッシュ化
- 有効期限算出
- メール用URL生成

---

### 8.5 保存処理

service は transaction 内で以下を実施する。

- 新規作成 or 更新
- token_hash 保存
- expires_at 保存
- 必要項目更新

---

### 8.6 メール送信

service は保存成功後に mailer を呼び出す。

送信内容：

- 宛先メールアドレス
- 本登録URL
- 言語
- テンプレート識別子

---

## 9. service 処理手順

1. 入力バリデーション
2. 既存ユーザ確認
3. 仮登録レコード取得
4. 処理区分判定
   - create
   - resend
   - reissue
5. 平文トークン生成
6. トークンハッシュ化
7. UUID / expires_at 決定
8. transaction 開始
9. 新規作成 or 更新
10. transaction commit
11. メール送信
12. 正常結果として `USER_REGISTRATION_REQUEST_CREATED` を返却

---

## 10. service 疑似コード

```go
func (s *userRegistrationService) Create(
    ctx context.Context,
    input CreateUserRegistrationInput,
) (*CreateUserRegistrationOutput, error) {
    if err := s.validate(input); err != nil {
        return nil, err
    }

    exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
    if err != nil {
        return nil, app_error.WrapInternal(err)
    }
    if exists {
        return nil, app_error.NewConflict("USER_ALREADY_REGISTERED")
    }

    now := s.clock.Now()

    req, err := s.userRegistrationRequestRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, app_error.WrapInternal(err)
    }

    plainToken, err := s.tokenGenerator.Generate()
    if err != nil {
        return nil, app_error.WrapInternal(err)
    }

    tokenHash, err := s.tokenHasher.Hash(plainToken)
    if err != nil {
        return nil, app_error.WrapInternal(err)
    }

    expiresAt := now.Add(24 * time.Hour)

    if req == nil {
        id, err := s.uuidGenerator.NewV7()
        if err != nil {
            return nil, app_error.WrapInternal(err)
        }

        newReq := model.UserRegistrationRequest{
            ID:         id,
            Email:      input.Email,
            TokenHash:  tokenHash,
            ExpiresAt:  expiresAt,
            VerifiedAt: nil,
            CreatedAt:  now,
        }

        err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
            return s.userRegistrationRequestRepo.Create(txCtx, newReq)
        })
        if err != nil {
            return nil, app_error.WrapInternal(err)
        }
    } else {
        req.TokenHash = tokenHash
        req.ExpiresAt = expiresAt

        err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
            return s.userRegistrationRequestRepo.UpdateToken(txCtx, *req)
        })
        if err != nil {
            return nil, app_error.WrapInternal(err)
        }
    }

    registerURL := s.registrationURLBuilder.Build(plainToken)

    err = s.mailer.SendUserRegistrationMail(ctx, mail.UserRegistrationMail{
        To:   input.Email,
        URL:  registerURL,
        Lang: input.Language,
    })
    if err != nil {
        return nil, app_error.WrapInternal(err)
    }

    return &CreateUserRegistrationOutput{
        Code: "USER_REGISTRATION_REQUEST_CREATED",
    }, nil
}
```

---

## 11. 境界設計で重要な点

### 11.1 `gin.Context` を service に渡さない

- handler → `context.Context` を渡す
- service は HTTP 非依存とする

---

### 11.2 表示文言は service に持たせない

- service は code を返す
- handler または response builder 側で i18n 解決する

---

### 11.3 validation の分担

- bind可能性チェック → handler
- 業務バリデーション → service

---

### 11.4 transaction 範囲

transaction 範囲は原則 DB更新までとする。  
メール送信は transaction 外で実施する。

理由：

- 外部I/O を transaction に含めない
- lock 長時間化を避ける
- DB整合性と処理性能を両立する

---

## 12. 推奨インターフェース

```go
type UserRepository interface {
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type UserRegistrationRequestRepository interface {
    FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error)
    Create(ctx context.Context, req model.UserRegistrationRequest) error
    UpdateToken(ctx context.Context, req model.UserRegistrationRequest) error
}

type TxManager interface {
    WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type TokenGenerator interface {
    Generate() (string, error)
}

type TokenHasher interface {
    Hash(value string) (string, error)
}

type UUIDGenerator interface {
    NewV7() (string, error)
}

type Clock interface {
    Now() time.Time
}

type Mailer interface {
    SendUserRegistrationMail(ctx context.Context, mail UserRegistrationMail) error
}

type Translator interface {
    Translate(lang string, code string) string
}
```

---

## 13. ディレクトリ案

```text
api/internal/
  handler/
    user_registration_handler.go

  service/
    user_registration_service.go

  repository/
    user_repository.go
    user_registration_request_repository.go
    tx_manager.go

  mail/
    mailer.go

  token/
    generator.go
    hasher.go

  uuid/
    generator.go

  clock/
    clock.go

  i18n/
    translator.go

  app_error/
```

---

## 14. 最終推奨

今回の仮登録APIは以下の責務で進める。

- handler
  - bind
  - language取得
  - service呼び出し
  - i18n解決
  - HTTP返却

- service
  - バリデーション
  - 既存ユーザ判定
  - 仮登録判定
  - トークン生成
  - 保存
  - メール送信
  - code返却

この方針により、正常系・異常系ともに `code + message` に統一できる。
