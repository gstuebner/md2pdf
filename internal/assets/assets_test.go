package assets

import (
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	for _, name := range []string{"document.gohtml", "templates/document.gohtml", "header.gohtml", "footer.gohtml"} {
		content, err := Template(name)
		if err != nil {
			t.Fatalf("Template(%q) error: %v", name, err)
		}
		if len(content) == 0 {
			t.Errorf("Template(%q) returned empty content", name)
		}
	}

	if _, err := Template("nonexistent.gohtml"); err == nil {
		t.Errorf("Template(nonexistent.gohtml) expected error, got nil")
	}
}

func TestTheme(t *testing.T) {
	css, err := Theme()
	if err != nil {
		t.Fatalf("Theme() error: %v", err)
	}
	s := string(css)
	if !strings.Contains(s, "@font-face") {
		t.Errorf("Theme() should contain fonts.css (@font-face)")
	}
	if !strings.Contains(s, "callout") {
		t.Errorf("Theme() should contain default.css (callout)")
	}
	if !strings.Contains(s, "chroma") {
		t.Errorf("Theme() should contain code.css (chroma)")
	}
}

func TestMermaidJS(t *testing.T) {
	js, err := MermaidJS()
	if err != nil {
		t.Fatalf("MermaidJS() error: %v", err)
	}
	if len(js) < 1000000 {
		t.Errorf("MermaidJS() size suspiciously small: %d bytes", len(js))
	}
}
