//go:build windows

package config

import (
	"syscall"
	"unsafe"
)

// localeNameMaxLength is LOCALE_NAME_MAX_LENGTH from winnls.h.
const localeNameMaxLength = 85

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultLocaleName = kernel32.NewProc("GetUserDefaultLocaleName")
)

// platformLocale returns the user's default locale name, e.g. "de-DE".
// Windows sets no LC_ALL or LANG, so without this call every document on a
// German Windows would be labelled in English. An empty string means the
// lookup failed; the caller then falls back to English.
func platformLocale() string {
	buf := make([]uint16, localeNameMaxLength)
	n, _, _ := procGetUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
