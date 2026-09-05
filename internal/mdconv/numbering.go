package mdconv

import (
	"regexp"

	"github.com/yuin/goldmark/ast"
)

// manualNumber matches a chapter number the author wrote into the heading
// themselves: "1 ", "1. ", "2) ", "1.2 ", "10.3.1. ".
//
// A bare number is only accepted with at most two digits, so that a heading
// like "2026 im Rückblick" is not mistaken for a numbering scheme.
var manualNumber = regexp.MustCompile(`^(\d{1,2}|\d+(\.\d+)+)[.)]?\s+\S`)

// manualNumberingShare is the fraction of top-level chapters that must carry
// their own number before md2pdf keeps its hands off.
const manualNumberingShare = 0.6

// detectManualNumbering reports whether the document numbers its own chapters.
// Only level-2 headings are examined, because that is the level the theme
// numbers first and the level on which an author's scheme shows.
func detectManualNumbering(doc *ast.Document, src []byte) bool {
	var total, numbered int
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok || h.Level != 2 {
			return ast.WalkContinue, nil
		}
		total++
		if manualNumber.MatchString(nodeText(h, src)) {
			numbered++
		}
		return ast.WalkContinue, nil
	})

	if total < 2 || numbered < 2 {
		return false
	}
	return float64(numbered)/float64(total) >= manualNumberingShare
}
