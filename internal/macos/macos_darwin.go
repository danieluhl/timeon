//go:build darwin

package macos

/*
#cgo darwin LDFLAGS: -framework Cocoa -framework ApplicationServices -framework CoreGraphics
#include <stdlib.h>
#include "macos.h"
*/
import "C"
import "unsafe"

func FrontmostApp() string {
	cstr := C.timeon_frontmost_app()
	if cstr == nil {
		return "Unknown"
	}
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

type FrontmostDiag struct {
	AX           string
	LSAppInfo    string
	SystemEvents string
	Workspace    string
	Selected     string
}

func DiagnoseFrontmost() FrontmostDiag {
	bufSize := C.size_t(256)
	ax := (*C.char)(C.malloc(bufSize))
	ls := (*C.char)(C.malloc(bufSize))
	se := (*C.char)(C.malloc(bufSize))
	ws := (*C.char)(C.malloc(bufSize))
	defer C.free(unsafe.Pointer(ax))
	defer C.free(unsafe.Pointer(ls))
	defer C.free(unsafe.Pointer(se))
	defer C.free(unsafe.Pointer(ws))

	C.timeon_frontmost_diag(ax, ls, se, ws, bufSize)

	return FrontmostDiag{
		AX:           C.GoString(ax),
		LSAppInfo:    C.GoString(ls),
		SystemEvents: C.GoString(se),
		Workspace:    C.GoString(ws),
		Selected:     FrontmostApp(),
	}
}

func AccessibilityTrusted() bool {
	return C.timeon_accessibility_trusted() != 0
}

func RequestAccessibility() {
	C.timeon_request_accessibility()
}

func IdleSeconds() float64 {
	return float64(C.timeon_idle_seconds())
}

func ScreenLocked() bool {
	return C.timeon_screen_locked() != 0
}
