package mdconv

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// calloutMarker matches the GitHub alert marker on the first line of a block
// quote, for example "[!WARNING]".
var calloutMarker = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$`)

var calloutTitles = map[string]map[string]string{
	"de": {
		"note":      "Hinweis",
		"tip":       "Tipp",
		"important": "Wichtig",
		"warning":   "Warnung",
		"caution":   "Achtung",
	},
	"en": {
		"note":      "Note",
		"tip":       "Tip",
		"important": "Important",
		"warning":   "Warning",
		"caution":   "Caution",
	},
}

func calloutTitle(kind, lang string) string {
	titles, ok := calloutTitles[lang]
	if !ok {
		titles = calloutTitles["de"]
	}
	if t, ok := titles[kind]; ok {
		return t
	}
	return kind
}

type calloutExtension struct {
	state *docState
}

func (e *calloutExtension) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&calloutTransformer{}, 100),
	))
	md.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&calloutRenderer{state: e.state}, 100),
	))
}

type calloutTransformer struct{}

func (t *calloutTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		bq, ok := n.(*ast.Blockquote)
		if !ok {
			return ast.WalkContinue, nil
		}
		para, ok := bq.FirstChild().(*ast.Paragraph)
		if !ok || para.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}
		firstLine := para.Lines().At(0)
		m := calloutMarker.FindSubmatch(bytes.TrimSpace(firstLine.Value(src)))
		if m == nil {
			return ast.WalkContinue, nil
		}

		dropFirstLine(para)
		if para.FirstChild() == nil {
			bq.RemoveChild(bq, para)
		}
		bq.SetAttributeString("data-callout", []byte(strings.ToLower(string(m[1]))))
		return ast.WalkContinue, nil
	})
}

// dropFirstLine removes the inline nodes that make up the first line of a
// paragraph. The marker may have been split across several text nodes, so the
// loop runs until it has removed the node that carries the line break.
func dropFirstLine(para *ast.Paragraph) {
	for c := para.FirstChild(); c != nil; {
		next := c.NextSibling()
		last := false
		if t, ok := c.(*ast.Text); ok && (t.SoftLineBreak() || t.HardLineBreak()) {
			last = true
		}
		para.RemoveChild(para, c)
		if last {
			return
		}
		c = next
	}
}

type calloutRenderer struct {
	state *docState
}

func (r *calloutRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindBlockquote, r.render)
}

func (r *calloutRenderer) render(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	raw, ok := n.AttributeString("data-callout")
	kind, _ := raw.([]byte)
	if !ok || len(kind) == 0 {
		if entering {
			_, _ = w.WriteString("<blockquote>\n")
		} else {
			_, _ = w.WriteString("</blockquote>\n")
		}
		return ast.WalkContinue, nil
	}

	if entering {
		_, _ = fmt.Fprintf(w, "<div class=\"callout callout-%s\">\n<p class=\"callout-title\">%s</p>\n<div class=\"callout-body\">\n",
			kind, calloutTitle(string(kind), r.state.lang))
	} else {
		_, _ = w.WriteString("</div>\n</div>\n")
	}
	return ast.WalkContinue, nil
}
