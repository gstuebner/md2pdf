package mdconv

import (
	"strings"
	"testing"
)

func TestMermaidBypassesHighlighting(t *testing.T) {
	res, err := Convert([]byte("```mermaid\ngraph TD;\n  A-->B;\n```\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	got := string(res.BodyHTML)
	if !strings.Contains(got, `<pre class="mermaid">graph TD;`) {
		t.Errorf("mermaid block not passed through verbatim:\n%s", got)
	}
	if strings.Contains(got, "chroma") {
		t.Errorf("mermaid block went through the highlighter:\n%s", got)
	}
	if !res.HasMermaid {
		t.Error("HasMermaid not set")
	}
}

func TestNoMermaidLeavesFlagUnset(t *testing.T) {
	res, err := Convert([]byte("```go\nvar x int\n```\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if res.HasMermaid {
		t.Error("HasMermaid set although the document has no diagram")
	}
}

func TestCodeBlockAttributes(t *testing.T) {
	res, err := Convert([]byte("```go\nvar a int\nvar b int\n```\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	got := string(res.BodyHTML)
	if !strings.Contains(got, `data-lang="go"`) {
		t.Errorf("language label missing:\n%s", got)
	}
	if !strings.Contains(got, `data-lines="2"`) {
		t.Errorf("line count wrong:\n%s", got)
	}
	if !strings.Contains(got, `class="chroma"`) {
		t.Errorf("chroma output missing:\n%s", got)
	}
	if !strings.Contains(got, `class="k"`) && !strings.Contains(got, `class="kd"`) {
		t.Errorf("no keyword classes, WithClasses may be off:\n%s", got)
	}
}

func TestUnknownLanguageFallsBack(t *testing.T) {
	res, err := Convert([]byte("```gibtesnicht\nirgendwas\n```\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	got := string(res.BodyHTML)
	if !strings.Contains(got, `data-lang="gibtesnicht"`) {
		t.Errorf("language attribute lost:\n%s", got)
	}
	if !strings.Contains(got, "irgendwas") {
		t.Errorf("code text lost:\n%s", got)
	}
}

func TestFenceWithoutLanguage(t *testing.T) {
	res, err := Convert([]byte("```\nnur text\n```\n"), Options{Loader: stubLoader{}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.BodyHTML), `data-lang=""`) {
		t.Errorf("expected empty language attribute:\n%s", res.BodyHTML)
	}
}
