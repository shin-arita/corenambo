package i18n

import "testing"

func TestTranslatorTranslateJA(t *testing.T) {
	tl := NewTranslator()

	msg := tl.Translate("ja", CodeBadRequest)

	if msg == "" {
		t.Fatal("empty message")
	}
}

func TestTranslatorTranslateEN(t *testing.T) {
	tl := NewTranslator()

	msg := tl.Translate("en", CodeBadRequest)

	if msg == "" {
		t.Fatal("empty message")
	}
}

func TestTranslatorFallback(t *testing.T) {
	tl := NewTranslator()

	msg := tl.Translate("xx", CodeBadRequest)

	if msg == "" {
		t.Fatal("fallback failed")
	}
}

func TestTranslatorUnknownCode(t *testing.T) {
	tl := NewTranslator()

	msg := tl.Translate("ja", "UNKNOWN_CODE")

	if msg != "UNKNOWN_CODE" {
		t.Fatal("should return code")
	}
}

func TestTranslateVerifyCodesJA(t *testing.T) {
	tl := NewTranslator()
	codes := []string{
		CodeUserRegistrationVerified,
		CodeInvalidRegistrationToken,
		CodeExpiredRegistrationToken,
		CodeUsedRegistrationToken,
		CodeDisplayNameRequired,
		CodeDisplayNameTooShort,
		CodeDisplayNameTooLong,
		CodeDisplayNameControlChar,
		CodeDisplayNameZeroWidth,
		CodeDisplayNameReserved,
		CodePasswordRequired,
		CodePasswordTooWeak,
		CodePasswordConfirmationRequired,
		CodePasswordConfirmationNotMatch,
		CodeAgreedToTermsRequired,
		CodeTooManyRequests,
	}
	for _, code := range codes {
		msg := tl.Translate("ja", code)
		if msg == "" || msg == code {
			t.Fatalf("missing ja message for code %q", code)
		}
	}
}

func TestTranslateVerifyCodesEN(t *testing.T) {
	tl := NewTranslator()
	codes := []string{
		CodeUserRegistrationVerified,
		CodeInvalidRegistrationToken,
		CodeExpiredRegistrationToken,
		CodeUsedRegistrationToken,
		CodeDisplayNameRequired,
		CodeDisplayNameTooShort,
		CodeDisplayNameTooLong,
		CodeDisplayNameControlChar,
		CodeDisplayNameZeroWidth,
		CodeDisplayNameReserved,
		CodePasswordRequired,
		CodePasswordTooWeak,
		CodePasswordConfirmationRequired,
		CodePasswordConfirmationNotMatch,
		CodeAgreedToTermsRequired,
		CodeTooManyRequests,
	}
	for _, code := range codes {
		msg := tl.Translate("en", code)
		if msg == "" || msg == code {
			t.Fatalf("missing en message for code %q", code)
		}
	}
}
