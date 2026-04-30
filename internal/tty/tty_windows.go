// +build windows

package tty

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.MustLoadDLL("kernel32.dll")
)

const (
	DA1_132_COLUMNS                         = 0
	DA1_PRINTER_PORT                        = 0
	DA1_SIXEL                               = 0
	DA1_SELECTIVE_ERASE                     = 0
	DA1_DRCS                                = 0
	DA1_SOFT_CHARACTER_SET                  = 0
	DA1_UDKS                                = 0
	DA1_USER_DEFINED_KEYS                   = 0
	DA1_NRCS                                = 0
	DA1_NATIONAL_REPLACEMENT_CHARACTER_SETS = 0
	DA1_SCS                                 = 0
	DA1_YUGOSLAVIAN                         = 0
	DA1_TECHNICAL_CHARACTER_SET             = 0
	DA1_WINDOWING_CAPABILITY                = 0
	DA1_HORIZONTAL_SCROLLING                = 0
	DA1_GREEK                               = 0
	DA1_TURKISH                             = 0
	DA1_ISO_LATIN2_CHARACTER_SET            = 0
	DA1_PCTERM                              = 0
	DA1_SOFT_KEY_MAP                        = 0
	DA1_ASCII_EMULATION                     = 0
	DA1_CLASS4                              = 0
	DA1_MAX                                 = 0
)

type DeviceAttributes1 [1]bool

// IsTty checks if the given fd is a tty
func IsTty(file *os.File) bool {
	var st uint32

	f := syscall.MustLoadDLL("kernel32.dll").MustFindProc("GetConsoleMode")
	r1, _, err := f.Call(file.Fd(), uintptr(unsafe.Pointer(&st)))

	return r1 != 0 && err != nil
}

func GetDeviceAttributes1(file *os.File) (DeviceAttributes1, error) {
	var da1 DeviceAttributes1

	return da1, nil
}
