package mdconv

import (
	"strings"
	"testing"
)

func TestSoleImageBecomesFigure(t *testing.T) {
	got := convertString(t, "![Alternativtext](bild.png)\n")
	if !strings.Contains(got, `<figure class="figure">`) {
		t.Errorf("no figure produced:\n%s", got)
	}
	if !strings.Contains(got, "<figcaption>Alternativtext</figcaption>") {
		t.Errorf("caption not taken from alt text:\n%s", got)
	}
}

func TestImageTitleWinsOverAltText(t *testing.T) {
	got := convertString(t, `![Alt](bild.png "Bildunterschrift")`+"\n")
	if !strings.Contains(got, "<figcaption>Bildunterschrift</figcaption>") {
		t.Errorf("title should win over alt text:\n%s", got)
	}
}

func TestImageWithoutTextHasNoCaption(t *testing.T) {
	got := convertString(t, "![](bild.png)\n")
	if strings.Contains(got, "<figcaption>") {
		t.Errorf("empty caption should be omitted:\n%s", got)
	}
}

func TestImageInRunningTextStaysInline(t *testing.T) {
	got := convertString(t, "Ein Satz mit ![Bild](bild.png) mittendrin.\n")
	if strings.Contains(got, "<figure") {
		t.Errorf("inline image must not become a figure:\n%s", got)
	}
	if !strings.Contains(got, "<p>") {
		t.Errorf("paragraph markup missing:\n%s", got)
	}
}

func TestImageSourceGoesThroughLoader(t *testing.T) {
	got := convertString(t, "![Bild](bild.png)\n")
	if !strings.Contains(got, `src="data:image/svg+xml;base64,STUB"`) {
		t.Errorf("loader was not used:\n%s", got)
	}
}

func TestRemoteImageIsLeftAlone(t *testing.T) {
	got := convertString(t, "![Bild](https://example.com/b.png)\n")
	if !strings.Contains(got, `src="https://example.com/b.png"`) {
		t.Errorf("remote URL was rewritten:\n%s", got)
	}
}
