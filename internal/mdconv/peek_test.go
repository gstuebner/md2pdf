package mdconv

import "testing"

func TestPeekMeta(t *testing.T) {
	cases := []struct {
		name       string
		src        string
		wantPreset string
		wantLang   string
	}{
		{"no front matter", "# Title\n\nText.\n", "", ""},
		{"preset and lang", "---\npreset: report\nlang: de\n---\n\n# Title\n", "report", "de"},
		{"crlf line endings", "---\r\npreset: modern\r\n---\r\n\r\n# Title\r\n", "modern", ""},
		{"unterminated block", "---\npreset: report\n\n# Title\n", "", ""},
		{"broken yaml", "---\npreset: [\n---\n\n# Title\n", "", ""},
		{"fence not at the start", "Text\n\n---\npreset: report\n---\n", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := PeekMeta([]byte(tc.src))
			if m.Preset != tc.wantPreset {
				t.Errorf("Preset = %q, want %q", m.Preset, tc.wantPreset)
			}
			if m.Lang != tc.wantLang {
				t.Errorf("Lang = %q, want %q", m.Lang, tc.wantLang)
			}
		})
	}
}

// PeekMeta must agree with the front matter Convert actually parses.
func TestPeekMetaMatchesConvert(t *testing.T) {
	src := []byte("---\ntitle: Handbuch\npreset: technical\nlang: de\n---\n\n# Handbuch\n\nText.\n")

	res, err := Convert(src, Options{TOCDepth: 3, Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	peeked := PeekMeta(src)

	if peeked.Preset != res.Meta.Preset || peeked.Lang != res.Meta.Lang || peeked.Title != res.Meta.Title {
		t.Errorf("PeekMeta = %+v, Convert = %+v", peeked, res.Meta)
	}
}
