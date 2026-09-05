package mdconv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type codeExtension struct {
	state *docState
}

func (e *codeExtension) Extend(md goldmark.Markdown) {
	md.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&codeRenderer{state: e.state}, 100),
	))
}

type codeRenderer struct {
	state *docState
}

func (r *codeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFenced)
	reg.Register(ast.KindCodeBlock, r.renderIndented)
}

func (r *codeRenderer) renderFenced(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	node := n.(*ast.FencedCodeBlock)
	lang := strings.ToLower(strings.TrimSpace(string(node.Language(src))))
	code := codeText(node, src)

	if lang == "mermaid" {
		r.state.hasMermaid = true
		_, _ = w.WriteString(`<pre class="mermaid">`)
		_, _ = w.Write(util.EscapeHTML([]byte(code)))
		_, _ = w.WriteString("</pre>\n")
		return ast.WalkSkipChildren, nil
	}

	writeCodeBlock(w, lang, code)
	return ast.WalkSkipChildren, nil
}

func (r *codeRenderer) renderIndented(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	writeCodeBlock(w, "", codeText(n, src))
	return ast.WalkSkipChildren, nil
}

// writeCodeBlock emits the wrapper the theme styles, with the language label and
// the line count it needs, followed by the Chroma output.
func writeCodeBlock(w util.BufWriter, lang, code string) {
	_, _ = fmt.Fprintf(w, `<div class="codeblock" data-lang="%s" data-lines="%d">`,
		util.EscapeHTML([]byte(lang)), countLines(code))
	if err := highlight(w, lang, code); err != nil {
		_, _ = w.WriteString(`<pre class="chroma"><code>`)
		_, _ = w.Write(util.EscapeHTML([]byte(code)))
		_, _ = w.WriteString("</code></pre>")
	}
	_, _ = w.WriteString("</div>\n")
}

// highlight writes Chroma output that carries CSS classes only; the colours come
// from theme/code.css so that they stay under the theme's control.
func highlight(w util.BufWriter, lang, code string) error {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	it, err := lexer.Tokenise(nil, code)
	if err != nil {
		return fmt.Errorf("tokenise %q: %w", lang, err)
	}
	f := chromahtml.New(chromahtml.WithClasses(true))
	if err := f.Format(w, styles.Fallback, it); err != nil {
		return fmt.Errorf("format %q: %w", lang, err)
	}
	return nil
}

func countLines(code string) int {
	trimmed := strings.TrimRight(code, "\n")
	if trimmed == "" {
		return 0
	}
	return strings.Count(trimmed, "\n") + 1
}
