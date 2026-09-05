// Package cmd implements the md2pdf command-line interface.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gstuebner/md2pdf/internal/config"
	"github.com/gstuebner/md2pdf/internal/pipeline"
	"github.com/spf13/cobra"
)

// version holds the program version, overridable via -ldflags -X.
var version = "dev"

// authors is shown in the help footer and alongside --version.
const authors = "Gregor Stübner & Claude (Anthropic)"

// credits is the one-line footer under the help output.
func credits() string {
	return fmt.Sprintf("md2pdf %s · %s", version, authors)
}

// errUsage marks errors caused by invalid CLI arguments (exit code 2).
var errUsage = errors.New("ungültige Argumente")

// runPipeline is a package-level indirection so tests can stub the pipeline
// without invoking a real browser.
var runPipeline = pipeline.Run

// newRootCmd builds the md2pdf root command. All flag values are written
// into o, which starts out already populated with config.DefaultOptions()
// by the caller. This split keeps the command construction testable without
// terminating a process.
func newRootCmd(o *config.Options) *cobra.Command {
	var (
		outputFlag     string
		cssFlag        []string
		titleFlag      string
		subtitleFlag   string
		kickerFlag     string
		docVersionFlag string
		authorFlag     string
		companyFlag    string
		dateFlag       string
		logoFlag       string
		langFlag       string
		noTOC          bool
		tocDepthFlag   int
		noCover        bool
		noNumbering    bool
		forceNumbering bool
		chapterPages   bool
		paperFlag      string
		landscape      bool
		marginFlag     string
		noHeader       bool
		noFooter       bool
		headerTemplate string
		footerTemplate string
		browserPath    string
		browserArgs    []string
		htmlOut        string
		timeoutFlag    time.Duration
		noOutline      bool
		quiet          bool
		showVersion    bool
	)

	cmd := &cobra.Command{
		Use:   "md2pdf <input.md> [flags]",
		Short: "Konvertiert eine Markdown-Datei in ein druckreifes PDF",
		Args: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("%w: genau eine Markdown-Datei erwartet, %d erhalten", errUsage, len(args))
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Fprintln(cmd.OutOrStdout(), credits())
				return nil
			}

			o.Input = args[0]
			o.Output = outputFlag
			if o.Output == "" {
				o.Output = deriveOutputPath(o.Input)
			}

			o.ExtraCSS = cssFlag

			if titleFlag != "" {
				o.Meta.Title = titleFlag
			}
			if subtitleFlag != "" {
				o.Meta.Subtitle = subtitleFlag
			}
			if kickerFlag != "" {
				o.Meta.Kicker = kickerFlag
			}
			if docVersionFlag != "" {
				o.Meta.Version = docVersionFlag
			}
			if authorFlag != "" {
				o.Meta.Author = authorFlag
			}
			if companyFlag != "" {
				o.Meta.Company = companyFlag
			}
			if dateFlag != "" {
				o.Meta.Date = dateFlag
			}
			if logoFlag != "" {
				o.Meta.Logo = logoFlag
			}
			if langFlag != "" {
				o.Meta.Lang = langFlag
			}

			o.TOC = !noTOC
			o.TOCDepth = tocDepthFlag
			o.Cover = !noCover
			o.NumberHeadings = !noNumbering
			o.ForceNumbering = forceNumbering
			o.ChapterPages = chapterPages

			paper, err := config.ParsePaper(paperFlag)
			if err != nil {
				return fmt.Errorf("%w: %v", errUsage, err)
			}
			o.Paper = paper
			o.Landscape = landscape

			margins, err := config.ParseMargins(marginFlag)
			if err != nil {
				return fmt.Errorf("%w: %v", errUsage, err)
			}
			o.Margins = margins

			o.Header = !noHeader
			o.Footer = !noFooter
			o.HeaderFile = headerTemplate
			o.FooterFile = footerTemplate

			o.BrowserPath = browserPath
			o.BrowserArgs = browserArgs

			o.HTMLOut = htmlOut
			o.Timeout = timeoutFlag
			o.Outline = !noOutline
			o.Quiet = quiet

			result, err := runPipeline(cmd.Context(), *o)
			if err != nil {
				return err
			}

			if !o.Quiet {
				for _, note := range result.Notes {
					fmt.Fprintf(cmd.ErrOrStderr(), "md2pdf: %s\n", note)
				}
				fmt.Fprintln(cmd.OutOrStdout(), formatSuccessLine(result))
			}
			return nil
		},
	}

	cmd.SetHelpTemplate(cmd.HelpTemplate() + "\n" + credits() + "\n")

	flags := cmd.Flags()
	flags.StringVarP(&outputFlag, "output", "o", "", "Ziel-PDF (Default: Eingabename mit .pdf)")
	flags.StringArrayVar(&cssFlag, "css", nil, "zusätzliche CSS-Datei(en), mehrfach angebbar")
	flags.StringVar(&titleFlag, "title", "", "überschreibt Frontmatter")
	flags.StringVar(&subtitleFlag, "subtitle", "", "überschreibt Frontmatter")
	flags.StringVar(&kickerFlag, "kicker", "", "überschreibt Frontmatter")
	flags.StringVar(&docVersionFlag, "doc-version", "", "Dokumentversion (nicht die Programmversion)")
	flags.StringVar(&authorFlag, "author", "", "überschreibt Frontmatter")
	flags.StringVar(&companyFlag, "company", "", "überschreibt Frontmatter")
	flags.StringVar(&dateFlag, "date", "", "überschreibt Frontmatter")
	flags.StringVar(&logoFlag, "logo", "", "überschreibt Frontmatter")
	flags.StringVar(&langFlag, "lang", "", "Dokumentsprache de|en, steuert Beschriftungen (Default: de)")
	flags.BoolVar(&noTOC, "no-toc", false, "kein Inhaltsverzeichnis")
	flags.IntVar(&tocDepthFlag, "toc-depth", config.DefaultOptions().TOCDepth, "Tiefe des Inhaltsverzeichnisses")
	flags.BoolVar(&noCover, "no-cover", false, "kein Deckblatt")
	flags.BoolVar(&noNumbering, "no-numbering", false, "keine Kapitelnummern")
	flags.BoolVar(&forceNumbering, "force-numbering", false, "auch dann nummerieren, wenn das Dokument eigene Kapitelnummern mitbringt")
	flags.BoolVar(&chapterPages, "chapter-pages", false, "jedes Kapitel (H2) auf neuer Seite")
	flags.StringVar(&paperFlag, "paper", "A4", "A4|A5|Letter|Legal")
	flags.BoolVar(&landscape, "landscape", false, "Querformat")
	flags.StringVar(&marginFlag, "margin", "25mm 20mm 20mm 20mm", `"20mm" | "25mm 20mm" | "25mm 20mm 20mm 20mm"`)
	flags.BoolVar(&noHeader, "no-header", false, "keine laufende Kopfzeile")
	flags.BoolVar(&noFooter, "no-footer", false, "keine Fußzeile (auch keine Seitenzahlen)")
	flags.StringVar(&headerTemplate, "header-template", "", "eigenes Chromium-Kopfzeilentemplate")
	flags.StringVar(&footerTemplate, "footer-template", "", "eigenes Chromium-Fußzeilentemplate")
	flags.StringVar(&browserPath, "browser-path", "", "Pfad zur Browser-Engine")
	flags.StringArrayVar(&browserArgs, "browser-arg", nil, "zusätzliches Chromium-Argument, mehrfach angebbar")
	flags.StringVar(&htmlOut, "html-out", "", "erzeugtes HTML zusätzlich speichern (Debug)")
	flags.DurationVar(&timeoutFlag, "timeout", config.DefaultOptions().Timeout, "Render-Timeout")
	flags.BoolVar(&noOutline, "no-outline", false, "keine PDF-Lesezeichen")
	flags.BoolVarP(&quiet, "quiet", "q", false, "keine Erfolgsmeldung")
	flags.BoolVarP(&showVersion, "version", "v", false, "Programmversion")

	return cmd
}

// deriveOutputPath replaces the input file's extension with .pdf, or
// appends .pdf if the input has no extension.
func deriveOutputPath(input string) string {
	ext := filepath.Ext(input)
	if ext == "" {
		return input + ".pdf"
	}
	return strings.TrimSuffix(input, ext) + ".pdf"
}

// formatSuccessLine renders the one-line success message, e.g.
// "handbuch.pdf · 842 KB · 1.9s".
func formatSuccessLine(r pipeline.Result) string {
	return fmt.Sprintf("%s · %s · %s", filepath.Base(r.OutputPath), formatSize(r.Size), formatDuration(r.Duration))
}

// formatSize renders a byte count as a human-readable KB/MB value with at
// most one decimal place, dropping a trailing ".0".
func formatSize(bytes int64) string {
	const kib = 1024.0
	const mib = kib * 1024.0

	switch {
	case float64(bytes) >= mib:
		return trimZeroDecimal(fmt.Sprintf("%.1f", float64(bytes)/mib)) + " MB"
	default:
		return trimZeroDecimal(fmt.Sprintf("%.1f", float64(bytes)/kib)) + " KB"
	}
}

func trimZeroDecimal(s string) string {
	return strings.TrimSuffix(s, ".0")
}

// formatDuration rounds d to the nearest 100 ms and renders it in seconds,
// e.g. "1.9s".
func formatDuration(d time.Duration) string {
	rounded := d.Round(100 * time.Millisecond)
	return fmt.Sprintf("%.1fs", rounded.Seconds())
}

// Execute runs the md2pdf CLI using the real process args and standard
// streams, returning the process exit code.
func Execute() int {
	return execute(os.Args[1:], os.Stdout, os.Stderr)
}

// execute drives the command with the given args and streams, keeping
// Execute testable without touching the real process state.
func execute(args []string, stdout, stderr io.Writer) int {
	o := config.DefaultOptions()
	cmd := newRootCmd(&o)
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err := cmd.Execute()
	if err == nil {
		return 0
	}

	switch {
	case errors.Is(err, errUsage):
		fmt.Fprintf(stderr, "md2pdf: %v\n", err)
		return 2
	case errors.Is(err, pipeline.ErrBrowserNotFound):
		// Wer einen Pfad explizit angegeben hat, braucht keine Installationshinweise,
		// sondern die Information, dass genau dieser Pfad nicht funktioniert.
		if o.BrowserPath != "" {
			fmt.Fprintf(stderr, "md2pdf: der angegebene Browser-Pfad %q ist nicht nutzbar.\n", o.BrowserPath)
		} else {
			fmt.Fprintln(stderr, pipeline.BrowserNotFoundHelp)
		}
		return 3
	case errors.Is(err, pipeline.ErrRenderTimeout):
		fmt.Fprintf(stderr, "md2pdf: %v\n", err)
		return 4
	default:
		fmt.Fprintf(stderr, "md2pdf: %v\n", err)
		return 1
	}
}
