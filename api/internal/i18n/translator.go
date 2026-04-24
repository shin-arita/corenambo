package i18n

type Translator interface {
	Translate(lang string, code string) string
}

type translator struct{}

func NewTranslator() Translator {
	return &translator{}
}

func (t *translator) Translate(lang string, code string) string {
	var dict map[string]string

	switch lang {
	case "en":
		dict = enMessages
	default:
		dict = jaMessages
	}

	if msg, ok := dict[code]; ok {
		return msg
	}

	return code
}
