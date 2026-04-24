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
}
