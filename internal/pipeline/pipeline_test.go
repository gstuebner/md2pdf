package pipeline

import (
	"strings"
	"testing"
)

func TestSentinelErrorsAreDistinct(t *testing.T) {
	if ErrBrowserNotFound == nil || ErrRenderTimeout == nil {
		t.Fatal("sentinel errors must not be nil")
	}
	if ErrBrowserNotFound.Error() == ErrRenderTimeout.Error() {
		t.Fatal("sentinel errors must have distinct messages")
	}
}

func TestBrowserNotFoundHelpMentionsManualOverride(t *testing.T) {
	if !strings.Contains(BrowserNotFoundHelp, "--browser-path") {
		t.Errorf("expected help text to mention --browser-path, got: %s", BrowserNotFoundHelp)
	}
}
