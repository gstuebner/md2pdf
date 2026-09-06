package config

import "testing"

func TestParsePreset(t *testing.T) {
	for _, name := range PresetNames() {
		p, err := ParsePreset(name)
		if err != nil {
			t.Fatalf("ParsePreset(%q): %v", name, err)
		}
		if p.Name != name {
			t.Errorf("ParsePreset(%q).Name = %q", name, p.Name)
		}
		if p.Description == "" {
			t.Errorf("preset %q has no description", name)
		}
		if _, err := ParseMargins(p.Margin); err != nil {
			t.Errorf("preset %q has an unparsable margin %q: %v", name, p.Margin, err)
		}
	}

	if _, err := ParsePreset("MODERN"); err != nil {
		t.Errorf("ParsePreset should be case-insensitive: %v", err)
	}
	if _, err := ParsePreset("nope"); err == nil {
		t.Error("ParsePreset(nope) expected an error, got nil")
	}
}

func TestParsePresetAcceptsPrefixes(t *testing.T) {
	cases := map[string]string{
		"c":    "classic",
		"m":    "modern",
		"t":    "technical",
		"r":    "report",
		"p":    "plain",
		"h":    "handout",
		"cla":  "classic",
		"REP":  "report",
		" te ": "technical",
	}
	for in, want := range cases {
		got, err := ParsePreset(in)
		if err != nil {
			t.Errorf("ParsePreset(%q): %v", in, err)
			continue
		}
		if got.Name != want {
			t.Errorf("ParsePreset(%q) = %q, want %q", in, got.Name, want)
		}
	}

	if _, err := ParsePreset(""); err == nil {
		t.Error("ParsePreset(\"\") expected an error, got nil")
	}
}

// Every preset must stay reachable by its first letter — that is what makes
// "-p c" work, and a new preset starting with the same letter would break it.
func TestPresetInitialsAreUnique(t *testing.T) {
	seen := map[string]string{}
	for _, p := range Presets {
		initial := p.Name[:1]
		if other, ok := seen[initial]; ok {
			t.Errorf("presets %q and %q share the initial %q, so -p %s is ambiguous", other, p.Name, initial, initial)
		}
		seen[initial] = p.Name
	}
}

func TestDefaultPresetExists(t *testing.T) {
	if _, err := ParsePreset(DefaultPreset); err != nil {
		t.Fatalf("DefaultPreset %q is not registered: %v", DefaultPreset, err)
	}
}

func TestPresetApplyFillsDefaults(t *testing.T) {
	p, err := ParsePreset("modern")
	if err != nil {
		t.Fatal(err)
	}

	o := DefaultOptions()
	if err := p.Apply(&o); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if o.Preset != "modern" {
		t.Errorf("Preset = %q, want modern", o.Preset)
	}
	if !o.Cover || !o.TOC || !o.NumberHeadings {
		t.Errorf("modern should switch cover, TOC and numbering on, got %+v", o)
	}
	if o.TOCDepth != 3 {
		t.Errorf("TOCDepth = %d, want 3", o.TOCDepth)
	}
}

func TestPresetApplyKeepsExplicitOptions(t *testing.T) {
	p, err := ParsePreset("classic") // cover, TOC and numbering all off
	if err != nil {
		t.Fatal(err)
	}

	o := DefaultOptions()
	o.Cover = true
	o.TOC = true
	o.TOCDepth = 5
	o.NumberHeadings = true
	o.ChapterPages = true
	o.Landscape = true
	o.Header = false
	o.Footer = false
	o.Margins = Margins{TopIn: 1, RightIn: 1, BottomIn: 1, LeftIn: 1}
	o.Explicit = Explicit{
		Cover: true, TOC: true, TOCDepth: true, NumberHeadings: true,
		ChapterPages: true, Landscape: true, Header: true, Footer: true, Margins: true,
	}

	if err := p.Apply(&o); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if !o.Cover || !o.TOC || !o.NumberHeadings || !o.ChapterPages || !o.Landscape {
		t.Errorf("explicit switches were overwritten by the preset: %+v", o)
	}
	if o.Header || o.Footer {
		t.Error("explicitly disabled header/footer were re-enabled by the preset")
	}
	if o.TOCDepth != 5 {
		t.Errorf("TOCDepth = %d, want the explicit 5", o.TOCDepth)
	}
	if o.Margins.TopIn != 1 {
		t.Errorf("explicit margins were overwritten: %+v", o.Margins)
	}
}
