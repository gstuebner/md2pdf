package assets

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
)

//go:embed templates theme fonts vendor
var FS embed.FS

// Template returns the contents of a template file from templates/<name>.
func Template(name string) ([]byte, error) {
	path := name
	if !strings.HasPrefix(path, "templates/") {
		path = "templates/" + path
	}
	b, err := FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read embedded template %q: %w", name, err)
	}
	return b, nil
}

// Theme returns the concatenated contents of fonts.css, default.css, and code.css in that order.
func Theme() ([]byte, error) {
	files := []string{"theme/fonts.css", "theme/default.css", "theme/code.css"}
	var buf bytes.Buffer
	for _, file := range files {
		data, err := FS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read embedded theme file %q: %w", file, err)
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

// MermaidJS returns the contents of vendor/mermaid.min.js.
func MermaidJS() ([]byte, error) {
	data, err := FS.ReadFile("vendor/mermaid.min.js")
	if err != nil {
		return nil, fmt.Errorf("read embedded mermaid.min.js: %w", err)
	}
	return data, nil
}
