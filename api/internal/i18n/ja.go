package i18n

var jaMessages = map[string]string{
	CodeUserRegistrationRequestCreated: "仮登録メールを送信しました。メールをご確認ください",
	CodeUserRegistrationVerified:       "本登録が完了しました",

	CodeBadRequest:          "リクエストが不正です",
	CodeValidationError:     "入力内容に誤りがあります",
	CodeTooManyRequests:     "リクエストが多すぎます。しばらく待ってから再試行してください",
	CodeInternalServerError: "システムエラーが発生しました",

	CodeUserAlreadyRegistered: "入力されたメールアドレスは既に登録されています",

	CodeRegistrationTokenValid:   "本登録トークンは有効です",
	CodeInvalidRegistrationToken: "トークンが不正です",
	CodeExpiredRegistrationToken: "トークンの有効期限が切れています",
	CodeUsedRegistrationToken:    "既に登録が完了しています",

	CodeEmailRequired:             "メールアドレスを入力してください",
	CodeEmailFormatInvalid:        "正しいメールアドレス形式で入力してください",
	CodeEmailConfirmationRequired: "確認用メールアドレスを入力してください",
	CodeEmailConfirmationNotMatch: "メールアドレスが一致しません",

	CodeDisplayNameRequired:          "ユーザ名を入力してください",
	CodeDisplayNameTooShort:          "ユーザ名は3文字以上で入力してください",
	CodeDisplayNameTooLong:           "ユーザ名は30文字以内で入力してください",
	CodeDisplayNameControlChar:       "ユーザ名に使用できない文字が含まれています",
	CodeDisplayNameZeroWidth:         "ユーザ名に使用できない文字が含まれています",
	CodeDisplayNameReserved:          "このユーザ名は使用できません",
	CodePasswordRequired:             "パスワードを入力してください",
	CodePasswordTooLong:              "パスワードは72文字以内で入力してください",
	CodePasswordTooWeak:              "パスワードは8文字以上で、英字と数字をそれぞれ1文字以上含めてください",
	CodePasswordConfirmationRequired: "確認用パスワードを入力してください",
	CodePasswordConfirmationNotMatch: "パスワードが一致しません",
	CodeAgreedToTermsRequired:        "利用規約への同意が必要です",

	CodeMailUserRegistrationSubject: "ユーザ仮登録のご案内",

	CodeMailUserRegistrationBody: `コレナンボ オークションをご利用いただきありがとうございます

以下のURLをクリックして、本登録を完了してください

{{.URL}}

※このURLの有効期限は{{.ExpiresMinutes}}分です
※本メールに心当たりがない場合は破棄してください`,

	CodeMailUserAlreadyRegisteredSubject: "ご案内",

	CodeMailUserAlreadyRegisteredBody: `このメールアドレスは既に登録されています

ログインページをご利用ください

※本メールに心当たりがない場合は破棄してください`,
}
