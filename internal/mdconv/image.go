package mdconv

import (
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type imageExtension struct {
	loader AssetLoader
}

func (e *imageExtension) Extend(md goldmark.Markdown) {
	md.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&imageRenderer{loader: e.loader}, 100),
	))
}

type imageRenderer struct {
	loader AssetLoader
}

func (r *imageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.render)
}

func (r *imageRenderer) render(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	img := n.(*ast.Image)

	dest := string(img.Destination)
	if r.loader != nil {
		resolved, err := r.loader.DataURI(dest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "md2pdf: Bild %q konnte nicht eingebettet werden: %v\n", dest, err)
		} else {
			dest = resolved
		}
	}

	_, _ = w.WriteString(`<img src="`)
	// Data URIs must not be URL-escaped: base64 uses "+", which would turn into
	// %2B and break the image.
	if strings.HasPrefix(dest, "data:") {
		_, _ = w.Write(util.EscapeHTML([]byte(dest)))
	} else {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape([]byte(dest), true)))
	}
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(util.EscapeHTML([]byte(nodeText(img, src))))
	_ = w.WriteByte('"')
	if len(img.Title) > 0 {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(util.EscapeHTML(img.Title))
		_ = w.WriteByte('"')
	}
	_, _ = w.WriteString(">")
	return ast.WalkSkipChildren, nil
}
