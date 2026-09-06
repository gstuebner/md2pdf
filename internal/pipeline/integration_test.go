package pipeline

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gstuebner/md2pdf/internal/browser"
	"github.com/gstuebner/md2pdf/internal/config"
)

// showcasePath is the kitchen-sink document that exercises every supported
// Markdown feature.
func showcasePath(t *testing.T) string {
	t.Helper()
	p := filepath.Join("..", "..", "testdata", "showcase.md")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("showcase document missing: %v", err)
	}
	return p
}

func requireBrowser(t *testing.T) {
	t.Helper()
	if _, err := browser.Find(""); err != nil {
		t.Skip("no Chromium found, skipping integration test")
	}
}

func TestRunProducesPDF(t *testing.T) {
	requireBrowser(t)

	out := filepath.Join(t.TempDir(), "showcase.pdf")
	o := config.DefaultOptions()
	o.Input = showcasePath(t)
	o.Output = out
	o.Timeout = 60 * time.Second

	res, err := Run(context.Background(), o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if res.OutputPath != out {
		t.Errorf("OutputPath = %q, want %q", res.OutputPath, out)
	}
	if res.Duration <= 0 {
		t.Error("Duration not measured")
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Errorf("output is not a PDF, starts with %q", data[:min(8, len(data))])
	}
	if len(data) < 20*1024 {
		t.Errorf("PDF is only %d bytes, expected more than 20 KB", len(data))
	}
	if int64(len(data)) != res.Size {
		t.Errorf("Size = %d, file has %d bytes", res.Size, len(data))
	}
}

func TestRunWritesDebugHTML(t *testing.T) {
	requireBrowser(t)

	dir := t.TempDir()
	o := config.DefaultOptions()
	o.Input = showcasePath(t)
	o.Output = filepath.Join(dir, "out.pdf")
	o.HTMLOut = filepath.Join(dir, "out.html")
	o.Preset = "modern"
	o.Timeout = 60 * time.Second

	if _, err := Run(context.Background(), o); err != nil {
		t.Fatalf("Run: %v", err)
	}

	html, err := os.ReadFile(o.HTMLOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="cover"`,
		`class="toc"`,
		`class="callout callout-warning"`,
		`class="codeblock"`,
		`<pre class="mermaid">`,
		`data:image/svg+xml;base64,`,
		`window.__md2pdfReady`,
	} {
		if !bytes.Contains(html, []byte(want)) {
			t.Errorf("generated HTML is missing %q", want)
		}
	}
}

func TestRunReportsMissingBrowser(t *testing.T) {
	o := config.DefaultOptions()
	o.Input = showcasePath(t)
	o.Output = filepath.Join(t.TempDir(), "out.pdf")
	o.BrowserPath = filepath.Join(t.TempDir(), "does-not-exist")

	_, err := Run(context.Background(), o)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, ErrBrowserNotFound) {
		t.Errorf("error = %v, want ErrBrowserNotFound", err)
	}
}

func TestRunReportsMissingInput(t *testing.T) {
	o := config.DefaultOptions()
	o.Input = filepath.Join(t.TempDir(), "fehlt.md")
	o.Output = filepath.Join(t.TempDir(), "out.pdf")

	if _, err := Run(context.Background(), o); err == nil {
		t.Fatal("expected an error for a missing input file")
	}
}

func TestManualNumberingDisablesAutoNumbers(t *testing.T) {
	requireBrowser(t)

	dir := t.TempDir()
	input := filepath.Join(dir, "handbuch.md")
	src := "# Handbuch\n\n## 1. Einleitung\n\nText.\n\n## 2. Installation\n\nText.\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	o := config.DefaultOptions()
	o.Input = input
	o.Output = filepath.Join(dir, "out.pdf")
	o.HTMLOut = filepath.Join(dir, "out.html")
	o.Preset = "modern"
	o.Timeout = 60 * time.Second

	res, err := Run(context.Background(), o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	html, err := os.ReadFile(o.HTMLOut)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(html, []byte(`class="numbered`)) {
		t.Error("automatic numbering stayed on although the document brings its own numbers")
	}
	if len(res.Notes) == 0 {
		t.Error("no note about the disabled numbering")
	}
}

func TestForceNumberingOverridesDetection(t *testing.T) {
	requireBrowser(t)

	dir := t.TempDir()
	input := filepath.Join(dir, "handbuch.md")
	src := "# Handbuch\n\n## 1. Einleitung\n\nText.\n\n## 2. Installation\n\nText.\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	o := config.DefaultOptions()
	o.Input = input
	o.Output = filepath.Join(dir, "out.pdf")
	o.HTMLOut = filepath.Join(dir, "out.html")
	o.ForceNumbering = true
	o.Preset = "modern"
	o.Timeout = 60 * time.Second

	res, err := Run(context.Background(), o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	html, err := os.ReadFile(o.HTMLOut)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(html, []byte(`class="numbered`)) {
		t.Error("--force-numbering did not override the detection")
	}
	if len(res.Notes) != 0 {
		t.Errorf("unexpected note: %v", res.Notes)
	}
}
