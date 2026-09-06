package config

import (
	"fmt"
	"sort"
	"strings"
)

// DefaultPreset is used when neither --preset nor the front matter names one.
const DefaultPreset = "classic"

// Preset bundles a built-in stylesheet with the structural defaults that go
// with it. A preset only fills in what the user did not set explicitly; every
// flag on the command line wins over the preset.
type Preset struct {
	Name        string
	Description string

	Cover          bool
	TOC            bool
	TOCDepth       int
	NumberHeadings bool
	ChapterPages   bool
	Landscape      bool
	Header         bool
	Footer         bool

	// Margin is a margin specification in the syntax ParseMargins accepts.
	Margin string
}

// Presets lists the built-in presets in the order --list-presets prints them.
var Presets = []Preset{
	{
		Name:        "classic",
		Description: "plain office document, no cover or contents (default)",
		Cover:       false, TOC: false, TOCDepth: 3,
		NumberHeadings: false, ChapterPages: false, Landscape: false,
		Header: true, Footer: true,
		Margin: "20mm",
	},
	{
		Name:        "modern",
		Description: "the original md2pdf look: cover, contents, numbered chapters",
		Cover:       true, TOC: true, TOCDepth: 3,
		NumberHeadings: true, ChapterPages: false, Landscape: false,
		Header: true, Footer: true,
		Margin: "25mm 20mm 20mm 20mm",
	},
	{
		Name:        "technical",
		Description: "dense manual: deep contents, chapters on their own pages",
		Cover:       true, TOC: true, TOCDepth: 4,
		NumberHeadings: true, ChapterPages: true, Landscape: false,
		Header: true, Footer: true,
		Margin: "25mm 20mm 20mm 20mm",
	},
	{
		Name:        "report",
		Description: "business report: serif body text, wide margins, no numbering",
		Cover:       true, TOC: true, TOCDepth: 2,
		NumberHeadings: false, ChapterPages: false, Landscape: false,
		Header: true, Footer: true,
		Margin: "30mm 25mm 25mm 25mm",
	},
	{
		Name:        "plain",
		Description: "greyscale and almost unstyled, a base for your own --css",
		Cover:       false, TOC: false, TOCDepth: 3,
		NumberHeadings: false, ChapterPages: false, Landscape: false,
		Header: false, Footer: true,
		Margin: "20mm",
	},
	{
		Name:        "handout",
		Description: "landscape handout: large type, one chapter per page",
		Cover:       false, TOC: false, TOCDepth: 2,
		NumberHeadings: false, ChapterPages: true, Landscape: true,
		Header: false, Footer: true,
		Margin: "18mm",
	},
}

// PresetNames returns the preset names in registry order.
func PresetNames() []string {
	names := make([]string, len(Presets))
	for i, p := range Presets {
		names[i] = p.Name
	}
	return names
}

// ParsePreset looks a preset up by name, case-insensitively. An unambiguous
// prefix is enough, so "-p c" selects classic. An exact name always wins over
// a prefix match.
func ParsePreset(name string) (Preset, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return Preset{}, fmt.Errorf("no preset given (available: %s)", availablePresets())
	}

	var matches []Preset
	for _, p := range Presets {
		if p.Name == key {
			return p, nil
		}
		if strings.HasPrefix(p.Name, key) {
			matches = append(matches, p)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return Preset{}, fmt.Errorf("unknown preset: %q (available: %s)", name, availablePresets())
	default:
		names := make([]string, len(matches))
		for i, p := range matches {
			names[i] = p.Name
		}
		sort.Strings(names)
		return Preset{}, fmt.Errorf("ambiguous preset: %q matches %s", name, strings.Join(names, ", "))
	}
}

// availablePresets lists the preset names alphabetically, for error messages.
func availablePresets() string {
	sorted := append([]string(nil), PresetNames()...)
	sort.Strings(sorted)
	return strings.Join(sorted, ", ")
}

// Apply writes the preset's structural defaults into o, skipping every field
// the user set on the command line.
func (p Preset) Apply(o *Options) error {
	o.Preset = p.Name

	if !o.Explicit.Cover {
		o.Cover = p.Cover
	}
	if !o.Explicit.TOC {
		o.TOC = p.TOC
	}
	if !o.Explicit.TOCDepth {
		o.TOCDepth = p.TOCDepth
	}
	if !o.Explicit.NumberHeadings {
		o.NumberHeadings = p.NumberHeadings
	}
	if !o.Explicit.ChapterPages {
		o.ChapterPages = p.ChapterPages
	}
	if !o.Explicit.Landscape {
		o.Landscape = p.Landscape
	}
	if !o.Explicit.Header {
		o.Header = p.Header
	}
	if !o.Explicit.Footer {
		o.Footer = p.Footer
	}
	if !o.Explicit.Margins && p.Margin != "" {
		m, err := ParseMargins(p.Margin)
		if err != nil {
			return fmt.Errorf("preset %q has an invalid margin %q: %w", p.Name, p.Margin, err)
		}
		o.Margins = m
	}
	return nil
}
