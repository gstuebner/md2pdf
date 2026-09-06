// Package mdconv turns Markdown into the HTML fragment and table of contents
// that the document template expects.
package mdconv

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"

	"github.com/gstuebner/md2pdf/internal/model"
)

// AssetLoader resolves an image reference to something the offline renderer can
// display, normally a data URI. It is declared here rather than imported so that
// mdconv stays independent of the render package.
type AssetLoader interface {
	DataURI(src string) (string, error)
}

type Options struct {
	TOCDepth int
	Loader   AssetLoader
	// Lang preselects the callout labels. Front matter overrides it.
	Lang string
	// LangFallback is used when neither Lang nor the front matter names a
	// language. Empty means "en".
	LangFallback string
}

type Result struct {
	Meta       model.Meta
	TOC        []model.TOCEntry
	BodyHTML   template.HTML
	HasMermaid bool
	// ManualNumbering is true when the document numbers its own chapters, so
	// that the theme must not add a second number in front of them.
	ManualNumbering bool
}

// docState carries values that are only known once the front matter has been
// parsed, but that the renderers need while rendering the same document.
type docState struct {
	hasMermaid bool
	lang       string
}

func Convert(src []byte, o Options) (Result, error) {
	if o.TOCDepth <= 0 {
		o.TOCDepth = 3
	}

	src = StripBOM(src)

	state := &docState{lang: o.Lang}
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
			extension.DefinitionList,
			extension.Footnote,
			&frontmatter.Extender{},
			&calloutExtension{state: state},
			&figureExtension{},
			&codeExtension{state: state},
			&imageExtension{loader: o.Loader},
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	pc := parser.NewContext()
	root := md.Parser().Parse(text.NewReader(src), parser.WithContext(pc))
	doc, ok := root.(*ast.Document)
	if !ok {
		return Result{}, fmt.Errorf("unexpected root node %T", root)
	}

	var res Result
	if fm := frontmatter.Get(pc); fm != nil {
		if err := fm.Decode(&res.Meta); err != nil {
			return Result{}, fmt.Errorf("parse front matter: %w", err)
		}
	}
	// Front matter only decides the language when the caller did not pass one,
	// so that a --lang flag keeps precedence over the document.
	if res.Meta.Lang != "" && o.Lang == "" {
		state.lang = res.Meta.Lang
	}
	if state.lang == "" {
		state.lang = o.LangFallback
	}

	if title := takeTitleHeading(doc, src); title != "" && res.Meta.Title == "" {
		res.Meta.Title = title
	}

	res.TOC = collectTOC(doc, src, o.TOCDepth)
	res.ManualNumbering = detectManualNumbering(doc, src)

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, src, doc); err != nil {
		return Result{}, fmt.Errorf("render markdown: %w", err)
	}

	res.BodyHTML = template.HTML(buf.String()) //nolint:gosec // the renderer escapes; raw HTML is opt-in by design
	res.HasMermaid = state.hasMermaid
	return res, nil
}

// takeTitleHeading removes a leading level-1 heading from the document and
// returns its text. It belongs on the cover page, not into the body.
func takeTitleHeading(doc *ast.Document, src []byte) string {
	first := firstBlock(doc)
	h, ok := first.(*ast.Heading)
	if !ok || h.Level != 1 {
		return ""
	}
	title := nodeText(h, src)
	doc.RemoveChild(doc, h)
	return title
}

// firstBlock skips over nodes that carry no rendered content, such as the front
// matter block the extension leaves behind.
func firstBlock(doc *ast.Document) ast.Node {
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind().String() {
		case "FrontMatter", "Frontmatter":
			continue
		}
		return c
	}
	return nil
}

func collectTOC(doc *ast.Document, src []byte, depth int) []model.TOCEntry {
	var out []model.TOCEntry
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok || h.Level > depth {
			return ast.WalkContinue, nil
		}
		id, ok := h.AttributeString("id")
		if !ok {
			return ast.WalkContinue, nil
		}
		idBytes, ok := id.([]byte)
		if !ok {
			return ast.WalkContinue, nil
		}
		out = append(out, model.TOCEntry{
			Level: h.Level,
			Text:  nodeText(h, src),
			ID:    string(idBytes),
		})
		return ast.WalkContinue, nil
	})
	return out
}

// nodeText returns the plain text of a node. ast.Node.Text is deprecated in
// current goldmark versions, so the text nodes are collected by hand.
func nodeText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := node.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
			if t.SoftLineBreak() || t.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(t.Value)
		case *ast.AutoLink:
			b.Write(t.URL(src))
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

// codeText returns the raw source of a code block.
func codeText(n ast.Node, src []byte) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}
