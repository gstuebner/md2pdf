package mdconv

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type figureExtension struct{}

func (e *figureExtension) Extend(md goldmark.Markdown) {
	md.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&figureRenderer{}, 100),
	))
}

type figureRenderer struct{}

func (r *figureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindParagraph, r.render)
}

func (r *figureRenderer) render(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	img := soleImage(n, src)
	if img == nil {
		if entering {
			_, _ = w.WriteString("<p>")
		} else {
			_, _ = w.WriteString("</p>\n")
		}
		return ast.WalkContinue, nil
	}

	if entering {
		_, _ = w.WriteString("<figure class=\"figure\">\n")
		return ast.WalkContinue, nil
	}

	caption := string(img.Title)
	if caption == "" {
		caption = nodeText(img, src)
	}
	if caption != "" {
		_, _ = w.WriteString("\n<figcaption>")
		_, _ = w.Write(util.EscapeHTML([]byte(caption)))
		_, _ = w.WriteString("</figcaption>")
	}
	_, _ = w.WriteString("\n</figure>\n")
	return ast.WalkContinue, nil
}

// soleImage reports the image of a paragraph that contains nothing but that
// image, ignoring surrounding whitespace.
func soleImage(n ast.Node, src []byte) *ast.Image {
	var found *ast.Image
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch v := c.(type) {
		case *ast.Image:
			if found != nil {
				return nil
			}
			found = v
		case *ast.Text:
			if len(bytes.TrimSpace(v.Segment.Value(src))) != 0 {
				return nil
			}
		default:
			return nil
		}
	}
	return found
}
