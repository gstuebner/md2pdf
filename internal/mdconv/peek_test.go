package mdconv

import (
	"strings"
	"testing"
)

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

// A UTF-8 BOM at the start of the file must not keep the first heading from
// being recognised — otherwise "# Title" ends up verbatim in the body and the
// cover page stays untitled.
func TestConvertHandlesUTF8BOM(t *testing.T) {
	const bomBytes = "\xef\xbb\xbf"

	res, err := Convert([]byte(bomBytes+"# Titel\n\nText.\n"), Options{TOCDepth: 3, Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Meta.Title != "Titel" {
		t.Errorf("Title = %q, want %q", res.Meta.Title, "Titel")
	}
	if strings.Contains(string(res.BodyHTML), "# Titel") {
		t.Errorf("heading markup leaked into the body:\n%s", res.BodyHTML)
	}

	// Front matter behind a BOM must parse as well, in Convert and in PeekMeta.
	src := []byte(bomBytes + "---\ntitle: Aus Frontmatter\npreset: report\n---\n\n# Heading\n")
	res, err = Convert(src, Options{TOCDepth: 3, Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Meta.Title != "Aus Frontmatter" || res.Meta.Preset != "report" {
		t.Errorf("front matter behind a BOM not parsed: %+v", res.Meta)
	}
	if got := PeekMeta(src); got.Preset != "report" {
		t.Errorf("PeekMeta behind a BOM = %+v, want preset report", got)
	}
}
