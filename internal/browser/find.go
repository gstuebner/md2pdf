// Package browser locates a Chromium-based browser engine and drives it via
// the Chrome DevTools Protocol to render HTML documents as PDF.
package browser

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ErrBrowserNotFound is returned by Find when no usable Chromium-based
// browser could be located by any of the lookup strategies.
var ErrBrowserNotFound = errors.New("keine Chromium-basierte Browser-Engine gefunden")

// NotFoundHelp is a multi-line, user-facing help message explaining how to
// install a supported browser or point md2pdf at one explicitly. Callers are
// responsible for printing it; this package never writes to stderr itself.
const NotFoundHelp = `md2pdf: keine Chromium-basierte Browser-Engine gefunden.

Installiere eine davon oder gib den Pfad explizit an:
  Arch/CachyOS:  paru -S chromium
  Debian/Ubuntu: sudo apt install chromium
  Windows:       Microsoft Edge ist vorinstalliert
  manuell:       md2pdf --browser-path /pfad/zu/chrome`

// candidateNames are executables looked up via exec.LookPath, in order.
var candidateNames = []string{
	"chromium",
	"chromium-browser",
	"google-chrome-stable",
	"google-chrome",
	"brave-browser",
	"microsoft-edge",
	"microsoft-edge-stable",
}

// Find locates a usable Chromium-based browser executable.
//
// Lookup order:
//  1. explicitPath, if non-empty. A path that does not exist or is not
//     executable is a hard error — no further lookup is attempted.
//  2. the MD2PDF_BROWSER environment variable, under the same rule.
//  3. exec.LookPath over a fixed list of well-known executable names.
//  4. a list of well-known absolute install paths for the current OS.
//
// If nothing is found, the returned error wraps ErrBrowserNotFound.
func Find(explicitPath string) (string, error) {
	if explicitPath != "" {
		return checkExecutable(explicitPath)
	}

	if envPath := os.Getenv("MD2PDF_BROWSER"); envPath != "" {
		return checkExecutable(envPath)
	}

	for _, name := range candidateNames {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	for _, path := range knownPathsFn() {
		if isExecutableFile(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("%w", ErrBrowserNotFound)
}

// checkExecutable verifies that path exists and is executable, returning it
// unchanged on success or a wrapped ErrBrowserNotFound on failure.
func checkExecutable(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("Browser-Pfad %q nicht nutzbar: %w", path, ErrBrowserNotFound)
	}
	if info.IsDir() {
		return "", fmt.Errorf("Browser-Pfad %q ist ein Verzeichnis: %w", path, ErrBrowserNotFound)
	}
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return "", fmt.Errorf("Browser-Pfad %q ist nicht ausführbar: %w", path, ErrBrowserNotFound)
	}
	return path, nil
}

// isExecutableFile reports whether path exists, is a regular file, and is
// executable (on non-Windows platforms).
func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return false
	}
	return true
}

// knownPathsFn returns well-known absolute install locations for the current
// operating system. It is a variable so tests can stub it out.
var knownPathsFn = knownPaths

// knownPaths returns well-known absolute install locations for the current
// operating system.
func knownPaths() []string {
	switch runtime.GOOS {
	case "windows":
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		programFiles := os.Getenv("ProgramFiles")
		localAppData := os.Getenv("LOCALAPPDATA")
		var paths []string
		if programFilesX86 != "" {
			paths = append(paths, filepath.Join(programFilesX86, "Microsoft", "Edge", "Application", "msedge.exe"))
		}
		if programFiles != "" {
			paths = append(paths, filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"))
		}
		if localAppData != "" {
			paths = append(paths, filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"))
		}
		return paths
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	default:
		return []string{
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/google-chrome-stable",
			"/snap/bin/chromium",
			"/var/lib/flatpak/exports/bin/org.chromium.Chromium",
		}
	}
}
