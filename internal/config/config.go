package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gstuebner/md2pdf/internal/model"
)

type Paper struct{ WidthIn, HeightIn float64 }
type Margins struct{ TopIn, RightIn, BottomIn, LeftIn float64 }

type Options struct {
	Input, Output string
	Meta          model.Meta // aus CLI-Flags; überschreibt Frontmatter feldweise
	ExtraCSS      []string

	TOC      bool // Default true
	TOCDepth int  // Default 3
	Cover    bool // Default true

	NumberHeadings bool // Default true  -> <body class="numbered">
	ForceNumbering bool // Default false -> Erkennung eigener Kapitelnummern übergehen
	ChapterPages   bool // Default false -> <body class="chapters">

	Paper     Paper
	Landscape bool
	Margins   Margins

	Header     bool   // Default true
	Footer     bool   // Default true
	HeaderFile string // optionales eigenes Chromium-Template
	FooterFile string

	BrowserPath string
	BrowserArgs []string

	HTMLOut string
	Timeout time.Duration // Default 60s
	Outline bool          // PDF-Lesezeichen, Default true
	Quiet   bool
}

// Predefined paper sizes in inches.
var PaperSizes = map[string]Paper{
	"a4":     {WidthIn: 8.27, HeightIn: 11.69},
	"a5":     {WidthIn: 5.83, HeightIn: 8.27},
	"letter": {WidthIn: 8.5, HeightIn: 11.0},
	"legal":  {WidthIn: 8.5, HeightIn: 14.0},
}

// DefaultPaper is A4.
var DefaultPaper = PaperSizes["a4"]

// DefaultMargins: top 25mm, right/bottom/left 20mm.
var DefaultMargins = Margins{
	TopIn:    25.0 / 25.4,
	RightIn:  20.0 / 25.4,
	BottomIn: 20.0 / 25.4,
	LeftIn:   20.0 / 25.4,
}

// DefaultOptions returns an Options struct populated with default values.
func DefaultOptions() Options {
	return Options{
		TOC:            true,
		TOCDepth:       3,
		Cover:          true,
		NumberHeadings: true,
		ChapterPages:   false,
		Paper:          DefaultPaper,
		Landscape:      false,
		Margins:        DefaultMargins,
		Header:         true,
		Footer:         true,
		Timeout:        60 * time.Second,
		Outline:        true,
		Quiet:          false,
	}
}

// ParsePaper resolves a paper format string (case-insensitive) to Paper dimensions.
func ParsePaper(s string) (Paper, error) {
	key := strings.ToLower(strings.TrimSpace(s))
	if p, ok := PaperSizes[key]; ok {
		return p, nil
	}
	return Paper{}, fmt.Errorf("unknown paper format: %q (supported: A4, A5, Letter, Legal)", s)
}

// ParseMargins parses margin strings with 1, 2, or 4 values (e.g. "20mm", "25mm 20mm", "25mm 20mm 20mm 20mm").
// Units supported: mm, cm, in, pt (defaults to mm if no unit specified).
func ParseMargins(s string) (Margins, error) {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 {
		return Margins{}, fmt.Errorf("margins specification is empty")
	}

	values := make([]float64, len(fields))
	for i, f := range fields {
		val, err := parseDimensionToInches(f)
		if err != nil {
			return Margins{}, fmt.Errorf("invalid margin value %q: %w", f, err)
		}
		if val < 0 {
			return Margins{}, fmt.Errorf("margin value cannot be negative: %q", f)
		}
		values[i] = val
	}

	switch len(values) {
	case 1:
		return Margins{
			TopIn:    values[0],
			RightIn:  values[0],
			BottomIn: values[0],
			LeftIn:   values[0],
		}, nil
	case 2:
		// 2 values: vertical horizontal
		return Margins{
			TopIn:    values[0],
			RightIn:  values[1],
			BottomIn: values[0],
			LeftIn:   values[1],
		}, nil
	case 4:
		// 4 values: top right bottom left
		return Margins{
			TopIn:    values[0],
			RightIn:  values[1],
			BottomIn: values[2],
			LeftIn:   values[3],
		}, nil
	default:
		return Margins{}, fmt.Errorf("margins specification requires 1, 2, or 4 values, got %d", len(values))
	}
}

func parseDimensionToInches(s string) (float64, error) {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)

	var numStr string
	var unit string

	if strings.HasSuffix(lower, "mm") {
		numStr = strings.TrimSpace(s[:len(s)-2])
		unit = "mm"
	} else if strings.HasSuffix(lower, "cm") {
		numStr = strings.TrimSpace(s[:len(s)-2])
		unit = "cm"
	} else if strings.HasSuffix(lower, "in") {
		numStr = strings.TrimSpace(s[:len(s)-2])
		unit = "in"
	} else if strings.HasSuffix(lower, "pt") {
		numStr = strings.TrimSpace(s[:len(s)-2])
		unit = "pt"
	} else {
		numStr = s
		unit = "mm"
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse number in %q: %w", s, err)
	}

	switch unit {
	case "mm":
		return val / 25.4, nil
	case "cm":
		return (val * 10.0) / 25.4, nil
	case "in":
		return val, nil
	case "pt":
		return val / 72.0, nil
	default:
		return 0, fmt.Errorf("unsupported unit in %q", s)
	}
}
