// Package render assembles the final HTML document and the header/footer
// snippets that Chromium prints from.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

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
	TOCTitle       string
	NumberHeadings bool
	ChapterPages   bool
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
	css, err := buildCSS(o.ExtraCSS)
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

	data := TemplateData{
		Doc:            doc,
		CSS:            template.CSS(css),
		MermaidJS:      mermaidJS,
		ShowCover:      o.Cover,
		ShowTOC:        o.TOC,
		TOCTitle:       tocTitle(doc.Meta.Lang),
		NumberHeadings: o.NumberHeadings,
		ChapterPages:   o.ChapterPages,
	}

	out, err := executeEmbeddedTemplate("document.gohtml", data)
	if err != nil {
		return nil, err
	}
	return out, nil
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

	pageWord, ofWord := "Seite", "von"
	if meta.Lang == "en" {
		pageWord, ofWord = "Page", "of"
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

func buildCSS(extraCSS []string) (string, error) {
	theme, err := assets.Theme()
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

func tocTitle(lang string) string {
	if lang == "en" {
		return "Contents"
	}
	return "Inhalt"
}
