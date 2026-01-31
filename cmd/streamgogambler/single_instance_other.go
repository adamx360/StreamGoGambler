//go:build !windows

package main

func AcquireSingleInstanceLock() bool {
	return true
}

func ReleaseSingleInstanceLock() {
	// No-op: single instance locking is only implemented on Windows
}
