package browser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"github.com/gstuebner/md2pdf/internal/config"
)

// ErrRenderTimeout is returned by Print when the document does not signal
// readiness (window.__md2pdfReady === true) within PrintOptions.Timeout.
var ErrRenderTimeout = errors.New("Rendern hat das Zeitlimit überschritten")

// PrintOptions holds everything Print needs to render an HTML document to
// PDF via Page.PrintToPDF.
type PrintOptions struct {
	HTML []byte

	Paper     config.Paper
	Margins   config.Margins
	Landscape bool

	HeaderHTML          string
	FooterHTML          string
	DisplayHeaderFooter bool

	Outline bool

	Timeout time.Duration

	BrowserPath string
	BrowserArgs []string
}

// Print renders o.HTML through a headless Chromium instance and returns the
// resulting PDF bytes.
func Print(ctx context.Context, o PrintOptions) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "md2pdf-print-*")
	if err != nil {
		return nil, fmt.Errorf("temporäres Verzeichnis konnte nicht angelegt werden: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "document.html")
	if err := os.WriteFile(htmlPath, o.HTML, 0o644); err != nil {
		return nil, fmt.Errorf("temporäre HTML-Datei konnte nicht geschrieben werden: %w", err)
	}

	absPath, err := filepath.Abs(htmlPath)
	if err != nil {
		return nil, fmt.Errorf("absoluter Pfad zur HTML-Datei konnte nicht ermittelt werden: %w", err)
	}
	fileURL := "file://" + filepath.ToSlash(absPath)

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(o.BrowserPath),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("font-render-hinting", "none"),
		chromedp.Flag("disable-lcd-text", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("allow-file-access-from-files", true),
	)
	if os.Geteuid() == 0 || os.Getenv("MD2PDF_NO_SANDBOX") != "" {
		allocOpts = append(allocOpts, chromedp.Flag("no-sandbox", true))
	}
	for _, arg := range o.BrowserArgs {
		allocOpts = append(allocOpts, browserArgFlag(arg))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	var pdfData []byte
	runErr := chromedp.Run(taskCtx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Poll("window.__md2pdfReady === true", nil, chromedp.WithPollingTimeout(o.Timeout)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			data, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(o.Paper.WidthIn).
				WithPaperHeight(o.Paper.HeightIn).
				WithMarginTop(o.Margins.TopIn).
				WithMarginRight(o.Margins.RightIn).
				WithMarginBottom(o.Margins.BottomIn).
				WithMarginLeft(o.Margins.LeftIn).
				WithLandscape(o.Landscape).
				WithDisplayHeaderFooter(o.DisplayHeaderFooter).
				WithHeaderTemplate(o.HeaderHTML).
				WithFooterTemplate(o.FooterHTML).
				WithGenerateTaggedPDF(true).
				WithGenerateDocumentOutline(o.Outline).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfData = data
			return nil
		}),
	)
	if runErr != nil {
		if errors.Is(runErr, chromedp.ErrPollingTimeout) {
			return nil, fmt.Errorf("Dokument hat window.__md2pdfReady nicht innerhalb von %s gesetzt: %w", o.Timeout, ErrRenderTimeout)
		}
		return nil, fmt.Errorf("Chromium-Rendering fehlgeschlagen: %w", runErr)
	}

	return pdfData, nil
}

// browserArgFlag turns a raw "--name", "--name=value", "name" or "name=value"
// user-supplied argument into an ExecAllocatorOption via chromedp.Flag.
func browserArgFlag(arg string) chromedp.ExecAllocatorOption {
	name := strings.TrimLeft(arg, "-")
	if idx := strings.IndexByte(name, '='); idx >= 0 {
		return chromedp.Flag(name[:idx], name[idx+1:])
	}
	return chromedp.Flag(name, true)
}
