package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gstuebner/md2pdf/internal/config"
)

func TestPresetFlagIsPassedThrough(t *testing.T) {
	o, err := runCmd(t, []string{"input.md", "--preset", "modern"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Preset != "modern" {
		t.Errorf("Preset = %q, want modern", o.Preset)
	}
}

// An abbreviated preset is resolved in the CLI, so the rest of the program
// only sees full names.
func TestPresetShorthandAndAbbreviation(t *testing.T) {
	o, err := runCmd(t, []string{"input.md", "-p", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Preset != "classic" {
		t.Errorf("Preset = %q, want classic", o.Preset)
	}

	o, err = runCmd(t, []string{"input.md", "--preset", "tech"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Preset != "technical" {
		t.Errorf("Preset = %q, want technical", o.Preset)
	}
}

func TestUnknownPresetIsAUsageError(t *testing.T) {
	_, err := runCmd(t, []string{"input.md", "--preset", "nope"})
	if err == nil {
		t.Fatal("expected an error for an unknown preset")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Errorf("error = %v, want it to mention the unknown preset", err)
	}
}

// The preset only fills in options the user left alone, so the CLI has to
// record which of them appeared on the command line.
func TestExplicitFlagsAreRecorded(t *testing.T) {
	o, err := runCmd(t, []string{"input.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Explicit != (config.Explicit{}) {
		t.Errorf("without flags nothing should be marked explicit, got %+v", o.Explicit)
	}

	o, err = runCmd(t, []string{
		"input.md",
		"--no-cover", "--no-toc", "--toc-depth", "4", "--no-numbering",
		"--chapter-pages", "--landscape", "--no-header", "--no-footer",
		"--margin", "10mm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := config.Explicit{
		Cover: true, TOC: true, TOCDepth: true, NumberHeadings: true,
		ChapterPages: true, Landscape: true, Header: true, Footer: true, Margins: true,
	}
	if o.Explicit != want {
		t.Errorf("Explicit = %+v, want %+v", o.Explicit, want)
	}
}

func TestListPresets(t *testing.T) {
	withStubPipeline(t, nil)

	o := config.DefaultOptions()
	c := newRootCmd(&o)
	var out bytes.Buffer
	c.SetArgs([]string{"--list-presets"})
	c.SetOut(&out)
	c.SetErr(&bytes.Buffer{})

	if err := c.Execute(); err != nil {
		t.Fatalf("--list-presets returned an error: %v", err)
	}
	for _, name := range config.PresetNames() {
		if !strings.Contains(out.String(), name) {
			t.Errorf("--list-presets does not mention %q:\n%s", name, out.String())
		}
	}
}
