package render

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxAssetSize is the size limit for locally loaded assets (20 MB).
const maxAssetSize = 20 * 1024 * 1024

// AssetLoader resolves an image source into a form Chromium can load without
// touching the filesystem again, typically a data URI.
type AssetLoader interface {
	// DataURI resolves src relative to the Markdown file's directory and returns
	// "data:image/png;base64,…". For http(s):// or data: sources, src comes back
	// unchanged. On error: a warning on stderr, src unchanged, no failure.
	DataURI(src string) (string, error)
}

// FileLoader implements AssetLoader by reading files relative to BaseDir.
type FileLoader struct {
	BaseDir string
	// Warn receives warning messages; defaults to os.Stderr when nil.
	Warn io.Writer
}

// DataURI implements AssetLoader.
func (l *FileLoader) DataURI(src string) (string, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:") {
		return src, nil
	}

	path := src
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.BaseDir, path)
	}

	info, err := os.Stat(path)
	if err != nil {
		l.warnf("md2pdf: warning: cannot read image %q: %v\n", src, err)
		return src, nil
	}
	if info.Size() > maxAssetSize {
		l.warnf("md2pdf: warning: image %q exceeds the 20 MB limit and is used unchanged\n", src)
		return src, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		l.warnf("md2pdf: warning: cannot read image %q: %v\n", src, err)
		return src, nil
	}

	mime, unknown := mimeType(path)
	if unknown {
		l.warnf("md2pdf: warning: unknown file extension on %q, using application/octet-stream\n", src)
	}

	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func (l *FileLoader) warnf(format string, args ...any) {
	w := l.Warn
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintf(w, format, args...)
}

func mimeType(path string) (mime string, unknown bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png", false
	case ".jpg", ".jpeg":
		return "image/jpeg", false
	case ".gif":
		return "image/gif", false
	case ".svg":
		return "image/svg+xml", false
	case ".webp":
		return "image/webp", false
	default:
		return "application/octet-stream", true
	}
}
