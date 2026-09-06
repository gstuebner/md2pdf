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
	for _, preset := range []string{"classic", "modern", "technical", "report", "plain", "handout"} {
		css, err := Theme(preset)
		if err != nil {
			t.Fatalf("Theme(%q) error: %v", preset, err)
		}
		s := string(css)
		if !strings.Contains(s, "@font-face") {
			t.Errorf("Theme(%q) should contain fonts.css (@font-face)", preset)
		}
		if !strings.Contains(s, "callout") {
			t.Errorf("Theme(%q) should contain base.css (callout)", preset)
		}
		if !strings.Contains(s, "--font-body") {
			t.Errorf("Theme(%q) should contain the preset stylesheet (--font-body)", preset)
		}
		if !strings.Contains(s, "chroma") {
			t.Errorf("Theme(%q) should contain code.css (chroma)", preset)
		}
	}

	if _, err := Theme("nonexistent"); err == nil {
		t.Errorf("Theme(nonexistent) expected error, got nil")
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
