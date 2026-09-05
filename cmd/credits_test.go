package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpShowsVersionAndAuthors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}

	out := stdout.String()
	last := lastNonEmptyLine(out)
	if !strings.Contains(last, "md2pdf "+version) {
		t.Errorf("last help line %q does not carry the version", last)
	}
	if !strings.Contains(last, authors) {
		t.Errorf("last help line %q does not carry the authors", last)
	}
}

func TestVersionFlagShowsVersionAndAuthors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := execute([]string{"--version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}

	out := strings.TrimSpace(stdout.String())
	if !strings.Contains(out, version) || !strings.Contains(out, authors) {
		t.Errorf("--version printed %q, want version and authors", out)
	}
}

func TestAuthorsSpelling(t *testing.T) {
	if authors != "Gregor Stübner & Claude (Anthropic)" {
		t.Errorf("authors = %q, unexpected spelling", authors)
	}
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}
