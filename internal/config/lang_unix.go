//go:build !windows

package config

// platformLocale is the fallback for systems whose locale is configured
// through the environment, which DetectLang has already consulted. Returning
// nothing keeps LC_ALL and LANG authoritative there.
func platformLocale() string { return "" }
