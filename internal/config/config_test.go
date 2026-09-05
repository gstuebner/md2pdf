package config

import (
	"math"
	"testing"
)

func TestParsePaper(t *testing.T) {
	tests := []struct {
		input       string
		wantWidth   float64
		wantHeight  float64
		expectError bool
	}{
		{"A4", 8.27, 11.69, false},
		{"a4", 8.27, 11.69, false},
		{"A5", 5.83, 8.27, false},
		{"Letter", 8.5, 11.0, false},
		{"legal", 8.5, 14.0, false},
		{"UnknownFormat", 0, 0, true},
		{"", 0, 0, true},
	}

	for _, tt := range tests {
		p, err := ParsePaper(tt.input)
		if tt.expectError {
			if err == nil {
				t.Errorf("ParsePaper(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParsePaper(%q) unexpected error: %v", tt.input, err)
			}
			if p.WidthIn != tt.wantWidth || p.HeightIn != tt.wantHeight {
				t.Errorf("ParsePaper(%q) = (%v, %v), want (%v, %v)", tt.input, p.WidthIn, p.HeightIn, tt.wantWidth, tt.wantHeight)
			}
		}
	}
}

func TestParseMargins(t *testing.T) {
	const eps = 1e-4

	tests := []struct {
		input       string
		wantTop     float64
		wantRight   float64
		wantBottom  float64
		wantLeft    float64
		expectError bool
	}{
		// Single value: applies to all 4
		{"20mm", 20.0 / 25.4, 20.0 / 25.4, 20.0 / 25.4, 20.0 / 25.4, false},
		// Without unit: defaults to mm
		{"20", 20.0 / 25.4, 20.0 / 25.4, 20.0 / 25.4, 20.0 / 25.4, false},
		// 2 values: vertical horizontal
		{"25mm 20mm", 25.0 / 25.4, 20.0 / 25.4, 25.0 / 25.4, 20.0 / 25.4, false},
		// 4 values: top right bottom left
		{"25mm 20mm 15mm 10mm", 25.0 / 25.4, 20.0 / 25.4, 15.0 / 25.4, 10.0 / 25.4, false},
		// Units: cm, in, pt
		{"1in", 1.0, 1.0, 1.0, 1.0, false},
		{"2.54cm", 1.0, 1.0, 1.0, 1.0, false},
		{"72pt", 1.0, 1.0, 1.0, 1.0, false},
		// Invalid counts
		{"", 0, 0, 0, 0, true},
		{"10mm 20mm 30mm", 0, 0, 0, 0, true},
		{"10mm 20mm 30mm 40mm 50mm", 0, 0, 0, 0, true},
		// Negative values
		{"-5mm", 0, 0, 0, 0, true},
		// Invalid unit or number
		{"foo", 0, 0, 0, 0, true},
		{"10px", 0, 0, 0, 0, true},
	}

	for _, tt := range tests {
		m, err := ParseMargins(tt.input)
		if tt.expectError {
			if err == nil {
				t.Errorf("ParseMargins(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseMargins(%q) unexpected error: %v", tt.input, err)
			}
			if math.Abs(m.TopIn-tt.wantTop) > eps ||
				math.Abs(m.RightIn-tt.wantRight) > eps ||
				math.Abs(m.BottomIn-tt.wantBottom) > eps ||
				math.Abs(m.LeftIn-tt.wantLeft) > eps {
				t.Errorf("ParseMargins(%q) = %+v, want Top=%v Right=%v Bottom=%v Left=%v",
					tt.input, m, tt.wantTop, tt.wantRight, tt.wantBottom, tt.wantLeft)
			}
		}
	}
}
