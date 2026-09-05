package mdconv

import "testing"

func TestDetectManualNumbering(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "eigene Nummern mit Punkt",
			src:  "## 1. Einleitung\n\n## 2. Installation\n\n## 3. Betrieb\n",
			want: true,
		},
		{
			name: "eigene Nummern ohne Punkt",
			src:  "## 1 Einleitung\n\n## 2 Installation\n",
			want: true,
		},
		{
			name: "eigene Nummern mit Klammer",
			src:  "## 1) Einleitung\n\n## 2) Installation\n",
			want: true,
		},
		{
			name: "mehrstellige Gliederung",
			src:  "## 1.1 Einleitung\n\n## 1.2 Installation\n",
			want: true,
		},
		{
			name: "ohne Nummern",
			src:  "## Einleitung\n\n## Installation\n\n## Betrieb\n",
			want: false,
		},
		{
			name: "Jahreszahlen sind keine Gliederung",
			src:  "## 2025 im Rückblick\n\n## 2026 im Ausblick\n",
			want: false,
		},
		{
			name: "einzelne nummerierte Überschrift reicht nicht",
			src:  "## 1. Einleitung\n\n## Installation\n\n## Betrieb\n\n## Anhang\n",
			want: false,
		},
		{
			name: "nur eine Überschrift insgesamt",
			src:  "## 1. Einleitung\n",
			want: false,
		},
		{
			name: "Unterkapitel zählen nicht mit",
			src:  "## Einleitung\n\n### 1. Schritt\n\n### 2. Schritt\n\n## Installation\n",
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := Convert([]byte(c.src), Options{Loader: stubLoader{}})
			if err != nil {
				t.Fatal(err)
			}
			if res.ManualNumbering != c.want {
				t.Errorf("ManualNumbering = %v, want %v", res.ManualNumbering, c.want)
			}
		})
	}
}

func TestManualNumberingIgnoresLeadingTitle(t *testing.T) {
	// Die erste H1 wandert aufs Deckblatt und darf die Erkennung nicht stören.
	res, err := Convert([]byte("# Handbuch\n\n## 1. Eins\n\n## 2. Zwei\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.ManualNumbering {
		t.Error("ManualNumbering = false, want true")
	}
}
