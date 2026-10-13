//go:build !windows

package i18n

// systemLocale non fai falta fóra de Windows: macOS e Linux indican o idioma
// coas variables LANG, LC_ALL ou LC_MESSAGES.
func systemLocale() string { return "" }
