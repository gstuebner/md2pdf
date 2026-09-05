package mdconv

import (
	"strings"
	"testing"
)

func convertString(t *testing.T, src string) string {
	t.Helper()
	res, err := Convert([]byte(src), Options{TOCDepth: 3, Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	return string(res.BodyHTML)
}

func TestCalloutKinds(t *testing.T) {
	cases := []struct{ marker, class, title string }{
		{"NOTE", "callout-note", "Hinweis"},
		{"TIP", "callout-tip", "Tipp"},
		{"IMPORTANT", "callout-important", "Wichtig"},
		{"WARNING", "callout-warning", "Warnung"},
		{"CAUTION", "callout-caution", "Achtung"},
	}
	for _, c := range cases {
		t.Run(c.marker, func(t *testing.T) {
			got := convertString(t, "> [!"+c.marker+"]\n> Inhalt.\n")
			if !strings.Contains(got, `class="callout `+c.class+`"`) {
				t.Errorf("missing class %q in:\n%s", c.class, got)
			}
			if !strings.Contains(got, ">"+c.title+"<") {
				t.Errorf("missing title %q in:\n%s", c.title, got)
			}
			if !strings.Contains(got, "Inhalt.") {
				t.Errorf("body text lost in:\n%s", got)
			}
			if strings.Contains(got, "[!") {
				t.Errorf("marker line not removed in:\n%s", got)
			}
		})
	}
}

func TestCalloutMarkerOnItsOwnParagraph(t *testing.T) {
	got := convertString(t, "> [!NOTE]\n>\n> Erst nach einer Leerzeile.\n")
	if !strings.Contains(got, "callout-note") {
		t.Fatalf("callout not recognised:\n%s", got)
	}
	if strings.Contains(got, "<p></p>") {
		t.Errorf("empty paragraph left behind:\n%s", got)
	}
	if !strings.Contains(got, "Erst nach einer Leerzeile.") {
		t.Errorf("body text lost:\n%s", got)
	}
}

func TestPlainBlockquoteStaysBlockquote(t *testing.T) {
	got := convertString(t, "> Ein normales Zitat.\n")
	if !strings.Contains(got, "<blockquote>") {
		t.Errorf("plain blockquote was rewritten:\n%s", got)
	}
	if strings.Contains(got, "callout") {
		t.Errorf("plain blockquote turned into a callout:\n%s", got)
	}
}

func TestUnknownMarkerIsNotACallout(t *testing.T) {
	got := convertString(t, "> [!HINWEIS]\n> Text.\n")
	if strings.Contains(got, "callout") {
		t.Errorf("unknown marker treated as callout:\n%s", got)
	}
}

func TestMarkerOnlyCountsOnFirstLine(t *testing.T) {
	got := convertString(t, "> Text zuerst.\n> [!NOTE]\n")
	if strings.Contains(got, "callout") {
		t.Errorf("marker below the first line must not trigger a callout:\n%s", got)
	}
}

func TestCalloutLanguageFollowsFrontMatter(t *testing.T) {
	got := convertString(t, "---\nlang: en\n---\n\n> [!WARNING]\n> Careful.\n")
	if !strings.Contains(got, ">Warning<") {
		t.Errorf("english callout title missing:\n%s", got)
	}
}
