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
