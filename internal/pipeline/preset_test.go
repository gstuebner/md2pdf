package pipeline

import (
	"testing"

	"github.com/gstuebner/md2pdf/internal/config"
	"github.com/gstuebner/md2pdf/internal/model"
)

func TestApplyPresetPrecedence(t *testing.T) {
	cases := []struct {
		name     string
		flag     string
		frontMat string
		want     string
	}{
		{"nothing given", "", "", config.DefaultPreset},
		{"front matter decides", "", "modern", "modern"},
		{"flag beats front matter", "report", "modern", "report"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := config.DefaultOptions()
			o.Preset = tc.flag
			if err := applyPreset(&o, model.Meta{Preset: tc.frontMat}); err != nil {
				t.Fatalf("applyPreset: %v", err)
			}
			if o.Preset != tc.want {
				t.Errorf("Preset = %q, want %q", o.Preset, tc.want)
			}
		})
	}
}

func TestApplyPresetRejectsUnknownName(t *testing.T) {
	o := config.DefaultOptions()
	if err := applyPreset(&o, model.Meta{Preset: "nope"}); err == nil {
		t.Fatal("expected an error for an unknown preset in the front matter")
	}
}

func TestApplyPresetSetsStructure(t *testing.T) {
	o := config.DefaultOptions()
	o.Preset = "classic"
	if err := applyPreset(&o, model.Meta{}); err != nil {
		t.Fatalf("applyPreset: %v", err)
	}
	if o.Cover || o.TOC || o.NumberHeadings {
		t.Errorf("classic should leave cover, TOC and numbering off, got %+v", o)
	}

	o = config.DefaultOptions()
	o.Preset = "handout"
	if err := applyPreset(&o, model.Meta{}); err != nil {
		t.Fatalf("applyPreset: %v", err)
	}
	if !o.Landscape || !o.ChapterPages {
		t.Errorf("handout should be landscape with one chapter per page, got %+v", o)
	}
}
