// Package render assembles the final HTML document and the header/footer
// snippets that Chromium prints from.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/gstuebner/md2pdf/internal/assets"
	"github.com/gstuebner/md2pdf/internal/config"
	"github.com/gstuebner/md2pdf/internal/model"
)

// TemplateData is the data document.gohtml is executed with.
type TemplateData struct {
	Doc            model.Document
	CSS            template.CSS
	MermaidJS      template.JS // empty unless Doc.HasMermaid
	ShowCover      bool
	ShowTOC        bool
	NumberHeadings bool
	ChapterPages   bool
	BodyClass      string
	Lang           string
	Labels         Labels
}

// Labels holds the wording the document template needs in the document's
// language. The program's own messages are always English; only what ends up
// inside the PDF is translated.
type Labels struct {
	Contents  string // heading above the table of contents
	Version   string // cover sheet: document version
	Date      string // cover sheet: date
	Author    string // cover sheet: author
	Publisher string // cover sheet: company
	Kicker    string // fallback line above the cover title
	Untitled  string // fallback cover title
	Document  string // fallback <title>
}

var labelsByLang = map[string]Labels{
	"de": {
		Contents:  "Inhalt",
		Version:   "Version",
		Date:      "Stand",
		Author:    "Autor",
		Publisher: "Herausgeber",
		Kicker:    "Dokumentation",
		Untitled:  "Ohne Titel",
		Document:  "Dokument",
	},
	"en": {
		Contents:  "Contents",
		Version:   "Version",
		Date:      "Date",
		Author:    "Author",
		Publisher: "Publisher",
		Kicker:    "Documentation",
		Untitled:  "Untitled",
		Document:  "Document",
	},
}

// LabelsFor returns the document labels for lang, falling back to English.
func LabelsFor(lang string) Labels {
	if l, ok := labelsByLang[lang]; ok {
		return l
	}
	return labelsByLang["en"]
}

// RunningHead is the data header.gohtml and footer.gohtml are executed with.
type RunningHead struct {
	Title         string
	Version       string
	Footer        string // Company, falls back to Title
	PageWord      string // "Seite" or "Page"
	OfWord        string // "von" or "of"
	MarginLeftMM  float64
	MarginRightMM float64
}

// Document renders the complete HTML document for doc using the options in o.
func Document(doc model.Document, o config.Options) ([]byte, error) {
	css, err := buildCSS(o.Preset, o.ExtraCSS)
	if err != nil {
		return nil, err
	}

	var mermaidJS template.JS
	if doc.HasMermaid {
		js, err := assets.MermaidJS()
		if err != nil {
			return nil, fmt.Errorf("load mermaid.js asset: %w", err)
		}
		mermaidJS = template.JS(js)
	}

	lang := doc.Meta.Lang
	if lang == "" {
		lang = "en"
	}

	data := TemplateData{
		Doc:            doc,
		CSS:            template.CSS(css),
		MermaidJS:      mermaidJS,
		ShowCover:      o.Cover,
		ShowTOC:        o.TOC,
		NumberHeadings: o.NumberHeadings,
		ChapterPages:   o.ChapterPages,
		BodyClass:      bodyClass(o),
		Lang:           lang,
		Labels:         LabelsFor(lang),
	}

	out, err := executeEmbeddedTemplate("document.gohtml", data)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// bodyClass builds the class attribute of <body>. The stylesheet keys chapter
// numbering and page breaks off these classes.
func bodyClass(o config.Options) string {
	var classes []string
	if o.NumberHeadings {
		classes = append(classes, "numbered")
	}
	if o.ChapterPages {
		classes = append(classes, "chapters")
	}
	return strings.Join(classes, " ")
}

// Header renders the running header for the given metadata and options.
func Header(meta model.Meta, o config.Options) (string, error) {
	return renderRunningHead("header.gohtml", o.HeaderFile, buildRunningHead(meta, o))
}

// Footer renders the running footer for the given metadata and options.
func Footer(meta model.Meta, o config.Options) (string, error) {
	return renderRunningHead("footer.gohtml", o.FooterFile, buildRunningHead(meta, o))
}

func buildRunningHead(meta model.Meta, o config.Options) RunningHead {
	footer := meta.Company
	if footer == "" {
		footer = meta.Title
	}

	pageWord, ofWord := "Page", "of"
	if meta.Lang == "de" {
		pageWord, ofWord = "Seite", "von"
	}

	return RunningHead{
		Title:         meta.Title,
		Version:       meta.Version,
		Footer:        footer,
		PageWord:      pageWord,
		OfWord:        ofWord,
		MarginLeftMM:  o.Margins.LeftIn * 25.4,
		MarginRightMM: o.Margins.RightIn * 25.4,
	}
}

func renderRunningHead(embeddedName, overrideFile string, data RunningHead) (string, error) {
	var tmplSrc []byte
	name := embeddedName

	if overrideFile != "" {
		b, err := os.ReadFile(overrideFile)
		if err != nil {
			return "", fmt.Errorf("read custom template %q: %w", overrideFile, err)
		}
		tmplSrc = b
		name = overrideFile
	} else {
		b, err := assets.Template(embeddedName)
		if err != nil {
			return "", fmt.Errorf("load embedded template %q: %w", embeddedName, err)
		}
		tmplSrc = b
	}

	tmpl, err := template.New(name).Parse(string(tmplSrc))
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.String(), nil
}

func executeEmbeddedTemplate(name string, data any) ([]byte, error) {
	tmplSrc, err := assets.Template(name)
	if err != nil {
		return nil, fmt.Errorf("load embedded template %q: %w", name, err)
	}

	tmpl, err := template.New(name).Parse(string(tmplSrc))
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.Bytes(), nil
}

func buildCSS(preset string, extraCSS []string) (string, error) {
	if preset == "" {
		preset = config.DefaultPreset
	}
	theme, err := assets.Theme(preset)
	if err != nil {
		return "", fmt.Errorf("load theme css: %w", err)
	}

	var buf bytes.Buffer
	buf.Write(theme)

	for _, path := range extraCSS {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read extra css file %q: %w", path, err)
		}
		buf.WriteByte('\n')
		buf.Write(data)
	}

	return buf.String(), nil
}
