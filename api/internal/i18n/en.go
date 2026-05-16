package i18n

var enMessages = map[string]string{
	CodeUserRegistrationRequestCreated: "A temporary registration email has been sent. Please check your email",
	CodeUserRegistrationVerified:       "Registration is complete",

	CodeBadRequest:          "The request is invalid",
	CodeValidationError:     "There are errors in the input",
	CodeTooManyRequests:     "Too many requests. Please wait and try again",
	CodeInternalServerError: "A system error has occurred",

	CodeUserAlreadyRegistered: "The entered email address is already registered",

	CodeRegistrationTokenValid:   "Registration token is valid",
	CodeInvalidRegistrationToken: "The token is invalid",
	CodeExpiredRegistrationToken: "The token has expired",
	CodeUsedRegistrationToken:    "Registration is already completed",

	CodeEmailRequired:             "Please enter your email address",
	CodeEmailFormatInvalid:        "Please enter a valid email address",
	CodeEmailConfirmationRequired: "Please enter the email confirmation",
	CodeEmailConfirmationNotMatch: "Email addresses do not match",

	CodeDisplayNameRequired:          "Please enter your username",
	CodePasswordRequired:             "Please enter your password",
	CodePasswordTooWeak:              "Password must be at least 8 characters long and include at least one letter and one number",
	CodePasswordConfirmationRequired: "Please enter the password confirmation",
	CodePasswordConfirmationNotMatch: "Passwords do not match",
	CodeAgreedToTermsRequired:        "You must agree to the terms of service",

	CodeMailUserRegistrationSubject: "User Registration",

	CodeMailUserRegistrationBody: `Thank you for using Corenambo Auction

Please click the URL below to complete your registration

{{.URL}}

* This URL will expire in {{.ExpiresMinutes}} minutes
* If you did not request this, please ignore this email`,

	CodeMailUserAlreadyRegisteredSubject: "Information",

	CodeMailUserAlreadyRegisteredBody: `This email address is already registered

Please use the login page

If you did not request this, please ignore this email`,
}
