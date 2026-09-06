package config

import (
	"os"
	"strings"
)

// SupportedLangs are the document languages md2pdf has labels for.
var SupportedLangs = []string{"de", "en"}

// DetectLang guesses the document language from the environment. It looks at
// LC_ALL and then LANG, takes the part before "_" or ".", and returns it when
// md2pdf has labels for it. Everything else, including an unset environment,
// yields "en".
func DetectLang() string {
	for _, key := range []string{"LC_ALL", "LANG"} {
		if lang, ok := langFromLocale(os.Getenv(key)); ok {
			return lang
		}
	}
	return "en"
}

func langFromLocale(value string) (string, bool) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", false
	}
	if i := strings.IndexAny(v, "_.@"); i >= 0 {
		v = v[:i]
	}
	v = strings.ToLower(v)
	for _, l := range SupportedLangs {
		if v == l {
			return l, true
		}
	}
	return "", false
}
