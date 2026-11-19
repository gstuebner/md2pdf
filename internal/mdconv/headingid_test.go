package mdconv

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1. Einleitung – Motivation und Zweck", "1-einleitung--motivation-und-zweck"},
		{"2. Erste Schritte", "2-erste-schritte"},
		{"3. Die Ansichten im Überblick", "3-die-ansichten-im-überblick"},
		{"4. Filtern und Suchen", "4-filtern-und-suchen"},
		{"5. Datensätze anlegen", "5-datensätze-anlegen"},
		{"6. Datensätze ansehen und bearbeiten", "6-datensätze-ansehen-und-bearbeiten"},
		{"7. Kontextmenü: Sammelaktionen für markierte Zeilen", "7-kontextmenü-sammelaktionen-für-markierte-zeilen"},
		{"8. Ansicht „Prüfen\" und SAGA-Abgleich", "8-ansicht-prüfen-und-saga-abgleich"},
		{"9. Schutz vor gleichzeitigen Änderungen", "9-schutz-vor-gleichzeitigen-änderungen"},
		{"10. Einstellungen", "10-einstellungen"},
		{"11. Begriffe und Datenregeln", "11-begriffe-und-datenregeln"},
		{"12. Kurzreferenz", "12-kurzreferenz"},
		{"Einführung", "einführung"},
		{"Tägliche Arbeit", "tägliche-arbeit"},
		{"Anmeldung schlägt fehl", "anmeldung-schlägt-fehl"},
		{"Straße & Hausnummer", "straße--hausnummer"},
		{"C# & .NET", "c--net"},
		{"API-Zugriff", "api-zugriff"},
		{"snake_case_name", "snake_case_name"},
		{"Café Crème", "café-crème"},
		{"Café Crème", "café-crème"},
		{"!!!", "heading"},
		{"A -- B", "a----b"},
		{"`code` im Titel", "code-im-titel"},
	}
	for _, c := range cases {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHeadingIDsDuplicates(t *testing.T) {
	gen := func(h *headingIDs, s string) string {
		return string(h.Generate([]byte(s), ast.KindHeading))
	}
	check := func(name string, got, want []string) {
		t.Helper()
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s: got %q, want %q", name, got, want)
				return
			}
		}
	}

	h := newHeadingIDs()
	check("three equal", []string{gen(h, "Beispiel"), gen(h, "Beispiel"), gen(h, "Beispiel")},
		[]string{"beispiel", "beispiel-1", "beispiel-2"})

	h = newHeadingIDs()
	check("literal suffix", []string{gen(h, "Beispiel"), gen(h, "Beispiel 1"), gen(h, "Beispiel")},
		[]string{"beispiel", "beispiel-1", "beispiel-2"})

	h = newHeadingIDs()
	h.Put([]byte("beispiel"))
	check("put", []string{gen(h, "Beispiel")}, []string{"beispiel-1"})
}

var (
	idAttr   = regexp.MustCompile(`id="([^"]*)"`)
	hrefAttr = regexp.MustCompile(`href="#([^"]*)"`)
)

func TestConvertUmlautAnchorLink(t *testing.T) {
	src := "## Überblick\n\n[hin](#überblick)\n"
	res, err := Convert([]byte(src), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	ids := idAttr.FindAllStringSubmatch(string(res.BodyHTML), -1)
	if len(ids) != 1 || ids[0][1] != "überblick" {
		t.Fatalf("heading ids = %v, want [überblick]\n%s", ids, res.BodyHTML)
	}
	hrefs := hrefAttr.FindAllStringSubmatch(string(res.BodyHTML), -1)
	if len(hrefs) != 1 {
		t.Fatalf("want one fragment link\n%s", res.BodyHTML)
	}
	target, err := url.PathUnescape(hrefs[0][1])
	if err != nil {
		t.Fatal(err)
	}
	if target != "überblick" {
		t.Errorf("link target = %q, want %q", target, "überblick")
	}
}

// TestConvertHeadingIDSource pins down what goldmark hands to Generate: the
// raw source of the heading line, before inline parsing and Typographer.
func TestConvertHeadingIDSource(t *testing.T) {
	cases := []struct{ md, want string }{
		{"## Ansicht „Prüfen\" und SAGA-Abgleich", "ansicht-prüfen-und-saga-abgleich"},
		{"## `code` im Titel", "code-im-titel"},
		{"## A -- B", "a----b"},
		{"## **Fett** im Titel ##", "fett-im-titel"},
		// Known deviation: GitHub gives "link-im-titel".
		{"## [Link](ziel) im Titel", "linkziel-im-titel"},
	}
	for _, c := range cases {
		res, err := Convert([]byte(c.md+"\n"), Options{Loader: stubLoader{}})
		if err != nil {
			t.Fatal(err)
		}
		m := idAttr.FindStringSubmatch(string(res.BodyHTML))
		if m == nil || m[1] != c.want {
			t.Errorf("%q: got %v, want id %q\n%s", c.md, m, c.want, res.BodyHTML)
		}
	}
}
