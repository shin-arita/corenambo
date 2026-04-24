package i18n

var jaMessages = map[string]string{
	CodeUserRegistrationRequestCreated: "仮登録メールを送信しました。メールをご確認ください。",

	CodeBadRequest:          "リクエストが不正です。",
	CodeValidationError:     "入力内容に誤りがあります。",
	CodeInternalServerError: "システムエラーが発生しました。",

	CodeUserAlreadyRegistered: "入力されたメールアドレスは既に登録されています。",

	CodeEmailRequired:             "メールアドレスを入力してください。",
	CodeEmailFormatInvalid:        "正しいメールアドレス形式で入力してください。",
	CodeEmailConfirmationRequired: "メールアドレスを入力してください。",
	CodeEmailConfirmationNotMatch: "メールアドレスが一致しません。",
}
