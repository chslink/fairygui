//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	moduser32 = windows.NewLazySystemDLL("user32.dll")
	modimm32  = windows.NewLazySystemDLL("imm32.dll")

	procFindWindowW        = moduser32.NewProc("FindWindowW")
	procGetForegroundWindow = moduser32.NewProc("GetForegroundWindow")
	procImmGetContext       = modimm32.NewProc("ImmGetContext")
	procImmReleaseContext   = modimm32.NewProc("ImmReleaseContext")
	procImmAssociateContext = modimm32.NewProc("ImmAssociateContext")
	procImmSetOpenStatus    = modimm32.NewProc("ImmSetOpenStatus")
)

// enableIME associates the Ebiten window with the default IME context so that
// CJK input methods can be activated.
func enableIME() {
	hwnd := findEbitenWindow()
	if hwnd == 0 {
		return
	}
	// Get or create IME context for this window
	hIMC, _, _ := procImmGetContext.Call(hwnd)
	if hIMC == 0 {
		// No IME context yet — associate with default (0 = default context)
		procImmAssociateContext.Call(hwnd, 0)
		hIMC, _, _ = procImmGetContext.Call(hwnd)
	}
	if hIMC == 0 {
		return
	}
	procImmReleaseContext.Call(hwnd, hIMC)
	// Enable IME open status
	procImmSetOpenStatus.Call(hIMC, 1)
}

func findEbitenWindow() uintptr {
	// Try window class name first
	name, _ := syscall.UTF16PtrFromString("Ebiten")
	hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(name)), 0)
	if hwnd != 0 {
		return hwnd
	}
	// Fallback: foreground window
	hwnd, _, _ = procGetForegroundWindow.Call()
	return hwnd
}
