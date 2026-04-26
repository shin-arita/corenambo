package i18n

var enMessages = map[string]string{
	CodeUserRegistrationRequestCreated: "A temporary registration email has been sent. Please check your email.",

	CodeBadRequest:          "The request is invalid.",
	CodeValidationError:     "There are errors in the input.",
	CodeInternalServerError: "A system error has occurred.",

	CodeUserAlreadyRegistered: "The entered email address is already registered.",

	CodeEmailRequired:             "Please enter your email address.",
	CodeEmailFormatInvalid:        "Please enter a valid email address.",
	CodeEmailConfirmationRequired: "Please enter the email confirmation.",
	CodeEmailConfirmationNotMatch: "Email addresses do not match.",

	CodeTokenInvalid:                    "The token is invalid.",
	CodeTokenExpired:                    "The token has expired.",
	CodeUserRegistrationAlreadyVerified: "Registration is already completed.",

	CodeMailUserRegistrationSubject: "User Registration",

	CodeMailUserRegistrationBody: `Thank you for using Corenambo Auction.

Please click the URL below to complete your registration.

{{.URL}}

* This URL will expire in {{.ExpiresMinutes}} minutes.
* If you did not request this, please ignore this email.`,

	CodeMailUserAlreadyRegisteredSubject: "Information",

	CodeMailUserAlreadyRegisteredBody: `This email address is already registered.

Please use the login page.

If you did not request this, please ignore this email.`,
}
