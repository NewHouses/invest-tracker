//go:build windows

package i18n

import (
	"syscall"
	"unsafe"
)

// systemLocale devolve a configuración rexional do usuario de Windows (p.e.
// "gl-ES" ou "en-US"), xa que en Windows non adoita haber variable LANG.
func systemLocale() string {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserDefaultLocaleName")
	if proc.Find() != nil {
		return ""
	}
	buf := make([]uint16, 85) // LOCALE_NAME_MAX_LENGTH
	if r, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))); r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
