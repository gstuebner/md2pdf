package cmd

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gstuebner/md2pdf/internal/config"
	"github.com/gstuebner/md2pdf/internal/pipeline"
)

// withStubPipeline temporarily replaces runPipeline and restores it after
// the test finishes.
func withStubPipeline(t *testing.T, fn func(ctx context.Context, o config.Options) (pipeline.Result, error)) {
	t.Helper()
	original := runPipeline
	runPipeline = fn
	t.Cleanup(func() { runPipeline = original })
}

func runCmd(t *testing.T, args []string) (config.Options, error) {
	t.Helper()
	withStubPipeline(t, func(ctx context.Context, o config.Options) (pipeline.Result, error) {
		return pipeline.Result{OutputPath: o.Output, Size: 1024, Duration: time.Second}, nil
	})
	o := config.DefaultOptions()
	c := newRootCmd(&o)
	c.SetArgs(args)
	c.SetOut(&bytes.Buffer{})
	c.SetErr(&bytes.Buffer{})
	err := c.Execute()
	return o, err
}

func TestFlagMappingDefaults(t *testing.T) {
	o, err := runCmd(t, []string{"input.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := config.DefaultOptions()
	want.Input = "input.md"
	want.Output = "input.pdf"

	if !reflect.DeepEqual(o, want) {
		t.Fatalf("options mismatch:\n got: %+v\nwant: %+v", o, want)
	}
}

func TestFlagMappingAllNoFlags(t *testing.T) {
	o, err := runCmd(t, []string{
		"input.md",
		"--no-toc", "--no-cover", "--no-numbering",
		"--no-header", "--no-footer", "--no-outline",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if o.TOC {
		t.Error("expected TOC to be false")
	}
	if o.Cover {
		t.Error("expected Cover to be false")
	}
	if o.NumberHeadings {
		t.Error("expected NumberHeadings to be false")
	}
	if o.Header {
		t.Error("expected Header to be false")
	}
	if o.Footer {
		t.Error("expected Footer to be false")
	}
	if o.Outline {
		t.Error("expected Outline to be false")
	}
}

func TestFlagMappingPaperA5(t *testing.T) {
	o, err := runCmd(t, []string{"input.md", "--paper", "A5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := config.PaperSizes["a5"]
	if o.Paper != want {
		t.Errorf("got paper %+v, want %+v", o.Paper, want)
	}
}

func TestFlagMappingMargin(t *testing.T) {
	o, err := runCmd(t, []string{"input.md", "--margin", "25mm 20mm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want, err := config.ParseMargins("25mm 20mm")
	if err != nil {
		t.Fatalf("ParseMargins failed: %v", err)
	}
	if o.Margins != want {
		t.Errorf("got margins %+v, want %+v", o.Margins, want)
	}
}

func TestFlagMappingMultipleCSS(t *testing.T) {
	o, err := runCmd(t, []string{"input.md", "--css", "a.css", "--css", "b.css"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a.css", "b.css"}
	if !reflect.DeepEqual(o.ExtraCSS, want) {
		t.Errorf("got ExtraCSS %v, want %v", o.ExtraCSS, want)
	}
}

func TestFlagMappingMetadataOnlyNonEmpty(t *testing.T) {
	o, err := runCmd(t, []string{"input.md", "--title", "Handbuch", "--doc-version", "2.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Meta.Title != "Handbuch" {
		t.Errorf("got title %q, want %q", o.Meta.Title, "Handbuch")
	}
	if o.Meta.Version != "2.1" {
		t.Errorf("got version %q, want %q", o.Meta.Version, "2.1")
	}
	if o.Meta.Author != "" {
		t.Errorf("expected empty author, got %q", o.Meta.Author)
	}
}

func TestDeriveOutputPath(t *testing.T) {
	cases := map[string]string{
		"handbuch.md":       "handbuch.pdf",
		"handbuch.markdown": "handbuch.pdf",
		"noext":             "noext.pdf",
		"dir/sub/file.md":   "dir/sub/file.pdf",
	}
	for in, want := range cases {
		if got := deriveOutputPath(in); got != want {
			t.Errorf("deriveOutputPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{842 * 1024, "842 KB"},
		{2 * 1024, "2 KB"},
		{1992294, "1.9 MB"}, // ~1.9 MiB
		{1024 * 1024, "1 MB"},
	}
	for _, c := range cases {
		if got := formatSize(c.bytes); got != c.want {
			t.Errorf("formatSize(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{1900 * time.Millisecond, "1.9s"},
		{1860 * time.Millisecond, "1.9s"},
		{1840 * time.Millisecond, "1.8s"},
		{0, "0.0s"},
	}
	for _, c := range cases {
		if got := formatDuration(c.d); got != c.want {
			t.Errorf("formatDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestFormatSuccessLine(t *testing.T) {
	r := pipeline.Result{OutputPath: "/tmp/handbuch.pdf", Size: 842 * 1024, Duration: 1900 * time.Millisecond}
	want := "handbuch.pdf · 842 KB · 1.9s"
	if got := formatSuccessLine(r); got != want {
		t.Errorf("formatSuccessLine() = %q, want %q", got, want)
	}
}

func TestExecuteExitCodes(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		withStubPipeline(t, func(ctx context.Context, o config.Options) (pipeline.Result, error) {
			return pipeline.Result{OutputPath: o.Output, Size: 100, Duration: time.Second}, nil
		})
		var out, errOut bytes.Buffer
		code := execute([]string{"input.md", "--quiet"}, &out, &errOut)
		if code != 0 {
			t.Fatalf("got exit code %d, want 0 (stderr: %s)", code, errOut.String())
		}
	})

	t.Run("invalid arg count", func(t *testing.T) {
		var out, errOut bytes.Buffer
		code := execute([]string{}, &out, &errOut)
		if code != 2 {
			t.Fatalf("got exit code %d, want 2", code)
		}
	})

	t.Run("invalid paper", func(t *testing.T) {
		var out, errOut bytes.Buffer
		code := execute([]string{"input.md", "--paper", "A3"}, &out, &errOut)
		if code != 2 {
			t.Fatalf("got exit code %d, want 2", code)
		}
	})

	t.Run("invalid margin", func(t *testing.T) {
		var out, errOut bytes.Buffer
		code := execute([]string{"input.md", "--margin", "banana"}, &out, &errOut)
		if code != 2 {
			t.Fatalf("got exit code %d, want 2", code)
		}
	})

	t.Run("browser not found", func(t *testing.T) {
		withStubPipeline(t, func(ctx context.Context, o config.Options) (pipeline.Result, error) {
			return pipeline.Result{}, pipeline.ErrBrowserNotFound
		})
		var out, errOut bytes.Buffer
		code := execute([]string{"input.md"}, &out, &errOut)
		if code != 3 {
			t.Fatalf("got exit code %d, want 3", code)
		}
		if !bytes.Contains(errOut.Bytes(), []byte("Chromium")) {
			t.Errorf("expected browser help text on stderr, got %q", errOut.String())
		}
	})

	t.Run("render timeout", func(t *testing.T) {
		withStubPipeline(t, func(ctx context.Context, o config.Options) (pipeline.Result, error) {
			return pipeline.Result{}, pipeline.ErrRenderTimeout
		})
		var out, errOut bytes.Buffer
		code := execute([]string{"input.md"}, &out, &errOut)
		if code != 4 {
			t.Fatalf("got exit code %d, want 4", code)
		}
	})

	t.Run("generic conversion error", func(t *testing.T) {
		withStubPipeline(t, func(ctx context.Context, o config.Options) (pipeline.Result, error) {
			return pipeline.Result{}, errors.New("input file not found")
		})
		var out, errOut bytes.Buffer
		code := execute([]string{"input.md"}, &out, &errOut)
		if code != 1 {
			t.Fatalf("got exit code %d, want 1", code)
		}
	})
}

func TestQuietSuppressesOutput(t *testing.T) {
	withStubPipeline(t, func(ctx context.Context, o config.Options) (pipeline.Result, error) {
		return pipeline.Result{OutputPath: o.Output, Size: 100, Duration: time.Second}, nil
	})
	var out, errOut bytes.Buffer
	code := execute([]string{"input.md", "--quiet"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("got exit code %d, want 0", code)
	}
	if out.Len() != 0 {
		t.Errorf("expected no stdout output with --quiet, got %q", out.String())
	}
}
