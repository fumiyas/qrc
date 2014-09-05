package tty

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
)

// IsTty checks if the given fd is a tty
func IsTty(file *os.File) bool {
	var st uint32

	f := syscall.MustLoadDLL("kernel32.dll").MustFindProc("GetConsoleMode")
	r1, _, err := f.Call(file.Fd(), uintptr(unsafe.Pointer(&st)))

	return r1 != 0 && err != nil
}

