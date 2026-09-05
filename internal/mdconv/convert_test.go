package mdconv

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden file")

// stubLoader keeps the golden file readable by not inlining real image bytes.
type stubLoader struct{}

func (stubLoader) DataURI(src string) (string, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		return src, nil
	}
	return "data:image/svg+xml;base64,STUB", nil
}

func TestConvertShowcaseGolden(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "showcase.md"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := Convert(src, Options{TOCDepth: 3, Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}

	golden := filepath.Join("..", "..", "testdata", "golden", "showcase.html")
	if *update {
		if err := os.WriteFile(golden, []byte(res.BodyHTML), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Log("golden file rewritten")
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("golden file missing, run: go test ./internal/mdconv/ -update (%v)", err)
	}
	if string(want) != string(res.BodyHTML) {
		t.Errorf("body HTML differs from golden file; run with -update to inspect the change")
	}
}

func TestConvertShowcaseMetadata(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "showcase.md"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := Convert(src, Options{TOCDepth: 3, Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}

	if res.Meta.Title != "Aurora Desk" {
		t.Errorf("Title = %q, want %q", res.Meta.Title, "Aurora Desk")
	}
	if res.Meta.Version != "2.4.0" {
		t.Errorf("Version = %q, want %q", res.Meta.Version, "2.4.0")
	}
	if res.Meta.Kicker != "Handbuch" {
		t.Errorf("Kicker = %q, want %q", res.Meta.Kicker, "Handbuch")
	}
	if !res.HasMermaid {
		t.Error("HasMermaid = false, want true")
	}
	if strings.Contains(string(res.BodyHTML), "<h1") {
		t.Error("leading H1 should have been moved to the cover, but is still in the body")
	}

	if len(res.TOC) == 0 {
		t.Fatal("TOC is empty")
	}
	first := res.TOC[0]
	if first.Level != 2 || first.Text != "Einführung" {
		t.Errorf("first TOC entry = %+v, want level 2 %q", first, "Einführung")
	}
	for _, e := range res.TOC {
		if e.ID == "" {
			t.Errorf("TOC entry %q has no anchor", e.Text)
		}
		if e.Level > 3 {
			t.Errorf("TOC entry %q exceeds depth 3", e.Text)
		}
	}
}
