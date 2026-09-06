package mdconv

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/gstuebner/md2pdf/internal/model"
)

// PeekMeta reads only the YAML front matter of src. The caller needs a few
// values — the style preset above all — before the full conversion runs,
// because they decide options that Convert itself depends on. Anything that
// does not parse yields a zero Meta; Convert reports the real error later.
func PeekMeta(src []byte) model.Meta {
	body, ok := frontMatterBlock(src)
	if !ok {
		return model.Meta{}
	}
	var m model.Meta
	if err := yaml.Unmarshal(body, &m); err != nil {
		return model.Meta{}
	}
	return m
}

// frontMatterBlock returns the bytes between the opening and closing "---"
// fence at the very start of the document.
func frontMatterBlock(src []byte) ([]byte, bool) {
	rest := bytes.TrimPrefix(src, []byte("\ufeff"))
	const fence = "---"

	lines := bytes.Split(rest, []byte("\n"))
	if len(lines) < 2 || string(bytes.TrimRight(lines[0], "\r \t")) != fence {
		return nil, false
	}
	for i := 1; i < len(lines); i++ {
		if string(bytes.TrimRight(lines[i], "\r \t")) == fence {
			return bytes.Join(lines[1:i], []byte("\n")), true
		}
	}
	return nil, false
}
