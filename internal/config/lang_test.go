package config

import "testing"

func TestDetectLang(t *testing.T) {
	cases := []struct {
		lcAll, lang string
		want        string
	}{
		{"", "", "en"},
		{"", "de_DE.UTF-8", "de"},
		{"", "en_US.UTF-8", "en"},
		{"", "de", "de"},
		{"de_AT@euro", "en_US.UTF-8", "de"},
		{"", "fr_FR.UTF-8", "en"},
		{"", "C", "en"},
		{"", "POSIX", "en"},
		// Windows reports locale names with a hyphen; DetectLang sees these
		// through platformLocale, langFromLocale must cope with the form.
		{"de-DE", "", "de"},
		{"en-GB", "", "en"},
	}

	for _, tc := range cases {
		t.Setenv("LC_ALL", tc.lcAll)
		t.Setenv("LANG", tc.lang)
		if got := DetectLang(); got != tc.want {
			t.Errorf("DetectLang() with LC_ALL=%q LANG=%q = %q, want %q", tc.lcAll, tc.lang, got, tc.want)
		}
	}
}

// langFromLocale has to handle both the Unix spelling and the one the Windows
// API returns, because platformLocale feeds it "de-DE".
func TestLangFromLocale(t *testing.T) {
	cases := map[string]string{
		"de_DE.UTF-8": "de",
		"de-DE":       "de",
		"de":          "de",
		"DE-de":       "de",
		"en-US":       "en",
		"de_AT@euro":  "de",
	}
	for in, want := range cases {
		got, ok := langFromLocale(in)
		if !ok || got != want {
			t.Errorf("langFromLocale(%q) = %q, %v; want %q, true", in, got, ok, want)
		}
	}

	for _, in := range []string{"", "  ", "C", "POSIX", "fr-FR", "zh_CN.UTF-8"} {
		if got, ok := langFromLocale(in); ok {
			t.Errorf("langFromLocale(%q) = %q, true; want no match", in, got)
		}
	}
}
