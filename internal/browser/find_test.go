package browser

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeExecutable creates a dummy executable file at the given path.
func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("failed to write dummy executable %q: %v", path, err)
	}
}

// writeNonExecutable creates a dummy non-executable file at the given path.
func writeNonExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("not executable"), 0o644); err != nil {
		t.Fatalf("failed to write dummy file %q: %v", path, err)
	}
}

func TestFind_ExplicitPathValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my-browser")
	writeExecutable(t, path)

	got, err := Find(path)
	if err != nil {
		t.Fatalf("Find(%q) returned unexpected error: %v", path, err)
	}
	if got != path {
		t.Fatalf("Find(%q) = %q, want %q", path, got, path)
	}
}

func TestFind_ExplicitPathInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist")

	_, err := Find(path)
	if err == nil {
		t.Fatal("Find with nonexistent explicit path: want error, got nil")
	}
	if !errors.Is(err, ErrBrowserNotFound) {
		t.Fatalf("Find with nonexistent explicit path: err = %v, want errors.Is ErrBrowserNotFound", err)
	}
}

func TestFind_ExplicitPathInvalidDoesNotFallThrough(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH lookup semantics differ on windows")
	}
	withNoKnownPaths(t)

	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "does-not-exist")

	pathDir := t.TempDir()
	writeExecutable(t, filepath.Join(pathDir, "chromium"))
	t.Setenv("PATH", pathDir)

	_, err := Find(invalidPath)
	if err == nil {
		t.Fatal("Find with invalid explicit path found something despite PATH having a valid candidate")
	}
	if !errors.Is(err, ErrBrowserNotFound) {
		t.Fatalf("err = %v, want errors.Is ErrBrowserNotFound", err)
	}
}

func TestFind_ExplicitPathNotExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable bit is not meaningful on windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "my-browser")
	writeNonExecutable(t, path)

	_, err := Find(path)
	if err == nil {
		t.Fatal("Find with non-executable explicit path: want error, got nil")
	}
	if !errors.Is(err, ErrBrowserNotFound) {
		t.Fatalf("err = %v, want errors.Is ErrBrowserNotFound", err)
	}
}

func TestFind_EnvVariable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env-browser")
	writeExecutable(t, path)

	t.Setenv("MD2PDF_BROWSER", path)

	got, err := Find("")
	if err != nil {
		t.Fatalf("Find(\"\") with MD2PDF_BROWSER set returned unexpected error: %v", err)
	}
	if got != path {
		t.Fatalf("Find(\"\") = %q, want %q", got, path)
	}
}

func TestFind_EnvVariableInvalidDoesNotFallThrough(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH lookup semantics differ on windows")
	}
	withNoKnownPaths(t)

	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "does-not-exist")
	t.Setenv("MD2PDF_BROWSER", invalidPath)

	pathDir := t.TempDir()
	writeExecutable(t, filepath.Join(pathDir, "chromium"))
	t.Setenv("PATH", pathDir)

	_, err := Find("")
	if err == nil {
		t.Fatal("Find with invalid MD2PDF_BROWSER found something despite PATH having a valid candidate")
	}
	if !errors.Is(err, ErrBrowserNotFound) {
		t.Fatalf("err = %v, want errors.Is ErrBrowserNotFound", err)
	}
}

func TestFind_PathLookup(t *testing.T) {
	withNoKnownPaths(t)

	pathDir := t.TempDir()
	binPath := filepath.Join(pathDir, "chromium-browser")
	writeExecutable(t, binPath)

	t.Setenv("MD2PDF_BROWSER", "")
	t.Setenv("PATH", pathDir)

	got, err := Find("")
	if err != nil {
		t.Fatalf("Find(\"\") via PATH returned unexpected error: %v", err)
	}
	if got != binPath {
		t.Fatalf("Find(\"\") = %q, want %q", got, binPath)
	}
}

func TestFind_NothingFound(t *testing.T) {
	withNoKnownPaths(t)

	emptyDir := t.TempDir()
	t.Setenv("MD2PDF_BROWSER", "")
	t.Setenv("PATH", emptyDir)

	_, err := Find("")
	if err == nil {
		t.Fatal("Find(\"\") with empty PATH and no MD2PDF_BROWSER: want error, got nil")
	}
	if !errors.Is(err, ErrBrowserNotFound) {
		t.Fatalf("err = %v, want errors.Is ErrBrowserNotFound", err)
	}
}

// withNoKnownPaths stubs out the OS-specific well-known browser install
// paths for the duration of the test, so tests are not affected by browsers
// actually installed on the machine running them.
func withNoKnownPaths(t *testing.T) {
	t.Helper()
	original := knownPathsFn
	knownPathsFn = func() []string { return nil }
	t.Cleanup(func() { knownPathsFn = original })
}
