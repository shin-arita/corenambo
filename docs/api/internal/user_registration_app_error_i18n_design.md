# ユーザ仮登録 app_error / i18n 設計

## 1. 概要

本設計は、ユーザ仮登録機能における app_error / i18n の責務分割を定義する。  
APIレスポンスを `code + message` で統一し、表示文言とエラー構造を分離することで、多言語対応・保守性・一貫性を高めることを目的とする。

---

## 2. 全体方針

本機能では以下の方針で設計する。

- エラー判定は `code` で行う
- 表示文言は i18n で解決する
- 正常系・異常系ともに `code + message` で統一する
- service は表示文言を直接持たない
- handler は `code` を受け取り、最終レスポンスへ変換する
- 項目エラーは `errors` で返却する

---

## 3. 役割分担

## 3.1 app_error の責務

app_error は、APIで扱うアプリケーションエラーの構造を統一する。

### 担当すること

- エラーコード保持
- HTTPステータス保持
- 項目エラー保持
- 内部エラーのラップ
- レスポンス変換用データの保持

### 担当しないこと

- 文言のベタ書き
- 翻訳辞書の保持
- 言語ファイルの管理
- 業務ルール判定

---

## 3.2 i18n の責務

i18n は、code から表示文言を解決する。

### 担当すること

- `code -> message` の変換
- 言語ごとの辞書管理
- デフォルト言語フォールバック
- 存在しない code の扱い

### 担当しないこと

- HTTPステータス管理
- エラー構造管理
- DBアクセス
- 業務判定

---

## 4. 依存関係

```text
Handler
  ↓
Service
  ↓
app_error
  ↓
i18n
```

実際の責務としては以下。

- service
  - `app_error` を返す
- handler
  - `app_error` を HTTPレスポンスへ変換する
- i18n
  - code に対応する message を解決する

---

## 5. レスポンス統一方針

## 5.1 正常系

```json
{
  "code": "USER_REGISTRATION_REQUEST_CREATED",
  "message": "仮登録メールを送信しました。メールをご確認ください。"
}
```

---

## 5.2 異常系

```json
{
  "code": "USER_ALREADY_REGISTERED",
  "message": "入力されたメールアドレスは既に登録されています。"
}
```

---

## 5.3 バリデーションエラー

```json
{
  "code": "VALIDATION_ERROR",
  "message": "入力内容に誤りがあります。",
  "errors": {
    "email": [
      {
        "code": "EMAIL_REQUIRED",
        "message": "メールアドレスを入力してください。"
      }
    ],
    "email_confirmation": [
      {
        "code": "EMAIL_CONFIRMATION_NOT_MATCH",
        "message": "メールアドレスが一致しません。"
      }
    ]
  }
}
```

---

## 6. app_error 設計

## 6.1 配置

```text
api/internal/app_error/
  app_error.go
  builder.go
  normalize.go
  validation_error.go
```

---

## 6.2 基本構造

```go
type AppError struct {
    Code       string
    Status     int
    FieldErrors map[string][]FieldError
    Cause      error
}
```

---

## 6.3 FieldError

```go
type FieldError struct {
    Code string
}
```

### 方針

- FieldError には原則 message を持たせない
- レスポンス変換時に i18n で解決する
- 内部表現は code 中心とする

---

## 6.4 AppError interface 案

```go
type Error interface {
    error
    Code() string
    StatusCode() int
    FieldErrors() map[string][]FieldError
    CauseError() error
}
```

---

## 6.5 AppError の役割

AppError は以下を統一する。

- 異常系コード
- HTTPステータス
- 項目エラー一覧
- 内部原因エラー

---

## 6.6 AppError 生成関数案

```go
func NewBadRequest(code string) *AppError
func NewConflict(code string) *AppError
func NewInternal(code string, cause error) *AppError
func NewValidation(fieldErrors map[string][]FieldError) *AppError
func WrapInternal(err error) *AppError
```

---

## 6.7 ValidationError の扱い

バリデーションエラーは `VALIDATION_ERROR` を親コードとして返す。

### 例

- 親
  - `VALIDATION_ERROR`
- 子
  - `EMAIL_REQUIRED`
  - `EMAIL_FORMAT_INVALID`
  - `EMAIL_CONFIRMATION_REQUIRED`
  - `EMAIL_CONFIRMATION_NOT_MATCH`

---

## 6.8 HTTPステータス対応

| code | status |
|---|---:|
| BAD_REQUEST | 400 |
| VALIDATION_ERROR | 422 |
| USER_ALREADY_REGISTERED | 409 |
| INTERNAL_SERVER_ERROR | 500 |

### 備考

- 項目個別コードは `VALIDATION_ERROR` 配下で返す
- HTTPステータスは親エラーで決定する

---

## 6.9 Normalize 方針

service や repository から返る error を handler で直接判定しないため、`Normalize` を用意する。

```go
func Normalize(err error) *AppError
```

### 役割

- `AppError` の場合はそのまま返す
- それ以外は `INTERNAL_SERVER_ERROR` に変換する

---

## 7. i18n 設計

## 7.1 配置

```text
api/internal/i18n/
  translator.go
  ja.go
  en.go
```

またはファイルベースにする場合：

```text
api/internal/i18n/
  translator.go
  locales/
    ja.json
    en.json
```

---

## 7.2 Translator interface

```go
type Translator interface {
    Translate(lang string, code string) string
}
```

---

## 7.3 基本方針

- code をキーに message を返す
- `lang` が未指定または未対応の場合は `ja`
- code 未定義時は安全なフォールバック文言を返す

---

## 7.4 Translate 動作方針

### 例

```go
msg := translator.Translate("ja", "EMAIL_REQUIRED")
```

結果：

```text
メールアドレスを入力してください。
```

---

## 7.5 フォールバック方針

### 言語フォールバック

- 指定言語が未対応 → `ja`

### code フォールバック

- code が未登録 → code 自体または汎用文言

推奨：

- 開発環境: code を返して気づけるようにする
- 本番環境: 汎用文言を返す

本設計では単純化のため以下を推奨する。

- 未定義 code は code 文字列を返す

---

## 7.6 管理対象コード例

### 正常系

| code | ja | en |
|---|---|---|
| USER_REGISTRATION_REQUEST_CREATED | 仮登録メールを送信しました。メールをご確認ください。 | A temporary registration email has been sent. Please check your email. |

### 共通エラー

| code | ja | en |
|---|---|---|
| BAD_REQUEST | リクエストが不正です。 | The request is invalid. |
| VALIDATION_ERROR | 入力内容に誤りがあります。 | There are errors in the input. |
| USER_ALREADY_REGISTERED | 入力されたメールアドレスは既に登録されています。 | The entered email address is already registered. |
| INTERNAL_SERVER_ERROR | システムエラーが発生しました。 | A system error has occurred. |

### 項目エラー

| code | ja | en |
|---|---|---|
| EMAIL_REQUIRED | メールアドレスを入力してください。 | Please enter your email address. |
| EMAIL_FORMAT_INVALID | 正しいメールアドレス形式で入力してください。 | Please enter a valid email address. |
| EMAIL_CONFIRMATION_REQUIRED | メールアドレスを入力してください。 | Please enter the email confirmation. |
| EMAIL_CONFIRMATION_NOT_MATCH | メールアドレスが一致しません。 | Email addresses do not match. |

---

## 8. レスポンス変換設計

## 8.1 正常レスポンス DTO

```go
type SuccessResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

---

## 8.2 異常レスポンス DTO

```go
type ErrorResponse struct {
    Code    string                         `json:"code"`
    Message string                         `json:"message"`
    Errors  map[string][]FieldErrorResponse `json:"errors,omitempty"`
}
```

---

## 8.3 FieldErrorResponse

```go
type FieldErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

---

## 8.4 AppError → Response 変換

```go
func (e *AppError) ToResponse(lang string, t i18n.Translator) ErrorResponse
```

### 役割

- 親 `code` を `message` に変換する
- 各 FieldError の `code` を `message` に変換する
- `ErrorResponse` を返す

---

## 8.5 変換イメージ

```go
func (e *AppError) ToResponse(lang string, t i18n.Translator) ErrorResponse {
    res := ErrorResponse{
        Code:    e.Code,
        Message: t.Translate(lang, e.Code),
    }

    if len(e.FieldErrors) > 0 {
        res.Errors = map[string][]FieldErrorResponse{}
        for field, errs := range e.FieldErrors {
            for _, fe := range errs {
                res.Errors[field] = append(res.Errors[field], FieldErrorResponse{
                    Code:    fe.Code,
                    Message: t.Translate(lang, fe.Code),
                })
            }
        }
    }

    return res
}
```

---

## 9. service との境界

## 9.1 service は何を返すか

service は以下を返す。

- 正常時
  - 成功 code
- 異常時
  - `AppError`

### 例

```go
return &CreateUserRegistrationOutput{
    Code: "USER_REGISTRATION_REQUEST_CREATED",
}, nil
```

```go
return nil, app_error.NewConflict("USER_ALREADY_REGISTERED")
```

---

## 9.2 service がやらないこと

- message の組み立て
- 日本語文言の返却
- 英語文言の返却
- HTTPレスポンス生成

---

## 10. handler との境界

## 10.1 handler の責務

- `Accept-Language` 取得
- translator による code -> message 解決
- 正常レスポンス生成
- AppError を ErrorResponse に変換して返却

---

## 10.2 handler 疑似コード

```go
func (h *UserRegistrationHandler) Create(c *gin.Context) {
    lang := c.GetHeader("Accept-Language")
    if lang == "" {
        lang = "ja"
    }

    var req CreateUserRegistrationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := app_error.NewBadRequest("BAD_REQUEST")
        c.JSON(appErr.StatusCode(), appErr.ToResponse(lang, h.translator))
        return
    }

    output, err := h.userRegistrationService.Create(c.Request.Context(), service.CreateUserRegistrationInput{
        Email:             req.Email,
        EmailConfirmation: req.EmailConfirmation,
        Language:          lang,
    })
    if err != nil {
        appErr := app_error.Normalize(err)
        c.JSON(appErr.StatusCode(), appErr.ToResponse(lang, h.translator))
        return
    }

    c.JSON(http.StatusCreated, SuccessResponse{
        Code:    output.Code,
        Message: h.translator.Translate(lang, output.Code),
    })
}
```

---

## 11. バリデーションエラー設計詳細

## 11.1 返却方針

- 親 `code`: `VALIDATION_ERROR`
- `errors` に項目別詳細を格納する

---

## 11.2 例

```go
fieldErrors := map[string][]app_error.FieldError{
    "email": {
        {Code: "EMAIL_REQUIRED"},
    },
    "email_confirmation": {
        {Code: "EMAIL_CONFIRMATION_NOT_MATCH"},
    },
}

return nil, app_error.NewValidation(fieldErrors)
```

---

## 11.3 メリット

- フロントで field ごとの制御がしやすい
- 文言変更しても code は安定する
- 多言語切り替えに強い

---

## 12. ログ・監視方針

## 12.1 表示用 message をログ基準にしない

ログ・監視は `code` 基準とする。

### 理由

- 文言変更の影響を受けない
- 多言語化の影響を受けない
- 集計しやすい

---

## 12.2 Cause の扱い

- `Cause` は内部ログ用とする
- APIレスポンスには含めない

---

## 13. テスト観点

## 13.1 app_error

- code と status が正しく保持される
- validation errors が正しく保持される
- Normalize が AppError / 非AppError を正しく処理する
- ToResponse で code -> message が正しく変換される

---

## 13.2 i18n

- ja で正しい文言が返る
- en で正しい文言が返る
- 未対応 lang で `ja` にフォールバックする
- 未定義 code でフォールバック動作する

---

## 13.3 handler 連携

- 正常時に success response が返る
- 異常時に error response が返る
- field errors がレスポンスに含まれる
- Accept-Language に応じて文言が切り替わる

---

## 14. ディレクトリ案

```text
api/internal/
  app_error/
    app_error.go
    builder.go
    normalize.go
    validation_error.go

  i18n/
    translator.go
    ja.go
    en.go
```

---

## 15. 最終推奨

今回の仮登録機能は以下で進める。

- app_error
  - エラー構造を統一する
  - HTTPステータスと code を持つ
  - field errors を保持する

- i18n
  - code から message を解決する
  - ja / en を管理する
  - フォールバックを提供する

- 正常系 / 異常系
  - どちらも `code + message` で統一する

この方針により、表示文言と業務ロジックを分離しつつ、実装・テスト・多言語対応を安定させることができる。
