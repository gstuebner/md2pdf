package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gstuebner/md2pdf/internal/browser"
	"github.com/gstuebner/md2pdf/internal/config"
	"github.com/gstuebner/md2pdf/internal/mdconv"
	"github.com/gstuebner/md2pdf/internal/model"
	"github.com/gstuebner/md2pdf/internal/render"
)

// browserTimeoutHeadroom is added to the render timeout so that a browser that
// hangs before the document is even loaded cannot block forever.
const browserTimeoutHeadroom = 30 * time.Second

func run(ctx context.Context, o config.Options) (Result, error) {
	start := time.Now()

	src, err := os.ReadFile(o.Input)
	if err != nil {
		return Result{}, fmt.Errorf("Eingabedatei konnte nicht gelesen werden: %w", err)
	}

	loader := &render.FileLoader{BaseDir: filepath.Dir(o.Input)}

	converted, err := mdconv.Convert(src, mdconv.Options{
		TOCDepth: o.TOCDepth,
		Loader:   loader,
		Lang:     o.Meta.Lang,
	})
	if err != nil {
		return Result{}, err
	}

	var notes []string

	// Ein Dokument, das seine Kapitel selbst nummeriert, darf keine zweite
	// Nummer vom Theme bekommen. ForceNumbering hebelt die Erkennung aus.
	if o.NumberHeadings && !o.ForceNumbering && converted.ManualNumbering {
		o.NumberHeadings = false
		notes = append(notes, "eigene Kapitelnummern erkannt, automatische Nummerierung ausgeschaltet (--force-numbering erzwingt sie)")
	}

	meta := applyMetaDefaults(mergeMeta(converted.Meta, o.Meta), loader)

	doc := model.Document{
		Meta:       meta,
		TOC:        converted.TOC,
		BodyHTML:   converted.BodyHTML,
		HasMermaid: converted.HasMermaid,
	}

	html, err := render.Document(doc, o)
	if err != nil {
		return Result{}, err
	}

	if o.HTMLOut != "" {
		if err := os.WriteFile(o.HTMLOut, html, 0o644); err != nil {
			return Result{}, fmt.Errorf("HTML konnte nicht geschrieben werden: %w", err)
		}
	}

	header, footer, err := runningHeads(meta, o)
	if err != nil {
		return Result{}, err
	}

	browserPath, err := browser.Find(o.BrowserPath)
	if err != nil {
		return Result{}, fmt.Errorf("%w", ErrBrowserNotFound)
	}

	printCtx, cancel := context.WithTimeout(ctx, o.Timeout+browserTimeoutHeadroom)
	defer cancel()

	pdf, err := browser.Print(printCtx, browser.PrintOptions{
		HTML:                html,
		Paper:               o.Paper,
		Margins:             o.Margins,
		Landscape:           o.Landscape,
		HeaderHTML:          header,
		FooterHTML:          footer,
		DisplayHeaderFooter: o.Header || o.Footer,
		Outline:             o.Outline,
		Timeout:             o.Timeout,
		BrowserPath:         browserPath,
		BrowserArgs:         o.BrowserArgs,
	})
	if err != nil {
		if errors.Is(err, browser.ErrRenderTimeout) {
			return Result{}, fmt.Errorf("%w", ErrRenderTimeout)
		}
		return Result{}, err
	}

	if err := os.WriteFile(o.Output, pdf, 0o644); err != nil {
		return Result{}, fmt.Errorf("PDF konnte nicht geschrieben werden: %w", err)
	}

	return Result{
		OutputPath: o.Output,
		Size:       int64(len(pdf)),
		Duration:   time.Since(start),
		Notes:      notes,
	}, nil
}

// runningHeads renders the Chromium header and footer templates. A disabled
// half stays empty so that Chromium draws nothing there.
func runningHeads(meta model.Meta, o config.Options) (header, footer string, err error) {
	if o.Header {
		header, err = render.Header(meta, o)
		if err != nil {
			return "", "", err
		}
	}
	if o.Footer {
		footer, err = render.Footer(meta, o)
		if err != nil {
			return "", "", err
		}
	}
	// With displayHeaderFooter on, an empty template makes Chromium fall back to
	// its own default; a blank element keeps the margin box empty instead.
	if header == "" {
		header = "<span></span>"
	}
	if footer == "" {
		footer = "<span></span>"
	}
	return header, footer, nil
}

// mergeMeta lets non-empty CLI values win over the front matter.
func mergeMeta(fromDoc, fromFlags model.Meta) model.Meta {
	out := fromDoc
	for _, f := range []struct {
		dst *string
		val string
	}{
		{&out.Title, fromFlags.Title},
		{&out.Subtitle, fromFlags.Subtitle},
		{&out.Kicker, fromFlags.Kicker},
		{&out.Version, fromFlags.Version},
		{&out.Author, fromFlags.Author},
		{&out.Company, fromFlags.Company},
		{&out.Date, fromFlags.Date},
		{&out.Logo, fromFlags.Logo},
		{&out.Lang, fromFlags.Lang},
	} {
		if f.val != "" {
			*f.dst = f.val
		}
	}
	return out
}

func applyMetaDefaults(m model.Meta, loader render.AssetLoader) model.Meta {
	if m.Lang == "" {
		m.Lang = "de"
	}
	if m.Kicker == "" {
		m.Kicker = "Dokumentation"
	}
	if m.Date == "" {
		m.Date = time.Now().Format("02.01.2006")
	}
	if m.Logo != "" {
		if uri, err := loader.DataURI(m.Logo); err == nil {
			m.Logo = uri
		}
	}
	return m
}
