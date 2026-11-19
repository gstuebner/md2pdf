package mdconv

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// slugify turns heading text into an anchor following GitHub's scheme
// (github-slugger): lower-case, drop everything except letters, combining
// marks, numbers, '-', '_' and spaces, then replace each space with '-'.
// Spaces are not collapsed and edges are not trimmed, so "A – B" becomes
// "a--b". An empty result falls back to "heading".
func slugify(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_' ||
			unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsNumber(r):
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "heading"
	}
	return b.String()
}

// headingIDs implements parser.IDs with GitHub-style anchors. Duplicates get
// "-1", "-2", … appended, skipping names that are already taken.
type headingIDs struct {
	used map[string]bool
}

var _ parser.IDs = (*headingIDs)(nil)

func newHeadingIDs() *headingIDs {
	return &headingIDs{used: map[string]bool{}}
}

// Generate receives the raw source of the heading line from goldmark.
func (h *headingIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	base := slugify(strings.TrimSpace(string(value)))
	if kind != ast.KindHeading && base == "heading" {
		base = "id"
	}
	id := base
	for i := 1; h.used[id]; i++ {
		id = base + "-" + strconv.Itoa(i)
	}
	h.used[id] = true
	return []byte(id)
}

// Put marks an explicitly assigned id as taken.
func (h *headingIDs) Put(value []byte) {
	h.used[string(value)] = true
}
