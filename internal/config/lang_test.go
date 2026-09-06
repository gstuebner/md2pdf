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
	}

	for _, tc := range cases {
		t.Setenv("LC_ALL", tc.lcAll)
		t.Setenv("LANG", tc.lang)
		if got := DetectLang(); got != tc.want {
			t.Errorf("DetectLang() with LC_ALL=%q LANG=%q = %q, want %q", tc.lcAll, tc.lang, got, tc.want)
		}
	}
}
