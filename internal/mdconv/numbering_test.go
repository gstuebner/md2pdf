package mdconv

import "testing"

func TestDetectManualNumbering(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "own numbers with a dot",
			src:  "## 1. Einleitung\n\n## 2. Installation\n\n## 3. Betrieb\n",
			want: true,
		},
		{
			name: "own numbers without a dot",
			src:  "## 1 Einleitung\n\n## 2 Installation\n",
			want: true,
		},
		{
			name: "own numbers with a bracket",
			src:  "## 1) Einleitung\n\n## 2) Installation\n",
			want: true,
		},
		{
			name: "multi-level outline",
			src:  "## 1.1 Einleitung\n\n## 1.2 Installation\n",
			want: true,
		},
		{
			name: "no numbers",
			src:  "## Einleitung\n\n## Installation\n\n## Betrieb\n",
			want: false,
		},
		{
			name: "years are not an outline",
			src:  "## 2025 im Rückblick\n\n## 2026 im Ausblick\n",
			want: false,
		},
		{
			name: "a single numbered heading is not enough",
			src:  "## 1. Einleitung\n\n## Installation\n\n## Betrieb\n\n## Anhang\n",
			want: false,
		},
		{
			name: "only one heading in total",
			src:  "## 1. Einleitung\n",
			want: false,
		},
		{
			name: "sub-chapters do not count",
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
	// The first H1 moves to the cover page and must not confuse the detection.
	res, err := Convert([]byte("# Handbuch\n\n## 1. Eins\n\n## 2. Zwei\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.ManualNumbering {
		t.Error("ManualNumbering = false, want true")
	}
}
