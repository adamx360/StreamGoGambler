//go:build windows

package main

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	procCreateMutexW = kernel32.NewProc("CreateMutexW")
	procCloseHandle  = kernel32.NewProc("CloseHandle")

	mutexHandle uintptr
)

const (
	errorAlreadyExists = 183
)

func AcquireSingleInstanceLock() bool {
	mutexName, err := syscall.UTF16PtrFromString("Global\\StreamGoGamblerSingleInstance")
	if err != nil {
		return true // On error, allow the app to run
	}

	handle, _, lastErr := procCreateMutexW.Call(
		0,
		1,
		uintptr(unsafe.Pointer(mutexName)),
	)

	if handle == 0 {
		return true // On error, allow the app to run
	}

	var errno syscall.Errno
	if errors.As(lastErr, &errno) && errno == errorAlreadyExists {
		_, _, _ = procCloseHandle.Call(handle)
		return false
	}

	mutexHandle = handle
	return true
}

func ReleaseSingleInstanceLock() {
	if mutexHandle != 0 {
		_, _, _ = procCloseHandle.Call(mutexHandle)
		mutexHandle = 0
	}
}
