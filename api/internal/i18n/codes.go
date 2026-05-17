package i18n

const (
	CodeUserRegistrationRequestCreated = "USER_REGISTRATION_REQUEST_CREATED"
	CodeUserRegistrationVerified       = "USER_REGISTRATION_VERIFIED"

	CodeBadRequest          = "BAD_REQUEST"
	CodeValidationError     = "VALIDATION_ERROR"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"

	CodeUserAlreadyRegistered = "USER_ALREADY_REGISTERED"

	CodeRegistrationTokenValid   = "REGISTRATION_TOKEN_VALID"
	CodeInvalidRegistrationToken = "INVALID_REGISTRATION_TOKEN"
	CodeExpiredRegistrationToken = "EXPIRED_REGISTRATION_TOKEN"
	CodeUsedRegistrationToken    = "USED_REGISTRATION_TOKEN"

	CodeEmailRequired             = "EMAIL_REQUIRED"
	CodeEmailFormatInvalid        = "EMAIL_FORMAT_INVALID"
	CodeEmailConfirmationRequired = "EMAIL_CONFIRMATION_REQUIRED"
	CodeEmailConfirmationNotMatch = "EMAIL_CONFIRMATION_NOT_MATCH"

	CodeDisplayNameRequired          = "DISPLAY_NAME_REQUIRED"
	CodeDisplayNameTooShort          = "DISPLAY_NAME_TOO_SHORT"
	CodeDisplayNameTooLong           = "DISPLAY_NAME_TOO_LONG"
	CodeDisplayNameControlChar       = "DISPLAY_NAME_CONTROL_CHAR"
	CodeDisplayNameZeroWidth         = "DISPLAY_NAME_ZERO_WIDTH"
	CodeDisplayNameReserved          = "DISPLAY_NAME_RESERVED"
	CodePasswordRequired             = "PASSWORD_REQUIRED"
	CodePasswordTooLong              = "PASSWORD_TOO_LONG"
	CodePasswordTooWeak              = "PASSWORD_TOO_WEAK"
	CodePasswordConfirmationRequired = "PASSWORD_CONFIRMATION_REQUIRED"
	CodePasswordConfirmationNotMatch = "PASSWORD_CONFIRMATION_NOT_MATCH"
	CodeAgreedToTermsRequired        = "AGREED_TO_TERMS_REQUIRED"

	CodeMailUserRegistrationSubject = "MAIL_USER_REGISTRATION_SUBJECT"
	CodeMailUserRegistrationBody    = "MAIL_USER_REGISTRATION_BODY"

	CodeMailUserAlreadyRegisteredSubject = "MAIL_USER_ALREADY_REGISTERED_SUBJECT"
	CodeMailUserAlreadyRegisteredBody    = "MAIL_USER_ALREADY_REGISTERED_BODY"
)
