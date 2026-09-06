// Package cmd implements the md2pdf command-line interface.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gstuebner/md2pdf/internal/config"
	"github.com/gstuebner/md2pdf/internal/pipeline"
	"github.com/spf13/cobra"
)

// version holds the program version. build.sh and the Makefile set it from
// the VERSION file at the repository root via -ldflags -X; a plain "go build"
// leaves it at "dev" and resolveVersion() then falls back to the module
// version the toolchain recorded.
var version = "dev"

// resolveVersion returns the version to display. A binary built with
// -ldflags -X reports exactly what was baked in. Without that — "go install
// github.com/gstuebner/md2pdf@v1.2.0", for instance — the module version from
// the build info is still better than "dev".
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return version
	}
	return info.Main.Version
}

// authors is shown in the help footer and alongside --version.
const authors = "Gregor Stübner & Claude (Anthropic)"

// credits is the one-line footer under the help output.
func credits() string {
	return fmt.Sprintf("md2pdf %s · %s", resolveVersion(), authors)
}

// errUsage marks errors caused by invalid CLI arguments (exit code 2).
var errUsage = errors.New("invalid arguments")

// usageHint follows every usage error, so that a bare "md2pdf" points the way
// to the help instead of leaving the reader with just a complaint.
const usageHint = "Run 'md2pdf -h' for usage and all available flags."

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
		presetFlag     string
		listPresets    bool
	)

	cmd := &cobra.Command{
		Use:   "md2pdf <input.md> [flags]",
		Short: "Convert a Markdown file into a print-ready PDF",
		Args: func(cmd *cobra.Command, args []string) error {
			if showVersion || listPresets {
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("%w: expected exactly one Markdown file, got %d", errUsage, len(args))
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
			if listPresets {
				fmt.Fprint(cmd.OutOrStdout(), presetList())
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

			// The style preset supplies the structural defaults, but only for
			// options the user did not set. Everything named on the command
			// line is recorded here and stays untouched.
			flags := cmd.Flags()
			o.Preset = presetFlag
			o.Explicit = config.Explicit{
				Cover:          flags.Changed("no-cover"),
				TOC:            flags.Changed("no-toc") || flags.Changed("toc-depth"),
				TOCDepth:       flags.Changed("toc-depth"),
				NumberHeadings: flags.Changed("no-numbering"),
				ChapterPages:   flags.Changed("chapter-pages"),
				Landscape:      flags.Changed("landscape"),
				Header:         flags.Changed("no-header"),
				Footer:         flags.Changed("no-footer"),
				Margins:        flags.Changed("margin"),
			}

			o.TOC = !noTOC
			// Naming a depth means you want a table of contents, even under a
			// preset that leaves it off. --no-toc still wins.
			if flags.Changed("toc-depth") && !flags.Changed("no-toc") {
				o.TOC = true
			}
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

			if marginFlag != "" {
				margins, err := config.ParseMargins(marginFlag)
				if err != nil {
					return fmt.Errorf("%w: %v", errUsage, err)
				}
				o.Margins = margins
			}

			// An abbreviated --preset is resolved right here, so that the rest
			// of the program only ever sees a full preset name.
			if presetFlag != "" {
				preset, err := config.ParsePreset(presetFlag)
				if err != nil {
					return fmt.Errorf("%w: %v", errUsage, err)
				}
				o.Preset = preset.Name
			}

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
	flags.StringVarP(&outputFlag, "output", "o", "", "target PDF (default: input name with .pdf)")
	flags.StringVarP(&presetFlag, "preset", "p", "", "style preset: "+strings.Join(config.PresetNames(), "|")+"; an unambiguous prefix is enough (default: "+config.DefaultPreset+")")
	flags.BoolVar(&listPresets, "list-presets", false, "list the built-in style presets and exit")
	flags.StringArrayVar(&cssFlag, "css", nil, "extra CSS file, repeatable; applied after the preset")
	flags.StringVar(&titleFlag, "title", "", "overrides the front matter")
	flags.StringVar(&subtitleFlag, "subtitle", "", "overrides the front matter")
	flags.StringVar(&kickerFlag, "kicker", "", "overrides the front matter")
	flags.StringVar(&docVersionFlag, "doc-version", "", "document version (not the program version)")
	flags.StringVar(&authorFlag, "author", "", "overrides the front matter")
	flags.StringVar(&companyFlag, "company", "", "overrides the front matter")
	flags.StringVar(&dateFlag, "date", "", "overrides the front matter")
	flags.StringVar(&logoFlag, "logo", "", "overrides the front matter")
	flags.StringVar(&langFlag, "lang", "", "document language de|en, controls the labels in the PDF (default: front matter, else locale, else en)")
	flags.BoolVar(&noTOC, "no-toc", false, "no table of contents")
	flags.IntVar(&tocDepthFlag, "toc-depth", 0, "depth of the table of contents; implies a table of contents (preset default)")
	flags.BoolVar(&noCover, "no-cover", false, "no cover page")
	flags.BoolVar(&noNumbering, "no-numbering", false, "no chapter numbers")
	flags.BoolVar(&forceNumbering, "force-numbering", false, "number chapters even when the document brings its own numbers")
	flags.BoolVar(&chapterPages, "chapter-pages", false, "start every chapter (H2) on a new page")
	flags.StringVar(&paperFlag, "paper", "A4", "A4|A5|Letter|Legal")
	flags.BoolVar(&landscape, "landscape", false, "landscape orientation")
	flags.StringVar(&marginFlag, "margin", "", `"20mm" | "25mm 20mm" | "25mm 20mm 20mm 20mm" (preset default)`)
	flags.BoolVar(&noHeader, "no-header", false, "no running header")
	flags.BoolVar(&noFooter, "no-footer", false, "no footer, and therefore no page numbers")
	flags.StringVar(&headerTemplate, "header-template", "", "custom Chromium header template")
	flags.StringVar(&footerTemplate, "footer-template", "", "custom Chromium footer template")
	flags.StringVar(&browserPath, "browser-path", "", "path to the browser engine")
	flags.StringArrayVar(&browserArgs, "browser-arg", nil, "extra Chromium argument, repeatable")
	flags.StringVar(&htmlOut, "html-out", "", "also write the generated HTML (debugging)")
	flags.DurationVar(&timeoutFlag, "timeout", config.DefaultOptions().Timeout, "render timeout")
	flags.BoolVar(&noOutline, "no-outline", false, "no PDF bookmarks")
	flags.BoolVarP(&quiet, "quiet", "q", false, "no success message")
	flags.BoolVarP(&showVersion, "version", "v", false, "program version")

	return cmd
}

// presetList renders the --list-presets output: one line per preset, name and
// description in two columns.
func presetList() string {
	width := 0
	for _, p := range config.Presets {
		if len(p.Name) > width {
			width = len(p.Name)
		}
	}
	var b strings.Builder
	for _, p := range config.Presets {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, p.Name, p.Description)
	}
	return b.String()
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
		fmt.Fprintln(stderr, usageHint)
		return 2
	case errors.Is(err, pipeline.ErrBrowserNotFound):
		// Wer einen Pfad explizit angegeben hat, braucht keine Installationshinweise,
		// sondern die Information, dass genau dieser Pfad nicht funktioniert.
		if o.BrowserPath != "" {
			fmt.Fprintf(stderr, "md2pdf: the browser path %q cannot be used.\n", o.BrowserPath)
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
