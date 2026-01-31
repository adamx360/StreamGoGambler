//go:build windows

package main

import (
	"syscall"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procFreeConsole      = kernel32.NewProc("FreeConsole")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
)

func HideConsole() {
	procFreeConsole.Call() //nolint:errcheck // return value not needed for FreeConsole
}

func HasConsole() bool {
	hwnd, _, _ := procGetConsoleWindow.Call()
	return hwnd != 0
}
