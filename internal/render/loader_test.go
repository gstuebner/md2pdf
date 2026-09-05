package render

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileLoaderLocalPNG(t *testing.T) {
	dir := t.TempDir()
	pngBytes := []byte("\x89PNG\r\n\x1a\nfake-png-content")
	imgPath := filepath.Join(dir, "logo.png")
	if err := os.WriteFile(imgPath, pngBytes, 0o644); err != nil {
		t.Fatalf("write test png: %v", err)
	}

	var warnBuf bytes.Buffer
	loader := &FileLoader{BaseDir: dir, Warn: &warnBuf}

	uri, err := loader.DataURI("logo.png")
	if err != nil {
		t.Fatalf("DataURI() error: %v", err)
	}
	if warnBuf.Len() != 0 {
		t.Fatalf("unexpected warning: %s", warnBuf.String())
	}

	wantPrefix := "data:image/png;base64,"
	if !strings.HasPrefix(uri, wantPrefix) {
		t.Fatalf("data uri missing expected prefix, got: %s", uri)
	}

	encoded := strings.TrimPrefix(uri, wantPrefix)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}
	if !bytes.Equal(decoded, pngBytes) {
		t.Fatalf("decoded bytes do not match original file content")
	}
}

func TestFileLoaderUnknownExtensionWarns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.xyz")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	var warnBuf bytes.Buffer
	loader := &FileLoader{BaseDir: dir, Warn: &warnBuf}

	uri, err := loader.DataURI("file.xyz")
	if err != nil {
		t.Fatalf("DataURI() error: %v", err)
	}
	if !strings.HasPrefix(uri, "data:application/octet-stream;base64,") {
		t.Fatalf("expected octet-stream mime, got: %s", uri)
	}
	if warnBuf.Len() == 0 {
		t.Fatalf("expected warning about unknown extension")
	}
}

func TestFileLoaderHTTPPassthrough(t *testing.T) {
	var warnBuf bytes.Buffer
	loader := &FileLoader{BaseDir: t.TempDir(), Warn: &warnBuf}

	for _, src := range []string{
		"https://example.com/image.png",
		"http://example.com/image.png",
		"data:image/png;base64,QUJD",
	} {
		uri, err := loader.DataURI(src)
		if err != nil {
			t.Fatalf("DataURI(%q) error: %v", src, err)
		}
		if uri != src {
			t.Fatalf("DataURI(%q) = %q, want unchanged", src, uri)
		}
	}
	if warnBuf.Len() != 0 {
		t.Fatalf("unexpected warning for passthrough sources: %s", warnBuf.String())
	}
}

func TestFileLoaderMissingFile(t *testing.T) {
	dir := t.TempDir()
	var warnBuf bytes.Buffer
	loader := &FileLoader{BaseDir: dir, Warn: &warnBuf}

	src := "does-not-exist.png"
	uri, err := loader.DataURI(src)
	if err != nil {
		t.Fatalf("DataURI() should not return an error, got: %v", err)
	}
	if uri != src {
		t.Fatalf("DataURI() = %q, want unchanged src %q", uri, src)
	}
	if warnBuf.Len() == 0 {
		t.Fatalf("expected warning for missing file")
	}
}

func TestFileLoaderOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "huge.png")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create huge file: %v", err)
	}
	if err := f.Truncate(maxAssetSize + 1); err != nil {
		f.Close()
		t.Fatalf("truncate huge file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close huge file: %v", err)
	}

	var warnBuf bytes.Buffer
	loader := &FileLoader{BaseDir: dir, Warn: &warnBuf}

	src := "huge.png"
	uri, err := loader.DataURI(src)
	if err != nil {
		t.Fatalf("DataURI() should not return an error, got: %v", err)
	}
	if uri != src {
		t.Fatalf("DataURI() = %q, want unchanged src %q", uri, src)
	}
	if warnBuf.Len() == 0 {
		t.Fatalf("expected warning for oversized file")
	}
}

func TestFileLoaderDefaultWarnWriter(t *testing.T) {
	dir := t.TempDir()
	loader := &FileLoader{BaseDir: dir}

	uri, err := loader.DataURI("does-not-exist.png")
	if err != nil {
		t.Fatalf("DataURI() error: %v", err)
	}
	if uri != "does-not-exist.png" {
		t.Fatalf("expected unchanged src, got %q", uri)
	}
}
