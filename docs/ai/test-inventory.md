# テスト一覧・カバレッジ分析

仮会員登録フロー全体のテスト一覧と不足分析。

---

## 1. 仮会員登録フォーム

**ファイル:** `frontend/src/pages/UserRegistrationPage.test.jsx`（19件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | タイトルと説明文が表示される | 画面初期表示の静的テキスト |
| 2 | メールアドレスと確認用のinputが表示される | フォームフィールド存在確認 |
| 3 | 送信ボタンが表示される | ボタン存在確認 |
| 4 | ログインリンクが表示される | ログインリンク存在確認 |
| 5 | 空のまま送信するとメールアドレスのエラーが表示される | クライアント空欄バリデーション |
| 6 | 無効なメールアドレスを入力するとエラーが表示される | クライアント形式バリデーション |
| 7 | 確認用メールアドレスが空のときエラーが表示される | 確認用フィールド空欄バリデーション |
| 8 | 確認用メールアドレスが一致しない場合エラーが表示される | 不一致バリデーション |
| 9 | 成功時に `/registration/complete` へ遷移する | API成功→Routes経由で実際にパスが変わることを確認 |
| 10 | 成功時に `state.email` に入力メールアドレスが渡される | 遷移時 `state.email` の値検証 ✅ |
| 11 | 成功時に `state.expiresMinutes` に APIの `expires_minutes` が渡される | `expires_minutes` → `expiresMinutes` camelCase変換の検証 ✅ |
| 12 | `expires_minutes` が 30 の場合 `state.expiresMinutes` が 30 になる | 値が動的に渡されることの検証 ✅ |
| 13 | VALIDATION_ERRORのとき、フィールドエラーが表示される | emailフィールドエラー表示 |
| 14 | その他のAPIエラー時にフォームエラーが表示される | サーバーエラーメッセージ |
| 15 | VALIDATION_ERRORのとき、email_confirmationフィールドエラーが表示される | 確認用フィールドサーバーエラー |
| 16 | VALIDATION_ERRORだが既知フィールドエラーがない場合はフォームエラーが表示される | 未知フィールドのフォールバックエラー |
| 17 | APIエラーにmessageがない場合はデフォルトエラーメッセージが表示される | messageなしフォールバック |
| 18 | 通信エラー時に通信エラーメッセージが表示される | fetch例外ハンドリング |
| 19 | 送信中はボタンが無効化される | ローディング中UI状態 |

**不足している観点:** なし

---

## 2. 完了画面

**ファイル:** `frontend/src/pages/UserRegistrationCompletePage.test.jsx`（12件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | タイトルが表示される | 静的テキスト存在 |
| 2 | ロゴが表示される | ロゴ存在 |
| 3 | ログインボタンは表示されない | ログインボタン除去確認 |
| 4 | メールアドレスが表示される | `state.email` 表示 |
| 5 | 本登録リンク送信メッセージが表示される | 説明文テキスト |
| 6 | expiresMinutesがある場合、有効期限が表示される | `expiresMinutes: 60` → `「60分」` の完全一致 |
| 7 | expiresMinutesがnullの場合、有効期限は表示されない | null時の条件レンダリング |
| 8 | expiresMinutes が 30 の場合、有効期限文言に「30分」が含まれる | 値が動的に埋め込まれることの検証 ✅ |
| 9 | 迷惑メールフォルダの注意書きが表示される | 注意事項テキスト |
| 10 | メイン指示文が表示される | 指示文テキスト |
| 11 | stateがない場合、`/registration` にリダイレクトされる | 直接アクセス→リダイレクト |
| 12 | stateがない場合、完了画面は表示されない | リダイレクト中の描画なし |

**不足している観点:** なし

---

## 3. 仮登録API（ハンドラ層）

**ファイル:** `api/internal/handler/handler_test.go`（27件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestNewUserRegistrationHandler | ハンドラ生成・nilでない |
| 2 | TestUserRegistrationHandlerCreateSuccess | 正常系→201 |
| 3 | TestUserRegistrationHandlerCreateDefaultLanguage | `Accept-Language`なし→デフォルトja |
| 4 | TestUserRegistrationHandlerCreateBadRequest | JSONパースエラー→400 |
| 5 | TestUserRegistrationHandlerCreateValidationError | バリデーションエラー→422 |
| 6 | TestUserRegistrationHandlerCreateConflict | 競合→409 |
| 7 | TestUserRegistrationHandlerCreateInternalError | 内部エラー→500 |
| 8 | TestUserRegistrationHandlerCreateRateLimitIP | IPレート制限→429（code=TOO_MANY_REQUESTS確認） |
| 9 | TestUserRegistrationHandlerCreateRateLimitEmail | メールレート制限→429（code=TOO_MANY_REQUESTS確認） |
| 10 | TestToResponse | CreateUserRegistrationOutput→HTTPレスポンス変換 |
| 11 | TestUserRegistrationHandlerCreateEnglishLanguage | `Accept-Language: en`→英語 |
| 12 | TestUserRegistrationHandlerCreateSuccessResponseBody | レスポンスボディに `expires_minutes` 含む |
| 13 | TestUserRegistrationHandlerCreateBodyUnderSizeLimit | 1MB未満→受け付ける |
| 14 | TestUserRegistrationHandlerCreateBodyExceedsSizeLimit | 1MB超→400 |
| 15 | TestUserRegistrationHandlerVerifySuccess | 正常系→201（クエリtokenがserviceに渡されることも検証） |
| 16 | TestUserRegistrationHandlerVerifyMissingQueryToken | クエリtokenなし→400 INVALID_REGISTRATION_TOKEN ✅ |
| 17 | TestUserRegistrationHandlerVerifyBadRequest | JSONパースエラー→400 |
| 18 | TestUserRegistrationHandlerVerifyTokenInvalid | 不正トークン→400 |
| 19 | TestUserRegistrationHandlerVerifyTokenExpired | 期限切れトークン→400 |
| 20 | TestUserRegistrationHandlerVerifyAlreadyVerified | トークン使用済み→409 |
| 21 | TestUserRegistrationHandlerVerifyUserAlreadyRegistered | メール重複→409 |
| 22 | TestUserRegistrationHandlerVerifyValidationError | バリデーションエラー→422 |
| 23 | TestUserRegistrationHandlerVerifyInternalError | 内部エラー→500 |
| 24 | TestUserRegistrationHandlerVerifyRateLimitIP | IPレート制限→429（code=TOO_MANY_REQUESTS確認） |
| 25 | TestUserRegistrationHandlerVerifyBodyExceedsSizeLimit | 1MB超→400 |
| 26 | TestUserRegistrationHandlerVerifyDefaultLanguage | `Accept-Language`なし→デフォルトja |
| 27 | TestNormalizeLanguage | 言語コード正規化（ja-JP→ja, en-US→en等） |

**不足している観点:** なし

---

## 4. レート制限

**ファイル:** `api/internal/handler/rate_limiter_test.go`（4件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestRateLimiterAllowIP | IPベース制限の許可/拒否 |
| 2 | TestRateLimiterAllowEmail | メールベース制限（正規化+SHA256キー） |
| 3 | TestRateLimiterAllowWhenLimitIsZero | limit=0のとき常に許可 |
| 4 | TestRateLimiterAllowWhenStoreError | Redisエラー→フェイルクローズド |

**不足している観点:** なし

---

## 5. トークン生成

**ファイル:** `api/internal/token/token_test.go`（3件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestDefaultGeneratorGenerate | 32バイト、base64、空でない |
| 2 | TestDefaultGeneratorGenerateError | `rand.Read` エラー時のエラー返却 |
| 3 | TestSHA256HasherHash | 既知値でのSHA256ハッシュ値検証 |

**不足している観点:** トークンの一意性（乱数依存のため省略可が妥当）

---

## 6. registration URL生成

**ファイル:** `api/internal/registrationurl/registrationurl_test.go`（5件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestStaticBuilderBuild | 正しいURL形式（ベースURL + パス + トークン） |
| 2 | TestNewStaticBuilder | コンストラクタ・nilでない |
| 3 | TestStaticBuilderBuild_TokenAppearsInURL | 生のトークンがURLに含まれる ✅ |
| 4 | TestStaticBuilderBuild_URLNotEndingWithTokenEquals | URLが `token=` で終わらない ✅ |
| 5 | TestStaticBuilderBuild_EmptyTokenProducesEmptySuffix | 空トークン→`token=`のみ（ドキュメント目的） |

**不足している観点:** なし

---

## 7. サービス層（token hash保存・mail_outbox payload）

**ファイル:** `api/internal/service/user_registration_service_test.go`（57件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestCreate | 正常系・ExpiresMinutes=60返却 |
| 2 | TestCreate_EmailEmpty | メール空→エラー |
| 3 | TestCreate_EmailFormatInvalid | 形式不正→エラー |
| 4 | TestCreate_EmailConfirmationNotMatch | 確認不一致→エラー |
| 5 | TestCreate_EmailConfirmationEmpty | 確認空→エラー |
| 6 | TestCreate_AlreadyVerified | 既に本登録済み→正常返却 |
| 7 | TestCreate_ResendNotAvailable | 再送間隔内→正常返却 |
| 8 | TestCreate_UpdateToken | 期限切れ→トークン更新 |
| 9 | TestCreate_DBError | DB検索エラー |
| 10 | TestCreate_TokenGenError | トークン生成エラー |
| 11 | TestCreate_TokenHashError | ハッシュエラー |
| 12 | TestCreate_FirstUUIDError | 1回目UUID生成エラー |
| 13 | TestCreate_SecondUUIDError | 2回目UUID生成エラー |
| 14 | TestCreate_CreateUserError | DB登録エラー |
| 15 | TestCreate_UpdateTokenError | トークン更新エラー |
| 16 | TestCreate_MarshalError | JSON marshalエラー |
| 17 | TestCreate_EmailNormalized | メール小文字正規化 |
| 18 | TestCreate_RawTokenPassedToURLBuilder | URL Builderに生トークンが渡される ✅ |
| 19 | TestCreate_DBStoresHashNotRawToken | DBにはハッシュのみ保存 ✅ |
| 20 | TestCreate_OutboxPayloadContainsTokenURL | outboxペイロードURLに生トークン含む ✅ |
| 21 | TestCreate_OutboxPayloadURLNotEmptyToken | outboxペイロードURLが `token=` で終わらない ✅ |
| 22 | TestCreate_EmptyTokenFails | 空トークン→エラー ✅ |
| 23 | TestCreate_OutboxMailTypeIsUserRegistration | outbox.mail_type = "user_registration" ✅ |
| 24 | TestCreate_OutboxStatusIsPending | outbox.status = "pending" ✅ |
| 25 | TestCreate_OutboxNextAttemptAtIsNow | outbox.next_attempt_at = 現在時刻 ✅ |
| 26 | TestCreate_AlreadyVerified_IgnoresTokenGenError | 認証済みユーザ→トークン生成エラーでも正常返却 ✅ |
| 27 | TestCreate_ResendNotAvailable_IgnoresTokenGenError | 再送間隔内→トークン生成エラーでも正常返却 ✅ |
| 28 | TestCreate_ConcurrentDuplicateEmail | 並行INSERT重複→ErrDuplicateEmail→正常返却 ✅ |
| 29 | TestVerify_Validation_DisplayNameRequired | display_name 必須エラー |
| 30 | TestVerify_Validation_DisplayNameWhitespace | display_name 空白のみ→エラー |
| 31 | TestVerify_Validation_PasswordRequired | password 必須エラー |
| 32 | TestVerify_Validation_PasswordConfirmationRequired | password_confirmation 必須エラー |
| 33 | TestVerify_Validation_PasswordConfirmationNotMatch | パスワード不一致エラー |
| 34 | TestVerify_Validation_PasswordTooShort | 8文字未満→PASSWORD_TOO_WEAK |
| 35 | TestVerify_Validation_PasswordNoLetter | 英字なし→PASSWORD_TOO_WEAK |
| 36 | TestVerify_Validation_PasswordNoDigit | 数字なし→PASSWORD_TOO_WEAK |
| 37 | TestVerify_Validation_PasswordMinLength | ちょうど8文字・英字+数字→合格 |
| 38 | TestVerify_Validation_AgreedToTermsRequired | agreed_to_terms=false→エラー |
| 39 | TestVerify | 正常系・全テーブルINSERT・verified_at更新 |
| 40 | TestVerify_EmptyToken | 空トークン→INVALID_REGISTRATION_TOKEN |
| 41 | TestVerify_TokenHashError | SHA-256ハッシュエラー→500 |
| 42 | TestVerify_PasswordHashError | bcryptエラー→500（tx内・全検証後） |
| 43 | TestVerify_TokenNotFound | トークン未存在→INVALID_REGISTRATION_TOKEN |
| 44 | TestVerify_AlreadyVerified | verified_at設定済み→USED_REGISTRATION_TOKEN |
| 45 | TestVerify_TokenExpired | expires_at超過→EXPIRED_REGISTRATION_TOKEN |
| 46 | TestVerify_EmailAlreadyExists | メール重複→USER_ALREADY_REGISTERED |
| 47 | TestVerify_FindEmailError | メール検索DBエラー→500 |
| 48 | TestVerify_UserCreateError | users INSERTエラー→500 |
| 49 | TestVerify_UserEmailCreateError | user_emails INSERTエラー→500 |
| 50 | TestVerify_UserCredentialCreateError | user_credentials INSERTエラー→500 |
| 51 | TestVerify_UpdateVerifiedAtError | verified_at UPDATEエラー→500 |
| 52 | TestVerify_DBError | txエラー→500 |
| 53 | TestVerify_FirstUUIDError | 1回目UUID生成エラー→500 |
| 54 | TestVerify_SecondUUIDError | 2回目UUID生成エラー→500 |
| 55 | TestVerify_DisplayNameTrimmed | display_name 前後空白トリム確認 |
| 56 | TestVerify_BcryptSkippedForInvalidToken | 不正トークン→bcrypt実行されない ✅ |
| 57 | TestHashPasswordBcryptError | bcryptGenerate エラー→エラー返却 |

**不足している観点:** なし

---

## 8. SMTPメール本文生成

**ファイル:** `api/internal/mail/smtp_mailer_test.go`（14件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestNewSMTPMailer | コンストラクタ（Host/Port/From） |
| 2 | TestSMTPMailerSendUserRegistrationMail | ヘッダ（From/To/Subject存在）・アドレス |
| 3 | TestSMTPMailerSendUserRegistrationMailWithAuth | User設定時にauth非nil |
| 4 | TestSMTPMailerSendUserRegistrationMailWithTLS | TLS接続・アドレス・auth |
| 5 | TestSMTPMailerSendUserRegistrationMailWithTLSNoAuth | TLS + User未設定→auth=nil |
| 6 | TestSMTPMailerSendUserRegistrationMailSendError | 送信エラーの伝播 |
| 7 | TestSMTPMailerSendUserRegistrationMailTemplateParseError | テンプレートパースエラー |
| 8 | TestSMTPMailerSendUserRegistrationMailTemplateExecuteError | テンプレート実行エラー |
| 9 | TestSMTPMailerSendUserRegistrationMail_URLWithTokenInBody | メール本文にトークン付きURL含む ✅ |
| 10 | TestSMTPMailerSendUserRegistrationMail_EmptyURLReturnsError | URL空→エラー ✅ |
| 11 | TestSMTPMailerSendUserRegistrationMailContentTypeHeader | `Content-Type: text/plain; charset=UTF-8` ヘッダ含む ✅ |
| 12 | TestSMTPMailerSendUserRegistrationMailJapaneseSubject | `Subject: ユーザ仮登録のご案内`（ja）含む ✅ |
| 13 | TestSMTPMailerSendUserAlreadyRegisteredMail | 既登録メール本文 |
| 14 | TestSMTPMailerSendUserAlreadyRegisteredMailSendError | 既登録メール送信エラー |

**不足している観点:** なし

---

## 9. DI・結合

**ファイル:** `api/internal/app/app_test.go`（4件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestNewUserRegistrationService | DIワイヤリング |
| 2 | TestNewUserRegistrationHandler | ハンドラワイヤリング |
| 3 | TestUserRegistrationHandlerCreate | sqlmock使用の統合テスト（仮登録 DB→HTTP） |
| 4 | TestUserRegistrationHandlerVerify | sqlmock使用の統合テスト（本登録 DB→HTTP） |

---

## 10. Noop Mailer

**ファイル:** `api/internal/mail/noop_mailer_test.go`（2件）

| # | テスト名 | 検証内容 |
|---|---------|---------|
| 1 | TestNoopMailerSendUserRegistrationMail | nilを返す |
| 2 | TestNoopMailerSendUserAlreadyRegisteredMail | nilを返す |

---

## 6つの重点観点 — カバレッジ確認

| 観点 | 担当テスト | 状態 |
|------|-----------|------|
| URLが `token=` で終わらない | `TestStaticBuilderBuild_URLNotEndingWithTokenEquals`<br>`TestCreate_OutboxPayloadURLNotEmptyToken`<br>`TestSMTPMailerSendUserRegistrationMail_URLWithTokenInBody` | ✅ |
| 生のトークンがURLに含まれる | `TestStaticBuilderBuild_TokenAppearsInURL`<br>`TestCreate_RawTokenPassedToURLBuilder`<br>`TestCreate_OutboxPayloadContainsTokenURL` | ✅ |
| DBにはハッシュのみ保存 | `TestCreate_DBStoresHashNotRawToken` | ✅ |
| outbox payloadにトークン付きURL | `TestCreate_OutboxPayloadContainsTokenURL`<br>`TestCreate_OutboxPayloadURLNotEmptyToken` | ✅ |
| SMTPメール本文にトークン付きURL | `TestSMTPMailerSendUserRegistrationMail_URLWithTokenInBody` | ✅ |
| 空トークンの防止 | `TestCreate_EmptyTokenFails`（service層）<br>`TestSMTPMailerSendUserRegistrationMail_EmptyURLReturnsError`（mailer層） | ✅ |

---

## 追加すべきテスト（優先度順）

| 優先度 | 対象 | テスト | 理由 |
|--------|------|---------|------|
| — | — | — | 現時点で追加すべきテストなし |
