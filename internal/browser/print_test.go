package browser

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/gstuebner/md2pdf/internal/config"
)

func TestPrint_RendersTrivialHTML(t *testing.T) {
	browserPath, err := Find("")
	if err != nil {
		t.Skipf("no browser found, skipping: %v", err)
	}

	html := []byte(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Test</title></head>
<body>
<p>Hello, PDF!</p>
<script>window.__md2pdfReady = true;</script>
</body>
</html>`)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pdf, err := Print(ctx, PrintOptions{
		HTML:                html,
		Paper:               config.DefaultPaper,
		Margins:             config.DefaultMargins,
		DisplayHeaderFooter: false,
		Outline:             true,
		Timeout:             10 * time.Second,
		BrowserPath:         browserPath,
	})
	if err != nil {
		t.Fatalf("Print returned unexpected error: %v", err)
	}

	if len(pdf) == 0 {
		t.Fatal("Print returned empty PDF bytes")
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("Print output does not start with %%PDF-, got: %q", pdf[:min(16, len(pdf))])
	}
}
