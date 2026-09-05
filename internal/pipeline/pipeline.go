// Package pipeline wires the conversion steps (Markdown parsing, HTML
// rendering, and PDF printing) together and exposes a single entry point
// for the CLI.
package pipeline

import (
	"context"
	"errors"
	"time"

	"github.com/gstuebner/md2pdf/internal/config"
)

// Result describes the outcome of a successful conversion.
type Result struct {
	OutputPath string
	Size       int64
	Duration   time.Duration
	// Notes holds remarks about decisions md2pdf made on its own, for the CLI
	// to show. They are not errors.
	Notes []string
}

var (
	// ErrBrowserNotFound indicates that no usable Chromium-based browser
	// engine could be located on the system.
	ErrBrowserNotFound = errors.New("no chromium-based browser found")
	// ErrRenderTimeout indicates that the headless browser did not finish
	// rendering the document within the configured timeout.
	ErrRenderTimeout = errors.New("rendering timed out")
)

// BrowserNotFoundHelp is printed to stderr when ErrBrowserNotFound occurs.
const BrowserNotFoundHelp = `md2pdf: keine Chromium-basierte Browser-Engine gefunden.

Installiere eine davon oder gib den Pfad explizit an:
  Arch/CachyOS:  paru -S chromium
  Debian/Ubuntu: sudo apt install chromium
  Windows:       Microsoft Edge ist vorinstalliert
  manuell:       md2pdf --browser-path /pfad/zu/chrome`

// Run converts the Markdown file described by o into a PDF.
func Run(ctx context.Context, o config.Options) (Result, error) {
	return run(ctx, o)
}
