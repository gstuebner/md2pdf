package mdconv

import (
	"strings"
	"testing"
)

func TestLeadingH1BecomesTitleAndLeavesBody(t *testing.T) {
	res, err := Convert([]byte("# Mein Titel\n\nText.\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Meta.Title != "Mein Titel" {
		t.Errorf("Title = %q, want %q", res.Meta.Title, "Mein Titel")
	}
	if strings.Contains(string(res.BodyHTML), "Mein Titel") {
		t.Errorf("title heading still in body:\n%s", res.BodyHTML)
	}
}

func TestFrontMatterTitleWinsAndH1Stays(t *testing.T) {
	res, err := Convert([]byte("---\ntitle: Aus Frontmatter\n---\n\n# Aus Überschrift\n\nText.\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Meta.Title != "Aus Frontmatter" {
		t.Errorf("Title = %q, want front matter value", res.Meta.Title)
	}
	if strings.Contains(string(res.BodyHTML), "Aus Überschrift") {
		t.Errorf("leading H1 is always removed from the body, even when the title comes from front matter:\n%s", res.BodyHTML)
	}
}

func TestH1BelowOtherContentStays(t *testing.T) {
	res, err := Convert([]byte("Ein Absatz.\n\n# Keine Titelzeile\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Meta.Title != "" {
		t.Errorf("Title = %q, want empty", res.Meta.Title)
	}
	if !strings.Contains(string(res.BodyHTML), "Keine Titelzeile") {
		t.Errorf("H1 below other content must stay in the body:\n%s", res.BodyHTML)
	}
}

func TestTOCDepthIsRespected(t *testing.T) {
	src := "## Zwei\n\n### Drei\n\n#### Vier\n"
	res, err := Convert([]byte(src), Options{TOCDepth: 2, Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.TOC) != 1 || res.TOC[0].Text != "Zwei" {
		t.Errorf("TOC = %+v, want only the level-2 heading", res.TOC)
	}
}

func TestNodeTextJoinsInlineMarkup(t *testing.T) {
	res, err := Convert([]byte("## Ein **fetter** Titel mit `Code`\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.TOC) != 1 {
		t.Fatalf("TOC = %+v", res.TOC)
	}
	if res.TOC[0].Text != "Ein fetter Titel mit Code" {
		t.Errorf("TOC text = %q, want plain text without markup", res.TOC[0].Text)
	}
}
