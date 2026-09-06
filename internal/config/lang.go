package config

import (
	"os"
	"strings"
)

// SupportedLangs are the document languages md2pdf has labels for.
var SupportedLangs = []string{"de", "en"}

// DetectLang guesses the document language from the system. It looks at LC_ALL
// and then LANG, which is how locales are configured on Unix; when neither is
// set it asks the operating system (on Windows that is the user's default
// locale, since Windows sets no such variables). Anything md2pdf has no labels
// for, and a system that reveals nothing, yields "en".
func DetectLang() string {
	for _, key := range []string{"LC_ALL", "LANG"} {
		if lang, ok := langFromLocale(os.Getenv(key)); ok {
			return lang
		}
	}
	if lang, ok := langFromLocale(platformLocale()); ok {
		return lang
	}
	return "en"
}

func langFromLocale(value string) (string, bool) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", false
	}
	// "de_DE.UTF-8" on Unix, "de-DE" from the Windows API — cut at the first
	// separator either way.
	if i := strings.IndexAny(v, "_-.@"); i >= 0 {
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
