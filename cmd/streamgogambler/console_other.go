//go:build !windows

package main

func HideConsole() {
	// No-op: console hiding is only needed on Windows
}

func HasConsole() bool {
	return false
}
