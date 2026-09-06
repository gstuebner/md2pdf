package render

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gstuebner/md2pdf/internal/assets"
	"github.com/gstuebner/md2pdf/internal/config"
	"github.com/gstuebner/md2pdf/internal/model"
)

func baseDoc() model.Document {
	return model.Document{
		Meta: model.Meta{
			Title: "Testdokument",
			Lang:  "de",
		},
		TOC: []model.TOCEntry{
			{Level: 1, Text: "Einführung", ID: "einfuehrung"},
		},
		BodyHTML: template.HTML("<h1 id=\"einfuehrung\">Einführung</h1><p>Text</p>"),
	}
}

func baseOptions() config.Options {
	o := config.DefaultOptions()
	return o
}

func TestDocumentCSSOrder(t *testing.T) {
	dir := t.TempDir()
	cssPath := filepath.Join(dir, "extra.css")
	marker := "/* extra-css-marker-xyz */"
	if err := os.WriteFile(cssPath, []byte(marker), 0o644); err != nil {
		t.Fatalf("write extra css: %v", err)
	}

	theme, err := assets.Theme(config.DefaultPreset)
	if err != nil {
		t.Fatalf("assets.Theme(%q): %v", config.DefaultPreset, err)
	}

	o := baseOptions()
	o.ExtraCSS = []string{cssPath}

	out, err := Document(baseDoc(), o)
	if err != nil {
		t.Fatalf("Document() error: %v", err)
	}

	outStr := string(out)
	themeIdx := strings.Index(outStr, string(theme))
	if themeIdx < 0 {
		t.Fatalf("output does not contain theme css verbatim")
	}
	markerIdx := strings.Index(outStr, marker)
	if markerIdx < 0 {
		t.Fatalf("output does not contain extra css marker")
	}
	if markerIdx < themeIdx {
		t.Fatalf("extra css (%d) appears before theme css (%d), want after", markerIdx, themeIdx)
	}
}

func TestDocumentMultipleExtraCSSOrder(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "a.css")
	second := filepath.Join(dir, "b.css")
	if err := os.WriteFile(first, []byte("/* marker-a */"), 0o644); err != nil {
		t.Fatalf("write a.css: %v", err)
	}
	if err := os.WriteFile(second, []byte("/* marker-b */"), 0o644); err != nil {
		t.Fatalf("write b.css: %v", err)
	}

	o := baseOptions()
	o.ExtraCSS = []string{first, second}

	out, err := Document(baseDoc(), o)
	if err != nil {
		t.Fatalf("Document() error: %v", err)
	}

	outStr := string(out)
	idxA := strings.Index(outStr, "/* marker-a */")
	idxB := strings.Index(outStr, "/* marker-b */")
	if idxA < 0 || idxB < 0 {
		t.Fatalf("markers missing: a=%d b=%d", idxA, idxB)
	}
	if idxB < idxA {
		t.Fatalf("css files out of order: a at %d, b at %d", idxA, idxB)
	}
}

func TestDocumentMissingExtraCSSIsHardError(t *testing.T) {
	o := baseOptions()
	o.ExtraCSS = []string{"/nonexistent/path/does-not-exist.css"}

	_, err := Document(baseDoc(), o)
	if err == nil {
		t.Fatalf("expected error for missing --css file, got nil")
	}
}

func TestDocumentMermaidInjectionOnlyWhenPresent(t *testing.T) {
	o := baseOptions()

	docWithout := baseDoc()
	docWithout.HasMermaid = false
	outWithout, err := Document(docWithout, o)
	if err != nil {
		t.Fatalf("Document() error: %v", err)
	}
	if strings.Contains(string(outWithout), "mermaid.initialize(") {
		t.Fatalf("output contains mermaid init code even though HasMermaid is false")
	}

	docWith := baseDoc()
	docWith.HasMermaid = true
	outWith, err := Document(docWith, o)
	if err != nil {
		t.Fatalf("Document() error: %v", err)
	}
	if !strings.Contains(string(outWith), "mermaid.initialize(") {
		t.Fatalf("output missing mermaid init code even though HasMermaid is true")
	}

	mermaidJS, err := assets.MermaidJS()
	if err != nil {
		t.Fatalf("assets.MermaidJS(): %v", err)
	}
	sample := string(mermaidJS[:64])
	if !strings.Contains(string(outWith), sample) {
		t.Fatalf("output does not contain embedded mermaid.min.js content")
	}
	if bytes.Contains(outWithout, mermaidJS[:64]) {
		t.Fatalf("output without mermaid unexpectedly contains mermaid.min.js content")
	}
}

func TestDocumentTOCTitle(t *testing.T) {
	cases := []struct {
		lang string
		want string
	}{
		{"de", "Inhalt"},
		{"en", "Contents"},
		{"", "Contents"},
		{"fr", "Contents"},
	}

	for _, tc := range cases {
		doc := baseDoc()
		doc.Meta.Lang = tc.lang
		o := baseOptions()

		out, err := Document(doc, o)
		if err != nil {
			t.Fatalf("Document() error for lang %q: %v", tc.lang, err)
		}
		want := `class="toc-heading">` + tc.want + `</h2>`
		if !strings.Contains(string(out), want) {
			t.Fatalf("lang %q: expected TOC title %q not found in output", tc.lang, tc.want)
		}
	}
}

func TestDocumentCoverAndTOCToggles(t *testing.T) {
	doc := baseDoc()

	o := baseOptions()
	o.Cover = false
	o.TOC = false
	out, err := Document(doc, o)
	if err != nil {
		t.Fatalf("Document() error: %v", err)
	}
	if strings.Contains(string(out), `class="cover"`) {
		t.Fatalf("cover present although ShowCover is false")
	}
	if strings.Contains(string(out), `class="toc"`) {
		t.Fatalf("toc present although ShowTOC is false")
	}
}

func TestHeaderContainsTitleAndVersion(t *testing.T) {
	meta := model.Meta{Title: "Mein Titel", Version: "1.2.3", Lang: "de"}
	o := baseOptions()

	out, err := Header(meta, o)
	if err != nil {
		t.Fatalf("Header() error: %v", err)
	}
	if !strings.Contains(out, "Mein Titel") {
		t.Fatalf("header missing title: %s", out)
	}
	if !strings.Contains(out, "1.2.3") {
		t.Fatalf("header missing version: %s", out)
	}
}

func TestFooterContainsPageNumberClasses(t *testing.T) {
	meta := model.Meta{Title: "Mein Titel", Company: "Beispiel GmbH", Lang: "de"}
	o := baseOptions()

	out, err := Footer(meta, o)
	if err != nil {
		t.Fatalf("Footer() error: %v", err)
	}
	if !strings.Contains(out, `class="pageNumber"`) {
		t.Fatalf("footer missing pageNumber class: %s", out)
	}
	if !strings.Contains(out, `class="totalPages"`) {
		t.Fatalf("footer missing totalPages class: %s", out)
	}
	if !strings.Contains(out, "Beispiel GmbH") {
		t.Fatalf("footer should show company: %s", out)
	}
	if !strings.Contains(out, "Seite") || !strings.Contains(out, "von") {
		t.Fatalf("footer missing German page words: %s", out)
	}
}

func TestFooterFallsBackToTitleWithoutCompany(t *testing.T) {
	meta := model.Meta{Title: "Nur Titel", Lang: "de"}
	o := baseOptions()

	out, err := Footer(meta, o)
	if err != nil {
		t.Fatalf("Footer() error: %v", err)
	}
	if !strings.Contains(out, "Nur Titel") {
		t.Fatalf("footer should fall back to title: %s", out)
	}
}

func TestFooterEnglishWords(t *testing.T) {
	meta := model.Meta{Title: "Title", Lang: "en"}
	o := baseOptions()

	out, err := Footer(meta, o)
	if err != nil {
		t.Fatalf("Footer() error: %v", err)
	}
	if !strings.Contains(out, "Page") || !strings.Contains(out, "of") {
		t.Fatalf("footer missing English page words: %s", out)
	}
}

func TestHeaderFooterMarginsInMM(t *testing.T) {
	meta := model.Meta{Title: "Titel", Lang: "de"}
	o := baseOptions()
	o.Margins = config.Margins{TopIn: 1, RightIn: 1, BottomIn: 1, LeftIn: 2}

	out, err := Header(meta, o)
	if err != nil {
		t.Fatalf("Header() error: %v", err)
	}
	// 2in -> 50.8mm, 1in -> 25.4mm
	if !strings.Contains(out, "50.8mm") {
		t.Fatalf("header missing left margin in mm: %s", out)
	}
	if !strings.Contains(out, "25.4mm") {
		t.Fatalf("header missing right margin in mm: %s", out)
	}
}

func TestHeaderFooterCustomTemplateFile(t *testing.T) {
	dir := t.TempDir()
	customHeader := filepath.Join(dir, "custom-header.gohtml")
	content := `<div>CUSTOM {{.Title}} / {{.Version}}</div>`
	if err := os.WriteFile(customHeader, []byte(content), 0o644); err != nil {
		t.Fatalf("write custom header: %v", err)
	}

	meta := model.Meta{Title: "Titel", Version: "9.9", Lang: "de"}
	o := baseOptions()
	o.HeaderFile = customHeader

	out, err := Header(meta, o)
	if err != nil {
		t.Fatalf("Header() error: %v", err)
	}
	if !strings.Contains(out, "CUSTOM Titel / 9.9") {
		t.Fatalf("expected custom header template to be used, got: %s", out)
	}
}

func TestHeaderCustomTemplateMissingFile(t *testing.T) {
	meta := model.Meta{Title: "Titel", Lang: "de"}
	o := baseOptions()
	o.HeaderFile = "/nonexistent/header.gohtml"

	_, err := Header(meta, o)
	if err == nil {
		t.Fatalf("expected error for missing custom header template")
	}
}
